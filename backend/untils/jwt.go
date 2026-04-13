package utils

import (
	"errors"
	"my-blog-backend/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 根据用户信息签发 Token
func GenerateToken(userID uint, username, role string) (string, error) {
	conf := config.GlobalConfig.JWT

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			// 从配置文件读取过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(conf.Expire) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 从配置文件读取密钥
	return token.SignedString([]byte(conf.Secret))
}

// ParseToken 解析并验证 Token
func ParseToken(tokenString string) (*Claims, error) {
	conf := config.GlobalConfig.JWT

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
