package service

import (
	"encoding/json"
	"my-blog-backend/config"
	"my-blog-backend/models"
	"os"
	"path/filepath"
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
