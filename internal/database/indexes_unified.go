package database

import (
	"context"
	"time"
)

// EnsureIndexes 確保所有必要的索引已建立（統一介面）
func EnsureIndexes(ctx context.Context, db Database) error {
	// Users collection 索引
	usersColl := db.Collection("users")

	// username 索引
	if err := usersColl.CreateIndex(ctx, map[string]interface{}{
		"username": 1,
	}, IndexOptions{
		Unique: true,
		Name:   "username_1",
	}); err != nil {
		// 忽略索引已存在的錯誤
		_ = err
	}

	// email 索引
	if err := usersColl.CreateIndex(ctx, map[string]interface{}{
		"email": 1,
	}, IndexOptions{
		Unique: true,
		Name:   "email_1",
	}); err != nil {
		_ = err
	}

	// access_key 索引（sparse，僅 MongoDB）
	if db.Type() == DatabaseTypeMongoDB {
		if err := usersColl.CreateIndex(ctx, map[string]interface{}{
			"access_key": 1,
		}, IndexOptions{
			Sparse: true,
			Name:   "access_key_1",
		}); err != nil {
			_ = err
		}
	}

	// Channels collection 索引
	channelsColl := db.Collection("channels")

	// owners 索引（MongoDB 使用陣列欄位，SQLite 使用關聯表）
	if db.Type() == DatabaseTypeMongoDB {
		if err := channelsColl.CreateIndex(ctx, map[string]interface{}{
			"owners": 1,
		}, IndexOptions{
			Name: "owners_1",
		}); err != nil {
			_ = err
		}
	}

	// last_modified 索引
	if err := channelsColl.CreateIndex(ctx, map[string]interface{}{
		"last_modified": -1,
	}, IndexOptions{
		Name: "last_modified_-1",
	}); err != nil {
		_ = err
	}

	// name 索引
	if err := channelsColl.CreateIndex(ctx, map[string]interface{}{
		"name": 1,
	}, IndexOptions{
		Name: "name_1",
	}); err != nil {
		_ = err
	}

	// tags 索引（MongoDB 使用陣列欄位，SQLite 使用關聯表）
	if db.Type() == DatabaseTypeMongoDB {
		if err := channelsColl.CreateIndex(ctx, map[string]interface{}{
			"tags": 1,
		}, IndexOptions{
			Name: "tags_1",
		}); err != nil {
			_ = err
		}
	}

	// contents._id 索引（僅 MongoDB）
	if db.Type() == DatabaseTypeMongoDB {
		if err := channelsColl.CreateIndex(ctx, map[string]interface{}{
			"contents._id": 1,
		}, IndexOptions{
			Name: "contents._id_1",
		}); err != nil {
			_ = err
		}
	}

	return nil
}

// EnsureIndexesWithTimeout 確保索引（帶超時）
func EnsureIndexesWithTimeout(db Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return EnsureIndexes(ctx, db)
}

