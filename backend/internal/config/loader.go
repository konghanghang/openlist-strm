package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Load loads configuration from file
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default config paths
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")              // 当前工作目录
		v.AddConfigPath("./configs")      // 项目根目录的 configs/
		v.AddConfigPath("../configs")     // 从 backend/ 目录向上查找
		v.AddConfigPath("/etc/openlist-strm")  // 系统目录
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal to config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// LoadOrDefault loads configuration or returns default
func LoadOrDefault(configPath string) (*Config, error) {
	cfg, err := Load(configPath)
	if err != nil {
		if os.IsNotExist(err) || configPath == "" {
			// Return default config
			return DefaultConfig(), nil
		}
		return nil, err
	}
	return cfg, nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Alist: AlistConfig{
			URL:         "http://localhost:5244",
			Token:       "",
			SignEnabled: false,
			Timeout:     30,
		},
		Schedule: ScheduleConfig{
			Enabled: false,
			Cron:    "0 2 * * *",
		},
		API: APIConfig{
			Enabled: true,
			Token:   "",
			Timeout: 300,
		},
		Web: WebConfig{
			Enabled:  true,
			Username: "admin",
			Password: "admin123",
		},
		Log: LogConfig{
			Level:      "info",
			File:       "",
			MaxSize:    100,
			MaxBackups: 3,
		},
		Database: DatabaseConfig{
			Path: "./data/openlist-strm.db",
		},
	}
}
