package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps gorm.DB
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	// Create database directory if not exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&File{}, &Task{}, &User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateFile creates a new file record
func (db *DB) CreateFile(file *File) error {
	return db.DB.Create(file).Error
}

// UpdateFile updates a file record
func (db *DB) UpdateFile(file *File) error {
	return db.DB.Save(file).Error
}

// GetFileByPath gets a file by path
func (db *DB) GetFileByPath(path string) (*File, error) {
	var file File
	err := db.DB.Where("path = ?", path).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// DeleteFileByPath deletes a file by path
func (db *DB) DeleteFileByPath(path string) error {
	return db.DB.Where("path = ?", path).Delete(&File{}).Error
}

// ListFiles lists all files
func (db *DB) ListFiles(limit, offset int) ([]*File, error) {
	var files []*File
	err := db.DB.Limit(limit).Offset(offset).Find(&files).Error
	return files, err
}

// CreateTask creates a new task record
func (db *DB) CreateTask(task *Task) error {
	return db.DB.Create(task).Error
}

// UpdateTask updates a task record
func (db *DB) UpdateTask(task *Task) error {
	return db.DB.Save(task).Error
}

// GetTaskByID gets a task by task ID
func (db *DB) GetTaskByID(taskID string) (*Task, error) {
	var task Task
	err := db.DB.Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks lists tasks with pagination
func (db *DB) ListTasks(limit, offset int) ([]*Task, error) {
	var tasks []*Task
	err := db.DB.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error
	return tasks, err
}

// GetTasksByStatus gets tasks by status
func (db *DB) GetTasksByStatus(status string, limit int) ([]*Task, error) {
	var tasks []*Task
	err := db.DB.Where("status = ?", status).Order("created_at DESC").Limit(limit).Find(&tasks).Error
	return tasks, err
}

// CreateUser creates a new user
func (db *DB) CreateUser(user *User) error {
	return db.DB.Create(user).Error
}

// GetUserByUsername gets a user by username
func (db *DB) GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (db *DB) UpdateUser(user *User) error {
	return db.DB.Save(user).Error
}
