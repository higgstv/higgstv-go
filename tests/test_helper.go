package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/higgstv/higgstv-go/internal/api"
	"github.com/higgstv/higgstv-go/internal/api/handlers"
	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/migration"
	"github.com/higgstv/higgstv-go/pkg/logger"
	"github.com/higgstv/higgstv-go/pkg/session"
)

// TestDBContext 測試資料庫上下文（每個測試獨立）
type TestDBContext struct {
	DB     database.Database
	Router *gin.Engine
}

// sanitizeTestName 清理測試名稱（移除不適合作為資料庫名稱的字元）
func sanitizeTestName(name string) string {
	// 移除測試套件前綴和斜線
	name = strings.TrimPrefix(name, "Test")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	
	// 限制長度（避免資料庫名稱過長）
	if len(name) > 50 {
		name = name[:50]
	}
	
	return name
}

// SetupTestDB 設定測試資料庫（每個測試使用獨立的資料庫連線）
// 返回 TestDBContext，包含獨立的資料庫連線和路由器
func SetupTestDB(t *testing.T) *TestDBContext {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 解析資料庫類型
	dbType, err := database.ParseDatabaseType(cfg.Database.Type)
	if err != nil {
		t.Fatalf("Invalid database type: %v", err)
	}

	// 建立測試資料庫連線（每個測試使用獨立的連線）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 為每個測試建立獨立的資料庫名稱（使用測試名稱，清理特殊字元）
	testName := sanitizeTestName(t.Name())
	testDBName := cfg.Database.Database + "_test_" + testName
	testURI := cfg.Database.URI
	
	if dbType == database.DatabaseTypeSQLite {
		// SQLite 測試使用獨立的記憶體資料庫（每個測試一個）
		// 使用測試名稱確保每個測試有獨立的資料庫
		testURI = "file::memory:?cache=private"
		testDBName = "test_" + testName
	}

	db, err := database.NewDatabase(ctx, database.DatabaseConfig{
		Type:     dbType,
		URI:      testURI,
		Database: testDBName,
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 清理資料庫（確保測試開始時是乾淨的）
	if err := cleanupDatabase(ctx, db); err != nil {
		t.Logf("Warning: Failed to cleanup database: %v", err)
		// 不中斷測試，但記錄警告
	}

	session.Init(cfg.Session.Secret)

	// 初始化 Logger（與 main.go 保持一致）
	if err := logger.Init(cfg.Server.Env); err != nil {
		t.Logf("Warning: Failed to initialize logger: %v", err)
		// 不中斷測試，但記錄警告
	}

	// 確保資料庫索引已建立（與 main.go 保持一致）
	if err := database.EnsureIndexesWithTimeout(db); err != nil {
		t.Logf("Warning: Failed to ensure database indexes: %v", err)
		// 不中斷測試，但記錄警告
	}

	// 執行資料庫遷移（與 main.go 保持一致）
	if err := migration.RunMigrationsWithTimeout(db); err != nil {
		t.Logf("Warning: Failed to run migrations: %v", err)
		// 不中斷測試，但記錄警告
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 註冊健康檢查端點（與 main.go 保持一致）
	router.GET("/health", handlers.HealthCheck(db))
	router.GET("/ready", handlers.ReadinessCheck(db))

	// 設定 API 路由
	api.SetupRoutes(router, db, cfg)

	return &TestDBContext{
		DB:     db,
		Router: router,
	}
}

// CleanupTestDB 清理測試資料庫
func CleanupTestDB(t *testing.T, ctx *TestDBContext) {
	if ctx != nil && ctx.DB != nil {
		// 清理資料庫（確保測試結束時清理）
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cleanupDatabase(cleanupCtx, ctx.DB)
		
		// 關閉資料庫連線
		_ = ctx.DB.Close(context.Background())
	}
}

// cleanupDatabase 清理資料庫（刪除所有資料）
func cleanupDatabase(ctx context.Context, db database.Database) error {
	switch db.Type() {
	case database.DatabaseTypeMongoDB:
		return cleanupMongoDB(ctx, db)
	case database.DatabaseTypeSQLite:
		return cleanupSQLite(ctx, db)
	default:
		return nil
	}
}

// cleanupMongoDB 清理 MongoDB 資料庫
func cleanupMongoDB(ctx context.Context, db database.Database) error {
	// 取得 MongoDB 資料庫實例
	mongoDB := db.(*database.MongoDBDatabase)
	mongoDatabase := mongoDB.GetDatabase()
	
	// 列出所有集合
	collections, err := mongoDatabase.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}
	
	// 刪除所有集合的資料
	for _, collName := range collections {
		coll := mongoDatabase.Collection(collName)
		// 使用 DeleteMany 刪除所有文件
		_, err := coll.DeleteMany(ctx, bson.M{})
		if err != nil {
			// 記錄錯誤但不中斷
			_ = err
		}
	}
	
	return nil
}

// cleanupSQLite 清理 SQLite 資料庫
func cleanupSQLite(ctx context.Context, db database.Database) error {
	// SQLite 清理：刪除所有表的資料（保留結構）
	// 由於使用記憶體資料庫，關閉連線即可，但為了確保，我們還是清理一下
	
	// 取得 SQLite 資料庫連線
	sqliteDB := db.(*database.SQLiteDatabase)
	sqlDB := sqliteDB.GetDB()
	
	// 禁用外鍵約束（暫時）
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA foreign_keys = OFF")
	
	// 刪除所有表的資料
	tables := []string{
		"users", "channels", "programs", "counters", "migrations",
		"user_channels", "channel_tags", "channel_owners", "channel_permissions",
		"program_tags", "channel_program_order",
	}
	
	for _, table := range tables {
		_, _ = sqlDB.ExecContext(ctx, "DELETE FROM "+table)
	}
	
	// 重新啟用外鍵約束
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	
	return nil
}

// getAuthCookie 輔助函數：註冊並登入，返回 Cookie（使用 TestDBContext）
func getAuthCookie(t *testing.T, ctx *TestDBContext, username, email, password string) string {
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
	ctx.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// 取得 Cookie
	cookies := w.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookies)
	return cookies
}
