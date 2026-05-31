package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type System struct {
	MotorID         byte    `json:"motor_id"`         // 当前电机ID
	MotorResolution float64 `json:"motor_resolution"` // 电机分辨率
	MotorDirection  int     `json:"motor_direction"`  // 电机方向，1 或 -1

	WeighID         byte    `json:"weigh_id"`         // 当前称重模块ID
	WeighResolution float64 `json:"weigh_resolution"` // 称重模块分辨率
	WeighDirection  int     `json:"weigh_direction"`  // 称重模块方向，1 或 -1

	LimitID        byte `json:"limit_id"`        // 当前限位模块ID
	LimitEnabled   bool `json:"limit_enabled"`   // 是否启用限位模块
	LimitDirection int  `json:"limit_direction"` // 限位模块方向，1 或 -1

	TempID      byte `json:"temp_id"`      // 当前温度模块ID
	TempEnabled bool `json:"temp_enabled"` // 是否启用温度模块

	ctx context.Context
}

func NewSystem() *System {
	return &System{
		MotorID:         1,          // 默认电机ID
		MotorResolution: 1312.47251, // 默认电机分辨率 (单位: mm)
		MotorDirection:  1,          // 默认电机方向

		WeighID:         2,   // 默认称重模块ID
		WeighResolution: 1.0, // 默认称重模块分辨率 (单位: kg)
		WeighDirection:  1,   // 默认称重模块方向

		LimitID:        4,     // 默认限位模块ID
		LimitEnabled:   false, // 默认不启用限位模块
		LimitDirection: 2,     // 默认限位模块方向：CW对应的端口

		TempID:      5,    // 默认温度模块ID
		TempEnabled: true, // 默认不启用温度模块
	}
}

func (s *System) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// 获取配置文件的存放路径
func (s *System) getConfigPath() (string, error) {
	homeDir, _ := os.UserHomeDir()

	// 定义文件夹路径 (不包含文件名)
	configDir := filepath.Join(homeDir, "PIMS", "MINIMTS", "system")

	// 确保目录存在
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// 返回最终的文件完整路径
	return filepath.Join(configDir, "config.json"), nil
}

// GetConfigFromLocalFile 从本地路径读取并解码配置
func (s *System) GetConfigFromLocalFile() (*System, error) {
	path, err := s.getConfigPath()
	if err != nil {
		return nil, err
	}
	fmt.Printf("尝试从本地文件加载系统配置: %s\n", path)

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 如果不存在，返回当前默认配置并保存一份
		_ = s.UpdateConfigToLocalFile(s)
		return s, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解码到当前结构体
	err = json.Unmarshal(data, s)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return s, nil
}

// UpdateConfigToLocalFile 将传入的配置保存到本地文件并更新内存状态
func (s *System) UpdateConfigToLocalFile(newConfig *System) error {
	path, err := s.getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %v", err)
	}

	// 将对象序列化为 JSON (带缩进以便阅读)
	data, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// 写入文件
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	// 同步更新内存中的当前对象值（排除 ctx）
	s.MotorID = newConfig.MotorID
	s.MotorResolution = newConfig.MotorResolution
	s.MotorDirection = newConfig.MotorDirection
	s.WeighID = newConfig.WeighID
	s.WeighResolution = newConfig.WeighResolution
	s.WeighDirection = newConfig.WeighDirection
	s.LimitID = newConfig.LimitID
	s.LimitEnabled = newConfig.LimitEnabled
	s.LimitDirection = newConfig.LimitDirection
	s.TempID = newConfig.TempID
	s.TempEnabled = newConfig.TempEnabled

	return nil
}
