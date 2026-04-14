package models

// DiaryItem 严格对应 Mizuki 的 DiaryItem 接口
type DiaryItem struct {
	ID       int      `json:"id"`
	Content  string   `json:"content"`
	Date     string   `json:"date"`     // ISO 8601 格式
	Images   []string `json:"images"`   // 可选
	Location string   `json:"location"` // 扩展字段
	Mood     string   `json:"mood"`     // 扩展字段
}

// DiaryRequest 前端传来的数据
type DiaryRequest struct {
	Content  string   `json:"content" binding:"required"`
	Images   []string `json:"images"`
	Location string   `json:"location"`
	Mood     string   `json:"mood"`
}
