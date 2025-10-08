package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/konghanghang/openlist-strm/internal/config"
)

// Notifier 媒体服务器通知接口
type Notifier interface {
	NotifyLibraryScan(ctx context.Context, strmPath string) error
}

// MediaServerNotifier 媒体服务器通知服务
type MediaServerNotifier struct {
	config *config.MediaServerConfig
	client *http.Client
}

// NewMediaServerNotifier 创建媒体服务器通知服务
func NewMediaServerNotifier(cfg *config.MediaServerConfig) *MediaServerNotifier {
	return &MediaServerNotifier{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NotifyLibraryScan 通知媒体服务器扫描库
func (n *MediaServerNotifier) NotifyLibraryScan(ctx context.Context, strmPath string) error {
	if !n.config.Enabled {
		log.Printf("[MediaServer] Notification disabled, skipping")
		return nil
	}

	var errs []error

	// 根据配置的类型通知对应的服务器
	switch n.config.Type {
	case "emby":
		if err := n.notifyEmby(ctx, strmPath); err != nil {
			errs = append(errs, fmt.Errorf("emby notification failed: %w", err))
		}
	case "jellyfin":
		if err := n.notifyJellyfin(ctx, strmPath); err != nil {
			errs = append(errs, fmt.Errorf("jellyfin notification failed: %w", err))
		}
	case "both":
		if err := n.notifyEmby(ctx, strmPath); err != nil {
			errs = append(errs, fmt.Errorf("emby notification failed: %w", err))
		}
		if err := n.notifyJellyfin(ctx, strmPath); err != nil {
			errs = append(errs, fmt.Errorf("jellyfin notification failed: %w", err))
		}
	default:
		return fmt.Errorf("unknown media server type: %s", n.config.Type)
	}

	// 如果有错误，返回第一个错误（但不影响其他服务器的通知）
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("[MediaServer] Error: %v", err)
		}
		return errs[0]
	}

	return nil
}

// notifyEmby 通知 Emby 扫描
func (n *MediaServerNotifier) notifyEmby(ctx context.Context, strmPath string) error {
	if n.config.Emby.URL == "" || n.config.Emby.APIKey == "" {
		log.Printf("[Emby] Skipping: URL or API Key not configured")
		return nil
	}

	var url string
	var requestBody []byte

	if n.config.Emby.ScanMode == "path" && len(n.config.Emby.PathMapping) > 0 {
		// 路径映射模式：扫描特定路径
		mappedPath := n.mapPath(strmPath, n.config.Emby.PathMapping)
		url = fmt.Sprintf("%s/Library/Media/Updated?api_key=%s", n.config.Emby.URL, n.config.Emby.APIKey)

		// 构建请求体
		updates := map[string]interface{}{
			"Updates": []map[string]string{
				{
					"Path":       mappedPath,
					"UpdateType": "Created",
				},
			},
		}
		var err error
		requestBody, err = json.Marshal(updates)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}

		log.Printf("[Emby] Notifying path scan: %s -> %s", strmPath, mappedPath)
	} else {
		// 全局扫描模式
		url = fmt.Sprintf("%s/Library/Refresh?api_key=%s", n.config.Emby.URL, n.config.Emby.APIKey)
		log.Printf("[Emby] Notifying full library scan")
	}

	// 发送 POST 请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if len(requestBody) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[Emby] WARNING: Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("emby returned status code: %d", resp.StatusCode)
	}

	log.Printf("[Emby] ✅ Notification sent successfully")
	return nil
}

// notifyJellyfin 通知 Jellyfin 扫描
func (n *MediaServerNotifier) notifyJellyfin(ctx context.Context, strmPath string) error {
	if n.config.Jellyfin.URL == "" || n.config.Jellyfin.APIKey == "" {
		log.Printf("[Jellyfin] Skipping: URL or API Key not configured")
		return nil
	}

	var url string
	var requestBody []byte

	if n.config.Jellyfin.ScanMode == "path" && len(n.config.Jellyfin.PathMapping) > 0 {
		// 路径映射模式：扫描特定路径
		mappedPath := n.mapPath(strmPath, n.config.Jellyfin.PathMapping)
		url = fmt.Sprintf("%s/Library/Media/Updated?api_key=%s", n.config.Jellyfin.URL, n.config.Jellyfin.APIKey)

		// 构建请求体
		updates := map[string]interface{}{
			"Updates": []map[string]string{
				{
					"Path":       mappedPath,
					"UpdateType": "Created",
				},
			},
		}
		var err error
		requestBody, err = json.Marshal(updates)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}

		log.Printf("[Jellyfin] Notifying path scan: %s -> %s", strmPath, mappedPath)
	} else {
		// 全局扫描模式
		url = fmt.Sprintf("%s/Library/Refresh?api_key=%s", n.config.Jellyfin.URL, n.config.Jellyfin.APIKey)
		log.Printf("[Jellyfin] Notifying full library scan")
	}

	// 发送 POST 请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if len(requestBody) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[Jellyfin] WARNING: Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jellyfin returned status code: %d", resp.StatusCode)
	}

	log.Printf("[Jellyfin] ✅ Notification sent successfully")
	return nil
}

// mapPath 将 OpenList-STRM 路径映射为媒体服务器路径
func (n *MediaServerNotifier) mapPath(strmPath string, mapping map[string]string) string {
	// 清理路径
	strmPath = filepath.Clean(strmPath)

	// 遍历所有映射规则，找到最长匹配
	var bestMatch string
	var mappedPath string

	for srcPath, dstPath := range mapping {
		srcPath = filepath.Clean(srcPath)
		// 检查是否匹配（精确匹配或前缀匹配）
		if len(srcPath) > len(bestMatch) {
			if strmPath == srcPath {
				bestMatch = srcPath
				mappedPath = dstPath
			} else if strings.HasPrefix(strmPath, srcPath+string(filepath.Separator)) {
				bestMatch = srcPath
				// 替换路径前缀
				relPath := strmPath[len(srcPath):]
				mappedPath = filepath.Join(dstPath, relPath)
			}
		}
	}

	// 如果找到映射，返回映射后的路径；否则返回原路径
	if mappedPath != "" {
		return mappedPath
	}
	return strmPath
}

// filepath.HasPrefix 检查路径是否有指定前缀（跨平台）
func HasPrefix(path, prefix string) bool {
	path = filepath.Clean(path)
	prefix = filepath.Clean(prefix)

	// 简单的前缀匹配
	if len(path) < len(prefix) {
		return false
	}

	return path[:len(prefix)] == prefix
}
