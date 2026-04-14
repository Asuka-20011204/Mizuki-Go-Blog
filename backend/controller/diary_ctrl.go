package controller

import (
	"my-blog-backend/models"
	"my-blog-backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DiaryController struct {
	DiaryService *service.DiaryService
}

func (dc *DiaryController) List(c *gin.Context) {
	data, err := dc.DiaryService.GetDiaries()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func (dc *DiaryController) Create(c *gin.Context) {
	var req models.DiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	if err := dc.DiaryService.SaveDiary(req); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "日记发布成功"})
}

func (dc *DiaryController) Delete(c *gin.Context) {
	// 从路由参数中获取 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的 ID 格式"})
		return
	}

	if err := dc.DiaryService.DeleteDiary(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "日记已物理删除并同步至前端"})
}
