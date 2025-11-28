package middleware

import (
	"github.com/gin-gonic/gin"
)

// APIVersion API 版本號
const APIVersion = "v1"

// APIVersionHeader API 版本 Header 名稱
const APIVersionHeader = "X-API-Version"

// APIVersionMiddleware API 版本中介層
func APIVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 設定 API 版本到回應 Header
		c.Header(APIVersionHeader, APIVersion)
		c.Next()
	}
}

// GetAPIVersion 取得 API 版本
func GetAPIVersion(c *gin.Context) string {
	version := c.GetHeader("X-API-Version")
	if version == "" {
		return APIVersion
	}
	return version
}

