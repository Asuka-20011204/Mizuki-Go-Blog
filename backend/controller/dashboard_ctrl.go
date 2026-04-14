package controller

import (
	"my-blog-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	DashboardService *service.DashboardService
}

func (dc *DashboardController) GetStats(c *gin.Context) {
	stats, err := dc.DashboardService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计数据失败"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
