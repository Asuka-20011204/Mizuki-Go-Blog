package config

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

// Config 总配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	System   SystemConfig   `yaml:"system"`
	Log      LogConfig      `yaml:"log"`
	InitData InitDataConfig `yaml:"init_data"`
}

type LogConfig struct {
	Dir   string `yaml:"dir"`   // 日志文件目录，如 "./logs"
	Level string `yaml:"level"` // debug / info / warn / error
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug 或 release
}

type DatabaseConfig struct {
	Dsn string `yaml:"dsn"` // MySQL 连接字符串
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"` // 过期时间(小时)
}

type SystemConfig struct {
	PostsDir      string `yaml:"posts_dir"`
	PreviewDir    string `yaml:"preview_dir"`
	FrontendDir   string `yaml:"frontend_dir"`
	DiaryJsonPath string `yaml:"diary_json_path"`
}

type InitDataConfig struct {
	OwnerUsername string `yaml:"owner_username"`
	OwnerPassword string `yaml:"owner_password"`
	OwnerQQ       string `yaml:"owner_qq"`
	OwnerPhone    string `yaml:"owner_phone"`
}

var GlobalConfig *Config

// LoadConfig 从指定路径加载配置，环境变量优先覆盖敏感字段
func LoadConfig(path string) {
	conf := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	err = yaml.Unmarshal(data, conf)
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	// 环境变量覆盖（敏感信息不从配置文件读取）
	if v := os.Getenv("DB_DSN"); v != "" {
		conf.Database.Dsn = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		conf.JWT.Secret = v
	}
	if v := os.Getenv("OWNER_PASSWORD"); v != "" {
		conf.InitData.OwnerPassword = v
	}

	GlobalConfig = conf
}
