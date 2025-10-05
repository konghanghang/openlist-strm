package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  host: "127.0.0.1"
  port: 9090

alist:
  url: "http://test-alist:5244"
  token: "test-token-123"
  sign_enabled: true
  timeout: 60

strm:
  output_dir: "/test/strm"
  concurrent: 20
  extensions:
    - mp4
    - mkv

mappings:
  - name: "Test Movies"
    source: "/movies"
    target: "/strm/movies"
    mode: "incremental"
    enabled: true

database:
  path: "/test/data.db"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load config
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config values
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %v, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %v, want 9090", cfg.Server.Port)
	}
	if cfg.Alist.URL != "http://test-alist:5244" {
		t.Errorf("Alist.URL = %v, want http://test-alist:5244", cfg.Alist.URL)
	}
	if cfg.Alist.Token != "test-token-123" {
		t.Errorf("Alist.Token = %v, want test-token-123", cfg.Alist.Token)
	}
	if cfg.STRM.Concurrent != 20 {
		t.Errorf("STRM.Concurrent = %v, want 20", cfg.STRM.Concurrent)
	}
	if len(cfg.Mappings) != 1 {
		t.Errorf("len(Mappings) = %v, want 1", len(cfg.Mappings))
	}
	if cfg.Mappings[0].Name != "Test Movies" {
		t.Errorf("Mappings[0].Name = %v, want Test Movies", cfg.Mappings[0].Name)
	}
}

func TestLoad_FileNotExists(t *testing.T) {
	_, err := Load("/non/existent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for non-existent file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `
server:
  host: "127.0.0.1"
  port: invalid_port
`

	if err := os.WriteFile(configFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := Load(configFile)
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}

func TestLoad_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "missing alist url",
			content: `
server:
  host: "127.0.0.1"
  port: 8080
alist:
  token: "test-token"
`,
			wantErr: true,
		},
		{
			name: "missing alist token",
			content: `
server:
  host: "127.0.0.1"
  port: 8080
alist:
  url: "http://localhost:5244"
`,
			wantErr: true,
		},
		{
			name: "invalid port",
			content: `
server:
  host: "127.0.0.1"
  port: 99999
alist:
  url: "http://localhost:5244"
  token: "test-token"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			_, err := Load(configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check default values
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("DefaultConfig().Server.Host = %v, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("DefaultConfig().Server.Port = %v, want 8080", cfg.Server.Port)
	}
	if cfg.STRM.Concurrent != 10 {
		t.Errorf("DefaultConfig().STRM.Concurrent = %v, want 10", cfg.STRM.Concurrent)
	}
	if len(cfg.STRM.Extensions) == 0 {
		t.Error("DefaultConfig().STRM.Extensions should not be empty")
	}
	if cfg.Log.Level != "info" {
		t.Errorf("DefaultConfig().Log.Level = %v, want info", cfg.Log.Level)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
				Alist:  AlistConfig{URL: "http://localhost:5244", Token: "token"},
				STRM:   STRMConfig{Concurrent: 10, Extensions: []string{"mp4"}},
				Mappings: []MappingConfig{
					{Source: "/src", Target: "/dst", Mode: "incremental"},
				},
				Database: DatabaseConfig{Path: "./data.db"},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: -1},
				Alist:  AlistConfig{URL: "http://localhost:5244", Token: "token"},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: 99999},
				Alist:  AlistConfig{URL: "http://localhost:5244", Token: "token"},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "missing alist url",
			config: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
				Alist:  AlistConfig{Token: "token"},
			},
			wantErr: true,
			errMsg:  "alist url is required",
		},
		{
			name: "missing alist token",
			config: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
				Alist:  AlistConfig{URL: "http://localhost:5244"},
			},
			wantErr: true,
			errMsg:  "alist token is required",
		},
		{
			name: "invalid mapping mode",
			config: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
				Alist:  AlistConfig{URL: "http://localhost:5244", Token: "token"},
				Mappings: []MappingConfig{
					{Source: "/src", Target: "/dst", Mode: "invalid"},
				},
			},
			wantErr: true,
			errMsg:  "mode must be 'incremental' or 'full'",
		},
		{
			name: "missing mapping source",
			config: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
				Alist:  AlistConfig{URL: "http://localhost:5244", Token: "token"},
				Mappings: []MappingConfig{
					{Target: "/dst", Mode: "incremental"},
				},
			},
			wantErr: true,
			errMsg:  "source path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestGetAddr(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "default",
			cfg: &Config{
				Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
			},
			want: "0.0.0.0:8080",
		},
		{
			name: "custom host and port",
			cfg: &Config{
				Server: ServerConfig{Host: "127.0.0.1", Port: 9090},
			},
			want: "127.0.0.1:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetAddr()
			if got != tt.want {
				t.Errorf("GetAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
