package service

import (
	"encoding/json"
	"my-blog-backend/config"
	"my-blog-backend/models"
	"os"
	"path/filepath"
)

type ConfigService struct{}

// getConfigFilePath 返回 settings.json 的完整路径
func (s *ConfigService) getConfigFilePath() string {
	return filepath.Join(config.GlobalConfig.System.FrontendDir, "src/data/settings.json")
}

// GetConfig 读取并解析配置文件
func (s *ConfigService) GetConfig() (*models.FullConfig, error) {
	filePath := s.getConfigFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg models.FullConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpdateConfig 将配置对象写入文件（带格式化）
func (s *ConfigService) UpdateConfig(cfg *models.FullConfig) error {
	filePath := s.getConfigFilePath()

	// 格式化 JSON，带缩进
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
