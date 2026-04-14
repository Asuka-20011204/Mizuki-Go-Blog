package controller

import (
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlbumController struct {
	AlbumService *service.AlbumService
}

func (ac *AlbumController) List(c *gin.Context) {
	albums, err := ac.AlbumService.ListAlbums()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, albums)
}

func (ac *AlbumController) Create(c *gin.Context) {
	var req struct {
		ID   string           `json:"id" binding:"required"`
		Info models.AlbumInfo `json:"info"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := ac.AlbumService.CreateAlbum(req.ID, req.Info); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "相册创建成功"})
}
