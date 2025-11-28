package repository

import (
	"context"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/models"
)

// ChannelRepository 頻道資料存取層
type ChannelRepository struct {
	collection *mongo.Collection
}

// NewChannelRepository 建立頻道 Repository
func NewChannelRepository(db *mongo.Database) *ChannelRepository {
	return &ChannelRepository{
		collection: db.Collection("channels"),
	}
}

// FindByID 依 ID 查詢頻道（支援 UUID binary 和字串兩種格式）
func (r *ChannelRepository) FindByID(ctx context.Context, id string) (*models.Channel, error) {
	var channel models.Channel
	
	// 先嘗試用字串查詢（新建立的頻道使用字串格式）
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&channel)
	if err == nil {
		return &channel, nil
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}
	
	// 如果字串查詢失敗，嘗試將 Base64 UUID 轉換為 binary 格式查詢
	// 這用於查詢舊資料庫中 UUID binary 格式的頻道
	if data, err := base64.URLEncoding.DecodeString(id); err == nil && len(data) == 16 {
		binaryID := primitive.Binary{
			Subtype: 4, // UUID subtype
			Data:    data,
		}
		err = r.collection.FindOne(ctx, bson.M{"_id": binaryID}).Decode(&channel)
		if err == nil {
			return &channel, nil
		}
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
	}
	
	// 兩種格式都找不到，回傳 nil
	return nil, nil
}

// Create 建立頻道
func (r *ChannelRepository) Create(ctx context.Context, channel *models.Channel) error {
	channel.Created = time.Now()
	channel.LastModified = time.Now()
	_, err := r.collection.InsertOne(ctx, channel)
	return err
}

// Update 更新頻道
func (r *ChannelRepository) Update(ctx context.Context, id string, update bson.M) error {
	update["last_modified"] = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	return err
}

// ListChannels 列出頻道（支援過濾和排序）
func (r *ChannelRepository) ListChannels(ctx context.Context, filter bson.M, sort bson.D, limit, skip int64) ([]models.Channel, error) {
	opts := options.Find()
	if sort != nil {
		opts.SetSort(sort)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// 記錄錯誤但不中斷執行
		}
	}()

	var channels []models.Channel
	if err := cursor.All(ctx, &channels); err != nil {
		return nil, err
	}
	return channels, nil
}

// IsAdmin 檢查使用者是否為頻道管理員
func (r *ChannelRepository) IsAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	var channel models.Channel
	err := r.collection.FindOne(ctx, bson.M{
		"_id": channelID,
		"$or": []bson.M{
			{"owners": userID},
			{"permission.user_id": userID, "permission.admin": true},
		},
	}).Decode(&channel)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddOwners 新增擁有者
func (r *ChannelRepository) AddOwners(ctx context.Context, channelID string, userIDs []string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": channelID},
		bson.M{
			"$addToSet": bson.M{
				"owners": bson.M{"$each": userIDs},
			},
			"$set": bson.M{"last_modified": time.Now()},
		},
	)
	return err
}

