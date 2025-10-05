package strm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/konghanghang/openlist-strm/internal/alist"
)

// AlistClient is an interface for Alist operations
type AlistClient interface {
	Ping(ctx context.Context) error
	ListFilesRecursive(ctx context.Context, dirPath string, extensions []string) ([]alist.FileItem, error)
	GetFileURL(ctx context.Context, filePath string) (string, error)
}

// Generator generates STRM files
type Generator struct {
	alistClient AlistClient
	concurrent  int
}

// NewGenerator creates a new STRM generator
func NewGenerator(alistClient AlistClient, concurrent int) *Generator {
	return &Generator{
		alistClient: alistClient,
		concurrent:  concurrent,
	}
}

// GenerateOptions represents options for generating STRM files
type GenerateOptions struct {
	SourcePath string
	TargetPath string
	Extensions []string
	Mode       string // incremental or full
}

// GenerateResult represents the result of generation
type GenerateResult struct {
	FilesCreated int
	FilesDeleted int
	FilesSkipped int
	Errors       []error
}

// Generate generates STRM files for a directory
func (g *Generator) Generate(ctx context.Context, opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Errors: []error{},
	}

	// Create target directory if not exists
	if err := os.MkdirAll(opts.TargetPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Full mode: clean target directory
	if opts.Mode == "full" {
		if err := cleanDirectory(opts.TargetPath); err != nil {
			return nil, fmt.Errorf("failed to clean directory: %w", err)
		}
	}

	// List all video files from Alist
	files, err := g.alistClient.ListFilesRecursive(ctx, opts.SourcePath, opts.Extensions)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Generate STRM files concurrently
	var wg sync.WaitGroup
	sem := make(chan struct{}, g.concurrent)
	mu := &sync.Mutex{}

	for _, file := range files {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(f alist.FileItem) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Generate STRM file
			created, err := g.generateSTRMFile(ctx, f, opts)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors, err)
				mu.Unlock()
			} else if created {
				mu.Lock()
				result.FilesCreated++
				mu.Unlock()
			} else {
				mu.Lock()
				result.FilesSkipped++
				mu.Unlock()
			}
		}(file)
	}

	wg.Wait()

	return result, nil
}

// generateSTRMFile generates a single STRM file
// Returns (created, error) where created is true if a new file was created
func (g *Generator) generateSTRMFile(ctx context.Context, file alist.FileItem, opts GenerateOptions) (bool, error) {
	// Calculate relative path
	relPath := strings.TrimPrefix(file.Name, opts.SourcePath)
	relPath = strings.TrimPrefix(relPath, "/")

	// Calculate target STRM file path
	strmPath := filepath.Join(opts.TargetPath, relPath)
	strmPath = changeExtension(strmPath, ".strm")

	// Create parent directory
	parentDir := filepath.Dir(strmPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	// Check if STRM file already exists (for incremental mode)
	if opts.Mode == "incremental" {
		if _, err := os.Stat(strmPath); err == nil {
			// File exists, skip
			return false, nil
		}
	}

	// Get file URL from Alist
	sourcePath := filepath.Join(opts.SourcePath, file.Name)
	fileURL, err := g.alistClient.GetFileURL(ctx, sourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to get URL for %s: %w", sourcePath, err)
	}

	// Write STRM file
	if err := os.WriteFile(strmPath, []byte(fileURL), 0644); err != nil {
		return false, fmt.Errorf("failed to write STRM file %s: %w", strmPath, err)
	}

	return true, nil
}

// cleanDirectory removes all files in a directory
func cleanDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}

// changeExtension changes the file extension
func changeExtension(filePath, newExt string) string {
	ext := filepath.Ext(filePath)
	return filePath[:len(filePath)-len(ext)] + newExt
}
