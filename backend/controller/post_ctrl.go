package controller

import (
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误: " + err.Error()})
		return
	}

	// 严格 MVC：调用 Service 层处理复杂的“写文件+存数据库”逻辑
	if err := pc.PostService.ProcessPostPublish(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败: " + err.Error()})
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

	// 调用 Service

	fileName, err := pc.PostService.SavePostResource(slug, file, isCover)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"code": 1,
		})
		return
	}

	// 严格适配 Mizuki 的 Vditor 返回格式
	c.JSON(http.StatusOK, gin.H{
		"msg":  "上传成功",
		"code": 0, // 0 表示成功
		"data": gin.H{
			"errFiles": []string{},
			"succMap": map[string]string{
				file.Filename: "/preview-cache/" + fileName,
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

	if err := pc.PostService.DeletePost(slug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章已从数据库和文件系统彻底删除"})
}

// GetPostDetail 获取单篇详情接口
func (pc *PostController) GetPostDetail(c *gin.Context) {
	slug := c.Param("slug")
	post, err := pc.PostService.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	c.JSON(http.StatusOK, post)
}
