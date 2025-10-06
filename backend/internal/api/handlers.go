package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/konghanghang/openlist-strm/internal/storage"
	"github.com/robfig/cron/v3"
)

// GenerateRequest represents a generate request
type GenerateRequest struct {
	Path string `json:"path"` // Optional, if empty, run all mappings
	Mode string `json:"mode"` // Optional, default to incremental
}

// GenerateResponse represents a generate response
type GenerateResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// TaskResponse represents a task response
type TaskResponse struct {
	TaskID       string     `json:"task_id"`
	ConfigName   string     `json:"config_name"`
	Mode         string     `json:"mode"`
	Status       string     `json:"status"`
	FilesCreated int        `json:"files_created"`
	FilesDeleted int        `json:"files_deleted"`
	FilesSkipped int        `json:"files_skipped"`
	Errors       string     `json:"errors,omitempty"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// ConfigResponse represents a config response
type ConfigResponse struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Target  string `json:"target"`
	Mode    string `json:"mode"`
	Enabled bool   `json:"enabled"`
}

// StatusResponse represents system status response
type StatusResponse struct {
	Version   string    `json:"version"`
	Uptime    int64     `json:"uptime"`
	StartTime time.Time `json:"start_time"`
}

var serverStartTime = time.Now()

// handleHealth handles health check
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "1.0.0",
	})
}

// handleGenerate handles generate STRM files
func (s *Server) handleGenerate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Default mode
	if req.Mode == "" {
		req.Mode = "incremental"
	}

	// Validate mode
	if req.Mode != "incremental" && req.Mode != "full" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "mode must be 'incremental' or 'full'",
		})
		return
	}

	taskID := uuid.New().String()
	traceID := taskID[:8]

	// Create context with trace ID
	ctx := context.WithValue(context.Background(), "trace_id", taskID)

	log.Printf("[TraceID: %s] API request received: path=%s, mode=%s", traceID, req.Path, req.Mode)

	// Run in background
	go func() {
		if req.Path == "" {
			// Run all mappings
			log.Printf("[TraceID: %s] Running all enabled mappings", traceID)
			s.scheduler.RunAll(ctx)
		} else {
			// Run specific mapping
			log.Printf("[TraceID: %s] Running specific mapping: %s", traceID, req.Path)
			s.scheduler.RunMappingByName(ctx, req.Path)
		}
	}()

	c.JSON(http.StatusOK, GenerateResponse{
		TaskID: taskID,
		Status: "running",
	})
}

// handleGetTask handles get task by ID
func (s *Server) handleGetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := s.db.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "task not found",
		})
		return
	}

	c.JSON(http.StatusOK, TaskResponse{
		TaskID:       task.TaskID,
		ConfigName:   task.ConfigName,
		Mode:         task.Mode,
		Status:       task.Status,
		FilesCreated: task.FilesCreated,
		FilesDeleted: task.FilesDeleted,
		FilesSkipped: task.FilesSkipped,
		Errors:       task.Errors,
		StartedAt:    task.StartedAt,
		CompletedAt:  task.CompletedAt,
	})
}

// handleListTasks handles list tasks with pagination
func (s *Server) handleListTasks(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize
	tasks, err := s.db.ListTasks(pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list tasks",
		})
		return
	}

	// Get total count
	total, err := s.db.CountTasks()
	if err != nil {
		total = 0
	}

	var response []TaskResponse
	for _, task := range tasks {
		response = append(response, TaskResponse{
			TaskID:       task.TaskID,
			ConfigName:   task.ConfigName,
			Mode:         task.Mode,
			Status:       task.Status,
			FilesCreated: task.FilesCreated,
			FilesDeleted: task.FilesDeleted,
			FilesSkipped: task.FilesSkipped,
			Errors:       task.Errors,
			StartedAt:    task.StartedAt,
			CompletedAt:  task.CompletedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":     response,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// handleGetConfigs handles get configurations (from database)
func (s *Server) handleGetConfigs(c *gin.Context) {
	mappings, err := s.db.ListMappings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list mappings"})
		return
	}

	var configs []MappingResponse
	for _, m := range mappings {
		configs = append(configs, MappingResponse{
			ID:         m.ID,
			Name:       m.Name,
			Source:     m.Source,
			Target:     m.Target,
			Extensions: strings.Split(m.Extensions, ","),
			Concurrent: m.Concurrent,
			Mode:       m.Mode,
			STRMMode:   m.STRMMode,
			CronExpr:   m.CronExpr,
			Enabled:    m.Enabled,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
	})
}

// handleGetStatus handles get system status
func (s *Server) handleGetStatus(c *gin.Context) {
	uptime := time.Since(serverStartTime).Seconds()

	c.JSON(http.StatusOK, StatusResponse{
		Version:   "1.0.0",
		Uptime:    int64(uptime),
		StartTime: serverStartTime,
	})
}

// WebhookRequest represents a webhook request
type WebhookRequest struct {
	Event  string `json:"event"`  // Event type: file.upload, file.delete, etc.
	Path   string `json:"path"`   // File or directory path
	Action string `json:"action"` // Action: add, update, delete
}

// WebhookResponse represents a webhook response
type WebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitempty"`
}

