package migration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/pkg/logger"
)

// Migration 遷移結構
type Migration struct {
	ID          string
	Description string
	Up          func(ctx context.Context, db database.Database) error
	Down        func(ctx context.Context, db database.Database) error
}

var migrations = []Migration{
	{
		ID:          "001_initial_schema",
		Description: "建立初始資料庫結構和索引",
		Up: func(ctx context.Context, db database.Database) error {
			// 這個遷移已經在 database/indexes_unified.go 中處理
			return nil
		},
		Down: func(ctx context.Context, db database.Database) error {
			// 不實作向下遷移
			return nil
		},
	},
}

// RunMigrations 執行所有未執行的遷移
func RunMigrations(ctx context.Context, db database.Database) error {
	migrationsColl := db.Collection("migrations")

	// 取得已執行的遷移
	executed, err := getExecutedMigrations(ctx, migrationsColl)
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	executedMap := make(map[string]bool)
	for _, id := range executed {
		executedMap[id] = true
	}

	// 執行未執行的遷移
	for _, migration := range migrations {
		if executedMap[migration.ID] {
			if logger.Logger != nil {
				logger.Logger.Info("Migration already executed",
					zap.String("id", migration.ID),
					zap.String("description", migration.Description),
				)
			}
			continue
		}

		if logger.Logger != nil {
			logger.Logger.Info("Running migration",
				zap.String("id", migration.ID),
				zap.String("description", migration.Description),
			)
		}

		// 執行遷移
		if err := migration.Up(ctx, db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.ID, err)
		}

		// 記錄遷移
		migrationDoc := map[string]interface{}{
			"_id":         migration.ID,
			"description": migration.Description,
			"executed_at": time.Now(),
		}
		if err := migrationsColl.InsertOne(ctx, migrationDoc); err != nil {
			// 如果是重複鍵錯誤，忽略（可能是在測試環境中重複執行）
			if !isDuplicateMigrationError(err) {
				return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
			}
			// 記錄已存在的遷移，繼續執行
			if logger.Logger != nil {
				logger.Logger.Info("Migration record already exists",
					zap.String("id", migration.ID),
				)
			}
		}

		if logger.Logger != nil {
			logger.Logger.Info("Migration completed",
				zap.String("id", migration.ID),
			)
		}
	}

	return nil
}

// isDuplicateMigrationError 檢查是否為重複遷移記錄錯誤
func isDuplicateMigrationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// MongoDB E11000 錯誤
	if strings.Contains(errStr, "E11000") || strings.Contains(errStr, "duplicate key") {
		return true
	}
	// SQLite UNIQUE constraint 錯誤
	if strings.Contains(errStr, "UNIQUE constraint") {
		return true
	}
	return false
}

// getExecutedMigrations 取得已執行的遷移 ID 列表
func getExecutedMigrations(ctx context.Context, coll database.Collection) ([]string, error) {
	var results []struct {
		ID string
	}
	
	err := coll.Find(ctx, database.Filter{}, nil, 0, 0, &results)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.ID
	}

	return ids, nil
}

// RunMigrationsWithTimeout 執行遷移（帶超時）
func RunMigrationsWithTimeout(db database.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return RunMigrations(ctx, db)
}

