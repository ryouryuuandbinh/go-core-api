// Package models chứa các cấu trúc dữ liệu thực thể của hệ thống
package models

import (
	"time"

	"gorm.io/gorm"
)

// Định nghĩa các hằng số cho Role để tránh gõ sai chính tả (Hardcode string) ở nhiều nơi
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User đại diện cho bảng 'users' trong database
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"index:idx_email_unique,unique,where:deleted_at IS NULL;not null" json:"email"`
	Password     string         `gorm:"not null" json:"-"` // Dấu - giúp ẩn field này khi trả về JSON
	FullName     string         `json:"full_name"`
	Avatar       string         `json:"avatar"`
	Phone        string         `json:"phone"`
	Role         string         `gorm:"default:'user'" json:"role"`
	TokenVersion int            `gorm:"default:1" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete
}
