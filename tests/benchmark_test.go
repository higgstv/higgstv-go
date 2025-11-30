package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/higgstv/higgstv-go/internal/api"
	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/pkg/session"
)

var benchmarkDB database.Database
var benchmarkRouter *gin.Engine

func initBenchmark() {
	cfg, _ := config.Load()
	
	// 解析資料庫類型
	dbType, err := database.ParseDatabaseType(cfg.Database.Type)
	if err != nil {
		// 如果解析失敗，預設使用 MongoDB
		dbType = database.DatabaseTypeMongoDB
	}

	// 建立測試資料庫連線
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	benchmarkDBName := cfg.Database.Database + "_benchmark"
	benchmarkURI := cfg.Database.URI
	if dbType == database.DatabaseTypeSQLite {
		// SQLite 測試使用記憶體資料庫
		benchmarkURI = "file::memory:?cache=shared"
	}

	benchmarkDB, err = database.NewDatabase(ctx, database.DatabaseConfig{
		Type:     dbType,
		URI:      benchmarkURI,
		Database: benchmarkDBName,
	})
	if err != nil {
		// 如果連線失敗，benchmark 測試會跳過
		return
	}

	session.Init(cfg.Session.Secret)
	gin.SetMode(gin.TestMode)
	benchmarkRouter = gin.New()
	api.SetupRoutes(benchmarkRouter, benchmarkDB, cfg)
}

// BenchmarkHealthCheck 效能測試：健康檢查
func BenchmarkHealthCheck(b *testing.B) {
	if benchmarkRouter == nil {
		initBenchmark()
	}

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkRouter.ServeHTTP(w, req)
	}
}

// BenchmarkSignIn 效能測試：登入
func BenchmarkSignIn(b *testing.B) {
	if benchmarkRouter == nil {
		initBenchmark()
	}

	payload := map[string]interface{}{
		"username": "benchuser",
		"password": "benchpass",
	}
	jsonData, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/apis/signin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		benchmarkRouter.ServeHTTP(w, req)
	}
}

// BenchmarkGetChannels 效能測試：取得頻道列表
func BenchmarkGetChannels(b *testing.B) {
	if benchmarkRouter == nil {
		initBenchmark()
	}

	req, _ := http.NewRequest("GET", "/apis/getchannels", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		benchmarkRouter.ServeHTTP(w, req)
	}
}

