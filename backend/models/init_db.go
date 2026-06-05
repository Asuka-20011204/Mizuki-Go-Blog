package models

import (
	"my-blog-backend/config"
	"my-blog-backend/logger"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func InitDatabase(db *gorm.DB) {
	// 1. 自动迁移表
	db.AutoMigrate(&Post{}, &User{})

	// 2. 初始化 Owner 用户
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		logger.Info("正在初始化默认所有者账号...")

		conf := config.GlobalConfig.InitData

		// 安全检查：密码不能为空且不少于 8 位
		if conf.OwnerPassword == "" {
			logger.Fatal("所有者初始化密码为空，请通过环境变量 OWNER_PASSWORD 设置至少 8 位密码")
		}
		if len(conf.OwnerPassword) < 8 {
			logger.Fatal("所有者初始化密码太短", "min_length", 8)
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(conf.OwnerPassword), bcrypt.DefaultCost)
		// 自动根据 QQ 号生成初始头像地址
		defaultAvatar := ""
		if conf.OwnerQQ != "" {
			defaultAvatar = "https://q1.qlogo.cn/g?b=qq&nk=" + conf.OwnerQQ + "&s=100"
		}

		owner := User{
			Username: conf.OwnerUsername,
			Password: string(hashedPassword),
			QQ:       conf.OwnerQQ,
			Phone:    conf.OwnerPhone,
			Avatar:   defaultAvatar,
			Role:     "owner",
		}

		if err := db.Create(&owner).Error; err != nil {
			logger.Fatal("初始化所有者失败", "error", err)
		}
		logger.Info("所有者账号初始化成功！")
	}
}
