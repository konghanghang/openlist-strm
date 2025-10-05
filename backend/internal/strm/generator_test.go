package strm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/konghanghang/openlist-strm/internal/alist"
)

// Mock Alist Client
type mockAlistClient struct {
	files      []alist.FileItem
	fileURLs   map[string]string
	shouldFail bool
}

func (m *mockAlistClient) Ping(ctx context.Context) error {
	if m.shouldFail {
		return os.ErrPermission
	}
	return nil
}

func (m *mockAlistClient) ListFilesRecursive(ctx context.Context, dirPath string, extensions []string) ([]alist.FileItem, error) {
	if m.shouldFail {
		return nil, os.ErrPermission
	}
	return m.files, nil
}

func (m *mockAlistClient) GetFileURL(ctx context.Context, filePath string) (string, error) {
	if m.shouldFail {
		return "", os.ErrPermission
	}
	if url, ok := m.fileURLs[filePath]; ok {
		return url, nil
	}
	return "http://example.com/" + filePath, nil
}

func TestNewGenerator(t *testing.T) {
	client := &mockAlistClient{}
	gen := NewGenerator(client, 10)

	if gen == nil {
		t.Fatal("NewGenerator() returned nil")
	}
	if gen.concurrent != 10 {
		t.Errorf("Generator.concurrent = %v, want 10", gen.concurrent)
	}
}

func TestGenerate_FullMode(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "strm")

	// Mock client with test files
	mockClient := &mockAlistClient{
		files: []alist.FileItem{
			{Name: "movie1.mp4", IsDir: false},
			{Name: "movie2.mkv", IsDir: false},
		},
		fileURLs: map[string]string{
			"/movies/movie1.mp4": "http://example.com/movies/movie1.mp4",
			"/movies/movie2.mkv": "http://example.com/movies/movie2.mkv",
		},
	}

	gen := NewGenerator(mockClient, 5)
	ctx := context.Background()

	opts := GenerateOptions{
		SourcePath: "/movies",
		TargetPath: targetDir,
		Extensions: []string{"mp4", "mkv"},
		Mode:       "full",
	}

	result, err := gen.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.FilesCreated != 2 {
		t.Errorf("FilesCreated = %v, want 2", result.FilesCreated)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}

	// Verify STRM files were created
	strmFile1 := filepath.Join(targetDir, "movie1.strm")
	strmFile2 := filepath.Join(targetDir, "movie2.strm")

	if _, err := os.Stat(strmFile1); os.IsNotExist(err) {
		t.Errorf("STRM file %s was not created", strmFile1)
	}
	if _, err := os.Stat(strmFile2); os.IsNotExist(err) {
		t.Errorf("STRM file %s was not created", strmFile2)
	}

	// Verify content
	content, err := os.ReadFile(strmFile1)
	if err != nil {
		t.Fatalf("Failed to read STRM file: %v", err)
	}
	if string(content) != "http://example.com/movies/movie1.mp4" {
		t.Errorf("STRM content = %s, want http://example.com/movies/movie1.mp4", string(content))
	}
}

func TestGenerate_IncrementalMode(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "strm")

	mockClient := &mockAlistClient{
		files: []alist.FileItem{
			{Name: "new.mp4", IsDir: false},
		},
		fileURLs: map[string]string{
			"/movies/new.mp4": "http://example.com/new",
		},
	}

	gen := NewGenerator(mockClient, 5)
	ctx := context.Background()

	opts := GenerateOptions{
		SourcePath: "/movies",
		TargetPath: targetDir,
		Extensions: []string{"mp4"},
		Mode:       "incremental",
	}

	// First run - create file
	result1, err := gen.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() first run error = %v", err)
	}

	if result1.FilesCreated != 1 {
		t.Errorf("First run: FilesCreated = %v, want 1", result1.FilesCreated)
	}

	// Verify STRM file was created
	strmFile := filepath.Join(targetDir, "new.strm")
	if _, err := os.Stat(strmFile); os.IsNotExist(err) {
		t.Fatalf("STRM file %s was not created in first run", strmFile)
	}

	// Read file content from first run
	content1, _ := os.ReadFile(strmFile)

	// Second run - should skip existing file
	result2, err := gen.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() second run error = %v", err)
	}

	if result2.FilesCreated != 0 {
		// Debug: list files in target directory
		entries, _ := os.ReadDir(targetDir)
		t.Logf("Files in targetDir after second run: %v", entries)
		t.Logf("STRM file path: %s", strmFile)
		t.Logf("STRM file exists: %v", fileExists(strmFile))
		content2, _ := os.ReadFile(strmFile)
		t.Logf("First run content: %s", string(content1))
		t.Logf("Second run content: %s", string(content2))
		t.Errorf("Second run: FilesCreated = %v, want 0 (file should be skipped)", result2.FilesCreated)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestGenerate_WithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "strm")

	mockClient := &mockAlistClient{
		files: []alist.FileItem{
			{Name: "action/movie1.mp4", IsDir: false},
			{Name: "comedy/movie2.mp4", IsDir: false},
		},
		fileURLs: map[string]string{
			"/movies/action/movie1.mp4": "http://example.com/action/movie1.mp4",
			"/movies/comedy/movie2.mp4": "http://example.com/comedy/movie2.mp4",
		},
	}

	gen := NewGenerator(mockClient, 5)
	ctx := context.Background()

	opts := GenerateOptions{
		SourcePath: "/movies",
		TargetPath: targetDir,
		Extensions: []string{"mp4"},
		Mode:       "full",
	}

	result, err := gen.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.FilesCreated != 2 {
		t.Errorf("FilesCreated = %v, want 2", result.FilesCreated)
	}

	// Verify directory structure is maintained
	strmFile1 := filepath.Join(targetDir, "action", "movie1.strm")
	strmFile2 := filepath.Join(targetDir, "comedy", "movie2.strm")

	if _, err := os.Stat(strmFile1); os.IsNotExist(err) {
		t.Errorf("STRM file %s was not created", strmFile1)
	}
	if _, err := os.Stat(strmFile2); os.IsNotExist(err) {
		t.Errorf("STRM file %s was not created", strmFile2)
	}
}

