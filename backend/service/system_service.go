package service

import (
	"fmt"
	"my-blog-backend/logger"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type SystemService struct{}

func (s *SystemService) RebuildProject() error {
	// 1. 获取前端绝对路径 (假设后端在 backend 目录，前端在根目录的 frontend)
	frontendPath, _ := filepath.Abs("../frontend")

	// 2. 清理缓存文件夹 (Astro 缓存通常在 .astro 和 dist)
	cacheDir := filepath.Join(frontendPath, ".astro")
	distDir := filepath.Join(frontendPath, "dist")
	_ = os.RemoveAll(cacheDir)
	_ = os.RemoveAll(distDir)

	// 3. 跨平台执行 pnpm build
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "pnpm build")
	} else {
		cmd = exec.Command("sh", "-c", "pnpm build")
	}

	cmd.Dir = frontendPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("构建失败: %v\n输出: %s", err, string(output))
	}

	logger.Info("前端重构成功", "output_size", len(output))
	return nil
}
