package models

// AlbumInfo 对应 info.json 的结构
type AlbumInfo struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Date        string   `json:"date,omitempty"`
	Location    string   `json:"location,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Layout      string   `json:"layout,omitempty"` // grid 或 masonry
	Columns     int      `json:"columns,omitempty"`
	Hidden      bool     `json:"hidden"`
	Mode        string   `json:"mode,omitempty"` // external 或 local
	Cover       string   `json:"cover,omitempty"`
	Photos      []Photo  `json:"photos,omitempty"` // 仅外链模式
}

type Photo struct {
	ID          string   `json:"id,omitempty"`
	Src         string   `json:"src"`
	Alt         string   `json:"alt,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// AlbumListItem 用于管理后台列表展示
type AlbumListItem struct {
	ID    string    `json:"id"` // 文件夹名
	Info  AlbumInfo `json:"info"`
	Cover string    `json:"cover"` // 封面访问路径
}
