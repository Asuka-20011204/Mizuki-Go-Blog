package controller

import (
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "重构任务执行成功"})
}
