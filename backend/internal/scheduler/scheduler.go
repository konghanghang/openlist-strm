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
		cron:        cron.New(cron.WithSeconds()), // Support second-level cron expressions
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

	log.Printf("[Scheduler] Loading mappings from database, found %d total mappings", len(mappings))

	registeredCount := 0
	for _, mapping := range mappings {
		log.Printf("[Scheduler] Processing mapping: name=%s, enabled=%v, cron_expr=%s",
			mapping.Name, mapping.Enabled, mapping.CronExpr)

		if mapping.Enabled && mapping.CronExpr != "" {
			if err := s.AddCronJob(mapping.ID, mapping.Name, mapping.CronExpr); err != nil {
				log.Printf("[Scheduler] ERROR: Failed to add cron job for mapping %s: %v", mapping.Name, err)
			} else {
				registeredCount++
			}
		} else {
			if !mapping.Enabled {
				log.Printf("[Scheduler] Skipping mapping %s: disabled", mapping.Name)
			} else if mapping.CronExpr == "" {
				log.Printf("[Scheduler] Skipping mapping %s: no cron expression", mapping.Name)
			}
		}
	}

	s.cron.Start()
	log.Printf("[Scheduler] Scheduler started successfully with %d cron jobs registered", registeredCount)

	// Log all registered cron jobs with their next execution times
	// Note: We need to get the times after cron.Start() for accurate Next values
	if registeredCount > 0 {
		log.Printf("[Scheduler] Active cron jobs:")
		s.mu.RLock()
		for id, entryID := range s.cronJobs {
			entry := s.cron.Entry(entryID)
			if !entry.Next.IsZero() {
				log.Printf("[Scheduler]   - Mapping ID %d: Next run at %s", id, entry.Next.Format("2006-01-02 15:04:05"))
			} else {
				log.Printf("[Scheduler]   - Mapping ID %d: Next run time not yet calculated", id)
			}
		}
		s.mu.RUnlock()
	}

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

// Context key for trace ID
type contextKey string

const traceIDKey contextKey = "trace_id"

// RunMapping runs a single mapping
func (s *Scheduler) RunMapping(ctx context.Context, mapping config.MappingConfig) error {
	// Get trace ID from context or generate new one
	var taskID string
	if ctxTaskID := ctx.Value(traceIDKey); ctxTaskID != nil {
		taskID = ctxTaskID.(string)
	} else {
		taskID = uuid.New().String()
	}
	traceID := taskID[:8] // Use first 8 chars as short trace ID

	// Create task record
	task := &storage.Task{
		TaskID:     taskID,
		ConfigName: mapping.Name,
		Mode:       mapping.Mode,
		Status:     "running",
		StartedAt:  time.Now(),
	}
	if err := s.db.CreateTask(task); err != nil {
		return fmt.Errorf("[TraceID: %s] failed to create task: %w", traceID, err)
	}

	log.Printf("[TraceID: %s] Task started: mapping=%s, mode=%s, source=%s, target=%s",
		traceID, mapping.Name, mapping.Mode, mapping.Source, mapping.Target)

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
	duration := now.Sub(task.StartedAt)

	if err != nil {
		task.Status = "failed"
		task.Errors = err.Error()
		s.db.UpdateTask(task)
		log.Printf("[TraceID: %s] Task FAILED: error=%v, duration=%v", traceID, err, duration)
		return fmt.Errorf("[TraceID: %s] generation failed: %w", traceID, err)
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
		log.Printf("[TraceID: %s] Task completed with %d errors", traceID, len(result.Errors))
	}

	if err := s.db.UpdateTask(task); err != nil {
		log.Printf("[TraceID: %s] WARNING: Failed to update task record: %v", traceID, err)
	}

	log.Printf("[TraceID: %s] Task COMPLETED: created=%d, deleted=%d, skipped=%d, errors=%d, duration=%v",
		traceID, result.FilesCreated, result.FilesDeleted, result.FilesSkipped, len(result.Errors), duration)

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

	log.Printf("[Scheduler] AddCronJob called: mappingID=%d, name=%s, expr=%s", mappingID, mappingName, cronExpr)

	// Remove existing job if any
	if entryID, exists := s.cronJobs[mappingID]; exists {
		log.Printf("[Scheduler] Removing existing cron job for mapping ID %d", mappingID)
		s.cron.Remove(entryID)
		delete(s.cronJobs, mappingID)
	}

	// Add new cron job
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		startTime := time.Now()
		log.Printf("[Scheduler] ========== Cron job TRIGGERED: mapping=%s (ID: %d) ==========", mappingName, mappingID)

		// RunMappingByName will create its own TraceID and log with it
		if err := s.RunMappingByName(context.Background(), mappingName); err != nil {
			// Extract TraceID from error message if present
			log.Printf("[Scheduler] Scheduled task FAILED for mapping %s: %v", mappingName, err)
		} else {
			duration := time.Since(startTime)
			log.Printf("[Scheduler] Scheduled task COMPLETED for mapping %s, duration: %v", mappingName, duration)
		}
		log.Printf("[Scheduler] ========== Cron job FINISHED: mapping=%s (ID: %d) ==========", mappingName, mappingID)
	})
	if err != nil {
		log.Printf("[Scheduler] ERROR: Failed to create cron job for mapping %s: %v", mappingName, err)
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cronJobs[mappingID] = entryID

	// Calculate next execution time manually
	// Support both 5-field (minute-based) and 6-field (second-based) cron expressions
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		log.Printf("[Scheduler] WARNING: Failed to parse cron expression for next time calculation: %v", err)
		log.Printf("[Scheduler] Cron job REGISTERED successfully: mapping=%s (ID: %d), expr=%s",
			mappingName, mappingID, cronExpr)
	} else {
		nextTime := schedule.Next(time.Now())
		log.Printf("[Scheduler] Cron job REGISTERED successfully: mapping=%s (ID: %d), expr=%s, next_run=%s",
			mappingName, mappingID, cronExpr, nextTime.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// RemoveCronJob removes a cron job for a mapping
func (s *Scheduler) RemoveCronJob(mappingID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[Scheduler] RemoveCronJob called for mapping ID: %d", mappingID)

	if entryID, exists := s.cronJobs[mappingID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronJobs, mappingID)
		log.Printf("[Scheduler] Cron job REMOVED successfully for mapping ID: %d", mappingID)
	} else {
		log.Printf("[Scheduler] No cron job found for mapping ID: %d, skipping removal", mappingID)
	}
}

// UpdateCronJob updates a cron job for a mapping
func (s *Scheduler) UpdateCronJob(mappingID uint, mappingName, cronExpr string, enabled bool) error {
	log.Printf("[Scheduler] UpdateCronJob called: mappingID=%d, name=%s, expr=%s, enabled=%v",
		mappingID, mappingName, cronExpr, enabled)

	// If disabled or no cron expression, remove the job
	if !enabled || cronExpr == "" {
		if !enabled {
			log.Printf("[Scheduler] Mapping %s is disabled, removing cron job", mappingName)
		} else {
			log.Printf("[Scheduler] Mapping %s has empty cron expression, removing cron job", mappingName)
		}
		s.RemoveCronJob(mappingID)
		return nil
	}

	// Otherwise, add/update the job
	log.Printf("[Scheduler] Updating cron job for mapping %s", mappingName)
	return s.AddCronJob(mappingID, mappingName, cronExpr)
}
