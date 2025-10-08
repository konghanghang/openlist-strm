package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/konghanghang/openlist-strm/internal/contextkeys"
	"github.com/konghanghang/openlist-strm/internal/storage"
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
	ctx := context.WithValue(context.Background(), contextkeys.TraceIDKey, taskID)

	log.Printf("[TraceID: %s] API request received: path=%s, mode=%s", traceID, req.Path, req.Mode)

	// Run in background
	go func() {
		if req.Path == "" {
			// Run all mappings
			log.Printf("[TraceID: %s] Running all enabled mappings", traceID)
			_ = s.scheduler.RunAll(ctx) // Error already logged by RunAll
		} else {
			// Run specific mapping
			log.Printf("[TraceID: %s] Running specific mapping: %s", traceID, req.Path)
			_ = s.scheduler.RunMappingByName(ctx, req.Path) // Error already logged
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
	Path       string `json:"path" binding:"required"` // 网盘原始路径（文件或目录）
	Event      string `json:"event"`                   // 事件类型（可选）
	ConfigName string `json:"config_name"`             // 指定配置名称（可选，优先使用）
	Mode       string `json:"mode"`                    // 执行模式：incremental/full（可选，覆盖配置）
	Source     string `json:"source"`                  // 来源标识（可选，用于日志）

	// 路径映射：网盘路径 -> Alist路径
	DrivePath string `json:"drive_path"` // 网盘路径前缀（可选）
	AlistPath string `json:"alist_path"` // Alist路径前缀（可选）
}

// WebhookResponse represents a webhook response
type WebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Skipped bool   `json:"skipped,omitempty"` // 是否跳过（未匹配到配置）
	TaskID  string `json:"task_id,omitempty"`
}

// convertPath 将网盘路径转换为 Alist 路径
func convertPath(originalPath, drivePath, alistPath string) (string, bool) {
	// 如果没有配置映射规则，直接返回原路径
	if drivePath == "" || alistPath == "" {
		return originalPath, false
	}

	// 清理路径
	originalPath = filepath.Clean(originalPath)
	drivePath = filepath.Clean(drivePath)
	alistPath = filepath.Clean(alistPath)

	// 检查是否匹配前缀
	if !strings.HasPrefix(originalPath, drivePath) {
		// 前缀不匹配，返回原路径
		return originalPath, false
	}

	// 去掉原前缀，得到相对路径
	relPath := strings.TrimPrefix(originalPath, drivePath)
	relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

	// 拼接新前缀
	if relPath == "" {
		return alistPath, true
	}

	return filepath.Join(alistPath, relPath), true
}

// matchPath 检查文件路径是否匹配源路径（支持目录和文件）
func matchPath(filePath, sourcePath string) bool {
	// 清理路径
	filePath = filepath.Clean(filePath)
	sourcePath = filepath.Clean(sourcePath)

	// 精确匹配
	if filePath == sourcePath {
		return true
	}

	// 前缀匹配 + 确保路径边界
	if strings.HasPrefix(filePath, sourcePath) {
		rest := filePath[len(sourcePath):]
		// 必须是路径分隔符开头，避免 /media/movies-hd 匹配 /media/movies
		return len(rest) > 0 && rest[0] == filepath.Separator
	}

	return false
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

	// 生成 TraceID
	taskID := uuid.New().String()
	traceID := taskID[:8]

	// 记录原始请求
	log.Printf("[TraceID: %s] Webhook received: path=%s, event=%s, source=%s, config_name=%s, mode=%s",
		traceID, req.Path, req.Event, req.Source, req.ConfigName, req.Mode)

	// 应用路径转换
	convertedPath, converted := convertPath(req.Path, req.DrivePath, req.AlistPath)
	if converted {
		log.Printf("[TraceID: %s] Path converted: %s -> %s (drive_path=%s, alist_path=%s)",
			traceID, req.Path, convertedPath, req.DrivePath, req.AlistPath)
	} else if req.DrivePath != "" && req.AlistPath != "" {
		log.Printf("[TraceID: %s] WARNING: Path conversion failed (prefix mismatch), using original path: %s",
			traceID, req.Path)
	}

	// 查找匹配的配置
	var matchedMappingName string
	var matchedMappingMode string

	// 优先使用指定的配置名称
	if req.ConfigName != "" {
		mapping, err := s.db.GetMappingByName(req.ConfigName)
		if err != nil {
			log.Printf("[TraceID: %s] ERROR: Config not found: %s", traceID, req.ConfigName)
			c.JSON(http.StatusBadRequest, WebhookResponse{
				Success: false,
				Message: fmt.Sprintf("config not found: %s", req.ConfigName),
			})
			return
		}
		if !mapping.Enabled {
			log.Printf("[TraceID: %s] WARNING: Config is disabled: %s", traceID, req.ConfigName)
			c.JSON(http.StatusOK, WebhookResponse{
				Success: true,
				Skipped: true,
				Message: fmt.Sprintf("config is disabled: %s", req.ConfigName),
			})
			return
		}
		matchedMappingName = mapping.Name
		matchedMappingMode = mapping.Mode
		log.Printf("[TraceID: %s] Using specified config: %s", traceID, matchedMappingName)
	} else {
		// 通过路径匹配查找配置
		mappings, err := s.db.ListEnabledMappings()
		if err != nil {
			log.Printf("[TraceID: %s] ERROR: Failed to list mappings: %v", traceID, err)
			c.JSON(http.StatusInternalServerError, WebhookResponse{
				Success: false,
				Message: "failed to list mappings",
			})
			return
		}

		for _, mapping := range mappings {
			if matchPath(convertedPath, mapping.Source) {
				matchedMappingName = mapping.Name
				matchedMappingMode = mapping.Mode
				log.Printf("[TraceID: %s] Matched config by path: %s (source=%s)",
					traceID, mapping.Name, mapping.Source)
				break
			}
		}

		if matchedMappingName == "" {
			log.Printf("[TraceID: %s] No matching config found for path: %s", traceID, convertedPath)
			c.JSON(http.StatusOK, WebhookResponse{
				Success: true,
				Skipped: true,
				Message: "no matching mapping found",
			})
			return
		}
	}

	// 确定执行模式（Webhook 的 mode 参数优先）
	execMode := matchedMappingMode
	if req.Mode != "" {
		if req.Mode == "incremental" || req.Mode == "full" {
			execMode = req.Mode
			log.Printf("[TraceID: %s] Mode overridden by webhook: %s -> %s",
				traceID, matchedMappingMode, execMode)
		} else {
			log.Printf("[TraceID: %s] WARNING: Invalid mode in webhook: %s, using config mode: %s",
				traceID, req.Mode, matchedMappingMode)
		}
	}

	// 创建 context
	ctx := context.WithValue(context.Background(), contextkeys.TraceIDKey, taskID)

	log.Printf("[TraceID: %s] Triggering generation: config=%s, mode=%s",
		traceID, matchedMappingName, execMode)

	// 后台执行任务
	go func() {
		_ = s.scheduler.RunMappingByName(ctx, matchedMappingName) // Error already logged
	}()

	c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "webhook received, generation triggered",
		TaskID:  taskID,
	})
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
