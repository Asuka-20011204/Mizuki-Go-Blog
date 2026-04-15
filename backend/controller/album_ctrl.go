package controller

import (
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"
	"os"
	"strings"

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

func (ac *AlbumController) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := ac.AlbumService.DeleteAlbum(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "相册已物理删除并同步至前端"})

}

// 获取相册文件列表
func (ac *AlbumController) GetFiles(c *gin.Context) {
	id := c.Param("id")
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
	var req struct {
		Filename string `json:"filename"`
	}
	c.ShouldBindJSON(&req)

	if err := ac.AlbumService.SetAlbumCover(id, req.Filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "封面设置成功"})
}

func (ac *AlbumController) DeletePhoto(c *gin.Context) {
	albumID := c.Param("id")
	filename := c.Param("filename")

	// 简单的参数校验
	if albumID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "相册ID和文件名不能为空",
		})
		return
	}

	// 可选：安全检查，防止路径遍历攻击（例如 filename 包含 ".." 或 "/"）
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "非法的文件名",
		})
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
				"error": "删除失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}
