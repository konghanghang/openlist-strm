package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/konghanghang/openlist-strm/internal/config"
)

var (
	// Default loggers
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
	Debug *log.Logger
)

// Init initializes the logger
func Init(cfg *config.LogConfig) error {
	var writers []io.Writer

	// Always log to stdout
	writers = append(writers, os.Stdout)

	// Log to file if configured
	if cfg.File != "" {
		// Create log directory
		dir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		logFile, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		writers = append(writers, logFile)
	}

	multiWriter := io.MultiWriter(writers...)

	// Initialize loggers
	Info = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(multiWriter, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	// Set log level
	if cfg.Level != "debug" {
		Debug.SetOutput(io.Discard)
	}

	return nil
}

// Close closes any open log files
func Close() {
	// For now, we don't need to explicitly close
	// because we're using the default loggers
}
