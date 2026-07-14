package backend

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Status struct {
	Load        float64 `json:"load"`        // 当前载荷
	Stress      float64 `json:"stress"`      // 当前应力值
	Disp        float64 `json:"disp"`        // 当前位移
	Strain      float64 `json:"strain"`      // 当前应变值
	VideoDisp   float64 `json:"videoDisp"`   // 当前视频位移值
	VideoStrain float64 `json:"videoStrain"` // 当前视频应变值
	Limit       int     `json:"limit"`       // 当前限位（0-无，1-端口1，2-端口2）
	Time        float64 `json:"time"`        // 实验时间，单位秒

	CameraConnected  bool `json:"cameraConnected"`  // 相机是否已连接
	MINIMTSConnected bool `json:"minimtsConnected"` // MINIMTS 设备是否已连接
}

const Debug = false

type MINIMTSService struct {
	MTSStatus Status       `json:"status"`
	statusMu  sync.RWMutex // 保护 MTSStatus 的读写锁

	comm *SerialCommunicator // 串口通讯器

	user    *User
	project *ProjectService
	system  *SystemService
	app     *application.App // Wails 应用实例
	camera  *HIKCameraService

	zeroDisp         float64 // 零位位移值，用于计算相对位移
	zeroLoad         float64 // 零位载荷值，用于计算相对载荷
	zeroVideoDisp    float64 // 零位视频位移值，用于计算相对视频位移
	displayDirection int     // 显示方向，1-正常，-1-反向

	isPolling bool      // 标记是否正在后台轮询
	startTime time.Time // 开始时间，用于计算实验时间

	cameraWin *application.WebviewWindow
	dataCache map[string][]float64 // 缓存数据

	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewMINIMTSService(system *SystemService, project *ProjectService, user *User) *MINIMTSService {
	project.SetUser(user)

	svc := &MINIMTSService{
		MTSStatus: Status{
			Load:        0,
			Stress:      0,
			Disp:        0,
			Strain:      0,
			VideoDisp:   0,
			VideoStrain: 0,
			Limit:       0,
			Time:        0,
		},
		dataCache:        make(map[string][]float64),
		system:           system,
		user:             user,
		project:          project,
		startTime:        time.Now(),
		displayDirection: 1,
		stopCh:           make(chan struct{}),
	}

	// 启动独立的前端数据推送协程（不依赖 MINIMTS 设备连接）
	go svc.uiPolling()

	return svc
}

func (m *MINIMTSService) SetApp(app *application.App) {
	m.app = app
}

func (m *MINIMTSService) SetHIKCameraService(camera *HIKCameraService) {
	m.camera = camera
}

func (m *MINIMTSService) CallExit() {
	m.CleanupHardware()
}

func (m *MINIMTSService) CleanupHardware() {
	if m.camera != nil {
		if err := m.camera.CloseHIKCamera(); err != nil {
			fmt.Println("关闭相机失败:", err)
		}
	}
	if err := m.CloseMINIMTS(); err != nil {
		fmt.Println("关闭MINIMTS失败:", err)
	}
}

// *************************************界面管理区 开始*************************************
func (m *MINIMTSService) CallHIKCameraWindow() error {
	app := application.Get()
	if app == nil {
		return fmt.Errorf("application not initialized")
	}

	if m.cameraWin != nil {
		err := m.cameraWin.Show()
		if err != nil {
			fmt.Printf("尝试显示隐藏窗口失败: %v, 正在尝试重新创建...\n", err)
			m.cameraWin = nil // 只有在窗口崩溃或被用户点×彻底关闭时才置空
		} else {
			m.cameraWin.Focus() // 将窗口提到最前并获取焦点
			return nil
		}
	}

	m.cameraWin = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Camera",
		Width:  1200,
		Height: 900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/#/camera",
	})

	return nil
}

func (m *MINIMTSService) CloseHIKCameraWindow() {
	if m.cameraWin != nil {
		m.cameraWin.Close()
		m.cameraWin = nil
	}
}

//*************************************界面管理区 结束*************************************

