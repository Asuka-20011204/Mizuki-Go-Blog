package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	count    int
	firstHit time.Time
}

// RateLimit 返回一个基于 IP 的简单限流中间件
// limit: 时间窗口内最大请求数
// window: 时间窗口长度
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string]*rateEntry)

	// 后台定期清理过期记录
	go func() {
		for {
			time.Sleep(window)
			mu.Lock()
			now := time.Now()
			for ip, entry := range visitors {
				if now.Sub(entry.firstHit) > window {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		entry, exists := visitors[ip]
		now := time.Now()

		if !exists || now.Sub(entry.firstHit) > window {
			// 新窗口，重置计数
			visitors[ip] = &rateEntry{count: 1, firstHit: now}
			mu.Unlock()
			c.Next()
			return
		}

		entry.count++
		if entry.count > limit {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}
