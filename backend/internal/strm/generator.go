package strm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/konghanghang/openlist-strm/internal/alist"
	"github.com/konghanghang/openlist-strm/internal/contextkeys"
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
}

// NewGenerator creates a new STRM generator
func NewGenerator(alistClient AlistClient) *Generator {
	return &Generator{
		alistClient: alistClient,
	}
}

// GenerateOptions represents options for generating STRM files
type GenerateOptions struct {
	SourcePath string
	TargetPath string
	Extensions []string
	Concurrent int    // concurrent for this task
	Mode       string // incremental or full
	STRMMode   string // alist_path or http_url
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
	// Extract trace ID from context if available
	traceID := getTraceID(ctx)

	result := &GenerateResult{
		Errors: []error{},
	}

	// Create target directory if not exists
	if err := os.MkdirAll(opts.TargetPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Full mode: clean target directory
	if opts.Mode == "full" {
		log.Printf("[TraceID: %s] Cleaning target directory: %s", traceID, opts.TargetPath)
		if err := cleanDirectory(opts.TargetPath); err != nil {
			return nil, fmt.Errorf("failed to clean directory: %w", err)
		}
	}

	// List all video files from Alist
	log.Printf("[TraceID: %s] Scanning source directory: %s", traceID, opts.SourcePath)
	files, err := g.alistClient.ListFilesRecursive(ctx, opts.SourcePath, opts.Extensions)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	log.Printf("[TraceID: %s] Found %d video files to process", traceID, len(files))

	// Deduplicate files by priority (when same filename with different extensions)
	files = deduplicateFilesByPriority(files, traceID)
	log.Printf("[TraceID: %s] After deduplication: %d files to process", traceID, len(files))

	// Validate concurrent value
	concurrent := opts.Concurrent
	if concurrent <= 0 {
		concurrent = 10 // Default to 10
	}

	// Generate STRM files concurrently
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrent)
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
			created, err := g.generateSTRMFile(ctx, f, opts, traceID)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors, err)
				mu.Unlock()
				log.Printf("[TraceID: %s] ❌ ERROR: %s -> %v", traceID, f.Path, err)
			} else if created {
				mu.Lock()
				result.FilesCreated++
				mu.Unlock()
				log.Printf("[TraceID: %s] ✅ CREATED: %s", traceID, f.Path)
			} else {
				mu.Lock()
				result.FilesSkipped++
				mu.Unlock()
				log.Printf("[TraceID: %s] ⏭️  SKIPPED: %s (already exists)", traceID, f.Path)
			}
		}(file)
	}

	wg.Wait()

	return result, nil
}

// generateSTRMFile generates a single STRM file
// Returns (created, error) where created is true if a new file was created
func (g *Generator) generateSTRMFile(ctx context.Context, file alist.FileItem, opts GenerateOptions, traceID string) (bool, error) {
	// Calculate relative path from full path
	relPath := strings.TrimPrefix(file.Path, opts.SourcePath)
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

	// Determine STRM file content based on mode
	var strmContent string
	if opts.STRMMode == "alist_path" {
		// MediaWarp mode: write Alist path
		strmContent = file.Path
	} else {
		// Direct URL mode: get actual file URL
		fileURL, err := g.alistClient.GetFileURL(ctx, file.Path)
		if err != nil {
			return false, fmt.Errorf("failed to get URL for %s: %w", file.Path, err)
		}
		strmContent = fileURL
	}

	// Write STRM file
	if err := os.WriteFile(strmPath, []byte(strmContent), 0644); err != nil {
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

// getTraceID extracts trace ID from context, returns "unknown" if not found
func getTraceID(ctx context.Context) string {
	// Get trace ID from context using proper key
	if traceID := ctx.Value(contextkeys.TraceIDKey); traceID != nil {
		if taskID, ok := traceID.(string); ok && len(taskID) >= 8 {
			return taskID[:8]
		}
	}
	return "unknown"
}

// getExtensionPriority returns the priority of a file extension (lower is better)
func getExtensionPriority(ext string) int {
	// Priority map: lower number = higher priority
	priorities := map[string]int{
		".mkv":  1,  // Best quality, supports multiple audio/subtitle tracks
		".mp4":  2,  // Good compatibility
		".avi":  3,  // Older format
		".mov":  4,  // Apple format
		".wmv":  5,  // Windows format
		".flv":  6,  // Flash video
		".m4v":  7,  // iTunes video
		".mpg":  8,  // MPEG
		".mpeg": 9,  // MPEG
		".3gp":  10, // Mobile format
		".webm": 11, // Web format
	}

	if priority, ok := priorities[strings.ToLower(ext)]; ok {
		return priority
	}
	return 999 // Unknown formats get lowest priority
}

// deduplicateFilesByPriority removes duplicate files (same name, different extension)
// keeping only the one with highest priority
func deduplicateFilesByPriority(files []alist.FileItem, traceID string) []alist.FileItem {
	if len(files) == 0 {
		return files
	}

	// Group files by base name (without extension)
	fileMap := make(map[string][]alist.FileItem)

	for _, file := range files {
		// Get base name without extension
		ext := filepath.Ext(file.Path)
		baseName := strings.TrimSuffix(file.Path, ext)

		fileMap[baseName] = append(fileMap[baseName], file)
	}

	// Select the best file for each base name
	result := make([]alist.FileItem, 0, len(fileMap))
	duplicateCount := 0

	for baseName, group := range fileMap {
		if len(group) == 1 {
			// No duplicates, just add it
			result = append(result, group[0])
		} else {
			// Multiple files with same base name, select by priority
			duplicateCount += len(group) - 1

			bestFile := group[0]
			bestPriority := getExtensionPriority(filepath.Ext(bestFile.Path))

			// Log duplicate files
			var extensions []string
			for _, f := range group {
				extensions = append(extensions, filepath.Ext(f.Path))
			}
			log.Printf("[TraceID: %s] 🔍 DUPLICATE: %s has multiple formats: %v",
				traceID, filepath.Base(baseName), extensions)

			for i := 1; i < len(group); i++ {
				currentPriority := getExtensionPriority(filepath.Ext(group[i].Path))
				if currentPriority < bestPriority {
					bestFile = group[i]
					bestPriority = currentPriority
				}
			}

			log.Printf("[TraceID: %s] ✅ SELECTED: %s (priority: %d)",
				traceID, filepath.Base(bestFile.Path), bestPriority)

			result = append(result, bestFile)
		}
	}

	if duplicateCount > 0 {
		log.Printf("[TraceID: %s] Removed %d duplicate files based on priority", traceID, duplicateCount)
	}

	return result
}
