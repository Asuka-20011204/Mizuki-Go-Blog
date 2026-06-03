package service

import (
	"errors"
	"my-blog-backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

// ListAdmins 列出所有管理员
func (s *UserService) ListAdmins() ([]models.UserInfo, error) {
	var users []models.User
	if err := s.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]models.UserInfo, len(users))
	for i, u := range users {
		result[i] = models.UserInfo{
			ID:       u.ID,
			Username: u.Username,
			QQ:       u.QQ,
			Phone:    u.Phone,
			Avatar:   u.Avatar,
			Role:     u.Role,
		}
	}
	return result, nil
}

// AddAdmin 添加管理员（只有 owner 可调用）
func (s *UserService) AddAdmin(req models.AddAdminRequest) error {
	// 检查用户名是否已被占用
	var existing models.User
	err := s.DB.Where("username = ?", req.Username).First(&existing).Error
	if err == nil {
		return errors.New("用户名已被占用")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	avatar := ""
	if req.QQ != "" {
		avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + req.QQ + "&s=100"
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		QQ:       req.QQ,
		Phone:    req.Phone,
		Avatar:   avatar,
		Role:     "admin",
	}

	return s.DB.Create(&user).Error
}

// DeleteAdmin 删除管理员（只有 owner 可调用）
func (s *UserService) DeleteAdmin(userID uint, requesterID uint) error {
	if userID == requesterID {
		return errors.New("不能删除自己的账号")
	}

	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.Role == "owner" {
		return errors.New("不能删除所有者账号")
	}

	return s.DB.Delete(&user).Error
}

// ChangePassword 修改自己的密码
func (s *UserService) ChangePassword(userID uint, req models.ChangePasswordRequest) error {
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("旧密码不正确")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	return s.DB.Model(&user).Update("password", string(hashedPassword)).Error
}
