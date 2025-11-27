package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/pkg/errors"
	"github.com/higgstv/higgstv-go/pkg/logger"
)

// ErrorHandler 錯誤處理中介層
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 處理錯誤
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// 記錄錯誤
			if logger.Logger != nil {
				logger.Logger.Error("Request Error",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Error(err),
				)
			}

			// 處理應用程式錯誤
			if appErr, ok := err.Err.(*errors.AppError); ok {
				response.Error(c, appErr.Code)
				return
			}

			// 處理驗證錯誤
			if err.Type == gin.ErrorTypeBind {
				response.Error(c, response.ErrorRequiredField)
				return
			}

			// 預設錯誤
			response.Error(c, response.ErrorServerError)
		}
	}
}

// NotFoundHandler 404 處理器
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, response.ErrorAccessDenied)
	}
}

