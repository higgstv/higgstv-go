package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/repository"
)

// setupTestDB 建立測試用資料庫連線
func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	// 使用測試資料庫
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	db := client.Database("higgstv_test")

	cleanup := func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	}

	return db, cleanup
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

