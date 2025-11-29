package config

import (
	"github.com/spf13/viper"
)

// Config 應用程式配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Session  SessionConfig  `mapstructure:"session"`
	Mail     MailConfig     `mapstructure:"mail"`
}

// ServerConfig 伺服器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

// DatabaseConfig 資料庫配置
type DatabaseConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

// SessionConfig Session 配置
type SessionConfig struct {
	Secret string `mapstructure:"secret"`
}

// MailConfig 郵件配置
type MailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	From         string `mapstructure:"from"`
	BaseURL      string `mapstructure:"base_url"`
}

// Load 載入配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 環境變數
	viper.SetEnvPrefix("HIGGSTV")
	viper.AutomaticEnv()

	// 預設值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.env", "development")
	viper.SetDefault("database.uri", "mongodb://localhost:27017")
	viper.SetDefault("database.database", "higgstv")
	viper.SetDefault("session.secret", "change-me-in-production")

	if err := viper.ReadInConfig(); err != nil {
		// 如果找不到配置檔，使用環境變數和預設值
		_ = err // 忽略錯誤，繼續使用預設值
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}

