package middleware

import (
	"my-blog-backend/models"
	utils "my-blog-backend/untils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func JWTAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权，请先登录"})
			c.Abort()
			return
		}

		// 2. 截取 Token
		tokenString := authHeader[7:]

		// 3. 解析 Token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效或已过期"})
			c.Abort()
			return
		}

		// 4. 检查 token 版本：如果用户改过密码，旧 token 失效
		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}
		if claims.TokenVersion < user.TokenVersion {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码已修改，请重新登录"})
			c.Abort()
			return
		}

		// 5. 将用户信息注入 Context，方便下游逻辑使用
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
