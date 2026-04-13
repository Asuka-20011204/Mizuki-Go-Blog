package main

import (
	"log"
	"my-blog-backend/controller"
	"my-blog-backend/models"
	"my-blog-backend/service" // 确保导入了 service 包

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 初始化数据库连接 (Model 层)
	dsn := "root:Wangzixu880314.@tcp(127.0.0.1:3306)/blog_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动同步表结构
	if err := db.AutoMigrate(&models.Post{}); err != nil {
		log.Fatalf("数据库迁移失败，%v", err)
	}

	// 2. 链式依赖注入 (MVC 核心)
	// 第一步：将 db 注入 Service
	postService := &service.PostService{DB: db}
	// 第二步：将 service 注入 Controller
	postCtrl := &controller.PostController{PostService: postService}
	systemService := &service.SystemService{}
	systemCtrl := &controller.SystemController{SystemService: systemService}

	// 3. 设置 Gin 路由
	r := gin.Default()

	// 配置 CORS 跨域
	r.Use(corsMiddleware())

	// 业务接口组
	api := r.Group("/api")
	{
		api.POST("/posts", postCtrl.HandleCreatePost)
		api.POST("/upload", postCtrl.HandleUpload)
	}

	// 管理接口组
	admin := r.Group("/api/admin")
	{
		admin.POST("/rebuild", systemCtrl.HandleRebuild)
		admin.GET("/posts", postCtrl.ListPosts)
		admin.GET("/posts/:slug", postCtrl.GetPostDetail) // 新增这一行
		admin.DELETE("/posts/:slug", postCtrl.DeletePost)
	}
	r.Static("/preview-cache", "../frontend/public/preview-cache")

	// 4. 启动服务器
	log.Println("Go 后端 MVC 服务已启动，监听端口 :8080")
	if err := r.Run(":8080"); err != nil {
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