func TestGenerate_ChineseFilename(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "strm")

	mockClient := &mockAlistClient{
		files: []alist.FileItem{
			{Name: "电影/测试电影.mp4", IsDir: false},
		},
		fileURLs: map[string]string{
			"/movies/电影/测试电影.mp4": "http://example.com/电影/测试电影.mp4",
		},
	}

	gen := NewGenerator(mockClient, 5)
	ctx := context.Background()

	opts := GenerateOptions{
		SourcePath: "/movies",
		TargetPath: targetDir,
		Extensions: []string{"mp4"},
		Mode:       "full",
	}

	result, err := gen.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.FilesCreated != 1 {
		t.Errorf("FilesCreated = %v, want 1", result.FilesCreated)
	}

	// Verify Chinese filename is handled correctly
	strmFile := filepath.Join(targetDir, "电影", "测试电影.strm")
	if _, err := os.Stat(strmFile); os.IsNotExist(err) {
		t.Errorf("STRM file %s was not created", strmFile)
	}
}

func TestGenerate_AlistError(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "strm")

	mockClient := &mockAlistClient{
		shouldFail: true,
	}

	gen := NewGenerator(mockClient, 5)
	ctx := context.Background()

	opts := GenerateOptions{
		SourcePath: "/movies",
		TargetPath: targetDir,
		Extensions: []string{"mp4"},
		Mode:       "full",
	}

	_, err := gen.Generate(ctx, opts)
	if err == nil {
		t.Error("Generate() expected error when Alist fails, got nil")
	}
}

func TestGenerate_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "strm")

	// Create many files to ensure context cancellation is tested
	var files []alist.FileItem
	for i := 0; i < 100; i++ {
		files = append(files, alist.FileItem{
			Name:  filepath.Join("movie", string(rune('a'+i))+".mp4"),
			IsDir: false,
		})
	}

	mockClient := &mockAlistClient{
		files: files,
	}

	gen := NewGenerator(mockClient, 5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := GenerateOptions{
		SourcePath: "/movies",
		TargetPath: targetDir,
		Extensions: []string{"mp4"},
		Mode:       "full",
	}

	_, err := gen.Generate(ctx, opts)
	if err != context.Canceled {
		t.Errorf("Generate() error = %v, want context.Canceled", err)
	}
}

func TestChangeExtension(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		newExt   string
		want     string
	}{
		{
			name:     "mp4 to strm",
			filePath: "/path/to/movie.mp4",
			newExt:   ".strm",
			want:     "/path/to/movie.strm",
		},
		{
			name:     "mkv to strm",
			filePath: "/path/to/video.mkv",
			newExt:   ".strm",
			want:     "/path/to/video.strm",
		},
		{
			name:     "no extension",
			filePath: "/path/to/file",
			newExt:   ".strm",
			want:     "/path/to/file.strm",
		},
		{
			name:     "multiple dots",
			filePath: "/path/to/file.name.mp4",
			newExt:   ".strm",
			want:     "/path/to/file.name.strm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := changeExtension(tt.filePath, tt.newExt)
			if got != tt.want {
				t.Errorf("changeExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "clean_test")

	// Create test directory with files
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("test"), 0644)
	os.MkdirAll(filepath.Join(testDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(testDir, "subdir", "file3.txt"), []byte("test"), 0644)

	// Clean directory
	err := cleanDirectory(testDir)
	if err != nil {
		t.Fatalf("cleanDirectory() error = %v", err)
	}

	// Verify directory is empty
	entries, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Directory should be empty, got %d entries", len(entries))
	}
}

func TestCleanDirectory_NonExistent(t *testing.T) {
	err := cleanDirectory("/non/existent/directory")
	if err != nil {
		t.Errorf("cleanDirectory() should not error on non-existent directory, got %v", err)
	}
}
