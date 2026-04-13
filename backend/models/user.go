package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;size:50" json:"username"`
	Password string `gorm:"not null" json:"-"` // 存储哈希值，JSON不回显
	QQ       string `gorm:"uniqueIndex;size:20" json:"qq"`
	Phone    string `gorm:"uniqueIndex;size:20" json:"phone"`
	Avatar   string `json:"avatar"`
	Role     string `gorm:"default:admin" json:"role"` // owner 或 admin
}

// LoginRequest 用于解析登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // 可以是用户名、QQ或手机
	Password string `json:"password" binding:"required"`
}
