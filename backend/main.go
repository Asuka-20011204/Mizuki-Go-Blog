package main

import (
	"fmt"
	"log"
	"my-blog-backend/config"
	"my-blog-backend/controller"
	"my-blog-backend/middleware"
	"my-blog-backend/models"
	"my-blog-backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	config.LoadConfig("config.yaml")

	// 2. 使用配置连接数据库
	dsn := config.GlobalConfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
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
	// 第二步：将 service 注入 Controller
	postCtrl := &controller.PostController{PostService: postService}
	authCtrl := &controller.AuthController{AuthService: authService}
	systemCtrl := &controller.SystemController{SystemService: systemService}
	diaryCtrl := &controller.DiaryController{DiaryService: diaryService}
	dashboardCtrl := &controller.DashboardController{DashboardService: dashboardService}
	albumCtrl := &controller.AlbumController{AlbumService: albumService}

	// 3. 设置 Gin 路由
	r := gin.Default()

	// 配置 CORS 跨域
	r.Use(corsMiddleware())

	// 业务接口组
	r.POST("/api/login", authCtrl.Login)
	// 管理接口组
	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTAuth())
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
	}
	r.Static("/preview-cache", "../frontend/public/preview-cache")

	// 4. 启动服务器
	log.Printf("Go 后端 MVC 服务已启动，监听端口 :%d", config.GlobalConfig.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// corsMiddleware 保持不变
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
