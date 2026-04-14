package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"my-blog-backend/config"
	"my-blog-backend/models"
	"os"
	"path/filepath"
	"time"
)

type DiaryService struct{}

// 核心：读取纯 JSON 文件
func (s *DiaryService) GetDiaries() ([]models.DiaryItem, error) {
	path := config.GlobalConfig.System.DiaryJsonPath

	// 如果文件不存在，返回空数组而不是报错
	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Warn("数组为空")
		return []models.DiaryItem{}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		slog.Error("读取文件失败", "error", err)
		return nil, err
	}

	var diaries []models.DiaryItem
	if err := json.Unmarshal(content, &diaries); err != nil {
		slog.Error("解析json文件失败", "error", err)
		return nil, fmt.Errorf("JSON 数据损坏: %v", err)
	}
	return diaries, nil
}

// SaveDiary 添加新日记
func (s *DiaryService) SaveDiary(req models.DiaryRequest) error {
	diaries, err := s.GetDiaries()
	if err != nil {
		return err
	}

	// 生成新 ID（遍历取最大值）
	maxID := 0
	for _, d := range diaries {
		if d.ID > maxID {
			maxID = d.ID
		}
	}
	newID := maxID + 1

	newItem := models.DiaryItem{
		ID:       newID,
		Content:  req.Content,
		Date:     time.Now().UTC().Format(time.RFC3339),
		Images:   req.Images,
		Location: req.Location,
		Mood:     req.Mood,
	}

	// 追加到末尾（保持升序：旧→新）
	diaries = append(diaries, newItem)
	//  持久化到后端自己的 JSON
	jsonData, _ := json.MarshalIndent(diaries, "", "  ")
	if err := os.WriteFile(config.GlobalConfig.System.DiaryJsonPath, jsonData, 0644); err != nil {
		slog.Error("写入json文件失败", "error", err)
		return err
	}
	err = s.SyncToFrontend(diaries)
	if err != nil {
		slog.Error("写回前端日记ts数据文件失败", "error", err)
		return err
	}
	slog.Info("日记保存成功,日记ID", "id", newID)
	return nil
}

// 将数组写回为合法的 TypeScript 文件
func (s *DiaryService) SyncToFrontend(diaries []models.DiaryItem) error {
	jsonData, _ := json.MarshalIndent(diaries, "", "\t")

	// 只输出数据，不包含任何逻辑函数
	tsDataContent := fmt.Sprintf("// 此文件由后端自动生成，请勿手动修改逻辑\nexport const diaryData = %s;", string(jsonData))

	// 修改路径，存为数据专用文件
	dataFilePath := filepath.Join(config.GlobalConfig.System.FrontendDir, "src/data/diary_data.ts")

	return os.WriteFile(dataFilePath, []byte(tsDataContent), 0644)
}

// DeleteDiary 根据 ID 删除日记
func (s *DiaryService) DeleteDiary(id int) error {
	// 1. 获取当前所有日记
	diaries, err := s.GetDiaries()
	if err != nil {
		return err
	}

	// 2. 过滤掉目标 ID
	newDiaries := []models.DiaryItem{}
	found := false
	for _, item := range diaries {
		if item.ID != id {
			newDiaries = append(newDiaries, item)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("未找到 ID 为 %d 的日记", id)
	}

	// 3. 持久化到后端 JSON 文件
	jsonData, _ := json.MarshalIndent(newDiaries, "", "  ")
	if err := os.WriteFile(config.GlobalConfig.System.DiaryJsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("更新 JSON 失败: %v", err)
	}

	// 4. 同步更新前端的 diary_data.ts
	// 这确保了即使不执行 Build，开发环境下的前端数据也是最新的
	return s.SyncToFrontend(newDiaries)
}
