package service

import (
	"log/slog"
	"my-blog-backend/config"
	"my-blog-backend/models"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DashboardService struct {
	DB *gorm.DB
}

type StatsResult struct {
	PostCount     int64 `json:"postCount"`
	CategoryCount int   `json:"categoryCount"`
	RunningDays   int   `json:"runningDays"`
	TotalWords    int   `json:"totalWords"`
}

func (s *DashboardService) GetStats() (StatsResult, error) {
	var stats StatsResult

	// 1. 获取文章总数
	s.DB.Model(&models.Post{}).Count(&stats.PostCount)

	// 2. 获取分类数量
	var categories []string
	s.DB.Model(&models.Post{}).Distinct("category").Pluck("category", &categories)
	stats.CategoryCount = len(categories)

	// 3. 计算运行天数 (以第一篇文章的创建时间为准，若无则按网站初始时间)
	var firstPost models.Post
	err := s.DB.Order("created_at asc").First(&firstPost).Error
	if err == nil {
		duration := time.Since(firstPost.CreatedAt)
		stats.RunningDays = int(duration.Hours()/24) + 1
	} else {
		stats.RunningDays = 1 // 初始状态
	}

	// 4. 统计总字数 (遍历所有 index.md 物理文件统计最准确)
	stats.TotalWords = s.calculateTotalWords()
	slog.Info("获取文章状态成功！", "stats", stats)

	return stats, nil
}

// 辅助函数：统计所有文章的物理字数
func (s *DashboardService) calculateTotalWords() int {
	total := 0
	postsBaseDir := config.GlobalConfig.System.PostsDir // 使用配置路径 [cite: 17, 30]

	filepath.Walk(postsBaseDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			content, err := os.ReadFile(path)
			if err == nil {
				// 简单统计：去除空格后的字符数
				text := string(content)
				// 过滤 Frontmatter (--- ... ---) 逻辑可在此细化
				total += len([]rune(text))
			}
		}
		return nil
	})
	return total
}
