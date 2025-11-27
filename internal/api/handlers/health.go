package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/higgstv/higgstv-go/internal/api/response"
)

// HealthCheck 健康檢查端點
// @Summary      健康檢查
// @Description  檢查伺服器和資料庫連線狀態
// @Tags         系統
// @Produce      json
// @Success      200 {object} map[string]interface{} "服務正常"
// @Router       /health [get]
func HealthCheck(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 檢查資料庫連線
		err := db.Client().Ping(ctx, nil)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{
			"status":  "ok",
			"service": "higgstv-api",
			"database": "connected",
		})
	}
}

// ReadinessCheck 就緒檢查端點（更詳細的檢查）
// @Summary      就緒檢查
// @Description  檢查服務是否準備好接受請求（包含資料庫查詢測試）
// @Tags         系統
// @Produce      json
// @Success      200 {object} map[string]interface{} "服務就緒"
// @Router       /ready [get]
func ReadinessCheck(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 檢查資料庫連線
		err := db.Client().Ping(ctx, nil)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		// 檢查資料庫是否可以執行簡單查詢（使用簡單的 count 操作）
		usersColl := db.Collection("users")
		_, err = usersColl.CountDocuments(ctx, bson.M{})
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

