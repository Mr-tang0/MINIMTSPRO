package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ProjectConfig 对应 ProjectSettings.vue 中的 form 对象
type ProjectConfig struct {
	Experimenter string `json:"experimenter"`
	SampleNo     string `json:"sampleNo"`
	TestDate     string `json:"testDate"`

	SampleShape   string  `json:"sampleShape"` // dogbone | cylinder
	Width         float64 `json:"width"`
	Thickness     float64 `json:"thickness"`
	Diameter      float64 `json:"diameter"`
	SectionLength float64 `json:"sectionLength"`

	Type          string  `json:"type"` // tension | compression
	Speed         float64 `json:"speed"`
	StopCondition string  `json:"stopCondition"`

	FilePath string `json:"filePath"` // 准静态保存路径
	FileName string `json:"fileName"` // 准静态文件名

	DicEnable       bool   `json:"dicEnable"`
	ExternalTrigger bool   `json:"externalTrigger"`
	TriggerType     string `json:"triggerType"`
	TriggerInterval int    `json:"triggerInterval"`
	PulseWidth      int    `json:"pulseWidth"`

	DicFolder   string `json:"dicFolder"`
	DicFileName string `json:"dicFileName"`

	VideoExtEnable bool    `json:"videoExtEnable"`
	MarkerA        string  `json:"markerA"`
	MarkerB        string  `json:"markerB"`
	PixLength      float64 `json:"pixLength"`
	PhysLength     float64 `json:"physLength"`
	PoissonEnable  bool    `json:"poissonEnable"`
}

type ProjectService struct {
	// ctx          context.Context
	ActiveConfig *ProjectConfig
	user         *User
}

func NewProjectService(user *User) *ProjectService {
	p := &ProjectService{
		ActiveConfig: &ProjectConfig{
			Experimenter:    user.Name,
			SampleNo:        "Sample001",
			TestDate:        time.Now().Format("2006-01-02 15:04:05"),
			SampleShape:     "dogbone",
			Width:           1.5,
			Thickness:       2.0,
			Diameter:        2.0,
			SectionLength:   10.0,
			Type:            "tension",
			Speed:           0.01,
			StopCondition:   "manual",
			FilePath:        "",
			FileName:        "test",
			DicEnable:       false,
			ExternalTrigger: false,
			TriggerType:     "interval",
			TriggerInterval: 1000,
			PulseWidth:      100,
			DicFolder:       "",
			DicFileName:     "dic",
			VideoExtEnable:  false,
			MarkerA:         "A",
			MarkerB:         "B",
			PixLength:       100.0,
			PhysLength:      10.0,
			PoissonEnable:   false,
		},
		user: user,
	}
	return p
}

func (p *ProjectService) SetUser(user *User) {
	p.user = user
}

// LoadProjectConfigFromFile 给定路径读取历史项目信息
func (p *ProjectService) LoadProjectConfigFromFile() (*ProjectConfig, error) {
	fmt.Println("当前用户:", p.user.Name)
	path, _ := os.UserHomeDir()
	path = filepath.Join(path, "PIMS", "MINIMTS", "localuser", fmt.Sprintf("%s_%s", p.user.Name, p.user.ID), "config.json")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("未找到历史项目配置文件: %s\n", path)
		defaultConfig := ProjectConfig{
			Experimenter:    p.user.Name,
			SampleNo:        "Sample001",
			TestDate:        time.Now().Format("2006-01-02 15:04:05"),
			SampleShape:     "dogbone",
			Width:           10.0,
			Thickness:       5.0,
			Diameter:        0.0,
			SectionLength:   50.0,
			Type:            "tension",
			Speed:           1.0,
			StopCondition:   "strain",
			FilePath:        "",
			FileName:        "test",
			DicEnable:       false,
			ExternalTrigger: false,
			TriggerType:     "interval",
			TriggerInterval: 1000,
			PulseWidth:      100,
			DicFolder:       "",
			DicFileName:     "dic",
			VideoExtEnable:  false,
			MarkerA:         "A",
			MarkerB:         "B",
			PixLength:       100.0,
			PhysLength:      10.0,
			PoissonEnable:   false,
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败: %v", err)
		}

		p.SaveProjectConfig(defaultConfig)
		return &defaultConfig, nil
	}

	fmt.Println("尝试加载历史项目配置..." + path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}

	p.ActiveConfig = &config
	fmt.Printf("已成功加载历史项目: %s\n", config.SampleNo)
	return &config, nil
}

// SaveProjectConfig 保存项目信息到指定位置
func (p *ProjectService) SaveProjectConfig(config ProjectConfig) error {
	config.TestDate = time.Now().Format("2006-01-02 15:04:05")

	fullPath, _ := os.UserHomeDir()
	fullPath = filepath.Join(fullPath, "PIMS", "MINIMTS", "localuser", fmt.Sprintf("%s_%s", p.user.Name, p.user.ID), "config.json")

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	p.ActiveConfig = &config
	fmt.Println("已保存项目: \n", p.ActiveConfig)
	return nil
}

// GetActiveConfig 获取后端当前持有的配置（供前端调用）
func (p *ProjectService) GetActiveConfig() (*ProjectConfig, error) {
	return p.LoadProjectConfigFromFile()
}

// SelectDirectory 选择目录（供前端调用）
func (p *ProjectService) SelectDirectory() (string, error) {
	app := application.Get()
	if app == nil {
		fmt.Println("app is nil")
		return "", fmt.Errorf("application not initialized")
	}

	path, err := app.Dialog.OpenFile().
		SetTitle("Select Folder").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()

	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}

	fmt.Println("Selected path:", path)
	return path, nil
}
