package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
)

// SQLiteChannelRepository SQLite 頻道 Repository
type SQLiteChannelRepository struct {
	db database.Database
}

// NewSQLiteChannelRepository 建立 SQLite 頻道 Repository
func NewSQLiteChannelRepository(db database.Database) *SQLiteChannelRepository {
	return &SQLiteChannelRepository{db: db}
}

// getDB 取得底層 SQL 資料庫連線
func (r *SQLiteChannelRepository) getDB() *sql.DB {
	sqliteDB := r.db.(*database.SQLiteDatabase)
	return sqliteDB.GetDB()
}

// FindByID 依 ID 查詢頻道
func (r *SQLiteChannelRepository) FindByID(ctx context.Context, id string) (*models.Channel, error) {
	db := r.getDB()

	// 查詢頻道基本資訊
	query := `SELECT id, type, name, desc, CAST(contents_seq AS TEXT) as contents_seq, cover_default, created, last_modified 
	          FROM channels WHERE id = ?`

	var channel models.Channel
	var coverDefault sql.NullString
	var contentsSeq sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&channel.ID,
		&channel.Type,
		&channel.Name,
		&channel.Desc,
		&contentsSeq,
		&coverDefault,
		&channel.Created,
		&channel.LastModified,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if coverDefault.Valid {
		channel.Cover = &models.ChannelCover{Default: coverDefault.String}
	}

	// 處理 contents_seq
	if contentsSeq.Valid {
		channel.ContentsSeq = contentsSeq.String
	} else {
		channel.ContentsSeq = ""
	}

	// 載入 tags
	tags, err := r.loadTags(ctx, id)
	if err != nil {
		return nil, err
	}
	channel.Tags = tags

	// 載入 owners
	owners, err := r.loadOwners(ctx, id)
	if err != nil {
		return nil, err
	}
	channel.Owners = owners

	// 載入 permissions
	permissions, err := r.loadPermissions(ctx, id)
	if err != nil {
		return nil, err
	}
	channel.Permission = permissions

	// 載入 programs (contents)
	programs, err := r.loadPrograms(ctx, id)
	if err != nil {
		return nil, err
	}
	channel.Contents = programs

	// 載入 contents_order
	order, err := r.loadContentsOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	channel.ContentsOrder = order

	return &channel, nil
}

