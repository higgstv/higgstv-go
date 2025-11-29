package migration

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/higgstv/higgstv-go/pkg/logger"
)

// Migration 遷移結構
type Migration struct {
	ID          string
	Description string
	Up          func(ctx context.Context, db *mongo.Database) error
	Down        func(ctx context.Context, db *mongo.Database) error
}

var migrations = []Migration{
	{
		ID:          "001_initial_schema",
		Description: "建立初始資料庫結構和索引",
		Up: func(ctx context.Context, db *mongo.Database) error {
			// 這個遷移已經在 database/indexes.go 中處理
			return nil
		},
		Down: func(ctx context.Context, db *mongo.Database) error {
			// 不實作向下遷移
			return nil
		},
	},
}

// RunMigrations 執行所有未執行的遷移
func RunMigrations(ctx context.Context, db *mongo.Database) error {
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
		_, err := migrationsColl.InsertOne(ctx, bson.M{
			"_id":         migration.ID,
			"description": migration.Description,
			"executed_at": time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
		}

		if logger.Logger != nil {
			logger.Logger.Info("Migration completed",
				zap.String("id", migration.ID),
			)
		}
	}

	return nil
}

// getExecutedMigrations 取得已執行的遷移 ID 列表
func getExecutedMigrations(ctx context.Context, coll *mongo.Collection) ([]string, error) {
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// 記錄錯誤但不中斷執行
			_ = err // 忽略錯誤，繼續執行
		}
	}()

	var results []struct {
		ID string `bson:"_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.ID
	}

	return ids, nil
}

// RunMigrationsWithTimeout 執行遷移（帶超時）
func RunMigrationsWithTimeout(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return RunMigrations(ctx, db)
}

