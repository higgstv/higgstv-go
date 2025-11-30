package repository

import (
	"context"
	"time"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
)

// MongoDBProgramRepository MongoDB 節目 Repository
type MongoDBProgramRepository struct {
	collection database.Collection
	countersColl database.Collection
}

// NewMongoDBProgramRepository 建立 MongoDB 節目 Repository
func NewMongoDBProgramRepository(db database.Database) *MongoDBProgramRepository {
	return &MongoDBProgramRepository{
		collection:   db.Collection("channels"), // Program 內嵌在 Channel 中
		countersColl: db.Collection("counters"),
	}
}

// GetNextProgramID 取得下一個節目 ID（使用 counter collection）
func (r *MongoDBProgramRepository) GetNextProgramID(ctx context.Context) (int, error) {
	// MongoDB 的 $inc 操作需要直接使用 MongoDB driver
	// 這裡簡化處理，實際應該使用 MongoDB 的 FindOneAndUpdate with $inc
	// 暫時返回錯誤，需要在 MongoDB Collection 實作中支援 $inc
	var counter struct {
		ID  string `bson:"_id"`
		Seq int    `bson:"seq"`
	}

	// 嘗試查詢現有計數器
	err := r.countersColl.FindOne(ctx, database.Filter{"_id": "program_id"}, &counter)
	if database.IsNotFound(err) {
		// 如果不存在，初始化為 1
		if err := r.countersColl.InsertOne(ctx, map[string]interface{}{
			"_id": "program_id",
			"seq": 1,
		}); err != nil {
			return 0, err
		}
		return 1, nil
	}

	if err != nil {
		return 0, err
	}

	// 更新計數器（需要手動實作 $inc）
	newSeq := counter.Seq + 1
	if err := r.countersColl.UpdateOne(ctx, database.Filter{"_id": "program_id"}, database.Update{
		Set: map[string]interface{}{"seq": newSeq},
	}); err != nil {
		return 0, err
	}

	return newSeq, nil
}

// AddProgram 新增節目到頻道
func (r *MongoDBProgramRepository) AddProgram(ctx context.Context, channelID string, program *models.Program) error {
	programID, err := r.GetNextProgramID(ctx)
	if err != nil {
		return err
	}
	program.ID = programID
	program.Created = time.Now()
	program.LastModified = time.Now()

	return r.collection.UpdateOne(ctx, database.Filter{"_id": channelID}, database.Update{
		Push: map[string]interface{}{"contents": program},
		Set: map[string]interface{}{
			"last_modified": time.Now(),
		},
	})
}

// UpdateProgram 更新節目
func (r *MongoDBProgramRepository) UpdateProgram(ctx context.Context, channelID string, programID int, update map[string]interface{}) error {
	// 確保 last_modified 被設定
	update["contents.$.last_modified"] = time.Now()
	update["last_modified"] = time.Now()

	return r.collection.UpdateOne(ctx, database.Filter{
		"_id":          channelID,
		"contents._id": programID,
	}, database.Update{
		Set: update,
	})
}

// DeletePrograms 刪除節目
func (r *MongoDBProgramRepository) DeletePrograms(ctx context.Context, channelID string, programIDs []int) error {
	return r.collection.UpdateOne(ctx, database.Filter{"_id": channelID}, database.Update{
		Pull: map[string]interface{}{
			"contents": map[string]interface{}{
				"_id": map[string]interface{}{"$in": programIDs},
			},
		},
		Set: map[string]interface{}{
			"last_modified": time.Now(),
		},
	})
}

// SetOrder 設定節目順序
func (r *MongoDBProgramRepository) SetOrder(ctx context.Context, channelID string, order []int) error {
	return r.collection.UpdateOne(ctx, database.Filter{"_id": channelID}, database.Update{
		Set: map[string]interface{}{
			"contents_order": order,
			"last_modified":  time.Now(),
		},
	})
}

