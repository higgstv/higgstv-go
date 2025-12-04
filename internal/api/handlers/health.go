package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/internal/database"
)

// HealthCheck 健康檢查端點
// @Summary      健康檢查
// @Description  檢查伺服器和資料庫連線狀態（支援 MongoDB 和 SQLite）
// @Tags         系統
// @Produce      json
// @Success      200 {object} map[string]interface{} "服務正常"
// @Router       /health [get]
func HealthCheck(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 檢查資料庫連線
		err := db.Ping(ctx)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{
			"status":   "ok",
			"service":  "higgstv-api",
			"database": "connected",
		})
	}
}

// ReadinessCheck 就緒檢查端點（更詳細的檢查）
// @Summary      就緒檢查
// @Description  檢查服務是否準備好接受請求（包含資料庫查詢測試，支援 MongoDB 和 SQLite）
// @Tags         系統
// @Produce      json
// @Success      200 {object} map[string]interface{} "服務就緒"
// @Router       /ready [get]
func ReadinessCheck(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 檢查資料庫連線
		err := db.Ping(ctx)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		// 檢查資料庫是否可以執行簡單查詢（使用簡單的 count 操作）
		usersColl := db.Collection("users")
		_, err = usersColl.CountDocuments(ctx, database.Filter{})
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{
			"status":   "ready",
			"service":  "higgstv-api",
			"database": "connected",
		})
	}
}

