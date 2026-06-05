package controller

import (
	"my-blog-backend/logger"
	"my-blog-backend/models"
	"my-blog-backend/service"
	"my-blog-backend/validator"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type AlbumController struct {
	AlbumService *service.AlbumService
}

func (ac *AlbumController) List(c *gin.Context) {
	albums, err := ac.AlbumService.ListAlbums()
	if err != nil {
		logger.Error("获取相册列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取相册列表失败"})
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
	if !validator.SafePath(req.ID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "相册 ID 包含非法字符"})
		return
	}
	if err := ac.AlbumService.CreateAlbum(req.ID, req.Info); err != nil {
		logger.Error("创建相册失败", "error", err, "id", req.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建相册失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "相册创建成功"})
}

func (ac *AlbumController) Delete(c *gin.Context) {
	id := c.Param("id")
	if !validator.SafePath(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "相册 ID 包含非法字符"})
		return
	}

	if err := ac.AlbumService.DeleteAlbum(id); err != nil {
		logger.Error("删除相册失败", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除相册失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "相册已物理删除并同步至前端"})

}

// 获取相册文件列表
func (ac *AlbumController) GetFiles(c *gin.Context) {
	id := c.Param("id")
	if !validator.SafePath(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "相册 ID 包含非法字符"})
		return
	}
	files, err := ac.AlbumService.GetAlbumFiles(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取相册目录失败"})
		return
	}
	c.JSON(http.StatusOK, files)
}

// 设置封面
func (ac *AlbumController) SetCover(c *gin.Context) {
	id := c.Param("id")
	if !validator.SafePath(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "相册 ID 包含非法字符"})
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if !validator.SafePath(req.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件名包含非法字符"})
		return
	}

	if err := ac.AlbumService.SetAlbumCover(id, req.Filename); err != nil {
		logger.Error("设置封面失败", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置封面失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "封面设置成功"})
}

func (ac *AlbumController) DeletePhoto(c *gin.Context) {
	albumID := c.Param("id")
	filename := c.Param("filename")

	if albumID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "相册ID和文件名不能为空"})
		return
	}
	if !validator.SafePath(albumID) || !validator.SafePath(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "包含非法字符"})
		return
	}

	// 调用服务层删除文件
	err := ac.AlbumService.DeletePhoto(albumID, filename)
	if err != nil {
		// 如果文件不存在，返回 404；其他错误返回 500
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "图片不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "删除照片失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}
