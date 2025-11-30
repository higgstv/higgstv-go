package database

import (
	"context"

	"github.com/higgstv/higgstv-go/internal/models"
)

// DatabaseType 資料庫類型
type DatabaseType string

const (
	DatabaseTypeMongoDB DatabaseType = "mongodb"
	DatabaseTypeSQLite  DatabaseType = "sqlite"
)

// Filter 查詢過濾器（通用於 MongoDB 和 SQLite）
type Filter map[string]interface{}

// Sort 排序規則
type Sort []SortField

// SortField 排序欄位
type SortField struct {
	Field string
	Order int // 1 = ASC, -1 = DESC
}

// Update 更新操作
type Update struct {
	Set      map[string]interface{} // $set 操作
	AddToSet map[string]interface{}  // $addToSet 操作（僅 MongoDB，SQLite 需要手動處理）
	Pull     map[string]interface{}  // $pull 操作（僅 MongoDB，SQLite 需要手動處理）
	Push     map[string]interface{}  // $push 操作（僅 MongoDB，SQLite 需要手動處理）
}

// Database 資料庫抽象介面
type Database interface {
	// Type 回傳資料庫類型
	Type() DatabaseType

	// Collection 取得集合/表操作介面
	Collection(name string) Collection

	// Close 關閉資料庫連線
	Close(ctx context.Context) error

	// Ping 測試連線
	Ping(ctx context.Context) error

	// BeginTx 開始交易（如果支援）
	BeginTx(ctx context.Context) (Tx, error)
}

// Collection 集合/表操作介面
type Collection interface {
	// FindOne 查詢單筆文件
	FindOne(ctx context.Context, filter Filter, result interface{}) error

	// Find 查詢多筆文件
	Find(ctx context.Context, filter Filter, sort Sort, limit, skip int64, results interface{}) error

	// InsertOne 新增單筆文件
	InsertOne(ctx context.Context, document interface{}) error

	// UpdateOne 更新單筆文件
	UpdateOne(ctx context.Context, filter Filter, update Update) error

	// DeleteOne 刪除單筆文件
	DeleteOne(ctx context.Context, filter Filter) error

	// CountDocuments 計算文件數量
	CountDocuments(ctx context.Context, filter Filter) (int64, error)

	// FindOneAndUpdate 查詢並更新單筆文件
	FindOneAndUpdate(ctx context.Context, filter Filter, update Update, returnAfter bool, result interface{}) error

	// CreateIndex 建立索引
	CreateIndex(ctx context.Context, keys map[string]interface{}, options IndexOptions) error

	// ListIndexes 列出索引
	ListIndexes(ctx context.Context) ([]IndexInfo, error)
}

// IndexOptions 索引選項
type IndexOptions struct {
	Unique bool
	Name   string
	Sparse bool // 僅 MongoDB，SQLite 不支援
}

// IndexInfo 索引資訊
type IndexInfo struct {
	Name   string
	Keys   map[string]interface{}
	Unique bool
}

// Tx 交易介面
type Tx interface {
	// Commit 提交交易
	Commit(ctx context.Context) error

	// Rollback 回滾交易
	Rollback(ctx context.Context) error
}

// DatabaseFactory 資料庫工廠
type DatabaseFactory interface {
	// CreateDatabase 根據配置建立資料庫連線
	CreateDatabase(ctx context.Context, config DatabaseConfig) (Database, error)
}

// DatabaseConfig 資料庫配置
type DatabaseConfig struct {
	Type     DatabaseType
	URI      string
	Database string
}

// UserRepository 使用者 Repository 介面（抽象層）
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Exists(ctx context.Context, username, email string) (bool, error)
	Create(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, userID, hashedPassword string) error
	SetAccessKey(ctx context.Context, email, accessKey string) error
	ChangePasswordWithAccessKey(ctx context.Context, email, accessKey, hashedPassword string) (bool, error)
	AddChannel(ctx context.Context, username, channelID string) error
	SetUnclassifiedChannel(ctx context.Context, username, channelID string) error
	GetUsersBasicInfo(ctx context.Context, userIDs []string) ([]models.UserBasicInfo, error)
}

// ChannelRepository 頻道 Repository 介面（抽象層）
type ChannelRepository interface {
	FindByID(ctx context.Context, id string) (*models.Channel, error)
	Create(ctx context.Context, channel *models.Channel) error
	Update(ctx context.Context, id string, update map[string]interface{}) error
	ListChannels(ctx context.Context, filter Filter, sort Sort, limit, skip int64) ([]models.Channel, error)
	IsAdmin(ctx context.Context, channelID, userID string) (bool, error)
	AddOwners(ctx context.Context, channelID string, userIDs []string) error
}

// ProgramRepository 節目 Repository 介面（抽象層）
type ProgramRepository interface {
	GetNextProgramID(ctx context.Context) (int, error)
	AddProgram(ctx context.Context, channelID string, program *models.Program) error
	UpdateProgram(ctx context.Context, channelID string, programID int, update map[string]interface{}) error
	DeletePrograms(ctx context.Context, channelID string, programIDs []int) error
	SetOrder(ctx context.Context, channelID string, order []int) error
}

// ErrNoDocuments 找不到文件的錯誤（對應 MongoDB 的 ErrNoDocuments）
var ErrNoDocuments = &NotFoundError{Message: "no documents found"}

// NotFoundError 找不到文件錯誤
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// IsNotFound 檢查是否為找不到文件錯誤
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}

