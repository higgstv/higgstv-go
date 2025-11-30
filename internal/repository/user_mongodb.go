package repository

import (
	"context"
	"strings"
	"time"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
)

// MongoDBUserRepository MongoDB 使用者 Repository
type MongoDBUserRepository struct {
	collection database.Collection
}

// NewMongoDBUserRepository 建立 MongoDB 使用者 Repository
func NewMongoDBUserRepository(db database.Database) *MongoDBUserRepository {
	return &MongoDBUserRepository{
		collection: db.Collection("users"),
	}
}

// FindByUsername 依使用者名稱查詢
func (r *MongoDBUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, database.Filter{"username": username}, &user)
	if database.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 依 Email 查詢
func (r *MongoDBUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, database.Filter{"email": email}, &user)
	if database.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Exists 檢查使用者是否存在
func (r *MongoDBUserRepository) Exists(ctx context.Context, username, email string) (bool, error) {
	// MongoDB 的 $or 查詢
	filter := database.Filter{
		"$or": []database.Filter{
			{"username": username},
			{"email": email},
		},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	return count > 0, err
}

// Create 建立使用者
func (r *MongoDBUserRepository) Create(ctx context.Context, user *models.User) error {
	user.Created = time.Now()
	user.LastModified = time.Now()

	// 確保 own_channels 和 unclassified_channel 欄位初始化為正確的類型
	if user.OwnChannels == nil {
		user.OwnChannels = []string{}
	}

	return r.collection.InsertOne(ctx, user)
}

// UpdatePassword 更新密碼
func (r *MongoDBUserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	return r.collection.UpdateOne(ctx, database.Filter{"_id": userID}, database.Update{
		Set: map[string]interface{}{
			"password":      hashedPassword,
			"last_modified": time.Now(),
		},
	})
}

// SetAccessKey 設定 access_key
func (r *MongoDBUserRepository) SetAccessKey(ctx context.Context, email, accessKey string) error {
	return r.collection.UpdateOne(ctx, database.Filter{"email": email}, database.Update{
		Set: map[string]interface{}{
			"access_key":    accessKey,
			"last_modified": time.Now(),
		},
	})
}

// ChangePasswordWithAccessKey 使用 access_key 重設密碼
func (r *MongoDBUserRepository) ChangePasswordWithAccessKey(ctx context.Context, email, accessKey, hashedPassword string) (bool, error) {
	// MongoDB 需要先查詢再更新來確認是否有修改
	user, err := r.FindByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if user == nil || user.AccessKey == nil || *user.AccessKey != accessKey {
		return false, nil
	}

	err = r.collection.UpdateOne(ctx, database.Filter{
		"email":      email,
		"access_key": accessKey,
	}, database.Update{
		Set: map[string]interface{}{
			"password":      hashedPassword,
			"access_key":    nil,
			"last_modified": time.Now(),
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddChannel 新增頻道到使用者的 own_channels
func (r *MongoDBUserRepository) AddChannel(ctx context.Context, username, channelID string) error {
	filter := database.Filter{"username": username}

	// 使用 $addToSet 新增頻道 ID
	err := r.collection.UpdateOne(ctx, filter, database.Update{
		AddToSet: map[string]interface{}{
			"own_channels": channelID,
		},
		Set: map[string]interface{}{
			"last_modified": time.Now(),
		},
	})
	if err != nil {
		// 如果錯誤是因為 own_channels 為 null 或非陣列，先初始化再重試
		errStr := err.Error()
		if strings.Contains(errStr, "non-array") || strings.Contains(errStr, "null") {
			// 先初始化 own_channels 為空陣列
			initErr := r.collection.UpdateOne(ctx, filter, database.Update{
				Set: map[string]interface{}{
					"own_channels": []string{},
				},
			})
			if initErr != nil {
				return initErr
			}
			// 重試 $addToSet
			return r.collection.UpdateOne(ctx, filter, database.Update{
				AddToSet: map[string]interface{}{
					"own_channels": channelID,
				},
				Set: map[string]interface{}{
					"last_modified": time.Now(),
				},
			})
		}
		return err
	}
	return nil
}

// SetUnclassifiedChannel 設定未分類頻道
func (r *MongoDBUserRepository) SetUnclassifiedChannel(ctx context.Context, username, channelID string) error {
	return r.collection.UpdateOne(ctx, database.Filter{"username": username}, database.Update{
		Set: map[string]interface{}{
			"unclassified_channel": channelID,
			"last_modified":         time.Now(),
		},
	})
}

// GetUsersBasicInfo 取得使用者基本資訊（用於 owners_info）
func (r *MongoDBUserRepository) GetUsersBasicInfo(ctx context.Context, userIDs []string) ([]models.UserBasicInfo, error) {
	filter := database.Filter{
		"_id": database.Filter{"$in": userIDs},
	}

	var users []models.User
	err := r.collection.Find(ctx, filter, nil, 0, 0, &users)
	if err != nil {
		return nil, err
	}

	result := make([]models.UserBasicInfo, len(users))
	for i, user := range users {
		result[i] = models.UserBasicInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
	}
	return result, nil
}

