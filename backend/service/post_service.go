package service

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"my-blog-backend/models"

	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 基础路径配置
const postsBaseDir = "../frontend/src/content/posts"
const previewDir = "../frontend/public/preview-cache"

type PostService struct {
	DB *gorm.DB
}

// 1. 获取所有文章列表 (对接管理后台表格)
func (s *PostService) GetAllPosts() ([]models.Post, error) {
	var posts []models.Post
	// 按发布时间倒序排列
	err := s.DB.Order("created_at desc").Find(&posts).Error
	return posts, err
}

// 3. 处理文章发布 (核心：同步数据库 + 写入 index.md)
func (s *PostService) ProcessPostPublish(req models.PostRequest) error {
	cleanBody := req.Content
	if strings.HasPrefix(strings.TrimSpace(cleanBody), "---") {
		parts := strings.SplitN(cleanBody, "---", 3)
		if len(parts) >= 3 {
			cleanBody = strings.TrimSpace(parts[2])
		}
	}

	publishedDate := time.Now().Format("2006-01-02")
	formattedTags := "[]"
	if len(req.Tags) > 0 {
		formattedTags = fmt.Sprintf("['%s']", strings.Join(req.Tags, "', '"))
	}

	// 保持你原有的 MD 模板
	mdTemplate := `---
title: "%s"
published: %s
description: "%s"
image: "%s"
tags: %s
category: "%s"
pinned: %v
lang: "%s"
draft: false
---

%s`

	// 使用清洗后的 cleanBody
	fullContent := fmt.Sprintf(mdTemplate, req.Title, publishedDate, req.Description, req.Image, formattedTags, req.Category, req.Pinned, req.Lang, cleanBody)
	// 确保文章目录存在 (index.md 模式)
	postDir := filepath.Join(postsBaseDir, req.Slug)
	if err := os.MkdirAll(postDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	savePath := filepath.Join(postDir, "index.md")
	if err := os.WriteFile(savePath, []byte(fullContent), 0644); err != nil {
		return fmt.Errorf("写入 MD 文件失败: %v", err)
	}

	// --- B. 数据库同步处理 ---
	post := models.Post{
		Title:       req.Title,
		Slug:        req.Slug,
		Description: req.Description,
		Content:     req.Content,
		Category:    req.Category,
		Tags:        strings.Join(req.Tags, ","),
		Image:       req.Image,
		Lang:        req.Lang,
		Pinned:      req.Pinned,
	}

	// 适配数据库：使用 Slug 作为唯一键进行更新或创建
	return s.DB.Unscoped().Where("slug = ?", req.Slug).Assign(post).FirstOrCreate(&post).Error
}

// 4. 同步物理删除与数据库删除
func (s *PostService) DeletePost(slug string) error {
	// 1. 删除磁盘文件
	postPath := filepath.Join(postsBaseDir, slug)
	_ = os.RemoveAll(postPath)

	// 2. 删除数据库记录
	return s.DB.Unscoped().Where("slug = ?", slug).Delete(&models.Post{}).Error
}

// 5. 保存资源 (图片) - 保持你原有的预览逻辑
func (s *PostService) SavePostResource(slug string, file *multipart.FileHeader) (string, error) {
	postDir := filepath.Join(postsBaseDir, slug)
	os.MkdirAll(postDir, 0755)
	os.MkdirAll(previewDir, 0755)

	// 动态获取文件后缀，不要硬编码
	ext := filepath.Ext(file.Filename) // 会得到 .png, .jpg, .webp 等

	newFileName := uuid.New().String() + ext

	// 生成保存路径
	dstPath := filepath.Join(postDir, newFileName)
	fmt.Printf("保存路径为：%s\n", dstPath)

	// 保存原始文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	io.Copy(out, src)

	// 拷贝到路径 B (预览)
	previewPath := filepath.Join(previewDir, newFileName)
	if err := s.copyFile(dstPath, previewPath); err != nil {
		log.Printf("预览文件拷贝失败: %v", err)
	}

	return newFileName, nil
}

// 辅助函数
func (s *PostService) copyFile(srcPath, dstPath string) error {
	sourceFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return destFile.Sync()
}

// GetPostBySlug 根据 Slug 获取单篇文章完整数据（用于编辑回显）
func (s *PostService) GetPostBySlug(slug string) (models.Post, error) {
	var post models.Post
	err := s.DB.Unscoped().Where("slug = ?", slug).First(&post).Error
	if err != nil {
		return post, err
	}

	postDir := filepath.Join(postsBaseDir, slug)
	savePath := filepath.Join(postDir, "index.md")

	content, err := os.ReadFile(savePath)
	if err == nil {
		fullContent := string(content)
		// --- 逻辑优化：剥离 Frontmatter ---
		// 寻找第二个 "---"
		parts := strings.SplitN(fullContent, "---", 3)
		if len(parts) >= 3 {
			// parts[0] 为空, parts[1] 为 header, parts[2] 为 body
			post.Content = strings.TrimSpace(parts[2])
		} else {
			post.Content = fullContent
		}
	}

	return post, nil
}
