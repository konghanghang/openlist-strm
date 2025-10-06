package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/konghanghang/openlist-strm/internal/alist"
	"github.com/konghanghang/openlist-strm/internal/config"
	"github.com/konghanghang/openlist-strm/internal/storage"
	"github.com/konghanghang/openlist-strm/internal/strm"
	"github.com/robfig/cron/v3"
)

// Scheduler manages task scheduling and execution
type Scheduler struct {
	cfg         *config.Config
	alistClient *alist.Client
	generator   *strm.Generator
	db          *storage.DB
	cron        *cron.Cron
	cronJobs    map[uint]cron.EntryID // mapping ID -> cron entry ID
	mu          sync.RWMutex          // protect cronJobs map
}

// New creates a new scheduler
func New(cfg *config.Config, alistClient *alist.Client, generator *strm.Generator, db *storage.DB) *Scheduler {
	return &Scheduler{
		cfg:         cfg,
		alistClient: alistClient,
		generator:   generator,
		db:          db,
		cron:        cron.New(),
		cronJobs:    make(map[uint]cron.EntryID),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	// Load all mappings with cron expressions and register them
	mappings, err := s.db.ListMappings()
	if err != nil {
		return fmt.Errorf("failed to list mappings: %w", err)
	}

	for _, mapping := range mappings {
		if mapping.Enabled && mapping.CronExpr != "" {
			if err := s.AddCronJob(mapping.ID, mapping.Name, mapping.CronExpr); err != nil {
				log.Printf("Failed to add cron job for mapping %s: %v", mapping.Name, err)
			}
		}
	}

	s.cron.Start()
	log.Printf("Scheduler started with %d cron jobs", len(s.cronJobs))
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		log.Println("Scheduler stopped")
	}
}

// RunAll runs all enabled mappings (from database)
func (s *Scheduler) RunAll(ctx context.Context) error {
	mappings, err := s.db.ListEnabledMappings()
	if err != nil {
		return fmt.Errorf("failed to list mappings: %w", err)
	}

	for _, mapping := range mappings {
		// Parse extensions from database (comma-separated string)
		extensions := strings.Split(mapping.Extensions, ",")
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}

		mappingConfig := config.MappingConfig{
			Name:       mapping.Name,
			Source:     mapping.Source,
			Target:     mapping.Target,
			Extensions: extensions,
			Concurrent: mapping.Concurrent,
			Mode:       mapping.Mode,
			STRMMode:   mapping.STRMMode,
			Enabled:    mapping.Enabled,
		}

		if err := s.RunMapping(ctx, mappingConfig); err != nil {
			log.Printf("Failed to run mapping %s: %v", mapping.Name, err)
		}
	}
	return nil
}

// RunMapping runs a single mapping
func (s *Scheduler) RunMapping(ctx context.Context, mapping config.MappingConfig) error {
	taskID := uuid.New().String()

	// Create task record
	task := &storage.Task{
		TaskID:     taskID,
		ConfigName: mapping.Name,
		Mode:       mapping.Mode,
		Status:     "running",
		StartedAt:  time.Now(),
	}
	if err := s.db.CreateTask(task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	log.Printf("Task %s started for mapping %s", taskID, mapping.Name)

	// Generate STRM files
	result, err := s.generator.Generate(ctx, strm.GenerateOptions{
		SourcePath: mapping.Source,
		TargetPath: mapping.Target,
		Extensions: mapping.Extensions,
		Concurrent: mapping.Concurrent,
		Mode:       mapping.Mode,
		STRMMode:   mapping.STRMMode,
	})

	// Update task record
	now := time.Now()
	task.CompletedAt = &now

	if err != nil {
		task.Status = "failed"
		task.Errors = err.Error()
		s.db.UpdateTask(task)
		return fmt.Errorf("generation failed: %w", err)
	}

	task.Status = "completed"
	task.FilesCreated = result.FilesCreated
	task.FilesDeleted = result.FilesDeleted
	task.FilesSkipped = result.FilesSkipped

	if len(result.Errors) > 0 {
		errMsg := ""
		for _, e := range result.Errors {
			errMsg += e.Error() + "; "
		}
		task.Errors = errMsg
	}

	if err := s.db.UpdateTask(task); err != nil {
		log.Printf("Failed to update task: %v", err)
	}

	log.Printf("Task %s completed: created=%d, deleted=%d, skipped=%d, errors=%d",
		taskID, result.FilesCreated, result.FilesDeleted, result.FilesSkipped, len(result.Errors))

	return nil
}

// RunMappingByName runs a mapping by name (from database)
func (s *Scheduler) RunMappingByName(ctx context.Context, name string) error {
	mapping, err := s.db.GetMappingByName(name)
	if err != nil {
		return fmt.Errorf("mapping not found: %s", name)
	}

	// Parse extensions from database (comma-separated string)
	extensions := strings.Split(mapping.Extensions, ",")
	for i := range extensions {
		extensions[i] = strings.TrimSpace(extensions[i])
	}

	mappingConfig := config.MappingConfig{
		Name:       mapping.Name,
		Source:     mapping.Source,
		Target:     mapping.Target,
		Extensions: extensions,
		Concurrent: mapping.Concurrent,
		Mode:       mapping.Mode,
		STRMMode:   mapping.STRMMode,
		Enabled:    mapping.Enabled,
	}

	return s.RunMapping(ctx, mappingConfig)
}

// GetTaskStatus gets task status by task ID
func (s *Scheduler) GetTaskStatus(taskID string) (*storage.Task, error) {
	return s.db.GetTaskByID(taskID)
}

// AddCronJob adds a cron job for a mapping
func (s *Scheduler) AddCronJob(mappingID uint, mappingName, cronExpr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job if any
	if entryID, exists := s.cronJobs[mappingID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronJobs, mappingID)
	}

	// Add new cron job
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		log.Printf("Running scheduled task for mapping: %s", mappingName)
		if err := s.RunMappingByName(context.Background(), mappingName); err != nil {
			log.Printf("Scheduled task failed for mapping %s: %v", mappingName, err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cronJobs[mappingID] = entryID
	log.Printf("Cron job added for mapping %s (ID: %d) with expression: %s", mappingName, mappingID, cronExpr)
	return nil
}

// RemoveCronJob removes a cron job for a mapping
func (s *Scheduler) RemoveCronJob(mappingID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.cronJobs[mappingID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronJobs, mappingID)
		log.Printf("Cron job removed for mapping ID: %d", mappingID)
	}
}

// UpdateCronJob updates a cron job for a mapping
func (s *Scheduler) UpdateCronJob(mappingID uint, mappingName, cronExpr string, enabled bool) error {
	// If disabled or no cron expression, remove the job
	if !enabled || cronExpr == "" {
		s.RemoveCronJob(mappingID)
		return nil
	}

	// Otherwise, add/update the job
	return s.AddCronJob(mappingID, mappingName, cronExpr)
}
