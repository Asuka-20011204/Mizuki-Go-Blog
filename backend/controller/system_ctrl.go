package controller

import (
	"my-blog-backend/logger"
	"my-blog-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemController struct {
	SystemService *service.SystemService
}

func (sc *SystemController) HandleRebuild(c *gin.Context) {
	err := sc.SystemService.RebuildProject()
	if err != nil {
		logger.Error("重构失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重构失败，请查看后台日志"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "重构任务执行成功"})
}