// Create 建立頻道
func (r *SQLiteChannelRepository) Create(ctx context.Context, channel *models.Channel) error {
	db := r.getDB()
	now := time.Now()
	channel.Created = now
	channel.LastModified = now

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 插入頻道基本資訊
	var coverDefault interface{}
	if channel.Cover != nil {
		coverDefault = channel.Cover.Default
	}

	query := `INSERT INTO channels (id, type, name, desc, contents_seq, cover_default, created, last_modified)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		channel.ID,
		channel.Type,
		channel.Name,
		channel.Desc,
		channel.ContentsSeq,
		coverDefault,
		channel.Created,
		channel.LastModified,
	)
	if err != nil {
		return err
	}

	// 插入 tags
	if err := r.insertTagsTx(ctx, tx, channel.ID, channel.Tags); err != nil {
		return err
	}

	// 插入 owners
	if err := r.insertOwnersTx(ctx, tx, channel.ID, channel.Owners); err != nil {
		return err
	}

	// 插入 permissions
	if err := r.insertPermissionsTx(ctx, tx, channel.ID, channel.Permission); err != nil {
		return err
	}

	// 插入 programs (contents)
	// 注意：Programs 應該通過 ProgramRepository 來新增，這裡不處理
	// 因為 Programs 需要先有 program ID，應該在建立頻道後再新增節目

	// 插入 contents_order
	if len(channel.ContentsOrder) > 0 {
		if err := r.insertContentsOrderTx(ctx, tx, channel.ID, channel.ContentsOrder); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Update 更新頻道
func (r *SQLiteChannelRepository) Update(ctx context.Context, id string, update map[string]interface{}) error {
	db := r.getDB()

	// 建立 UPDATE 語句
	setParts := []string{"last_modified = ?"}
	args := []interface{}{time.Now()}

	for key, value := range update {
		// 處理特殊欄位
		if key == "cover.default" {
			setParts = append(setParts, "cover_default = ?")
			args = append(args, value)
		} else if key == "tags" {
			// 刪除舊的 tags 並插入新的
			if _, err := db.ExecContext(ctx, "DELETE FROM channel_tags WHERE channel_id = ?", id); err != nil {
				return err
			}
			if tags, ok := value.([]int); ok {
				if err := r.insertTags(ctx, id, tags); err != nil {
					return err
				}
			}
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}
	}

	if len(setParts) == 1 {
		// 只有 last_modified，不需要更新
		return nil
	}

	query := fmt.Sprintf("UPDATE channels SET %s WHERE id = ?", strings.Join(setParts, ", "))
	args = append(args, id)

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// ListChannels 列出頻道（支援過濾和排序）
func (r *SQLiteChannelRepository) ListChannels(ctx context.Context, filter database.Filter, sort database.Sort, limit, skip int64) ([]models.Channel, error) {
	db := r.getDB()

	// 建立 WHERE 子句
	whereParts := []string{}
	args := []interface{}{}

	// 處理過濾條件
	for key, value := range filter {
		switch key {
		case "owners":
			// 使用 EXISTS 子查詢
			whereParts = append(whereParts, "EXISTS (SELECT 1 FROM channel_owners WHERE channel_id = channels.id AND user_id = ?)")
			args = append(args, value)
		case "name":
			// 處理正則表達式（簡化為 LIKE）
			if nameFilter, ok := value.(database.Filter); ok {
				if regex, ok := nameFilter["$regex"].(string); ok {
					whereParts = append(whereParts, "name LIKE ?")
					args = append(args, "%"+regex+"%")
				} else {
					whereParts = append(whereParts, "name = ?")
					args = append(args, value)
				}
			} else if nameFilter, ok := value.(map[string]interface{}); ok {
				if regex, ok := nameFilter["$regex"].(string); ok {
					whereParts = append(whereParts, "name LIKE ?")
					args = append(args, "%"+regex+"%")
				} else {
					whereParts = append(whereParts, "name = ?")
					args = append(args, value)
				}
			} else {
				whereParts = append(whereParts, "name = ?")
				args = append(args, value)
			}
		case "type":
			// 處理 $nin 操作（not in）
			if typeFilter, ok := value.(database.Filter); ok {
				if ninValues, ok := typeFilter["$nin"].([]string); ok {
					// 建立 NOT IN 子句
					placeholders := make([]string, len(ninValues))
					for i, v := range ninValues {
						placeholders[i] = "?"
						args = append(args, v)
					}
					whereParts = append(whereParts, fmt.Sprintf("type NOT IN (%s)", strings.Join(placeholders, ", ")))
				} else {
					// 預設為等於操作
					whereParts = append(whereParts, "type = ?")
					args = append(args, value)
				}
			} else if typeFilter, ok := value.(map[string]interface{}); ok {
				if ninValues, ok := typeFilter["$nin"].([]string); ok {
					// 建立 NOT IN 子句
					placeholders := make([]string, len(ninValues))
					for i, v := range ninValues {
						placeholders[i] = "?"
						args = append(args, v)
					}
					whereParts = append(whereParts, fmt.Sprintf("type NOT IN (%s)", strings.Join(placeholders, ", ")))
				} else {
					// 預設為等於操作
					whereParts = append(whereParts, "type = ?")
					args = append(args, value)
				}
			} else {
				// 預設為等於操作
				whereParts = append(whereParts, "type = ?")
				args = append(args, value)
			}
		case "contents.0":
			// 處理 has_contents 參數（檢查是否有節目）
			if contentsFilter, ok := value.(database.Filter); ok {
				if _, ok := contentsFilter["$exists"]; ok {
					// 使用 EXISTS 子查詢檢查是否有節目
					whereParts = append(whereParts, "EXISTS (SELECT 1 FROM channel_program_order WHERE channel_id = channels.id)")
				}
			} else if contentsFilter, ok := value.(map[string]interface{}); ok {
				if _, ok := contentsFilter["$exists"]; ok {
					// 使用 EXISTS 子查詢檢查是否有節目
					whereParts = append(whereParts, "EXISTS (SELECT 1 FROM channel_program_order WHERE channel_id = channels.id)")
				}
			}
		default:
			whereParts = append(whereParts, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// 建立 ORDER BY 子句
	orderClause := ""
	if len(sort) > 0 {
		orderParts := []string{}
		for _, s := range sort {
			order := "ASC"
			if s.Order < 0 {
				order = "DESC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", s.Field, order))
		}
		orderClause = "ORDER BY " + strings.Join(orderParts, ", ")
	}

	// 建立 LIMIT 和 OFFSET 子句
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", limit)
		if skip > 0 {
			limitClause += fmt.Sprintf(" OFFSET %d", skip)
		}
	}

	query := fmt.Sprintf(`SELECT id, type, name, desc, CAST(contents_seq AS TEXT) as contents_seq, cover_default, created, last_modified 
	                      FROM channels %s %s %s`, whereClause, orderClause, limitClause)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var channels []models.Channel
	for rows.Next() {
		var channel models.Channel
		var coverDefault sql.NullString
		var contentsSeq sql.NullString

		if err := rows.Scan(
			&channel.ID,
			&channel.Type,
			&channel.Name,
			&channel.Desc,
			&contentsSeq,
			&coverDefault,
			&channel.Created,
			&channel.LastModified,
		); err != nil {
			return nil, err
		}

		if coverDefault.Valid {
			channel.Cover = &models.ChannelCover{Default: coverDefault.String}
		}

		// 處理 contents_seq
		if contentsSeq.Valid {
			channel.ContentsSeq = contentsSeq.String
		} else {
			channel.ContentsSeq = ""
		}

		// 載入關聯資料（可選，根據需求決定是否載入）
		// 為了效能，這裡暫時不載入，需要時再載入

		channels = append(channels, channel)
	}

	return channels, rows.Err()
}

// IsAdmin 檢查使用者是否為頻道管理員
func (r *SQLiteChannelRepository) IsAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	db := r.getDB()
	query := `SELECT COUNT(*) FROM (
		SELECT 1 FROM channel_owners WHERE channel_id = ? AND user_id = ?
		UNION
		SELECT 1 FROM channel_permissions WHERE channel_id = ? AND user_id = ? AND admin = 1
	)`

	var count int64
	err := db.QueryRowContext(ctx, query, channelID, userID, channelID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddOwners 新增擁有者
func (r *SQLiteChannelRepository) AddOwners(ctx context.Context, channelID string, userIDs []string) error {
	db := r.getDB()

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 插入擁有者
	query := `INSERT OR IGNORE INTO channel_owners (channel_id, user_id) VALUES (?, ?)`
	for _, userID := range userIDs {
		if _, err := tx.ExecContext(ctx, query, channelID, userID); err != nil {
			return err
		}
	}

	// 更新 last_modified
	if _, err := tx.ExecContext(ctx, "UPDATE channels SET last_modified = ? WHERE id = ?", time.Now(), channelID); err != nil {
		return err
	}

	return tx.Commit()
}

// 輔助方法：載入 tags
func (r *SQLiteChannelRepository) loadTags(ctx context.Context, channelID string) ([]int, error) {
	db := r.getDB()
	query := `SELECT tag FROM channel_tags WHERE channel_id = ? ORDER BY tag`

	rows, err := db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var tags []int
	for rows.Next() {
		var tag int
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// 輔助方法：載入 owners
func (r *SQLiteChannelRepository) loadOwners(ctx context.Context, channelID string) ([]string, error) {
	db := r.getDB()
	query := `SELECT user_id FROM channel_owners WHERE channel_id = ? ORDER BY user_id`

	rows, err := db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var owners []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		owners = append(owners, userID)
	}

	return owners, rows.Err()
}

// 輔助方法：載入 permissions
func (r *SQLiteChannelRepository) loadPermissions(ctx context.Context, channelID string) ([]models.ChannelPermission, error) {
	db := r.getDB()
	query := `SELECT user_id, admin, read, write FROM channel_permissions WHERE channel_id = ?`

	rows, err := db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var permissions []models.ChannelPermission
	for rows.Next() {
		var perm models.ChannelPermission
		var admin, read, write int
		if err := rows.Scan(&perm.UserID, &admin, &read, &write); err != nil {
			return nil, err
		}
		perm.Admin = admin != 0
		perm.Read = read != 0
		perm.Write = write != 0
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

// 輔助方法：載入 programs
func (r *SQLiteChannelRepository) loadPrograms(ctx context.Context, channelID string) ([]models.Program, error) {
	db := r.getDB()
	query := `SELECT id, name, desc, duration, type, youtube_id, created, last_modified 
	          FROM programs WHERE channel_id = ? ORDER BY id`

	rows, err := db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var programs []models.Program
	var programIDs []int
	for rows.Next() {
		var program models.Program
		if err := rows.Scan(
			&program.ID,
			&program.Name,
			&program.Desc,
			&program.Duration,
			&program.Type,
			&program.YouTubeID,
			&program.Created,
			&program.LastModified,
		); err != nil {
			return nil, err
		}

		programs = append(programs, program)
		programIDs = append(programIDs, program.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 批量載入所有節目的 tags
	if len(programIDs) > 0 {
		tagsMap, err := r.loadProgramTagsBatch(ctx, programIDs)
		if err != nil {
			return nil, err
		}
		// 將 tags 分配到對應的節目
		for i := range programs {
			programs[i].Tags = tagsMap[programs[i].ID]
		}
	}

	return programs, nil
}

// 輔助方法：批量載入 program tags
func (r *SQLiteChannelRepository) loadProgramTagsBatch(ctx context.Context, programIDs []int) (map[int][]int, error) {
	if len(programIDs) == 0 {
		return make(map[int][]int), nil
	}

	db := r.getDB()
	placeholders := make([]string, len(programIDs))
	args := make([]interface{}, len(programIDs))
	for i, id := range programIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT program_id, tag FROM program_tags WHERE program_id IN (%s) ORDER BY program_id, tag`,
		strings.Join(placeholders, ","),
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	tagsMap := make(map[int][]int)
	for rows.Next() {
		var programID int
		var tag int
		if err := rows.Scan(&programID, &tag); err != nil {
			return nil, err
		}
		tagsMap[programID] = append(tagsMap[programID], tag)
	}

	return tagsMap, rows.Err()
}

