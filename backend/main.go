package main

import (
	"fmt"
	"my-blog-backend/config"
	"my-blog-backend/controller"
	"my-blog-backend/logger"
	"my-blog-backend/middleware"
	"my-blog-backend/models"
	"my-blog-backend/service"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	config.LoadConfig("config.yaml")

	// 1.2. 设置 Gin 运行模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 1.5. 初始化日志
	if err := logger.Init(config.GlobalConfig.Log.Dir, config.GlobalConfig.Log.Level); err != nil {
		fmt.Printf("日志初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 使用配置连接数据库
	dsn := config.GlobalConfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("数据库连接失败", "error", err)
	}

	// 3. 初始化数据 (包含创建初始用户)
	models.InitDatabase(db)

	// 2. 链式依赖注入 (MVC 核心)
	// 第一步：将 db 注入 Service
	postService := &service.PostService{DB: db}
	authService := &service.AuthService{DB: db}
	systemService := &service.SystemService{}
	diaryService := &service.DiaryService{}
	dashboardService := &service.DashboardService{DB: db}
	albumService := &service.AlbumService{}
	configService := &service.ConfigService{}
	userService := &service.UserService{DB: db}

	// 第二步：将 service 注入 Controller
	postCtrl := &controller.PostController{PostService: postService}
	authCtrl := &controller.AuthController{AuthService: authService}
	systemCtrl := &controller.SystemController{SystemService: systemService}
	diaryCtrl := &controller.DiaryController{DiaryService: diaryService}
	dashboardCtrl := &controller.DashboardController{DashboardService: dashboardService}
	albumCtrl := &controller.AlbumController{AlbumService: albumService}
	configCtrl := &controller.ConfigController{ConfigService: configService}
	userCtrl := &controller.UserController{UserService: userService}
	// 3. 设置 Gin 路由
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.MaxMultipartMemory = 10 << 20 // 上传文件最大 10MB

	// 自定义 Gin 中间件：Logger（输出到文件） + Recovery
	r.Use(logger.GinLogger(), gin.Recovery())

	// 安全响应头（在 CORS 之前）
	r.Use(securityHeaders())
	// 配置 CORS 跨域
	r.Use(corsMiddleware())

	// 业务接口组（登录限流：每分钟最多 5 次尝试）
	r.POST("/api/login", middleware.RateLimit(5, time.Minute), authCtrl.Login)
	// 管理接口组
	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTAuth(db))
	{
		admin.POST("/posts", postCtrl.HandleCreatePost)
		admin.POST("/upload", postCtrl.HandleUpload)
		admin.POST("/rebuild", systemCtrl.HandleRebuild)
		admin.GET("/posts", postCtrl.ListPosts)
		admin.GET("/posts/:slug", postCtrl.GetPostDetail)
		admin.DELETE("/posts/:slug", postCtrl.DeletePost)
		admin.GET("/diaries", diaryCtrl.List)
		admin.POST("/diaries", diaryCtrl.Create)
		admin.DELETE("/diaries/:id", diaryCtrl.Delete)
		admin.GET("/stats", dashboardCtrl.GetStats)
		admin.GET("/albums", albumCtrl.List)
		admin.POST("/albums", albumCtrl.Create)
		admin.DELETE("/albums/:id", albumCtrl.Delete)
		admin.GET("/albums/:id/files", albumCtrl.GetFiles)
		admin.POST("/albums/:id/set-cover", albumCtrl.SetCover)
		admin.DELETE("/albums/:id/files/:filename", albumCtrl.DeletePhoto)
		admin.GET("/config", configCtrl.GetConfig)
		admin.POST("/config", configCtrl.UpdateConfig)
		admin.GET("/users", userCtrl.ListAdmins)
		admin.POST("/users", userCtrl.AddAdmin)
		admin.DELETE("/users/:id", userCtrl.DeleteAdmin)
		admin.PUT("/users/password", userCtrl.ChangePassword)

	}
	// 处理带尾斜杠的请求：匹配不到路由时，去掉斜杠重试
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 1. 尾斜杠 -> 去斜杠重试
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			c.Request.URL.Path = path[:len(path)-1]
			r.HandleContext(c)
			return
		}

		// 2. 尝试托管 frontend/dist 下的静态文件
		staticRoot, _ := filepath.Abs(filepath.Join("..", "frontend", "dist"))
		// 去掉路径开头的分隔符，确保是相对路径，防止 filepath.Join 丢弃 staticRoot
		cleanPath := strings.TrimLeft(filepath.Clean(path), "/\\")
		requestedPath := filepath.Join(staticRoot, cleanPath)
		// 安全检查：确保不越出 staticRoot
		if !strings.HasPrefix(requestedPath, staticRoot+string(filepath.Separator)) && requestedPath != staticRoot {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// 打开文件/目录
		info, err := os.Stat(requestedPath)
		if err == nil {
			if info.IsDir() {
				// 目录 -> 尝试 index.html
				indexPath := filepath.Join(requestedPath, "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					c.File(indexPath)
					return
				}
			} else {
				c.File(requestedPath)
				return
			}
		}

		// 3. 都不匹配 -> 返回 404 页面或 JSON
		notFoundPage := filepath.Join(staticRoot, "404.html")
		if _, err := os.Stat(notFoundPage); err == nil {
			c.File(notFoundPage)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	r.Static("/preview-cache", "../frontend/public/preview-cache")

	// 4. 启动服务器
	logger.Info("Go 后端 MVC 服务已启动", "port", config.GlobalConfig.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)); err != nil {
		logger.Fatal("服务启动失败", "error", err)
	}
}

// corsMiddleware 限制跨域来源，支持环境变量 CORS_ORIGIN 配置
func corsMiddleware() gin.HandlerFunc {
	allowOrigin := os.Getenv("CORS_ORIGIN")
	if allowOrigin == "" {
		allowOrigin = "http://localhost:4321" // 开发默认值
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// 匹配允许的来源
		if origin == allowOrigin || allowOrigin == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, Cache-Control")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// securityHeaders 添加基本安全响应头
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}
