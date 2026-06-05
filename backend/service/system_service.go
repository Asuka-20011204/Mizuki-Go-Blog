package service

import (
	"context"
	"fmt"
	"my-blog-backend/logger"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
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

	// 3. 跨平台执行 pnpm build（5 分钟超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", "pnpm build")
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", "pnpm build")
	}

	cmd.Dir = frontendPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("构建超时（超过5分钟），请检查前端项目")
		}
		return fmt.Errorf("构建失败")
	}

	logger.Info("前端重构成功", "output_size", len(output))
	return nil
}
