package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/api"
	"github.com/higgstv/higgstv-go/internal/api/handlers"
	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/migration"
	"github.com/higgstv/higgstv-go/pkg/logger"
	"github.com/higgstv/higgstv-go/pkg/session"
)

var testDB *mongo.Database
var testRouter *gin.Engine
var testConfig *config.Config

// SetupTestDB 設定測試資料庫
func SetupTestDB(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	testConfig = cfg

	client, err := mongo.Connect(nil, options.Client().ApplyURI(cfg.Database.URI))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	testDB = client.Database(cfg.Database.Database + "_test")
	session.Init(cfg.Session.Secret)

	// 初始化 Logger（與 main.go 保持一致）
	if err := logger.Init(cfg.Server.Env); err != nil {
		t.Logf("Warning: Failed to initialize logger: %v", err)
		// 不中斷測試，但記錄警告
	}

	// 確保資料庫索引已建立（與 main.go 保持一致）
	if err := database.EnsureIndexesWithTimeout(testDB); err != nil {
		t.Logf("Warning: Failed to ensure database indexes: %v", err)
		// 不中斷測試，但記錄警告
	}

	// 執行資料庫遷移（與 main.go 保持一致）
	if err := migration.RunMigrationsWithTimeout(testDB); err != nil {
		t.Logf("Warning: Failed to run migrations: %v", err)
		// 不中斷測試，但記錄警告
	}

	gin.SetMode(gin.TestMode)
	testRouter = gin.New()

	// 註冊健康檢查端點（與 main.go 保持一致）
	testRouter.GET("/health", handlers.HealthCheck(testDB))
	testRouter.GET("/ready", handlers.ReadinessCheck(testDB))

	// 設定 API 路由
	api.SetupRoutes(testRouter, testDB, cfg)
}

// CleanupTestDB 清理測試資料庫
func CleanupTestDB(t *testing.T) {
	if testDB != nil {
		_ = testDB.Drop(nil)
	}
}

// getAuthCookie 輔助函數：註冊並登入，返回 Cookie
func getAuthCookie(t *testing.T, router *gin.Engine, username, email, password string) string {
	// 註冊
	signupPayload := map[string]interface{}{
		"invitation_code": "sixpens",
		"username":        username,
		"email":           email,
		"password":        password,
	}
	jsonData, _ := json.Marshal(signupPayload)
	req, _ := http.NewRequest("POST", "/apis/signup", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// 取得 Cookie
	cookies := w.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookies)
	return cookies
}

