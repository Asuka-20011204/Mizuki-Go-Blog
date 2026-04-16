package controller

import (
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConfigController struct {
	ConfigService *service.ConfigService
}

// GetConfig 返回当前配置
func (cc *ConfigController) GetConfig(c *gin.Context) {
	cfg, err := cc.ConfigService.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取配置文件失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// UpdateConfig 更新配置
func (cc *ConfigController) UpdateConfig(c *gin.Context) {
	var req models.FullConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误: " + err.Error()})
		return
	}

	// 可以在这里添加额外的校验逻辑，例如校验 URL 格式、必填字段等

	if err := cc.ConfigService.UpdateConfig(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置文件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已更新"})
}
