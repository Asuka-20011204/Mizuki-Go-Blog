package service

import (
	"encoding/json"
	"my-blog-backend/config"
	"my-blog-backend/models"
	"os"
	"path/filepath"
	"strings"
)

type AlbumService struct{}

func (s *AlbumService) getAlbumsBaseDir() string {
	// 对应 public/images/albums/
	return filepath.Join(config.GlobalConfig.System.FrontendDir, "public/images/albums")
}

// ListAlbums 扫描目录获取所有相册
func (s *AlbumService) ListAlbums() ([]models.AlbumListItem, error) {
	baseDir := s.getAlbumsBaseDir()
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	var albums []models.AlbumListItem
	for _, entry := range entries {
		if entry.IsDir() {
			albumID := entry.Name()
			infoPath := filepath.Join(baseDir, albumID, "info.json")

			// 读取 info.json
			var info models.AlbumInfo
			infoData, err := os.ReadFile(infoPath)
			if err == nil {
				json.Unmarshal(infoData, &info)
			}

			albums = append(albums, models.AlbumListItem{
				ID:    albumID,
				Info:  info,
				Cover: "/images/albums/" + albumID + "/cover.jpg", // Web 访问路径
			})
		}
	}
	return albums, nil
}

// CreateAlbum 创建相册目录及初始化 info.json
func (s *AlbumService) CreateAlbum(id string, info models.AlbumInfo) error {
	albumDir := filepath.Join(s.getAlbumsBaseDir(), id)
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		return err
	}

	infoData, _ := json.MarshalIndent(info, "", "  ")
	return os.WriteFile(filepath.Join(albumDir, "info.json"), infoData, 0644)
}

// DeleteAlbum 物理删除相册文件夹
func (s *AlbumService) DeleteAlbum(id string) error {
	return os.RemoveAll(filepath.Join(s.getAlbumsBaseDir(), id))
}

// GetAlbumFiles 获取相册目录下的所有图片文件名
func (s *AlbumService) GetAlbumFiles(id string) ([]string, error) {
	albumDir := filepath.Join(s.getAlbumsBaseDir(), id)

	// 检查目录是否存在
	if _, err := os.Stat(albumDir); os.IsNotExist(err) {
		return []string{}, nil // 目录不存在返回空数组
	}

	entries, err := os.ReadDir(albumDir)
	if err != nil {
		return nil, err
	}

	// 显式初始化为切片，确保序列化为 JSON 时是 [] 而不是 null
	files := make([]string, 0)

	for _, entry := range entries {
		// 过滤逻辑：只保留图片，排除目录和 info.json
		if !entry.IsDir() && entry.Name() != "info.json" {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" {
				files = append(files, entry.Name())
			}
		}
	}
	return files, nil
}

// SetAlbumCover 将指定文件设为封面 (复制为 cover.jpg)
func (s *AlbumService) SetAlbumCover(id string, filename string) error {
	albumDir := filepath.Join(s.getAlbumsBaseDir(), id)
	srcPath := filepath.Join(albumDir, filename)
	dstPath := filepath.Join(albumDir, "cover.jpg")

	// 读取源文件并写入目标文件（实现物理覆盖）
	input, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, input, 0644)
}

// DeletePhoto 删除相册内单张照片
func (s *AlbumService) DeletePhoto(id string, filename string) error {
	return os.Remove(filepath.Join(s.getAlbumsBaseDir(), id, filename))
}
