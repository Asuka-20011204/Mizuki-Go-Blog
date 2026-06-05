package controller

import (
	"my-blog-backend/logger"
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DiaryController struct {
	DiaryService *service.DiaryService
}

func (dc *DiaryController) List(c *gin.Context) {
	data, err := dc.DiaryService.GetDiaries()
	if err != nil {
		logger.Error("获取日记列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日记失败"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (dc *DiaryController) Create(c *gin.Context) {
	var req models.DiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := dc.DiaryService.SaveDiary(req); err != nil {
		logger.Error("保存日记失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存日记失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "日记发布成功"})
}

func (dc *DiaryController) Delete(c *gin.Context) {
	// 从路由参数中获取 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID 格式"})
		return
	}

	if err := dc.DiaryService.DeleteDiary(id); err != nil {
		logger.Error("删除日记失败", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除日记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "日记已物理删除并同步至前端"})
}
