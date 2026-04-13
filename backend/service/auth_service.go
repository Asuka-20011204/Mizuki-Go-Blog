package service

import (
	"errors"
	"my-blog-backend/models"
	utils "my-blog-backend/untils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

func (s *AuthService) Authenticate(account, password string) (string, *models.User, error) {
	var user models.User
	// 核心解耦逻辑：一个 SQL 匹配三种身份标识
	err := s.DB.Where("username = ? OR qq = ? OR phone = ?", account, account, account).First(&user).Error
	if err != nil {
		return "", nil, errors.New("账号或密码错误")
	}

	// 验证哈希密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("账号或密码错误")
	}

	// 生成 JWT
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", nil, errors.New("生成令牌失败")
	}

	return token, &user, nil
}
