package controller

import (
	"my-blog-backend/service"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	DashboardService *service.DashboardService
}

func (dc *DashboardController) GetStats(c *gin.Context) {
	stats, err := dc.DashboardService.GetStats()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取统计数据失败"})
		return
	}
	c.JSON(200, stats)
}
