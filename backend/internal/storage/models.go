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

func (User) TableName() string {
	return "users"
}
