package storage

import (
	"time"
)

// File represents a file record in database
type File struct {
	ID         uint      `gorm:"primarykey"`
	Path       string    `gorm:"uniqueIndex;not null"`
	Size       int64     `gorm:"not null"`
	ModifiedAt time.Time `gorm:"not null"`
	Hash       string    `gorm:"index"`
	STRMPath   string    `gorm:"index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Task represents a task execution record
type Task struct {
	ID           uint   `gorm:"primarykey"`
	TaskID       string `gorm:"uniqueIndex;not null"`
	ConfigName   string `gorm:"index"`
	Mode         string // incremental or full
	Status       string `gorm:"index"` // running, completed, failed
	FilesCreated int
	FilesDeleted int
	FilesSkipped int
	Errors       string `gorm:"type:text"`
	StartedAt    time.Time
	CompletedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Mapping represents a path mapping configuration
type Mapping struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"uniqueIndex;not null"` // 配置名称
	Source    string `gorm:"not null"`             // Alist 源路径
	Target    string `gorm:"not null"`             // STRM 目标路径
	Mode      string `gorm:"default:incremental"`  // incremental or full
	Enabled   bool   `gorm:"default:true"`         // 是否启用
	CreatedAt time.Time
	UpdatedAt time.Time
}

// User represents a user account
type User struct {
	ID           uint   `gorm:"primarykey"`
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"default:user"` // admin or user
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies custom table names
func (File) TableName() string {
	return "files"
}

func (Task) TableName() string {
	return "tasks"
}

func (Mapping) TableName() string {
	return "mappings"
}

func (User) TableName() string {
	return "users"
}
