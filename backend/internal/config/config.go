package config

import (
	"fmt"
	"time"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Alist    AlistConfig    `mapstructure:"alist"`
	Schedule ScheduleConfig `mapstructure:"schedule"`
	API      APIConfig      `mapstructure:"api"`
	Web      WebConfig      `mapstructure:"web"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// AlistConfig represents Alist server configuration
type AlistConfig struct {
	URL         string        `mapstructure:"url"`
	Token       string        `mapstructure:"token"`
	SignEnabled bool          `mapstructure:"sign_enabled"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

// MappingConfig represents path mapping configuration (internal use, not from YAML)
type MappingConfig struct {
	Name       string
	Source     string
	Target     string
	Extensions []string
	Concurrent int
	Mode       string
	STRMMode   string
	Enabled    bool
}

// ScheduleConfig represents schedule configuration
type ScheduleConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Cron    string `mapstructure:"cron"`
}

// APIConfig represents API configuration
type APIConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Token   string        `mapstructure:"token"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// WebConfig represents Web UI configuration
type WebConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Alist.URL == "" {
		return fmt.Errorf("alist url is required")
	}

	if c.Alist.Token == "" {
		return fmt.Errorf("alist token is required")
	}

	if c.Database.Path == "" {
		c.Database.Path = "./data/openlist-strm.db"
	}

	return nil
}

// GetAddr returns the server address
func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
