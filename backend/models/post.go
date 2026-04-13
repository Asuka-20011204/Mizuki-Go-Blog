package models

import "gorm.io/gorm"

// 数据库模型
type Post struct {
	gorm.Model
	Title       string `json:"title"`
	Slug        string `json:"slug" gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string `json:"description"`
	Content     string `json:"content" gorm:"type:longtext"`
	Category    string `json:"category"`
	Tags        string `json:"tags"` // 存储为逗号分隔字符串
	Image       string `json:"image"`
	Lang        string `json:"lang"`
	Pinned      bool   `json:"pinned"`
}

// Data Transfer Object (用于 API 交互)
type PostRequest struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Image       string   `json:"image"`
	Lang        string   `json:"lang"`
	Pinned      bool     `json:"pinned"`
}
