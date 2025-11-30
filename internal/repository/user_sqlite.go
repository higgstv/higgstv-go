package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
)

// SQLiteUserRepository SQLite 使用者 Repository
type SQLiteUserRepository struct {
	db database.Database
}

// NewSQLiteUserRepository 建立 SQLite 使用者 Repository
func NewSQLiteUserRepository(db database.Database) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// getDB 取得底層 SQL 資料庫連線
func (r *SQLiteUserRepository) getDB() *sql.DB {
	// 透過反射或類型斷言取得底層 *sql.DB
	// 這需要在 database 包中提供方法
	sqliteDB := r.db.(*database.SQLiteDatabase)
	return sqliteDB.GetDB()
}

// FindByUsername 依使用者名稱查詢
func (r *SQLiteUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	db := r.getDB()
	query := `SELECT id, username, email, password, access_key, unclassified_channel, created, last_modified 
	          FROM users WHERE username = ?`
	
	var user models.User
	var accessKey sql.NullString
	var unclassifiedChannel sql.NullString
	
	err := db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&accessKey,
		&unclassifiedChannel,
		&user.Created,
		&user.LastModified,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if accessKey.Valid {
		user.AccessKey = &accessKey.String
	}
	if unclassifiedChannel.Valid {
		user.UnclassifiedChannel = &unclassifiedChannel.String
	}

	// 載入 own_channels
	channels, err := r.loadOwnChannels(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.OwnChannels = channels

	return &user, nil
}

// FindByEmail 依 Email 查詢
func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	db := r.getDB()
	query := `SELECT id, username, email, password, access_key, unclassified_channel, created, last_modified 
	          FROM users WHERE email = ?`
	
	var user models.User
	var accessKey sql.NullString
	var unclassifiedChannel sql.NullString
	
	err := db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&accessKey,
		&unclassifiedChannel,
		&user.Created,
		&user.LastModified,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if accessKey.Valid {
		user.AccessKey = &accessKey.String
	}
	if unclassifiedChannel.Valid {
		user.UnclassifiedChannel = &unclassifiedChannel.String
	}

	// 載入 own_channels
	channels, err := r.loadOwnChannels(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.OwnChannels = channels

	return &user, nil
}

// Exists 檢查使用者是否存在
func (r *SQLiteUserRepository) Exists(ctx context.Context, username, email string) (bool, error) {
	db := r.getDB()
	query := `SELECT COUNT(*) FROM users WHERE username = ? OR email = ?`
	
	var count int64
	err := db.QueryRowContext(ctx, query, username, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create 建立使用者
func (r *SQLiteUserRepository) Create(ctx context.Context, user *models.User) error {
	db := r.getDB()
	now := time.Now()
	user.Created = now
	user.LastModified = now

	query := `INSERT INTO users (id, username, email, password, access_key, unclassified_channel, created, last_modified)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	var accessKey interface{}
	if user.AccessKey != nil {
		accessKey = *user.AccessKey
	}
	var unclassifiedChannel interface{}
	if user.UnclassifiedChannel != nil {
		unclassifiedChannel = *user.UnclassifiedChannel
	}

	_, err := db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		accessKey,
		unclassifiedChannel,
		user.Created,
		user.LastModified,
	)
	if err != nil {
		return err
	}

	// 插入 own_channels
	if len(user.OwnChannels) > 0 {
		return r.insertOwnChannels(ctx, user.ID, user.OwnChannels)
	}

	return nil
}

// UpdatePassword 更新密碼
func (r *SQLiteUserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	db := r.getDB()
	query := `UPDATE users SET password = ?, last_modified = ? WHERE id = ?`
	
	_, err := db.ExecContext(ctx, query, hashedPassword, time.Now(), userID)
	return err
}

// SetAccessKey 設定 access_key
func (r *SQLiteUserRepository) SetAccessKey(ctx context.Context, email, accessKey string) error {
	db := r.getDB()
	query := `UPDATE users SET access_key = ?, last_modified = ? WHERE email = ?`
	
	_, err := db.ExecContext(ctx, query, accessKey, time.Now(), email)
	return err
}

// ChangePasswordWithAccessKey 使用 access_key 重設密碼
func (r *SQLiteUserRepository) ChangePasswordWithAccessKey(ctx context.Context, email, accessKey, hashedPassword string) (bool, error) {
	db := r.getDB()
	query := `UPDATE users SET password = ?, access_key = NULL, last_modified = ? 
	          WHERE email = ? AND access_key = ?`
	
	result, err := db.ExecContext(ctx, query, hashedPassword, time.Now(), email, accessKey)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// AddChannel 新增頻道到使用者的 own_channels
func (r *SQLiteUserRepository) AddChannel(ctx context.Context, username, channelID string) error {
	db := r.getDB()
	
	// 先取得使用者 ID
	var userID string
	err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err == sql.ErrNoRows {
		return database.ErrNoDocuments
	}
	if err != nil {
		return err
	}

	// 插入到 user_channels 表（使用 INSERT OR IGNORE 避免重複）
	query := `INSERT OR IGNORE INTO user_channels (user_id, channel_id) VALUES (?, ?)`
	_, err = db.ExecContext(ctx, query, userID, channelID)
	if err != nil {
		return err
	}

	// 更新 last_modified
	_, err = db.ExecContext(ctx, "UPDATE users SET last_modified = ? WHERE id = ?", time.Now(), userID)
	return err
}

// SetUnclassifiedChannel 設定未分類頻道
func (r *SQLiteUserRepository) SetUnclassifiedChannel(ctx context.Context, username, channelID string) error {
	db := r.getDB()
	query := `UPDATE users SET unclassified_channel = ?, last_modified = ? WHERE username = ?`
	
	result, err := db.ExecContext(ctx, query, channelID, time.Now(), username)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return database.ErrNoDocuments
	}

	return nil
}

// GetUsersBasicInfo 取得使用者基本資訊（用於 owners_info）
func (r *SQLiteUserRepository) GetUsersBasicInfo(ctx context.Context, userIDs []string) ([]models.UserBasicInfo, error) {
	if len(userIDs) == 0 {
		return []models.UserBasicInfo{}, nil
	}

	db := r.getDB()
	
	// 建立 IN 查詢的佔位符
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id, username, email FROM users WHERE id IN (` + 
		joinStrings(placeholders, ",") + `)`
	
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []models.UserBasicInfo
	for rows.Next() {
		var info models.UserBasicInfo
		if err := rows.Scan(&info.ID, &info.Username, &info.Email); err != nil {
			return nil, err
		}
		results = append(results, info)
	}

	return results, rows.Err()
}

// loadOwnChannels 載入使用者的 own_channels
func (r *SQLiteUserRepository) loadOwnChannels(ctx context.Context, userID string) ([]string, error) {
	db := r.getDB()
	query := `SELECT channel_id FROM user_channels WHERE user_id = ? ORDER BY channel_id`
	
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var channels []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, err
		}
		channels = append(channels, channelID)
	}

	return channels, rows.Err()
}

// insertOwnChannels 插入 own_channels
func (r *SQLiteUserRepository) insertOwnChannels(ctx context.Context, userID string, channelIDs []string) error {
	if len(channelIDs) == 0 {
		return nil
	}

	db := r.getDB()
	query := `INSERT OR IGNORE INTO user_channels (user_id, channel_id) VALUES (?, ?)`
	
	for _, channelID := range channelIDs {
		if _, err := db.ExecContext(ctx, query, userID, channelID); err != nil {
			return err
		}
	}

	return nil
}

// joinStrings 輔助函數：連接字串陣列
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

