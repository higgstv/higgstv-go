package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/pkg/logger"
)

// RateLimiter 簡單的記憶體型 Rate Limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // 每時間窗口的請求數
	window   time.Duration // 時間窗口
	cleanup  *time.Ticker
}

type visitor struct {
	lastSeen time.Time
	count    int
}

// NewRateLimiter 建立 Rate Limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		cleanup:  time.NewTicker(1 * time.Minute),
	}

	// 定期清理過期的訪問記錄
	go func() {
		for range rl.cleanup.C {
			rl.cleanupVisitors()
		}
	}()

	return rl
}

// Allow 檢查是否允許請求
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	now := time.Now()

	if !exists {
		rl.visitors[ip] = &visitor{
			lastSeen: now,
			count:    1,
		}
		return true
	}

	// 如果超過時間窗口，重置計數
	if now.Sub(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = now
		return true
	}

	// 檢查是否超過限制
	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}

// cleanupVisitors 清理過期的訪問記錄
func (rl *RateLimiter) cleanupVisitors() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.window*2 {
			delete(rl.visitors, ip)
		}
	}
}

var defaultRateLimiter = NewRateLimiter(100, 1*time.Minute) // 每分鐘 100 次請求

// RateLimit Rate Limiting 中介層
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !defaultRateLimiter.Allow(ip) {
			if logger.Logger != nil {
				logger.Logger.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.String("path", c.Request.URL.Path),
				)
			}
			response.Error(c, response.ErrorAccessDenied)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitCustom 自訂 Rate Limiting 中介層
func RateLimitCustom(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			if logger.Logger != nil {
				logger.Logger.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.String("path", c.Request.URL.Path),
				)
			}
			response.Error(c, response.ErrorAccessDenied)
			c.Abort()
			return
		}

		c.Next()
	}
}

