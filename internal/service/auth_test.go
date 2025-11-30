package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/repository"
)

// setupTestDB 建立測試用資料庫連線（支援 MongoDB 和 SQLite）
// 每個測試使用獨立的資料庫連線，確保測試隔離
func setupTestDB(t *testing.T) (database.Database, func()) {
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

	// 為每個測試建立獨立的資料庫名稱（使用測試名稱）
	testName := sanitizeTestName(t.Name())
	testDBName := cfg.Database.Database + "_test_" + testName
	testURI := cfg.Database.URI
	
	if dbType == database.DatabaseTypeSQLite {
		// SQLite 測試使用獨立的記憶體資料庫（每個測試一個）
		testURI = "file::memory:?cache=private"
		testDBName = "test_" + testName
	}

	db, err := database.NewDatabase(ctx, database.DatabaseConfig{
		Type:     dbType,
		URI:      testURI,
		Database: testDBName,
	})
	require.NoError(t, err)

	// 清理資料庫（確保測試開始時是乾淨的）
	_ = cleanupDatabase(ctx, db)

	// 執行遷移和索引建立
	_ = database.EnsureIndexesWithTimeout(db)
	// 注意：這裡不執行 migration，因為 service 層測試不需要完整的遷移

	cleanup := func() {
		// 清理資料庫
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cleanupDatabase(cleanupCtx, db)
		
		// 關閉資料庫連線
		_ = db.Close(context.Background())
	}

	return db, cleanup
}

// sanitizeTestName 清理測試名稱（移除不適合作為資料庫名稱的字元）
func sanitizeTestName(name string) string {
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
	mongoDB := db.(*database.MongoDBDatabase)
	mongoDatabase := mongoDB.GetDatabase()
	
	collections, err := mongoDatabase.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return err
	}
	
	for _, collName := range collections {
		coll := mongoDatabase.Collection(collName)
		_, _ = coll.DeleteMany(ctx, bson.M{})
	}
	
	return nil
}

// cleanupSQLite 清理 SQLite 資料庫
func cleanupSQLite(ctx context.Context, db database.Database) error {
	sqliteDB := db.(*database.SQLiteDatabase)
	sqlDB := sqliteDB.GetDB()
	
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA foreign_keys = OFF")
	
	tables := []string{
		"users", "channels", "programs", "counters", "migrations",
		"user_channels", "channel_tags", "channel_owners", "channel_permissions",
		"program_tags", "channel_program_order",
	}
	
	for _, table := range tables {
		_, _ = sqlDB.ExecContext(ctx, "DELETE FROM "+table)
	}
	
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	
	return nil
}

func TestAuthService_SignUp(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)

	ctx := context.Background()

	t.Run("成功註冊", func(t *testing.T) {
		user, err := authService.SignUp(ctx, "sixpens", "testuser", "test@example.com", "password123")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("邀請碼錯誤", func(t *testing.T) {
		_, err := authService.SignUp(ctx, "wrong-code", "testuser2", "test2@example.com", "password123")
		assert.Error(t, err)
		assert.Equal(t, "invalid invitation code", err.Error())
	})

	t.Run("使用者已存在", func(t *testing.T) {
		// 先建立一個使用者
		_, err := authService.SignUp(ctx, "sixpens", "existinguser", "existing@example.com", "password123")
		require.NoError(t, err)

		// 嘗試用相同使用者名稱註冊
		_, err = authService.SignUp(ctx, "sixpens", "existinguser", "another@example.com", "password123")
		assert.Error(t, err)
		assert.Equal(t, "user already exists", err.Error())
	})
}

func TestAuthService_SignIn(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo)

	ctx := context.Background()

	// 先註冊一個使用者
	_, err := authService.SignUp(ctx, "sixpens", "testuser", "test@example.com", "password123")
	require.NoError(t, err)

	t.Run("成功登入", func(t *testing.T) {
		user, err := authService.SignIn(ctx, "testuser", "password123")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("使用者不存在", func(t *testing.T) {
		_, err := authService.SignIn(ctx, "nonexistent", "password123")
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})

	t.Run("密碼錯誤", func(t *testing.T) {
		_, err := authService.SignIn(ctx, "testuser", "wrongpassword")
		assert.Error(t, err)
		assert.Equal(t, "invalid password", err.Error())
	})
}

