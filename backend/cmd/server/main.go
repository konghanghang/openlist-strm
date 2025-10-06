package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/konghanghang/openlist-strm/internal/alist"
	"github.com/konghanghang/openlist-strm/internal/api"
	"github.com/konghanghang/openlist-strm/internal/config"
	"github.com/konghanghang/openlist-strm/internal/logger"
	"github.com/konghanghang/openlist-strm/internal/scheduler"
	"github.com/konghanghang/openlist-strm/internal/storage"
	"github.com/konghanghang/openlist-strm/internal/strm"
	"github.com/konghanghang/openlist-strm/internal/web"
)

var (
	version = "1.0.0"

	configPath = flag.String("config", "", "Path to configuration file")
	showVer    = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	if *showVer {
		fmt.Printf("OpenList-STRM v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info.Printf("Starting OpenList-STRM v%s", version)
	logger.Info.Printf("Config loaded from: %s", *configPath)

	// Initialize database
	db, err := storage.New(cfg.Database.Path)
	if err != nil {
		logger.Error.Printf("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error.Printf("Failed to close database: %v", err)
		}
	}()
	logger.Info.Printf("Database initialized: %s", cfg.Database.Path)

	// Note: Mappings are now managed via Web UI and stored in database only
	// No YAML sync is performed - use Web UI to create your first mapping

	// Create Alist client
	alistClient := alist.NewClient(
		cfg.Alist.URL,
		cfg.Alist.Token,
		cfg.Alist.SignEnabled,
		cfg.Alist.Timeout,
	)
	logger.Info.Printf("Alist client created: %s", cfg.Alist.URL)

	// Test Alist connection
	ctx := context.Background()
	if err := alistClient.Ping(ctx); err != nil {
		logger.Warn.Printf("Failed to ping Alist server: %v", err)
	} else {
		logger.Info.Println("Alist server is accessible")
	}

	// Create STRM generator (concurrency is now per-mapping)
	generator := strm.NewGenerator(alistClient)
	logger.Info.Println("STRM generator created")

	// Create and start scheduler
	sched := scheduler.New(cfg, alistClient, generator, db)
	if err := sched.Start(); err != nil {
		logger.Error.Printf("Failed to start scheduler: %v", err)
		os.Exit(1)
	}
	defer sched.Stop()

	// Create API server
	apiServer := api.NewServer(cfg, sched, db)

	// Register Web UI routes
	if cfg.Web.Enabled {
		if err := web.RegisterRoutes(apiServer.GetRouter()); err != nil {
			logger.Error.Printf("Failed to register web routes: %v", err)
		} else {
			logger.Info.Println("Web UI enabled")
		}
	}

	// Start API server in background
	go func() {
		logger.Info.Printf("API server starting on %s", cfg.GetAddr())
		if err := apiServer.Run(); err != nil {
			logger.Error.Printf("API server error: %v", err)
		}
	}()

	// Run initial generation if mappings exist (optional, can be disabled)
	// if len(cfg.Mappings) > 0 {
	// 	logger.Info.Println("Running initial generation...")
	// 	if err := sched.RunAll(ctx); err != nil {
	// 		logger.Error.Printf("Initial generation failed: %v", err)
	// 	} else {
	// 		logger.Info.Println("Initial generation completed")
	// 	}
	// }

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info.Printf("OpenList-STRM is running on http://%s", cfg.GetAddr())
	logger.Info.Println("Press Ctrl+C to exit")
	<-sigChan

	logger.Info.Println("Shutting down gracefully...")
}
