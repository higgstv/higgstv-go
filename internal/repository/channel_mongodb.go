package repository

import (
	"context"
	"time"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
)

// MongoDBChannelRepository MongoDB 頻道 Repository
type MongoDBChannelRepository struct {
	collection database.Collection
}

// NewMongoDBChannelRepository 建立 MongoDB 頻道 Repository
func NewMongoDBChannelRepository(db database.Database) *MongoDBChannelRepository {
	return &MongoDBChannelRepository{
		collection: db.Collection("channels"),
	}
}

// FindByID 依 ID 查詢頻道
func (r *MongoDBChannelRepository) FindByID(ctx context.Context, id string) (*models.Channel, error) {
	var channel models.Channel
	err := r.collection.FindOne(ctx, database.Filter{"_id": id}, &channel)
	if database.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// Create 建立頻道
func (r *MongoDBChannelRepository) Create(ctx context.Context, channel *models.Channel) error {
	channel.Created = time.Now()
	channel.LastModified = time.Now()
	return r.collection.InsertOne(ctx, channel)
}

// Update 更新頻道
func (r *MongoDBChannelRepository) Update(ctx context.Context, id string, update map[string]interface{}) error {
	update["last_modified"] = time.Now()
	return r.collection.UpdateOne(ctx, database.Filter{"_id": id}, database.Update{
		Set: update,
	})
}

// ListChannels 列出頻道（支援過濾和排序）
func (r *MongoDBChannelRepository) ListChannels(ctx context.Context, filter database.Filter, sort database.Sort, limit, skip int64) ([]models.Channel, error) {
	var channels []models.Channel
	err := r.collection.Find(ctx, filter, sort, limit, skip, &channels)
	return channels, err
}

// IsAdmin 檢查使用者是否為頻道管理員
func (r *MongoDBChannelRepository) IsAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	var channel models.Channel
	err := r.collection.FindOne(ctx, database.Filter{
		"_id": channelID,
		"$or": []database.Filter{
			{"owners": userID},
			{"permission.user_id": userID, "permission.admin": true},
		},
	}, &channel)
	if database.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddOwners 新增擁有者
func (r *MongoDBChannelRepository) AddOwners(ctx context.Context, channelID string, userIDs []string) error {
	return r.collection.UpdateOne(ctx, database.Filter{"_id": channelID}, database.Update{
		AddToSet: map[string]interface{}{
			"owners": map[string]interface{}{"$each": userIDs},
		},
		Set: map[string]interface{}{
			"last_modified": time.Now(),
		},
	})
}

