package backend

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
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
	Temp        float64 `json:"temp"`        // 当前温度
	Limit       int     `json:"limit"`       // 当前限位（0-无，1-端口1，2-端口2）
	Time        float64 `json:"time"`        // 实验时间，单位秒
}

const Debug = false

type MINIMTSService struct {
	MTSStatus Status       `json:"status"`
	statusMu  sync.RWMutex // 保护 MTSStatus 的读写锁

	comm *SerialCommunicator // 串口通讯器

	user    *User
	project *Project
	system  *System
	ctx     context.Context
	app     *application.App // Wails 应用实例

	zeroDisp float64 // 零位位移值，用于计算相对位移
	zeroLoad float64 // 零位载荷值，用于计算相对载荷

	isPolling bool      // 标记是否正在后台轮询
	startTime time.Time // 开始时间，用于计算实验时间
}

func NewMINIMTSService() *MINIMTSService {
	ctx := context.Background()
	system := NewSystem()
	system.SetContext(ctx)
	user := User{}
	project := NewProject(&user)

	return &MINIMTSService{
		MTSStatus: Status{
			Load:        0,
			Stress:      0,
			Disp:        0,
			Strain:      0,
			VideoDisp:   0,
			VideoStrain: 0,
			Temp:        0,
			Limit:       0,
		},
		ctx:       ctx,
		system:    system,
		user:      &user,
		project:   project,
		startTime: time.Now(),
	}
}

func (m *MINIMTSService) SetApp(app *application.App) {
	m.app = app
}

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

	// 启动后台高速查询协程
	m.isPolling = true
	go m.Polling()
	go m.UIPolling()

	return nil
}

func (m *MINIMTSService) CloseMINIMTS() error {
	m.isPolling = false // 停止轮询

	if m.comm != nil {
		err := m.comm.Disconnect()
		m.comm = nil
		return err
	}
	return nil
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
	SencorsEnable := map[string]bool{"load": true, "temp": true, "disp": true, "limit": true}

	ErrorCounter := func(sensor string, err error) {
		if err == nil {
			ErrorCounts[sensor] = 0
		} else {
			ErrorCounts[sensor] += 1
			fmt.Println("错误计数器:", sensor, "累计错误次数:", ErrorCounts[sensor])
			// 如果错误累计次数达到阈值，该传感器被禁用,除非重新连接，协程重启
			if ErrorCounts[sensor] >= 3 {
				fmt.Println("传感器", sensor, "已禁用")
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
		if SencorsEnable["temp"] {
			ErrorCounter("temp", m.updateTempData())
		}
		if SencorsEnable["limit"] {
			ErrorCounter("limit", m.updateLimitStatus())
		}

		// 每25ms更新一次数据
		time.Sleep(25 * time.Millisecond)
	}

}

// 收集所有传感器数据打包给前端:子线程，固定5Hz
func (m *MINIMTSService) UIPolling() {
	for m.isPolling {
		m.statusMu.RLock()
		m.MTSStatus.Time = float64((time.Now().UnixMilli() - m.startTime.UnixMilli())) / 1000.0
		m.app.Event.Emit("update_status", m.MTSStatus)
		m.statusMu.RUnlock()
		time.Sleep(100 * time.Millisecond)
	}
}

// updateMotorPosition 针对电机协议 0x3e 92 的精准读取
// 3E 92 01 00 D1
func (m *MINIMTSService) updateMotorPosition() error {
	cmd := []byte{0x3e, 0x92, m.system.MotorID, 0x00}

	resp, err := m.comm.Send3EBus(cmd, 14, 0x3e)
	if err != nil {
		return err
	}

	if len(resp) == 14 && resp[1] == 0x92 {
		dataPart := resp[5:13]
		motorAngle := int64(binary.LittleEndian.Uint64(dataPart))
		length := float64(motorAngle) / 100.0
		trueVal := float64(m.system.MotorDirection) * length / m.system.MotorResolution

		// 写入缓存
		m.statusMu.Lock()
		m.MTSStatus.Disp = trueVal
		m.MTSStatus.Strain = trueVal / m.project.ActiveConfig.SectionLength
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
	resp, err := m.comm.SendModbus(m.system.WeighID, 0x03, 0x00, 0x02)
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

	load := float64(m.system.WeighDirection) * float64(rawLoad) / m.system.WeighResolution

	// 写入缓存
	m.statusMu.Lock()
	m.MTSStatus.Load = load
	if m.project.ActiveConfig.SampleShape == "dogbone" {
		m.MTSStatus.Stress = load / (m.project.ActiveConfig.Thickness * m.project.ActiveConfig.Width)
	} else if m.project.ActiveConfig.SampleShape == "cylinder" {
		m.MTSStatus.Stress = load / (math.Pi * m.project.ActiveConfig.Diameter * m.project.ActiveConfig.Diameter)
	}
	if Debug {
		fmt.Println("Load:", load)
	}
	m.statusMu.Unlock()

	return nil
}

// 获取温度数据，临时使用
func (m *MINIMTSService) updateTempData() error {
	//读modbus寄存器获取温度数据, 功能码0x04
	resp, err := m.comm.SendModbus(m.system.TempID, 0x04, 0x00, 0x01)
	if err != nil {
		return err
	}
	// 解析温度数据
	if len(resp) >= 7 {
		dataPart := resp[3:5]
		//16位有符号整数
		rawVal := int16(binary.BigEndian.Uint16(dataPart))
		m.statusMu.Lock()
		m.MTSStatus.Temp = float64(rawVal) / 10.0
		if Debug {
			fmt.Println("温度:", m.MTSStatus.Temp)
		}
		m.statusMu.Unlock()
	}

	return nil

}

// 获取限位状态, modbus, 功能码0x01, 读取离散输入
func (m *MINIMTSService) updateLimitStatus() error {
	resp, err := m.comm.SendModbus(m.system.LimitID, 0x04, 0x00, 0x02)
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

	cmd := []byte{0x3e, 0x88, m.system.MotorID, 0x00}
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

	cmd := []byte{0x3e, 0x80, m.system.MotorID, 0x00}
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
	cmd := []byte{0x3e, 0x81, m.system.MotorID, 0x00}
	resp, err := m.comm.Send3EBus(cmd, 5, 0x3e)
	if err != nil {
		return err
	}

	if cmd[0] == resp[0] && cmd[3] == resp[3] {
		fmt.Printf("电机停止成功\n")
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
		if m.MTSStatus.Limit == m.system.LimitDirection && speed > 0 {
			//正向限位，禁止正向运动
			fmt.Printf("正向限位，禁止正向运动\n")
			return fmt.Errorf("正向限位，禁止正向运动")
		} else if m.MTSStatus.Limit != m.system.LimitDirection && speed < 0 {
			//反向限位，禁止反向运动
			fmt.Printf("反向限位，禁止反向运动\n")
			return fmt.Errorf("反向限位，禁止反向运动")
		}
	}

	if speed == 0 {
		return m.MotorStop()
	}

	// 计算速率 (使用 system 里的 MotorResolution)
	rate := int32(float64(m.system.MotorDirection) * speed * m.system.MotorResolution * 100)
	// 构造指令头 (4字节: 3e A2 ID 04)
	header := []byte{0x3e, 0xA2, m.system.MotorID, 0x04}
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
		return m.JogMove(TestSpeed)
	} else if TestTpye == "compression" {
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
