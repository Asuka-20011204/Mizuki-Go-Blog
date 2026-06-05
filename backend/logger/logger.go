package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

var log *slog.Logger

// Init 初始化日志器，同时输出到文件和控制台
// dir: 日志文件目录，如 "./logs"
// level: 最低日志级别 (debug / info / warn / error)
func Init(dir, level string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 文件名按天分割，同一天追加写入
	filename := fmt.Sprintf("backend-%s.log", time.Now().Format("2006-01-02"))
	filePath := filepath.Join(dir, filename)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, file)

	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	handler := slog.NewTextHandler(multiWriter, opts)
	log = slog.New(handler)
	slog.SetDefault(log) // 让其他代码直接调 slog.Info 也能走同一个 handler

	return nil
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ---------- 便捷函数 ----------

func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	log.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

// Fatal 记录错误日志后退出程序
func Fatal(msg string, args ...any) {
	log.Error(msg, args...)
	os.Exit(1)
}

// ---------- Gin 中间件 ----------

// GinLogger 返回一个 Gin 中间件，以结构化格式记录每个 HTTP 请求
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		attrs := []any{
			"method", method,
			"path", path,
			"status", status,
			"latency", latency.String(),
			"ip", clientIP,
		}
		if query != "" {
			attrs = append(attrs, "query", query)
		}

		switch {
		case len(c.Errors) > 0:
			attrs = append(attrs, "errors", c.Errors.String())
			log.Error("HTTP", attrs...)
		case status >= 500:
			log.Error("HTTP", attrs...)
		case status >= 400:
			log.Warn("HTTP", attrs...)
		default:
			log.Info("HTTP", attrs...)
		}
	}
}