func (m *MINIMTSService) GetMINIMTSDevices() []string {
	fmt.Println("GetMINIMTSDevices")
	ports, err := m.comm.ListAvailablePorts()
	if err != nil {
		return nil
	}
	return ports
}

func (m *MINIMTSService) OpenMINIMTS(name string) error {

	if m.IsOpened() {
		m.CloseMINIMTS()
		return fmt.Errorf("设备已打开")
	}

	fmt.Printf("尝试打开串口: %s\n", name)

	// 创建串口通讯器
	m.comm = &SerialCommunicator{
		MaxRetries: 3,
		Timeout:    500 * time.Millisecond,
	}

	err := m.comm.Connect(name, 115200)
	if err != nil {
		return fmt.Errorf("无法打开端口 %s: %v", name, err)
	}

	fmt.Printf("串口 %s 打开成功\n", name)
	m.startTime = time.Now()
	m.setMINIMTSConnected(true)

	// 启动后台高速查询协程
	m.isPolling = true
	go m.Polling()

	return nil
}

func (m *MINIMTSService) CloseMINIMTS() error {
	m.isPolling = false // 停止轮询
	m.setMINIMTSConnected(false)

	if m.comm != nil {
		err := m.comm.Disconnect()
		m.comm = nil
		return err
	}
	return nil
}

// Stop 停止所有后台协程（应用退出时调用）
func (m *MINIMTSService) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopCh)
	})
	m.isPolling = false
}

// IsOpened 检查设备是否已打开
func (m *MINIMTSService) IsOpened() bool {
	return m.comm != nil && m.comm.IsConnected()
}

func (m *MINIMTSService) GetMINIMTSStatus() (Status, error) {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	return m.MTSStatus, nil
}

// Polling 轮询采集数据，固定25Hz
func (m *MINIMTSService) Polling() {
	fmt.Println("后台数据采集协程已启动...")
	defer fmt.Println("后台数据采集协程已退出")

	//错误累计器
	ErrorCounts := make(map[string]int)
	SencorsEnable := map[string]bool{"load": true, "disp": true, "limit": true}

	ErrorCounter := func(sensor string, err error) {
		if err == nil {
			ErrorCounts[sensor] = 0
		} else {
			ErrorCounts[sensor] += 1
			fmt.Println("错误计数器:", sensor, "累计错误次数:", ErrorCounts[sensor])
			// 如果错误累计次数达到阈值，该传感器被禁用,除非重新连接，协程重启
			if ErrorCounts[sensor] >= 3 {
				fmt.Println("传感器", sensor, "已禁用")
				m.app.Event.Emit("system_message", "传感器"+sensor+"已禁用")
				SencorsEnable[sensor] = false
			}
		}
	}

	for m.isPolling {
		// 检查设备是否打开
		if !m.IsOpened() {
			m.isPolling = false
			if m.app != nil {
				m.app.Event.Emit("system_message", "设备连接已断开，请重新连接")
			} else {
				m.app = application.Get()
				fmt.Println("m.app is nil")
				m.app.Event.Emit("system_message", "设备连接已断开，请重新连接")
			}
			return
		}

		if SencorsEnable["disp"] {
			ErrorCounter("disp", m.updateMotorPosition())
		}
		if SencorsEnable["load"] {
			ErrorCounter("load", m.updateLoadData())
		}
		if SencorsEnable["limit"] {
			ErrorCounter("limit", m.updateLimitStatus())
		}
		m.MTSStatus.Time = float64((time.Now().UnixMilli() - m.startTime.UnixMilli())) / 1000.0

		//保存到数据缓存
		m.SaveDataToCache()

		// 每25ms更新一次数据
		time.Sleep(20 * time.Millisecond)
	}

}

// setMINIMTSConnected 设置 MINIMTS 设备连接状态
func (m *MINIMTSService) setMINIMTSConnected(connected bool) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()
	m.MTSStatus.MINIMTSConnected = connected
}

// setCameraConnected 设置相机连接状态
func (m *MINIMTSService) setCameraConnected(connected bool) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()
	m.MTSStatus.CameraConnected = connected
}

