package database

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// isIndexExistsError 檢查是否為索引已存在的錯誤
func isIndexExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "IndexKeySpecsConflict") ||
		strings.Contains(errStr, "already exists") ||
		strings.Contains(errStr, "duplicate key")
}

// EnsureIndexes 確保所有必要的索引已建立
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	// Users collection 索引
	usersColl := db.Collection("users")

	// 檢查並建立 username 索引
	usernameIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"username": 1,
		},
		Options: options.Index().SetUnique(true).SetName("username_1"),
	}
	_, err := usersColl.Indexes().CreateOne(ctx, usernameIndex)
	if err != nil && !isIndexExistsError(err) {
		return err
	}

	// 檢查並建立 email 索引
	emailIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"email": 1,
		},
		Options: options.Index().SetUnique(true).SetName("email_1"),
	}
	_, err = usersColl.Indexes().CreateOne(ctx, emailIndex)
	if err != nil && !isIndexExistsError(err) {
		return err
	}

	// 檢查並建立 access_key 索引
	accessKeyIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"access_key": 1,
		},
		Options: options.Index().SetSparse(true).SetName("access_key_1"),
	}
	_, err = usersColl.Indexes().CreateOne(ctx, accessKeyIndex)
	if err != nil && !isIndexExistsError(err) {
		return err
	}

	// Channels collection 索引
	channelsColl := db.Collection("channels")

	indexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"owners": 1},
			Options: options.Index().SetName("owners_1"),
		},
		{
			Keys:    map[string]interface{}{"last_modified": -1},
			Options: options.Index().SetName("last_modified_-1"),
		},
		{
			Keys:    map[string]interface{}{"name": 1},
			Options: options.Index().SetName("name_1"),
		},
		{
			Keys:    map[string]interface{}{"tags": 1},
			Options: options.Index().SetName("tags_1"),
		},
		{
			Keys:    map[string]interface{}{"contents._id": 1},
			Options: options.Index().SetName("contents._id_1"),
		},
	}

	for _, index := range indexes {
		_, err := channelsColl.Indexes().CreateOne(ctx, index)
		if err != nil && !isIndexExistsError(err) {
			return err
		}
	}

	// Counters collection 索引（用於 Program ID）
	// 注意：_id 欄位預設就是唯一的，不需要額外建立 unique 索引
	// _id 索引會自動建立，不需要手動建立

	return nil
}

// EnsureIndexesWithTimeout 確保索引（帶超時）
func EnsureIndexesWithTimeout(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return EnsureIndexes(ctx, db)
}
