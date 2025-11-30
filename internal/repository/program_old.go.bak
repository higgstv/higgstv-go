package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/models"
)

// ProgramRepository 節目資料存取層
type ProgramRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

// NewProgramRepository 建立節目 Repository
func NewProgramRepository(db *mongo.Database) *ProgramRepository {
	return &ProgramRepository{
		collection: db.Collection("channels"), // Program 內嵌在 Channel 中
		db:         db,
	}
}

// GetNextProgramID 取得下一個節目 ID（使用 counter collection）
func (r *ProgramRepository) GetNextProgramID(ctx context.Context) (int, error) {
	counterColl := r.db.Collection("counters")
	result := counterColl.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "program_id"},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	)

	var counter struct {
		ID  string `bson:"_id"`
		Seq int    `bson:"seq"`
	}
	if err := result.Decode(&counter); err != nil {
		// 如果不存在，初始化為 1
		if err == mongo.ErrNoDocuments {
			_, err := counterColl.InsertOne(ctx, bson.M{"_id": "program_id", "seq": 1})
			if err != nil {
				return 0, err
			}
			return 1, nil
		}
		return 0, err
	}
	return counter.Seq, nil
}

// AddProgram 新增節目到頻道
func (r *ProgramRepository) AddProgram(ctx context.Context, channelID string, program *models.Program) error {
	programID, err := r.GetNextProgramID(ctx)
	if err != nil {
		return err
	}
	program.ID = programID
	program.Created = time.Now()
	program.LastModified = time.Now()

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": channelID},
		bson.M{
			"$push": bson.M{"contents": program},
			"$set":  bson.M{"last_modified": time.Now()},
		},
	)
	return err
}

// UpdateProgram 更新節目
func (r *ProgramRepository) UpdateProgram(ctx context.Context, channelID string, programID int, update bson.M) error {
	// 確保 last_modified 被設定
	update["contents.$.last_modified"] = time.Now()
	update["last_modified"] = time.Now()
	
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":          channelID,
			"contents._id": programID,
		},
		bson.M{"$set": update},
	)
	return err
}

// DeletePrograms 刪除節目
func (r *ProgramRepository) DeletePrograms(ctx context.Context, channelID string, programIDs []int) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": channelID},
		bson.M{
			"$pull": bson.M{
				"contents": bson.M{"_id": bson.M{"$in": programIDs}},
			},
			"$set": bson.M{"last_modified": time.Now()},
		},
	)
	return err
}

// SetOrder 設定節目順序
func (r *ProgramRepository) SetOrder(ctx context.Context, channelID string, order []int) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": channelID},
		bson.M{
			"$set": bson.M{
				"contents_order": order,
				"last_modified": time.Now(),
			},
		},
	)
	return err
}

