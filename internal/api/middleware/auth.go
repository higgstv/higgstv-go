package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/pkg/session"
)

// RequireAuth 需要登入的中介層
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !session.IsLoggedIn(c) {
			// 未登入時返回錯誤響應
			response.Error(c, response.ErrorRequireLogin)
			c.Abort()
			return
		}
		c.Next()
	}
}

