package controller

import (
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserService *service.UserService
}

// ListAdmins GET /api/admin/users
func (uc *UserController) ListAdmins(c *gin.Context) {
	users, err := uc.UserService.ListAdmins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// AddAdmin POST /api/admin/users
func (uc *UserController) AddAdmin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有所有者才能添加管理员"})
		return
	}

	var req models.AddAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式不正确"})
		return
	}

	if err := uc.UserService.AddAdmin(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "管理员添加成功"})
}

// DeleteAdmin DELETE /api/admin/users/:id
func (uc *UserController) DeleteAdmin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有所有者才能删除管理员"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	requesterID, _ := c.Get("userID")
	if err := uc.UserService.DeleteAdmin(uint(userID), requesterID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "管理员已删除"})
}

// ChangePassword PUT /api/admin/users/password
func (uc *UserController) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式不正确"})
		return
	}

	userID, _ := c.Get("userID")
	if err := uc.UserService.ChangePassword(userID.(uint), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}
