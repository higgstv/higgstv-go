package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBDatabase MongoDB 資料庫實作
type MongoDBDatabase struct {
	client   *mongo.Client
	database *mongo.Database
}

// GetDatabase 取得底層 MongoDB 資料庫連線（供測試清理使用）
func (d *MongoDBDatabase) GetDatabase() *mongo.Database {
	return d.database
}

// NewMongoDBDatabase 建立 MongoDB 資料庫連線
func NewMongoDBDatabase(ctx context.Context, config DatabaseConfig) (*MongoDBDatabase, error) {
	clientOptions := options.Client().ApplyURI(config.URI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 測試連線
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(config.Database)

	return &MongoDBDatabase{
		client:   client,
		database: db,
	}, nil
}

// Type 回傳資料庫類型
func (d *MongoDBDatabase) Type() DatabaseType {
	return DatabaseTypeMongoDB
}

// Collection 取得集合操作介面
func (d *MongoDBDatabase) Collection(name string) Collection {
	return &MongoDBCollection{
		collection: d.database.Collection(name),
	}
}

// Close 關閉資料庫連線
func (d *MongoDBDatabase) Close(ctx context.Context) error {
	return d.client.Disconnect(ctx)
}

// Ping 測試連線
func (d *MongoDBDatabase) Ping(ctx context.Context) error {
	return d.client.Ping(ctx, nil)
}

// BeginTx 開始交易（MongoDB 4.0+ 支援）
// 注意：MongoDB 交易需要使用 WithTransaction 模式，這裡提供簡化版本
func (d *MongoDBDatabase) BeginTx(ctx context.Context) (Tx, error) {
	// MongoDB 交易通常使用 WithTransaction 模式，而不是手動管理
	// 這裡返回一個 no-op 交易，實際交易應該在操作層面處理
	return &MongoDBTx{}, nil
}

// MongoDBTx MongoDB 交易實作（簡化版本）
type MongoDBTx struct{}

// Commit 提交交易（no-op，實際交易在 WithTransaction 中處理）
func (t *MongoDBTx) Commit(ctx context.Context) error {
	return nil
}

// Rollback 回滾交易（no-op，實際交易在 WithTransaction 中處理）
func (t *MongoDBTx) Rollback(ctx context.Context) error {
	return nil
}

// MongoDBCollection MongoDB 集合實作
type MongoDBCollection struct {
	collection *mongo.Collection
}

// FindOne 查詢單筆文件
func (c *MongoDBCollection) FindOne(ctx context.Context, filter Filter, result interface{}) error {
	bsonFilter := convertFilterToBSON(filter)
	err := c.collection.FindOne(ctx, bsonFilter).Decode(result)
	if err == mongo.ErrNoDocuments {
		return ErrNoDocuments
	}
	return err
}

// Find 查詢多筆文件
func (c *MongoDBCollection) Find(ctx context.Context, filter Filter, sort Sort, limit, skip int64, results interface{}) error {
	bsonFilter := convertFilterToBSON(filter)
	opts := options.Find()

	if sort != nil {
		bsonSort := convertSortToBSON(sort)
		opts.SetSort(bsonSort)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}

	cursor, err := c.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	return cursor.All(ctx, results)
}

// InsertOne 新增單筆文件
func (c *MongoDBCollection) InsertOne(ctx context.Context, document interface{}) error {
	_, err := c.collection.InsertOne(ctx, document)
	return err
}

// UpdateOne 更新單筆文件
func (c *MongoDBCollection) UpdateOne(ctx context.Context, filter Filter, update Update) error {
	bsonFilter := convertFilterToBSON(filter)
	bsonUpdate := convertUpdateToBSON(update)

	_, err := c.collection.UpdateOne(ctx, bsonFilter, bsonUpdate)
	return err
}

// DeleteOne 刪除單筆文件
func (c *MongoDBCollection) DeleteOne(ctx context.Context, filter Filter) error {
	bsonFilter := convertFilterToBSON(filter)
	_, err := c.collection.DeleteOne(ctx, bsonFilter)
	return err
}

// CountDocuments 計算文件數量
func (c *MongoDBCollection) CountDocuments(ctx context.Context, filter Filter) (int64, error) {
	bsonFilter := convertFilterToBSON(filter)
	return c.collection.CountDocuments(ctx, bsonFilter)
}

// FindOneAndUpdate 查詢並更新單筆文件
func (c *MongoDBCollection) FindOneAndUpdate(ctx context.Context, filter Filter, update Update, returnAfter bool, result interface{}) error {
	bsonFilter := convertFilterToBSON(filter)
	bsonUpdate := convertUpdateToBSON(update)

	opts := options.FindOneAndUpdate()
	if returnAfter {
		opts.SetReturnDocument(options.After)
	} else {
		opts.SetReturnDocument(options.Before)
	}
	opts.SetUpsert(true)

	err := c.collection.FindOneAndUpdate(ctx, bsonFilter, bsonUpdate, opts).Decode(result)
	if err == mongo.ErrNoDocuments {
		return ErrNoDocuments
	}
	return err
}

// CreateIndex 建立索引
func (c *MongoDBCollection) CreateIndex(ctx context.Context, keys map[string]interface{}, indexOpts IndexOptions) error {
	bsonKeys := bson.M{}
	for k, v := range keys {
		bsonKeys[k] = v
	}

	opts := options.Index()
	if indexOpts.Unique {
		opts.SetUnique(true)
	}
	if indexOpts.Name != "" {
		opts.SetName(indexOpts.Name)
	}
	if indexOpts.Sparse {
		opts.SetSparse(true)
	}

	indexModel := mongo.IndexModel{
		Keys:    bsonKeys,
		Options: opts,
	}

	_, err := c.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// ListIndexes 列出索引
func (c *MongoDBCollection) ListIndexes(ctx context.Context) ([]IndexInfo, error) {
	cursor, err := c.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var indexes []IndexInfo
	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			continue
		}

		keys, ok := index["key"].(bson.M)
		if !ok {
			continue
		}

		name, _ := index["name"].(string)
		unique, _ := index["unique"].(bool)

		indexes = append(indexes, IndexInfo{
			Name:   name,
			Keys:   convertBSONToMap(keys),
			Unique: unique,
		})
	}

	return indexes, nil
}

// convertFilterToBSON 將 Filter 轉換為 BSON
func convertFilterToBSON(filter Filter) bson.M {
	if filter == nil {
		return bson.M{}
	}
	result := bson.M{}
	for k, v := range filter {
		result[k] = v
	}
	return result
}

// convertSortToBSON 將 Sort 轉換為 BSON
func convertSortToBSON(sort Sort) bson.D {
	if sort == nil {
		return bson.D{}
	}
	result := bson.D{}
	for _, field := range sort {
		result = append(result, bson.E{Key: field.Field, Value: field.Order})
	}
	return result
}

// convertUpdateToBSON 將 Update 轉換為 BSON
func convertUpdateToBSON(update Update) bson.M {
	result := bson.M{}
	if len(update.Set) > 0 {
		result["$set"] = update.Set
	}
	if len(update.AddToSet) > 0 {
		result["$addToSet"] = update.AddToSet
	}
	if len(update.Pull) > 0 {
		result["$pull"] = update.Pull
	}
	if len(update.Push) > 0 {
		result["$push"] = update.Push
	}
	return result
}

// convertBSONToMap 將 BSON 轉換為 map
func convertBSONToMap(bsonM bson.M) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range bsonM {
		result[k] = v
	}
	return result
}