// handleWebhook handles webhook notifications from external systems
func (s *Server) handleWebhook(c *gin.Context) {
	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	// Validate event and path
	if req.Path == "" {
		c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Message: "path is required",
		})
		return
	}

	// Find matching mapping configuration from database
	mappings, err := s.db.ListEnabledMappings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, WebhookResponse{
			Success: false,
			Message: "failed to list mappings",
		})
		return
	}

	var matchedMapping *string
	for _, mapping := range mappings {
		// Check if the webhook path matches the mapping source
		if len(req.Path) >= len(mapping.Source) && req.Path[:len(mapping.Source)] == mapping.Source {
			matchedMapping = &mapping.Name
			break
		}
	}

	if matchedMapping == nil {
		c.JSON(http.StatusOK, WebhookResponse{
			Success: true,
			Message: "no matching mapping found, skipping",
		})
		return
	}

	// Trigger generation in background
	taskID := uuid.New().String()
	traceID := taskID[:8]
	ctx := context.WithValue(context.Background(), "trace_id", taskID)

	log.Printf("[TraceID: %s] Webhook received: event=%s, path=%s, action=%s, matched_mapping=%s",
		traceID, req.Event, req.Path, req.Action, *matchedMapping)

	go func() {
		s.scheduler.RunMappingByName(ctx, *matchedMapping)
	}()

	c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "webhook received, generation triggered",
		TaskID:  taskID,
	})
}

// toTaskResponse converts storage.Task to TaskResponse
func toTaskResponse(task *storage.Task) TaskResponse {
	return TaskResponse{
		TaskID:       task.TaskID,
		ConfigName:   task.ConfigName,
		Mode:         task.Mode,
		Status:       task.Status,
		FilesCreated: task.FilesCreated,
		FilesDeleted: task.FilesDeleted,
		FilesSkipped: task.FilesSkipped,
		Errors:       task.Errors,
		StartedAt:    task.StartedAt,
		CompletedAt:  task.CompletedAt,
	}
}

// MappingRequest represents a mapping create/update request
type MappingRequest struct {
	Name       string   `json:"name" binding:"required"`
	Source     string   `json:"source" binding:"required"`
	Target     string   `json:"target" binding:"required"`
	Extensions []string `json:"extensions" binding:"required"`
	Concurrent int      `json:"concurrent"`
	Mode       string   `json:"mode"`
	STRMMode   string   `json:"strm_mode"`
	CronExpr   string   `json:"cron_expr"`
	Enabled    *bool    `json:"enabled"`
}

// MappingResponse represents a mapping response
type MappingResponse struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	Source     string   `json:"source"`
	Target     string   `json:"target"`
	Extensions []string `json:"extensions"`
	Concurrent int      `json:"concurrent"`
	Mode       string   `json:"mode"`
	STRMMode   string   `json:"strm_mode"`
	CronExpr   string   `json:"cron_expr"`
	Enabled    bool     `json:"enabled"`
}