// SetVideoData 由 HIKCameraService 调用，推送视频位移/应变数据并保存到缓存
func (m *MINIMTSService) SetVideoData(disp, strain float64) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()
	m.MTSStatus.VideoDisp = disp - m.zeroVideoDisp
	m.MTSStatus.VideoStrain = strain
	m.MTSStatus.Time = float64(time.Now().UnixMilli()-m.startTime.UnixMilli()) / 1000.0

	// 保存视频数据到缓存
	m.dataCache["time"] = append(m.dataCache["time"], m.MTSStatus.Time)
	m.dataCache["videoDisp"] = append(m.dataCache["videoDisp"], m.MTSStatus.VideoDisp)
	m.dataCache["videoStrain"] = append(m.dataCache["videoStrain"], m.MTSStatus.VideoStrain)
}

// uiPolling 独立前端数据推送协程，始终运行，10Hz
// 不依赖任何设备连接状态，有数据就发，无数据就发零
func (m *MINIMTSService) uiPolling() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.statusMu.RLock()
			status := m.MTSStatus
			m.statusMu.RUnlock()
			if m.app != nil {
				m.app.Event.Emit("update_status", status)
			}
		case <-m.stopCh:
			return
		}
	}
}

func (m *MINIMTSService) SaveDataToCache() {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()

	// 将当前状态数据添加到缓存
	if m.dataCache == nil {
		m.dataCache = make(map[string][]float64)
	}

	// 添加时间、载荷、应力、位移、应变、视频位移、视频应变数据
	m.dataCache["time"] = append(m.dataCache["time"], m.MTSStatus.Time)
	m.dataCache["load"] = append(m.dataCache["load"], m.MTSStatus.Load)
	m.dataCache["stress"] = append(m.dataCache["stress"], m.MTSStatus.Stress)
	m.dataCache["disp"] = append(m.dataCache["disp"], m.MTSStatus.Disp)
	m.dataCache["strain"] = append(m.dataCache["strain"], m.MTSStatus.Strain)
	m.dataCache["videoDisp"] = append(m.dataCache["videoDisp"], m.MTSStatus.VideoDisp)
	m.dataCache["videoStrain"] = append(m.dataCache["videoStrain"], m.MTSStatus.VideoStrain)
}

func (m *MINIMTSService) ClearDataCache() error {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()
	// 清空所有缓存数据
	m.dataCache = make(map[string][]float64)
	m.startTime = time.Now()
	return nil
}

func (m *MINIMTSService) SaveDataTCsv() error {
	filePath := m.project.ActiveConfig.FilePath
	fileName := m.project.ActiveConfig.FileName

	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}
	if fileName == "" {
		return fmt.Errorf("文件名称不能为空")
	}

	// 确保目录存在
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成唯一的 CSV 文件名
	csvPath := getUniqueFilePath(filePath, fileName, ".csv")

	// 生成唯一的 JSON 文件名
	jsonPath := getUniqueFilePath(filePath, fileName+"_config", ".json")

	// 写入 CSV 文件
	if err := m.writeCSV(csvPath); err != nil {
		return fmt.Errorf("写入 CSV 文件失败: %v", err)
	}
	fmt.Printf("CSV 文件已保存至: %s\n", csvPath)

	// 写入 JSON 配置文件
	if err := m.writeConfigJSON(jsonPath); err != nil {
		return fmt.Errorf("写入 JSON 配置文件失败: %v", err)
	}
	fmt.Printf("JSON 配置文件已保存至: %s\n", jsonPath)

	return nil
}

