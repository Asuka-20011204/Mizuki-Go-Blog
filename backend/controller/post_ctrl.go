package controller

import (
	"my-blog-backend/logger"
	"my-blog-backend/models"
	"my-blog-backend/service"
	"my-blog-backend/validator"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	// 严格 MVC：Controller 持有 Service 实例，而不是直接操作 DB
	PostService *service.PostService
}

// HandleCreatePost 处理文章创建与更新
func (pc *PostController) HandleCreatePost(c *gin.Context) {
	var req models.PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if !validator.SafeSlug(req.Slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug 包含非法字符"})
		return
	}

	// 严格 MVC：调用 Service 层处理复杂的"写文件+存数据库"逻辑
	if err := pc.PostService.ProcessPostPublish(req); err != nil {
		logger.Error("文章发布失败", "error", err, "slug", req.Slug)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章已发布并同步至数据库"})
}

// HandleUpload 处理图片/文件上传
func (pc *PostController) HandleUpload(c *gin.Context) {
	// 适配 Mizuki：Vditor 上传图片时，slug 有可能是在 URL 参数中或者 PostForm 中
	slug := c.PostForm("slug")
	if slug == "" {
		slug = c.Query("slug") // 尝试从 URL 参数获取
	}

	// 如果还是没有 slug，Mizuki 允许上传到预览目录
	if slug == "" {
		slug = "preview-cache"
	}

	// 校验 slug 合法性
	if slug != "preview-cache" && slug != "diary" && !strings.HasPrefix(slug, "albums/") && !validator.SafeSlug(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Slug 包含非法字符", "code": 1})
		return
	}
	// 校验 albums 路径
	if strings.HasPrefix(slug, "albums/") {
		albumID := strings.TrimPrefix(slug, "albums/")
		if !validator.SafePath(albumID) {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "相册 ID 包含非法字符", "code": 1})
			return
		}
	}

	// 获取 is_cover 参数
	isCover := c.PostForm("is_cover") == "true"

	file, err := c.FormFile("file") // 严格匹配前端 append('file', ...)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  "未接收到文件",
			"code": 1, // Vditor 非0表示失败
		})
		return
	}

	// 文件大小限制 10MB
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "文件大小不能超过 10MB", "code": 1})
		return
	}

	// 只允许图片类型
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "只允许上传图片文件", "code": 1})
		return
	}

	// 调用 Service

	fileName, err := pc.PostService.SavePostResource(slug, file, isCover)
	if err != nil {
		logger.Error("文件上传失败", "error", err, "slug", slug)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "上传失败", "code": 1})
		return
	}

	// 严格适配 Mizuki 的 Vditor 返回格式
	c.JSON(http.StatusOK, gin.H{
		"msg":  "上传成功",
		"code": 0, // 0 表示成功
		"data": gin.H{
			"errFiles": []string{},
			"succMap": map[string]string{
				file.Filename: fileName,
			},
		},
	})
}

// ListPosts 获取所有文章列表
func (pc *PostController) ListPosts(c *gin.Context) {
	posts, err := pc.PostService.GetAllPosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// DeletePost 处理同步删除请求
func (pc *PostController) DeletePost(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug 不能为空"})
		return
	}
	if !validator.SafeSlug(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug 包含非法字符"})
		return
	}

	if err := pc.PostService.DeletePost(slug); err != nil {
		logger.Error("文章删除失败", "error", err, "slug", slug)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章已从数据库和文件系统彻底删除"})
}

// GetPostDetail 获取单篇详情接口
func (pc *PostController) GetPostDetail(c *gin.Context) {
	slug := c.Param("slug")
	if !validator.SafeSlug(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug 包含非法字符"})
		return
	}
	post, err := pc.PostService.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	c.JSON(http.StatusOK, post)
}
