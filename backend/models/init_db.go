package models

import (
	"log"
	"my-blog-backend/config"

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
		log.Println("正在初始化默认所有者账号...")

		conf := config.GlobalConfig.InitData
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
			log.Fatalf("初始化所有者失败: %v", err)
		}
		log.Println("所有者账号初始化成功！")
	}
}
