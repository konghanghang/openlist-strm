package alist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:5244", "test-token", false, 30)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.baseURL != "http://localhost:5244" {
		t.Errorf("baseURL = %v, want http://localhost:5244", client.baseURL)
	}
	if client.token != "test-token" {
		t.Errorf("token = %v, want test-token", client.token)
	}
	if client.signEnable != false {
		t.Errorf("signEnable = %v, want false", client.signEnable)
	}
}

func TestNewClient_TrimTrailingSlash(t *testing.T) {
	client := NewClient("http://localhost:5244/", "test-token", false, 30)
	if client.baseURL != "http://localhost:5244" {
		t.Errorf("baseURL = %v, want http://localhost:5244 (trailing slash should be removed)", client.baseURL)
	}
}

func TestPing_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ping" {
			t.Errorf("Ping request path = %v, want /ping", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	err := client.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
}

func TestPing_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	err := client.Ping(context.Background())
	if err == nil {
		t.Error("Ping() expected error for server error, got nil")
	}
}

func TestListFiles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/fs/list" {
			t.Errorf("Request path = %v, want /api/fs/list", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Request method = %v, want POST", r.Method)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "test-token" {
			t.Errorf("Authorization header = %v, want test-token", auth)
		}

		// Return mock response
		resp := ListResponse{
			Code:    200,
			Message: "success",
			Data: &struct {
				Content  []FileItem `json:"content"`
				Total    int        `json:"total"`
				Readme   string     `json:"readme,omitempty"`
				Write    bool       `json:"write,omitempty"`
				Provider string     `json:"provider,omitempty"`
			}{
				Content: []FileItem{
					{Name: "movie1.mp4", Size: 1024, IsDir: false},
					{Name: "movie2.mkv", Size: 2048, IsDir: false},
					{Name: "subdir", Size: 0, IsDir: true},
				},
				Total: 3,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	files, err := client.ListFiles(context.Background(), "/movies")
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 3 {
		t.Errorf("len(files) = %v, want 3", len(files))
	}
	if files[0].Name != "movie1.mp4" {
		t.Errorf("files[0].Name = %v, want movie1.mp4", files[0].Name)
	}
	if files[2].IsDir != true {
		t.Errorf("files[2].IsDir = %v, want true", files[2].IsDir)
	}
}

func TestListFiles_EmptyDirectory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ListResponse{
			Code:    200,
			Message: "success",
			Data:    nil, // Empty directory
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	files, err := client.ListFiles(context.Background(), "/empty")
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("len(files) = %v, want 0", len(files))
	}
}

func TestListFiles_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ListResponse{
			Code:    500,
			Message: "internal server error",
			Data:    nil,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	_, err := client.ListFiles(context.Background(), "/movies")
	if err == nil {
		t.Error("ListFiles() expected error for API error, got nil")
	}
}

