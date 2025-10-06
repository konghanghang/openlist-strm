package alist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

// Client represents an Alist API client
type Client struct {
	baseURL    string
	token      string
	signEnable bool
	timeout    time.Duration
	httpClient *http.Client
}

// NewClient creates a new Alist client
func NewClient(baseURL, token string, signEnable bool, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		token:      token,
		signEnable: signEnable,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout * time.Second,
		},
	}
}

// ListFiles lists files in the specified path
func (c *Client) ListFiles(ctx context.Context, dirPath string) ([]FileItem, error) {
	req := ListRequest{
		Path:    dirPath,
		Refresh: false,
	}

	var resp ListResponse
	if err := c.doRequest(ctx, "POST", "/api/fs/list", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("alist API error: %s (code: %d)", resp.Message, resp.Code)
	}

	if resp.Data == nil {
		return []FileItem{}, nil
	}

	return resp.Data.Content, nil
}

// ListFilesRecursive lists all files recursively
func (c *Client) ListFilesRecursive(ctx context.Context, dirPath string, extensions []string) ([]FileItem, error) {
	var result []FileItem

	if err := c.listFilesRecursiveHelper(ctx, dirPath, extensions, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// listFilesRecursiveHelper is a helper function for recursive listing
func (c *Client) listFilesRecursiveHelper(ctx context.Context, dirPath string, extensions []string, result *[]FileItem) error {
	files, err := c.ListFiles(ctx, dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir {
			// Recursively list subdirectory
			subPath := path.Join(dirPath, file.Name)
			if err := c.listFilesRecursiveHelper(ctx, subPath, extensions, result); err != nil {
				return err
			}
		} else if file.IsVideo(extensions) {
			// Set full path for the file
			file.Path = path.Join(dirPath, file.Name)
			// Add video file to result
			*result = append(*result, file)
		}
	}

	return nil
}

// GetFileURL gets the direct URL of a file
func (c *Client) GetFileURL(ctx context.Context, filePath string) (string, error) {
	req := GetRequest{
		Path: filePath,
	}

	var resp GetResponse
	if err := c.doRequest(ctx, "POST", "/api/fs/get", req, &resp); err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	if resp.Code != 200 {
		return "", fmt.Errorf("alist API error: %s (code: %d)", resp.Message, resp.Code)
	}

	if resp.Data == nil {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	// Build direct URL
	if resp.Data.RawURL != "" {
		return resp.Data.RawURL, nil
	}

	// Fallback: construct URL from base URL and path
	fileURL := fmt.Sprintf("%s/d%s", c.baseURL, filePath)
	if c.signEnable && resp.Data.Sign != "" {
		fileURL += "?sign=" + resp.Data.Sign
	}

	return fileURL, nil
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, endpoint string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respData, &errResp); err == nil {
			return fmt.Errorf("HTTP %d: %s (code: %d)", resp.StatusCode, errResp.Message, errResp.Code)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respData))
	}

	// Unmarshal response
	if respBody != nil {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Ping checks if the Alist server is accessible
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/ping", nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping Alist server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Alist server returned status %d", resp.StatusCode)
	}

	return nil
}
