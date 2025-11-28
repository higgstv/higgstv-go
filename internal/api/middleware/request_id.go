package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey Request ID 的 Context Key
const RequestIDKey = "X-Request-ID"

// RequestID 為每個請求產生唯一的 Request ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDKey, requestID)
		c.Next()
	}
}