func TestListFilesRecursive_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req ListRequest
		json.NewDecoder(r.Body).Decode(&req)

		var resp ListResponse
		if req.Path == "/movies" {
			// Root directory
			resp = ListResponse{
				Code:    200,
				Message: "success",
				Data: &struct {
					Content  []FileItem `json:"content"`
					Total    int        `json:"total"`
					Readme   string     `json:"readme,omitempty"`
					Write    bool       `json:"write,omitempty"`
					Provider string     `json:"provider,omitempty"`
				}{
					Content: []FileItem{
						{Name: "movie1.mp4", Size: 1024, IsDir: false},
						{Name: "action", Size: 0, IsDir: true},
					},
					Total: 2,
				},
			}
		} else if req.Path == "/movies/action" {
			// Subdirectory
			resp = ListResponse{
				Code:    200,
				Message: "success",
				Data: &struct {
					Content  []FileItem `json:"content"`
					Total    int        `json:"total"`
					Readme   string     `json:"readme,omitempty"`
					Write    bool       `json:"write,omitempty"`
					Provider string     `json:"provider,omitempty"`
				}{
					Content: []FileItem{
						{Name: "movie2.mp4", Size: 2048, IsDir: false},
						{Name: "readme.txt", Size: 100, IsDir: false},
					},
					Total: 2,
				},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	files, err := client.ListFilesRecursive(context.Background(), "/movies", []string{"mp4", "mkv"})
	if err != nil {
		t.Fatalf("ListFilesRecursive() error = %v", err)
	}

	// Should find 2 video files (movie1.mp4 and movie2.mp4)
	// readme.txt should be filtered out
	if len(files) != 2 {
		t.Errorf("len(files) = %v, want 2", len(files))
	}

	if callCount != 2 {
		t.Errorf("API call count = %v, want 2 (root + subdirectory)", callCount)
	}
}

func TestListFilesRecursive_FilterByExtension(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ListResponse{
			Code:    200,
			Message: "success",
			Data: &struct {
				Content  []FileItem `json:"content"`
				Total    int        `json:"total"`
				Readme   string     `json:"readme,omitempty"`
				Write    bool       `json:"write,omitempty"`
				Provider string     `json:"provider,omitempty"`
			}{
				Content: []FileItem{
					{Name: "movie.mp4", Size: 1024, IsDir: false},
					{Name: "video.mkv", Size: 2048, IsDir: false},
					{Name: "image.jpg", Size: 500, IsDir: false},
					{Name: "readme.txt", Size: 100, IsDir: false},
				},
				Total: 4,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	files, err := client.ListFilesRecursive(context.Background(), "/media", []string{"mp4", "mkv"})
	if err != nil {
		t.Fatalf("ListFilesRecursive() error = %v", err)
	}

	// Should only find mp4 and mkv files
	if len(files) != 2 {
		t.Errorf("len(files) = %v, want 2", len(files))
	}
}

func TestGetFileURL_WithRawURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GetResponse{
			Code:    200,
			Message: "success",
			Data: &struct {
				Name     string    `json:"name"`
				Size     int64     `json:"size"`
				IsDir    bool      `json:"is_dir"`
				Modified time.Time `json:"modified"`
				Sign     string    `json:"sign,omitempty"`
				Thumb    string    `json:"thumb,omitempty"`
				Type     int       `json:"type,omitempty"`
				RawURL   string    `json:"raw_url,omitempty"`
				Provider string    `json:"provider,omitempty"`
			}{
				Name:   "movie.mp4",
				Size:   1024,
				IsDir:  false,
				RawURL: "http://cdn.example.com/movies/movie.mp4",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	url, err := client.GetFileURL(context.Background(), "/movies/movie.mp4")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}

	if url != "http://cdn.example.com/movies/movie.mp4" {
		t.Errorf("url = %v, want http://cdn.example.com/movies/movie.mp4", url)
	}
}

func TestGetFileURL_WithoutRawURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GetResponse{
			Code:    200,
			Message: "success",
			Data: &struct {
				Name     string    `json:"name"`
				Size     int64     `json:"size"`
				IsDir    bool      `json:"is_dir"`
				Modified time.Time `json:"modified"`
				Sign     string    `json:"sign,omitempty"`
				Thumb    string    `json:"thumb,omitempty"`
				Type     int       `json:"type,omitempty"`
				RawURL   string    `json:"raw_url,omitempty"`
				Provider string    `json:"provider,omitempty"`
			}{
				Name:   "movie.mp4",
				Size:   1024,
				IsDir:  false,
				RawURL: "", // No raw URL
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	url, err := client.GetFileURL(context.Background(), "/movies/movie.mp4")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}

	expectedURL := server.URL + "/d/movies/movie.mp4"
	if url != expectedURL {
		t.Errorf("url = %v, want %v", url, expectedURL)
	}
}

func TestGetFileURL_WithSign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GetResponse{
			Code:    200,
			Message: "success",
			Data: &struct {
				Name     string    `json:"name"`
				Size     int64     `json:"size"`
				IsDir    bool      `json:"is_dir"`
				Modified time.Time `json:"modified"`
				Sign     string    `json:"sign,omitempty"`
				Thumb    string    `json:"thumb,omitempty"`
				Type     int       `json:"type,omitempty"`
				RawURL   string    `json:"raw_url,omitempty"`
				Provider string    `json:"provider,omitempty"`
			}{
				Name:   "movie.mp4",
				Size:   1024,
				IsDir:  false,
				RawURL: "",
				Sign:   "abc123",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", true, 30)
	url, err := client.GetFileURL(context.Background(), "/movies/movie.mp4")
	if err != nil {
		t.Fatalf("GetFileURL() error = %v", err)
	}

	expectedURL := server.URL + "/d/movies/movie.mp4?sign=abc123"
	if url != expectedURL {
		t.Errorf("url = %v, want %v", url, expectedURL)
	}
}

func TestGetFileURL_FileNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GetResponse{
			Code:    200,
			Message: "success",
			Data:    nil, // File not found
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	_, err := client.GetFileURL(context.Background(), "/movies/notfound.mp4")
	if err == nil {
		t.Error("GetFileURL() expected error for file not found, got nil")
	}
}

func TestGetFileURL_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GetResponse{
			Code:    500,
			Message: "internal server error",
			Data:    nil,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	_, err := client.GetFileURL(context.Background(), "/movies/movie.mp4")
	if err == nil {
		t.Error("GetFileURL() expected error for API error, got nil")
	}
}

func TestFileItem_IsVideo(t *testing.T) {
	tests := []struct {
		name       string
		fileItem   FileItem
		extensions []string
		want       bool
	}{
		{
			name:       "mp4 video",
			fileItem:   FileItem{Name: "movie.mp4", IsDir: false},
			extensions: []string{"mp4", "mkv"},
			want:       true,
		},
		{
			name:       "mkv video",
			fileItem:   FileItem{Name: "video.mkv", IsDir: false},
			extensions: []string{"mp4", "mkv"},
			want:       true,
		},
		{
			name:       "non-video file",
			fileItem:   FileItem{Name: "readme.txt", IsDir: false},
			extensions: []string{"mp4", "mkv"},
			want:       false,
		},
		{
			name:       "directory",
			fileItem:   FileItem{Name: "folder", IsDir: true},
			extensions: []string{"mp4", "mkv"},
			want:       false,
		},
		{
			name:       "case sensitive",
			fileItem:   FileItem{Name: "movie.MP4", IsDir: false},
			extensions: []string{"mp4"},
			want:       false, // Current implementation is case-sensitive
		},
		{
			name:       "no extension",
			fileItem:   FileItem{Name: "movie", IsDir: false},
			extensions: []string{"mp4"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fileItem.IsVideo(tt.extensions)
			if got != tt.want {
				t.Errorf("IsVideo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_Cancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.Ping(ctx)
	if err == nil {
		t.Error("Ping() expected error for cancelled context, got nil")
	}
}

func TestHTTP_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := ErrorResponse{
			Code:    401,
			Message: "unauthorized",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, 30)
	_, err := client.ListFiles(context.Background(), "/movies")
	if err == nil {
		t.Error("ListFiles() expected error for HTTP error, got nil")
	}
}