// getUniqueFilePath 生成唯一的文件路径，处理文件名冲突
func getUniqueFilePath(dir, baseName, ext string) string {
	counter := 0
	var filePath string

	for {
		if counter == 0 {
			filePath = filepath.Join(dir, baseName+ext)
		} else {
			filePath = filepath.Join(dir, fmt.Sprintf("%s(%d)%s", baseName, counter, ext))
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		counter++
	}

	return filePath
}

// writeCSV 将 dataCache 写入 CSV 文件
func (m *MINIMTSService) writeCSV(path string) error {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()

	if len(m.dataCache) == 0 {
		return fmt.Errorf("数据缓存为空")
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取所有 key 作为表头
	headers := make([]string, 0, len(m.dataCache))
	for key := range m.dataCache {
		headers = append(headers, key)
	}

	// 写入表头
	writer := csv.NewWriter(file)
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 确定最大行数
	maxLen := 0
	for _, values := range m.dataCache {
		if len(values) > maxLen {
			maxLen = len(values)
		}
	}

	// 逐行写入数据
	for i := 0; i < maxLen; i++ {
		row := make([]string, len(headers))
		for j, key := range headers {
			values := m.dataCache[key]
			if i < len(values) {
				row[j] = fmt.Sprintf("%.6f", values[i])
			} else {
				row[j] = ""
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// writeConfigJSON 将项目配置和系统配置写入 JSON 文件
func (m *MINIMTSService) writeConfigJSON(path string) error {
	type CombinedConfig struct {
		ProjectConfig *ProjectConfig `json:"projectConfig"`
		SystemConfig  *SystemConfig  `json:"systemConfig"`
	}

	combined := CombinedConfig{
		ProjectConfig: m.project.ActiveConfig,
		SystemConfig:  m.system.config,
	}

	data, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// CallDataToZero 数据归零函数
func (m *MINIMTSService) CallDataToZero(dataType string) error {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	switch dataType {
	case "load":
		// 更新 zeroLoad 为当前载荷值
		m.zeroLoad += m.MTSStatus.Load
		fmt.Printf("载荷归零: zeroLoad = %.6f\n", m.zeroLoad)
	case "disp":
		// 更新 zeroDisp 为当前位移值
		m.zeroDisp += m.MTSStatus.Disp
		fmt.Printf("位移归零: zeroDisp = %.6f\n", m.zeroDisp)
	case "videoDisp":
		m.zeroVideoDisp += m.MTSStatus.VideoDisp
		fmt.Printf("视频位移归零: zeroVideoDisp = %.6f\n", m.zeroVideoDisp)
	case "time":
		m.startTime = time.Now()
		fmt.Printf("时间归零: startTime = %v\n", m.startTime)

	default:
		return fmt.Errorf("invalid data type: %s", dataType)
	}
	return nil
}

// updateMotorPosition 针对电机协议 0x3e 92 的精准读取
// 3E 92 01 00 D1
func (m *MINIMTSService) updateMotorPosition() error {
	cmd := []byte{0x3e, 0x92, m.system.config.MotorID, 0x00}

	resp, err := m.comm.Send3EBus(cmd, 14, 0x3e)
	if err != nil {
		return err
	}

	if len(resp) == 14 && resp[1] == 0x92 {
		dataPart := resp[5:13]
		motorAngle := int64(binary.LittleEndian.Uint64(dataPart))
		length := float64(motorAngle) / 100.0
		trueVal := float64(m.system.config.MotorDirection) * length / m.system.config.MotorResolution

		// 写入缓存
		m.statusMu.Lock()
		m.MTSStatus.Disp = (trueVal - m.zeroDisp) * float64(m.displayDirection)
		m.MTSStatus.Strain = (trueVal - m.zeroDisp) / m.project.ActiveConfig.SectionLength
		if Debug {
			fmt.Println("Motor:", trueVal)
		}
		m.statusMu.Unlock()
	}
	return nil
}

// updateLoadData 读取载荷数据
// 02 03 00 00 00 02 C4 38
func (m *MINIMTSService) updateLoadData() error {
	resp, err := m.comm.SendModbus(m.system.config.WeighID, 0x03, 0x00, 0x02)
	if err != nil {
		return err
	}

	// 数据在索引 3 开始
	if len(resp) < 7 {
		return fmt.Errorf("invalid response length")
	}
	// 电机默认使用 BigEndian
	dataSwapped := []byte{resp[4], resp[3], resp[6], resp[5]}
	rawLoad := int32(binary.LittleEndian.Uint32(dataSwapped))

	load := float64(m.system.config.WeighDirection) * float64(rawLoad) / m.system.config.WeighResolution

	// 写入缓存
	m.statusMu.Lock()
	m.MTSStatus.Load = (load - m.zeroLoad) * float64(m.displayDirection)
	if m.project.ActiveConfig.SampleShape == "dogbone" {
		m.MTSStatus.Stress = (load - m.zeroLoad) / (m.project.ActiveConfig.Thickness * m.project.ActiveConfig.Width)
	} else if m.project.ActiveConfig.SampleShape == "cylinder" {
		m.MTSStatus.Stress = (load - m.zeroLoad) / (math.Pi * m.project.ActiveConfig.Diameter * m.project.ActiveConfig.Diameter)
	}
	if Debug {
		fmt.Println("Load:", load)
	}
	m.statusMu.Unlock()

	return nil
}

// 获取温度数据，临时使用
// func (m *MINIMTSService) updateTempData() error {
// 	// 读modbus寄存器获取温度数据, 功能码0x04
// 	resp, err := m.comm.SendModbus(m.system.TempID, 0x04, 0x00, 0x01)
// 	if err != nil {
// 		return err
// 	}
// 	// 解析温度数据
// 	if len(resp) >= 7 {
// 		dataPart := resp[3:5]
// 		//16位有符号整数
// 		rawVal := int16(binary.BigEndian.Uint16(dataPart))
// 		m.statusMu.Lock()
// 		m.MTSStatus.Temp = float64(rawVal) / 10.0
// 		if Debug {
// 			fmt.Println("温度:", m.MTSStatus.Temp)
// 		}
// 		m.statusMu.Unlock()
// 	}

// 	return nil

// }

// 获取限位状态, modbus, 功能码0x01, 读取离散输入
func (m *MINIMTSService) updateLimitStatus() error {
	resp, err := m.comm.SendModbus(m.system.config.LimitID, 0x04, 0x00, 0x02)
	// fmt.Println("获取限位状态:", resp)
	if err != nil {
		fmt.Println("获取限位状态失败:", err)
		return err
	}

	if len(resp) >= 8 {
		m.statusMu.Lock()
		if resp[4] == 0x01 {
			if m.MTSStatus.Limit == 0 {
				fmt.Println("限位1触发")
				m.app.Event.Emit("system_message", "限位1触发")
				m.MTSStatus.Limit = 1
				m.MotorStop()
			}
		} else if resp[6] == 0x01 {
			if m.MTSStatus.Limit == 0 {
				fmt.Println("限位2触发")
				m.app.Event.Emit("system_message", "限位2触发")
				m.MTSStatus.Limit = 2
				m.MotorStop()
			}
		} else {
			if m.MTSStatus.Limit != 0 {
				fmt.Println("限位解除")
				m.app.Event.Emit("system_message", "限位解除")
				m.MTSStatus.Limit = 0
			}
		}

		if Debug {
			fmt.Println("Limit:", m.MTSStatus.Limit)
		}
		m.statusMu.Unlock()
	}
	return nil
}

// MotorOn 电机使能
func (m *MINIMTSService) MotorOn() error {
	if m.system == nil {
		return fmt.Errorf("系统配置未初始化")
	}

	cmd := []byte{0x3e, 0x88, m.system.config.MotorID, 0x00}
	resp, err := m.comm.Send3EBus(cmd, 5, 0x3e)
	if err != nil {
		return err
	}

	if cmd[0] == resp[0] && cmd[3] == resp[3] {
		fmt.Printf("电机使能成功\n")
		return nil
	}
	return fmt.Errorf("电机使能失败")
}

// MotorOff 电机禁止使能
func (m *MINIMTSService) MotorOff() error {
	if m.system == nil {
		return fmt.Errorf("系统配置未初始化")
	}

	cmd := []byte{0x3e, 0x80, m.system.config.MotorID, 0x00}
	resp, err := m.comm.Send3EBus(cmd, 5, 0x3e)
	if err != nil {
		return err
	}

	if cmd[0] == resp[0] && cmd[3] == resp[3] {
		fmt.Printf("电机禁止使能成功\n")
		return nil
	}

	return fmt.Errorf("电机禁止使能失败")
}

// MotorStop 电机停止
func (m *MINIMTSService) MotorStop() error {
	if m.system == nil {
		return fmt.Errorf("系统配置未初始化")
	}
	cmd := []byte{0x3e, 0x81, m.system.config.MotorID, 0x00}
	resp, err := m.comm.Send3EBus(cmd, 5, 0x3e)
	if err != nil {
		return err
	}

	if cmd[0] == resp[0] && cmd[3] == resp[3] {
		return nil
	}

	return fmt.Errorf("电机停止失败")
}

// JogMove 电机速度模式移动
func (m *MINIMTSService) JogMove(speed float64) error {
	if !m.IsOpened() {
		return fmt.Errorf("设备未连接")
	}
	if m.system == nil {
		return fmt.Errorf("系统配置未初始化")
	}

	if m.MTSStatus.Limit != 0 {
		if m.MTSStatus.Limit == m.system.config.LimitDirection && speed > 0 {
			//正向限位，禁止正向运动
			fmt.Printf("正向限位，禁止正向运动\n")
			return fmt.Errorf("正向限位，禁止正向运动")
		} else if m.MTSStatus.Limit != m.system.config.LimitDirection && speed < 0 {
			//反向限位，禁止反向运动
			fmt.Printf("反向限位，禁止反向运动\n")
			return fmt.Errorf("反向限位，禁止反向运动")
		}
	}

	if speed == 0 {
		return m.MotorStop()
	}

	// 计算速率 (使用 system 里的 MotorResolution)
	rate := int32(float64(m.system.config.MotorDirection) * speed * m.system.config.MotorResolution * 100)
	// 构造指令头 (4字节: 3e A2 ID 04)
	header := []byte{0x3e, 0xA2, m.system.config.MotorID, 0x04}
	headerWithCheck := m.comm.Calculate3ECheckSum(header)

	// 构造数据体 (4字节速率值)
	// 按照协议，速度通常采用小端序
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(rate))
	dataWithCheck := m.comm.Calculate3ECheckSum(data)

	// 4. 拼接完整指令
	fullCmd := append(headerWithCheck, dataWithCheck...)

	// 5. 发送并同步获取回复
	resp, err := m.comm.Send3EBusRaw(fullCmd, 5, 0x3e)
	if err != nil {
		fmt.Printf("发送Jog命令失败: %v\n", err)
		return err
	}

	respCheck := m.comm.Calculate3ECheckSum(resp[:4])

	if respCheck[4] != resp[4] {
		fmt.Println("Jog运动失败")
		return fmt.Errorf("Jog运动失败")
	} else {
		return nil
	}
}

// StartMeasurement 开始测量逻辑
func (m *MINIMTSService) StartMeasurement() error {
	if m.system == nil {
		return fmt.Errorf("系统配置未初始化")
	}
	if m.project == nil {
		return fmt.Errorf("项目配置未初始化")
	}

	m.MotorOn()

	TestSpeed := m.project.ActiveConfig.Speed
	TestTpye := m.project.ActiveConfig.Type
	fmt.Printf("开始测量: 类型=%s, 速度=%.2f\n", TestTpye, TestSpeed)

	if TestTpye == "tension" {
		m.displayDirection = 1
		return m.JogMove(TestSpeed)
	} else if TestTpye == "compression" {
		m.displayDirection = -1
		return m.JogMove(-TestSpeed)
	} else {
		fmt.Printf("未知的测试类型: %s\n", TestTpye)
		return fmt.Errorf("未知的测试类型: %s", TestTpye)
	}
}

// StopMeasurement 停止测量
func (m *MINIMTSService) StopMeasurement() error {
	return m.MotorStop()
}

// EmergencyStop 紧急停止
func (m *MINIMTSService) EmergencyStop() error {
	m.MotorStop()
	return m.MotorOff()
}
