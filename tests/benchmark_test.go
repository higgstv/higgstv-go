package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/api"
	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/pkg/session"
)

var benchmarkDB *mongo.Database
var benchmarkRouter *gin.Engine

func initBenchmark() {
	cfg, _ := config.Load()
	client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.Database.URI))
	benchmarkDB = client.Database(cfg.Database.Database + "_benchmark")
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