// handleCreateMapping handles creating a new mapping
func (s *Server) handleCreateMapping(c *gin.Context) {
	var req MappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Mode == "" {
		req.Mode = "incremental"
	}
	if req.STRMMode == "" {
		req.STRMMode = "alist_path"
	}
	if req.Concurrent <= 0 {
		req.Concurrent = 10
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	// Validate mode
	if req.Mode != "incremental" && req.Mode != "full" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'incremental' or 'full'"})
		return
	}
	if req.STRMMode != "alist_path" && req.STRMMode != "http_url" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "strm_mode must be 'alist_path' or 'http_url'"})
		return
	}

	// Validate cron expression if provided
	// Support both 5-field (minute-based) and 6-field (second-based) cron expressions
	if req.CronExpr != "" {
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(req.CronExpr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid cron expression: %v", err)})
			return
		}
	}

	mapping := &storage.Mapping{
		Name:       req.Name,
		Source:     req.Source,
		Target:     req.Target,
		Extensions: strings.Join(req.Extensions, ","),
		Concurrent: req.Concurrent,
		Mode:       req.Mode,
		STRMMode:   req.STRMMode,
		CronExpr:   req.CronExpr,
		Enabled:    enabled,
	}

	if err := s.db.CreateMapping(mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create mapping"})
		return
	}

	// Add cron job if enabled and has cron expression
	if mapping.Enabled && mapping.CronExpr != "" {
		if err := s.scheduler.AddCronJob(mapping.ID, mapping.Name, mapping.CronExpr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("mapping created but failed to add cron job: %v", err)})
			return
		}
	}

	c.JSON(http.StatusCreated, MappingResponse{
		ID:         mapping.ID,
		Name:       mapping.Name,
		Source:     mapping.Source,
		Target:     mapping.Target,
		Extensions: req.Extensions,
		Concurrent: mapping.Concurrent,
		Mode:       mapping.Mode,
		STRMMode:   mapping.STRMMode,
		CronExpr:   mapping.CronExpr,
		Enabled:    mapping.Enabled,
	})
}

// handleUpdateMapping handles updating a mapping
func (s *Server) handleUpdateMapping(c *gin.Context) {
	id := c.Param("id")
	var mappingID uint
	if _, err := fmt.Sscanf(id, "%d", &mappingID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	var req MappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing mapping
	existing, err := s.db.GetMappingByID(mappingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	// Update fields
	existing.Name = req.Name
	existing.Source = req.Source
	existing.Target = req.Target
	existing.Extensions = strings.Join(req.Extensions, ",")
	if req.Concurrent > 0 {
		existing.Concurrent = req.Concurrent
	}
	if req.Mode != "" {
		if req.Mode != "incremental" && req.Mode != "full" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'incremental' or 'full'"})
			return
		}
		existing.Mode = req.Mode
	}
	if req.STRMMode != "" {
		if req.STRMMode != "alist_path" && req.STRMMode != "http_url" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "strm_mode must be 'alist_path' or 'http_url'"})
			return
		}
		existing.STRMMode = req.STRMMode
	}

	// Validate and update cron expression
	// Support both 5-field (minute-based) and 6-field (second-based) cron expressions
	if req.CronExpr != existing.CronExpr {
		if req.CronExpr != "" {
			parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			if _, err := parser.Parse(req.CronExpr); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid cron expression: %v", err)})
				return
			}
		}
		existing.CronExpr = req.CronExpr
	}

	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := s.db.UpdateMapping(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update mapping"})
		return
	}

	// Update cron job
	if err := s.scheduler.UpdateCronJob(existing.ID, existing.Name, existing.CronExpr, existing.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("mapping updated but failed to update cron job: %v", err)})
		return
	}

	c.JSON(http.StatusOK, MappingResponse{
		ID:         existing.ID,
		Name:       existing.Name,
		Source:     existing.Source,
		Target:     existing.Target,
		Extensions: strings.Split(existing.Extensions, ","),
		Concurrent: existing.Concurrent,
		Mode:       existing.Mode,
		STRMMode:   existing.STRMMode,
		CronExpr:   existing.CronExpr,
		Enabled:    existing.Enabled,
	})
}

// handleDeleteMapping handles deleting a mapping
func (s *Server) handleDeleteMapping(c *gin.Context) {
	id := c.Param("id")
	var mappingID uint
	if _, err := fmt.Sscanf(id, "%d", &mappingID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mapping id"})
		return
	}

	// Remove cron job first
	s.scheduler.RemoveCronJob(mappingID)

	if err := s.db.DeleteMapping(mappingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mapping deleted successfully"})
}
