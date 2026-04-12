package config

import (
	"encoding/json"
	"errors"
	"os"
)

// ServerConfig 用来保存 HTTP 服务本身的配置。
type ServerConfig struct {
	Host                string `json:"host"`
	Port                int    `json:"port"`
	ReadTimeoutSeconds  int    `json:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `json:"write_timeout_seconds"`
}

// DatabaseConfig 用来保存 MySQL 连接参数。
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"database_name"`
	Charset      string `json:"charset"`
}

// AuthConfig 保存后台登录和默认管理员配置。
type AuthConfig struct {
	TokenSecret             string `json:"token_secret"`
	DefaultAdminUsername    string `json:"default_admin_username"`
	DefaultAdminPassword    string `json:"default_admin_password"`
	DefaultAdminDisplayName string `json:"default_admin_display_name"`
}

// ProviderConfig 表示一个 AI Provider 的接入配置。
type ProviderConfig struct {
	Type    string `json:"type"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
	GroupID string `json:"group_id"`
	AppID   string `json:"app_id"`
}

// AIRouting 用来定义“某个功能默认走哪个 Provider”。
type AIRouting struct {
	AnalyzeTitle string `json:"analyze_title"`
	Summary      string `json:"summary"`
	Chat         string `json:"chat"`
	Moderate     string `json:"moderate"`
	Image        string `json:"image"`
	Embedding    string `json:"embedding"`
	TTS          string `json:"tts"`
}

// AIConfig 保存 AI 模块所有配置。
type AIConfig struct {
	Routing   AIRouting                 `json:"routing"`
	Providers map[string]ProviderConfig `json:"providers"`
}

// OSSConfig 保存阿里云 OSS 的连接信息。
type OSSConfig struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Bucket          string `json:"bucket"`
	Prefix          string `json:"prefix"`
}

// SyncConfig 保存同步任务配置。
type SyncConfig struct {
	Enabled         bool `json:"enabled"`
	IntervalMinutes int  `json:"interval_minutes"`
}

// AppConfig 是整个项目的总配置。
type AppConfig struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	AI       AIConfig       `json:"ai"`
	OSS      OSSConfig      `json:"oss"`
	Sync     SyncConfig     `json:"sync"`
}

// Load 会优先读取 APP_CONFIG 指定的配置文件。
// 如果没有指定，就默认读取 configs/config.local.json。
// 如果本地配置文件不存在，就直接返回内置默认配置。
func Load() (AppConfig, error) {
	configPath := os.Getenv("APP_CONFIG")
	if configPath == "" {
		configPath = "configs/config.local.json"
	}

	defaultConfig := Default()

	_, err := os.Stat(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return defaultConfig, nil
	}

	if err != nil {
		return AppConfig{}, err
	}

	rawBytes, err := os.ReadFile(configPath)
	if err != nil {
		return AppConfig{}, err
	}

	if err := json.Unmarshal(rawBytes, &defaultConfig); err != nil {
		return AppConfig{}, err
	}

	return defaultConfig, nil
}

// Default 返回一份本地开发能直接使用的默认配置。
func Default() AppConfig {
	return AppConfig{
		Server: ServerConfig{
			Host:                "0.0.0.0",
			Port:                8080,
			ReadTimeoutSeconds:  15,
			WriteTimeoutSeconds: 15,
		},
		Database: DatabaseConfig{
			Host:         "127.0.0.1",
			Port:         3306,
			Username:     "root",
			Password:     "123456",
			DatabaseName: "ai_blog",
			Charset:      "utf8mb4",
		},
		Auth: AuthConfig{
			TokenSecret:             "replace-with-your-own-long-secret",
			DefaultAdminUsername:    "admin",
			DefaultAdminPassword:    "admin123456",
			DefaultAdminDisplayName: "超级管理员",
		},
		AI: AIConfig{
			Routing: AIRouting{
				AnalyzeTitle: "glm",
				Summary:      "glm",
				Chat:         "minimax",
				Moderate:     "xiaomi",
				Image:        "openai",
				Embedding:    "ollama",
				TTS:          "minimax",
			},
			Providers: map[string]ProviderConfig{
				"glm": {
					Type:    "glm",
					BaseURL: "https://open.bigmodel.cn/api/paas/v4",
					Model:   "glm-5",
				},
				"minimax": {
					Type:    "minimax",
					BaseURL: "https://api.minimaxi.com/anthropic/v1",
					Model:   "MiniMax-M2.7",
				},
				"xiaomi": {
					Type:    "xiaomi",
					BaseURL: "https://api.xiaomimimo.com/v1",
					Model:   "mimo-v2-pro",
				},
				"openai": {
					Type:    "openai",
					BaseURL: "https://api.openai.com/v1",
					Model:   "gpt-4.1-mini",
				},
				"ollama": {
					Type:    "ollama",
					BaseURL: "http://127.0.0.1:11434",
					Model:   "qwen2.5:7b",
				},
			},
		},
		OSS: OSSConfig{
			Endpoint: "",
			Region:   "",
			Bucket:   "",
			Prefix:   "blog",
		},
		Sync: SyncConfig{
			Enabled:         false,
			IntervalMinutes: 60,
		},
	}
}
