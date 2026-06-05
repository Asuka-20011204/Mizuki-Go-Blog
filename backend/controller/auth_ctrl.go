package controller

import (
	"my-blog-backend/logger"
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	AuthService *service.AuthService
}

func (ac *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("登录请求格式错误", "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式不正确"})
		return
	}

	token, user, err := ac.AuthService.Authenticate(req.Account, req.Password)
	if err != nil {
		logger.Warn("登录失败", "account", req.Account, "ip", c.ClientIP(), "reason", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	logger.Info("登录成功", "account", req.Account, "username", user.Username, "role", user.Role, "ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"username": user.Username,
			"role":     user.Role,
			"avatar":   user.Avatar,
		},
	})
}
