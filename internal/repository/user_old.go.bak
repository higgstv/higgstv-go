package repository

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/models"
)

// UserRepository 使用者資料存取層
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository 建立使用者 Repository
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// FindByUsername 依使用者名稱查詢
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 依 Email 查詢
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Exists 檢查使用者是否存在
func (r *UserRepository) Exists(ctx context.Context, username, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"$or": []bson.M{
			{"username": username},
			{"email": email},
		},
	})
	return count > 0, err
}

// Create 建立使用者
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.Created = time.Now()
	user.LastModified = time.Now()
	
	// 確保 own_channels 和 unclassified_channel 欄位初始化為正確的類型
	if user.OwnChannels == nil {
		user.OwnChannels = []string{}
	}
	
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// UpdatePassword 更新密碼
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"password":      hashedPassword,
				"last_modified": time.Now(),
			},
		},
	)
	return err
}

// SetAccessKey 設定 access_key
func (r *UserRepository) SetAccessKey(ctx context.Context, email, accessKey string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{
			"$set": bson.M{
				"access_key":    accessKey,
				"last_modified": time.Now(),
			},
		},
	)
	return err
}

// ChangePasswordWithAccessKey 使用 access_key 重設密碼
func (r *UserRepository) ChangePasswordWithAccessKey(ctx context.Context, email, accessKey, hashedPassword string) (bool, error) {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"email":      email,
			"access_key": accessKey,
		},
		bson.M{
			"$set": bson.M{
				"password":      hashedPassword,
				"access_key":    nil,
				"last_modified": time.Now(),
			},
		},
	)
	if err != nil {
		return false, err
	}
	return result.ModifiedCount > 0, nil
}

// AddChannel 新增頻道到使用者的 own_channels
func (r *UserRepository) AddChannel(ctx context.Context, username, channelID string) error {
	filter := bson.M{"username": username}
	
	// 先嘗試使用 $addToSet 新增頻道 ID
	result, err := r.collection.UpdateOne(
		ctx,
		filter,
		bson.M{
			"$addToSet": bson.M{"own_channels": channelID},
			"$set":      bson.M{"last_modified": time.Now()},
		},
	)
	if err != nil {
		// 如果錯誤是因為 own_channels 為 null 或非陣列，先初始化再重試
		errStr := err.Error()
		if strings.Contains(errStr, "non-array") || strings.Contains(errStr, "null") {
			// 先初始化 own_channels 為空陣列
			_, initErr := r.collection.UpdateOne(
				ctx,
				filter,
				bson.M{"$set": bson.M{"own_channels": []string{}}},
			)
			if initErr != nil {
				return initErr
			}
			// 重試 $addToSet
			result, err = r.collection.UpdateOne(
				ctx,
				filter,
				bson.M{
					"$addToSet": bson.M{"own_channels": channelID},
					"$set":      bson.M{"last_modified": time.Now()},
				},
			)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// SetUnclassifiedChannel 設定未分類頻道
func (r *UserRepository) SetUnclassifiedChannel(ctx context.Context, username, channelID string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"username": username},
		bson.M{
			"$set": bson.M{
				"unclassified_channel": channelID,
				"last_modified":         time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// GetUsersBasicInfo 取得使用者基本資訊（用於 owners_info）
func (r *UserRepository) GetUsersBasicInfo(ctx context.Context, userIDs []string) ([]models.UserBasicInfo, error) {
	opts := options.Find().SetProjection(bson.M{
		"username": 1,
		"email":    1,
	})
	cursor, err := r.collection.Find(ctx, bson.M{
		"_id": bson.M{"$in": userIDs},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// 記錄錯誤但不中斷執行
			_ = err // 忽略錯誤，繼續執行
		}
	}()

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
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

