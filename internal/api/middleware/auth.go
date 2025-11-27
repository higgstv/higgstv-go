package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/higgstv/higgstv-go/pkg/session"
)

// RequireAuth 需要登入的中介層
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !session.IsLoggedIn(c) {
			// 根據原 API 行為，未登入時直接結束（無回應）
			c.Abort()
			return
		}
		c.Next()
	}
}

