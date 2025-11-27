package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogging 請求日誌中介層
func RequestLogging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		requestID := c.GetString(RequestIDKey)
		logger.Info("HTTP Request",
			zap.String("request_id", requestID),
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Duration("latency", latency),
		)

		// 記錄錯誤
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request Error",
					zap.String("path", path),
					zap.String("method", method),
					zap.Error(err),
				)
			}
		}
	}
}