// 輔助方法：載入 contents_order
func (r *SQLiteChannelRepository) loadContentsOrder(ctx context.Context, channelID string) ([]int, error) {
	db := r.getDB()
	query := `SELECT program_id FROM channel_program_order WHERE channel_id = ? ORDER BY order_index`

	rows, err := db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var order []int
	for rows.Next() {
		var programID int
		if err := rows.Scan(&programID); err != nil {
			return nil, err
		}
		order = append(order, programID)
	}

	return order, rows.Err()
}

// 輔助方法：插入 tags（交易版本）
func (r *SQLiteChannelRepository) insertTagsTx(ctx context.Context, tx *sql.Tx, channelID string, tags []int) error {
	if len(tags) == 0 {
		return nil
	}
	query := `INSERT INTO channel_tags (channel_id, tag) VALUES (?, ?)`
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, query, channelID, tag); err != nil {
			return err
		}
	}
	return nil
}

// 輔助方法：插入 tags
func (r *SQLiteChannelRepository) insertTags(ctx context.Context, channelID string, tags []int) error {
	db := r.getDB()
	query := `INSERT INTO channel_tags (channel_id, tag) VALUES (?, ?)`
	for _, tag := range tags {
		if _, err := db.ExecContext(ctx, query, channelID, tag); err != nil {
			return err
		}
	}
	return nil
}

