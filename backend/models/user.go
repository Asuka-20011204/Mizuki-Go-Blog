package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;size:50" json:"username"`
	Password     string `gorm:"not null" json:"-"` // 存储哈希值，JSON不回显
	QQ           string `gorm:"uniqueIndex;size:20" json:"qq"`
	Phone        string `gorm:"uniqueIndex;size:20" json:"phone"`
	Avatar       string `json:"avatar"`
	Role         string `gorm:"default:admin" json:"role"`  // owner 或 admin
	TokenVersion int    `gorm:"default:1" json:"-"`         // 改密码/删用户时自增，旧 token 即失效
}

// LoginRequest 用于解析登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // 可以是用户名、QQ或手机
	Password string `json:"password" binding:"required"`
}

// AddAdminRequest 添加管理员请求
type AddAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	QQ       string `json:"qq"`
	Phone    string `json:"phone"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UserInfo 返回给前端的用户信息（不含密码）
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	QQ       string `json:"qq"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}
