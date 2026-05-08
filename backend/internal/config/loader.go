package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load loads configuration from file
func Load(configPath string) (*Config, error) {
	cfg, _, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadWithSource loads configuration and returns the resolved config path.
func LoadWithSource(configPath string) (*Config, string, error) {
	return loadConfig(configPath)
}

func loadConfig(configPath string) (*Config, string, error) {
	v := viper.New()

	resolvedPath, err := resolveConfigPath(configPath)
	if err != nil {
		return nil, "", err
	}
	v.SetConfigFile(resolvedPath)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal to config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, "", fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, v.ConfigFileUsed(), nil
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

func resolveConfigPath(configPath string) (string, error) {
	if configPath != "" {
		return filepath.Clean(configPath), nil
	}

	candidates := defaultConfigCandidates()
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to stat config file %s: %w", candidate, err)
		}
	}

	return "", fmt.Errorf("config file not found, searched: %s", strings.Join(candidates, ", "))
}

func defaultConfigCandidates() []string {
	workDir, _ := os.Getwd()
	execPath, _ := os.Executable()
	return buildDefaultConfigCandidates(workDir, execPath)
}

func buildDefaultConfigCandidates(workDir, execPath string) []string {
	seen := make(map[string]struct{})
	var candidates []string

	addDirCandidates := func(dir string) {
		if dir == "" {
			return
		}
		for _, name := range []string{"config.yaml", "config.yml"} {
			candidate := filepath.Clean(filepath.Join(dir, name))
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			candidates = append(candidates, candidate)
		}
	}

	addDirCandidates(workDir)
	addDirCandidates(filepath.Join(workDir, "configs"))
	addDirCandidates(filepath.Join(workDir, ".."))
	addDirCandidates(filepath.Join(workDir, "..", "configs"))

	if execPath != "" {
		execDir := filepath.Dir(execPath)
		addDirCandidates(execDir)
		addDirCandidates(filepath.Join(execDir, "configs"))
		addDirCandidates(filepath.Join(execDir, ".."))
		addDirCandidates(filepath.Join(execDir, "..", "configs"))
	}

	addDirCandidates("/app/configs")
	addDirCandidates("/etc/openlist-strm")

	return candidates
}
