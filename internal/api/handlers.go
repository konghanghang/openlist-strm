package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/konghang/openlist-strm/internal/storage"
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

	ctx := context.Background()
	taskID := uuid.New().String()

	// Run in background
	go func() {
		if req.Path == "" {
			// Run all mappings
			s.scheduler.RunAll(ctx)
		} else {
			// Run specific mapping
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

// handleListTasks handles list tasks
func (s *Server) handleListTasks(c *gin.Context) {
	limit := 20
	offset := 0

	tasks, err := s.db.ListTasks(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list tasks",
		})
		return
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
		"tasks": response,
		"total": len(response),
	})
}

// handleGetConfigs handles get configurations
func (s *Server) handleGetConfigs(c *gin.Context) {
	var configs []ConfigResponse
	for _, m := range s.cfg.Mappings {
		configs = append(configs, ConfigResponse{
			Name:    m.Name,
			Source:  m.Source,
			Target:  m.Target,
			Mode:    m.Mode,
			Enabled: m.Enabled,
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

	// Find matching mapping configuration
	var matchedMapping *string
	for _, mapping := range s.cfg.Mappings {
		if !mapping.Enabled {
			continue
		}
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
	ctx := context.Background()
	taskID := uuid.New().String()

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