// 輔助方法：插入 owners（交易版本）
func (r *SQLiteChannelRepository) insertOwnersTx(ctx context.Context, tx *sql.Tx, channelID string, owners []string) error {
	if len(owners) == 0 {
		return nil
	}
	query := `INSERT INTO channel_owners (channel_id, user_id) VALUES (?, ?)`
	for _, owner := range owners {
		var err error
		if tx != nil {
			_, err = tx.ExecContext(ctx, query, channelID, owner)
		} else {
			_, err = r.getDB().ExecContext(ctx, query, channelID, owner)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// 輔助方法：插入 permissions（交易版本）
func (r *SQLiteChannelRepository) insertPermissionsTx(ctx context.Context, tx *sql.Tx, channelID string, permissions []models.ChannelPermission) error {
	if len(permissions) == 0 {
		return nil
	}
	query := `INSERT INTO channel_permissions (channel_id, user_id, admin, read, write) VALUES (?, ?, ?, ?, ?)`
	for _, perm := range permissions {
		admin := 0
		if perm.Admin {
			admin = 1
		}
		read := 0
		if perm.Read {
			read = 1
		}
		write := 0
		if perm.Write {
			write = 1
		}
		var err error
		if tx != nil {
			_, err = tx.ExecContext(ctx, query, channelID, perm.UserID, admin, read, write)
		} else {
			_, err = r.getDB().ExecContext(ctx, query, channelID, perm.UserID, admin, read, write)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// 輔助方法：插入 contents_order（交易版本）
func (r *SQLiteChannelRepository) insertContentsOrderTx(ctx context.Context, tx *sql.Tx, channelID string, order []int) error {
	if len(order) == 0 {
		return nil
	}
	query := `INSERT INTO channel_program_order (channel_id, program_id, order_index) VALUES (?, ?, ?)`
	for i, programID := range order {
		var err error
		if tx != nil {
			_, err = tx.ExecContext(ctx, query, channelID, programID, i)
		} else {
			_, err = r.getDB().ExecContext(ctx, query, channelID, programID, i)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
