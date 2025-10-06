package alist

import "time"

// ListRequest represents a request to list files
type ListRequest struct {
	Path     string `json:"path"`
	Password string `json:"password,omitempty"`
	Page     int    `json:"page,omitempty"`
	PerPage  int    `json:"per_page,omitempty"`
	Refresh  bool   `json:"refresh,omitempty"`
}

// ListResponse represents the response from list API
type ListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Content  []FileItem `json:"content"`
		Total    int        `json:"total"`
		Readme   string     `json:"readme,omitempty"`
		Write    bool       `json:"write,omitempty"`
		Provider string     `json:"provider,omitempty"`
	} `json:"data"`
}

// FileItem represents a file or directory
type FileItem struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	Modified time.Time `json:"modified"`
	Sign     string    `json:"sign,omitempty"`
	Thumb    string    `json:"thumb,omitempty"`
	Type     int       `json:"type,omitempty"`
	Path     string    `json:"-"` // Full path (not from API, set by client)
}

// GetRequest represents a request to get file info
type GetRequest struct {
	Path     string `json:"path"`
	Password string `json:"password,omitempty"`
}

// GetResponse represents the response from get API
type GetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Name     string    `json:"name"`
		Size     int64     `json:"size"`
		IsDir    bool      `json:"is_dir"`
		Modified time.Time `json:"modified"`
		Sign     string    `json:"sign,omitempty"`
		Thumb    string    `json:"thumb,omitempty"`
		Type     int       `json:"type,omitempty"`
		RawURL   string    `json:"raw_url,omitempty"`
		Provider string    `json:"provider,omitempty"`
	} `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// IsVideo checks if the file is a video based on extension
func (f *FileItem) IsVideo(extensions []string) bool {
	if f.IsDir {
		return false
	}

	for _, ext := range extensions {
		if len(f.Name) > len(ext) &&
			f.Name[len(f.Name)-len(ext)-1:] == "."+ext {
			return true
		}
	}
	return false
}
