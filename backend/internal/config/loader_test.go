package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadWithSourceExplicitPath(t *testing.T) {
	configFile := writeTestConfig(t, t.TempDir(), "custom.yaml")

	cfg, usedPath, err := LoadWithSource(configFile)
	if err != nil {
		t.Fatalf("LoadWithSource returned error: %v", err)
	}

	if cfg.Server.Port != 18080 {
		t.Fatalf("unexpected server port: got %d want %d", cfg.Server.Port, 18080)
	}

	if !sameFilePath(t, usedPath, configFile) {
		t.Fatalf("unexpected config path: got %s want %s", usedPath, configFile)
	}
}

func TestLoadWithSourceDefaultSearchPath(t *testing.T) {
	workDir := t.TempDir()
	configDir := filepath.Join(workDir, "configs")
	configFile := writeTestConfig(t, configDir, "config.yaml")

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("failed to restore working directory: %v", chdirErr)
		}
	})

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	cfg, usedPath, err := LoadWithSource("")
	if err != nil {
		t.Fatalf("LoadWithSource returned error: %v", err)
	}

	if cfg.Alist.URL != "http://127.0.0.1:5244" {
		t.Fatalf("unexpected alist url: got %s", cfg.Alist.URL)
	}

	if !sameFilePath(t, usedPath, configFile) {
		t.Fatalf("unexpected config path: got %s want %s", usedPath, configFile)
	}
}

func TestLoadWithSourceReportsSearchedPaths(t *testing.T) {
	workDir := t.TempDir()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(oldWD); chdirErr != nil {
			t.Fatalf("failed to restore working directory: %v", chdirErr)
		}
	})

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	_, _, err = LoadWithSource("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "config file not found, searched:") {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(err.Error(), filepath.Join(workDir, "config.yaml")) {
		t.Fatalf("searched paths missing workdir config.yaml: %v", err)
	}
}

func TestBuildDefaultConfigCandidates(t *testing.T) {
	workDir := "/tmp/project/backend"
	execPath := "/opt/openlist-strm/bin/openlist-strm"

	got := buildDefaultConfigCandidates(workDir, execPath)
	want := []string{
		filepath.Clean("/tmp/project/backend/config.yaml"),
		filepath.Clean("/tmp/project/backend/config.yml"),
		filepath.Clean("/tmp/project/backend/configs/config.yaml"),
		filepath.Clean("/tmp/project/backend/configs/config.yml"),
		filepath.Clean("/tmp/project/config.yaml"),
		filepath.Clean("/tmp/project/config.yml"),
		filepath.Clean("/tmp/project/configs/config.yaml"),
		filepath.Clean("/tmp/project/configs/config.yml"),
		filepath.Clean("/opt/openlist-strm/bin/config.yaml"),
		filepath.Clean("/opt/openlist-strm/bin/config.yml"),
		filepath.Clean("/opt/openlist-strm/bin/configs/config.yaml"),
		filepath.Clean("/opt/openlist-strm/bin/configs/config.yml"),
		filepath.Clean("/opt/openlist-strm/config.yaml"),
		filepath.Clean("/opt/openlist-strm/config.yml"),
		filepath.Clean("/opt/openlist-strm/configs/config.yaml"),
		filepath.Clean("/opt/openlist-strm/configs/config.yml"),
		filepath.Clean("/app/configs/config.yaml"),
		filepath.Clean("/app/configs/config.yml"),
		filepath.Clean("/etc/openlist-strm/config.yaml"),
		filepath.Clean("/etc/openlist-strm/config.yml"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected candidates:\n got: %v\nwant: %v", got, want)
	}
}

func writeTestConfig(t *testing.T, dir, name string) string {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	configFile := filepath.Join(dir, name)
	content := `server:
  host: "127.0.0.1"
  port: 18080
alist:
  url: "http://127.0.0.1:5244"
  token: "test-token"
  sign_enabled: false
  timeout: 30
api:
  enabled: true
  token: ""
  timeout: 300
web:
  enabled: true
  username: "admin"
  password: "admin123"
log:
  level: "info"
  file: ""
  max_size: 100
  max_backups: 3
database:
  path: "./data/openlist-strm.db"
media_server:
  enabled: false
  type: ""
`

	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	return configFile
}

func sameFilePath(t *testing.T, left, right string) bool {
	t.Helper()

	leftInfo, err := os.Stat(left)
	if err != nil {
		t.Fatalf("Stat returned error for %s: %v", left, err)
	}

	rightInfo, err := os.Stat(right)
	if err != nil {
		t.Fatalf("Stat returned error for %s: %v", right, err)
	}

	return os.SameFile(leftInfo, rightInfo)
}
