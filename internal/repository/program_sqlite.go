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

// SQLiteProgramRepository SQLite 節目 Repository
type SQLiteProgramRepository struct {
	db database.Database
}

// NewSQLiteProgramRepository 建立 SQLite 節目 Repository
func NewSQLiteProgramRepository(db database.Database) *SQLiteProgramRepository {
	return &SQLiteProgramRepository{db: db}
}

// getDB 取得底層 SQL 資料庫連線
func (r *SQLiteProgramRepository) getDB() *sql.DB {
	sqliteDB := r.db.(*database.SQLiteDatabase)
	return sqliteDB.GetDB()
}

// getNextProgramIDTx 在交易中取得下一個節目 ID（內部方法）
func (r *SQLiteProgramRepository) getNextProgramIDTx(ctx context.Context, tx *sql.Tx) (int, error) {
	// 更新計數器
	query := `UPDATE counters SET seq = seq + 1 WHERE id = 'program_id'`
	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	var seq int
	if rowsAffected == 0 {
		// 如果不存在，初始化為 1
		_, err = tx.ExecContext(ctx, `INSERT INTO counters (id, seq) VALUES ('program_id', 1)`)
		if err != nil {
			return 0, err
		}
		seq = 1
	} else {
		// 查詢當前的 seq 值
		err = tx.QueryRowContext(ctx, `SELECT seq FROM counters WHERE id = 'program_id'`).Scan(&seq)
		if err != nil {
			return 0, err
		}
	}

	return seq, nil
}

// GetNextProgramID 取得下一個節目 ID（使用 counter collection）
func (r *SQLiteProgramRepository) GetNextProgramID(ctx context.Context) (int, error) {
	db := r.getDB()

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	seq, err := r.getNextProgramIDTx(ctx, tx)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return seq, nil
}

// AddProgram 新增節目到頻道
func (r *SQLiteProgramRepository) AddProgram(ctx context.Context, channelID string, program *models.Program) error {
	db := r.getDB()

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 取得下一個節目 ID（使用同一個交易）
	programID, err := r.getNextProgramIDTx(ctx, tx)
	if err != nil {
		return err
	}
	program.ID = programID
	program.Created = time.Now()
	program.LastModified = time.Now()

	// 插入節目
	query := `INSERT INTO programs (id, channel_id, name, desc, duration, type, youtube_id, created, last_modified)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		program.ID,
		channelID,
		program.Name,
		program.Desc,
		program.Duration,
		program.Type,
		program.YouTubeID,
		program.Created,
		program.LastModified,
	)
	if err != nil {
		return err
	}

	// 插入 tags
	if len(program.Tags) > 0 {
		if err := r.insertProgramTagsTx(ctx, tx, program.ID, program.Tags); err != nil {
			return err
		}
	}

	// 更新頻道的 last_modified
	if _, err := tx.ExecContext(ctx, "UPDATE channels SET last_modified = ? WHERE id = ?", time.Now(), channelID); err != nil {
		return err
	}

	return tx.Commit()
}

// MigrateProgram 遷移節目（保留原有 ID，用於資料遷移）
// 返回值：(是否實際插入了新節目, 錯誤)
func (r *SQLiteProgramRepository) MigrateProgram(ctx context.Context, channelID string, program *models.Program) (bool, error) {
	db := r.getDB()

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 檢查節目是否已存在
	var existingID int
	var existingChannelID string
	err = tx.QueryRowContext(ctx, "SELECT id, channel_id FROM programs WHERE id = ?", program.ID).Scan(&existingID, &existingChannelID)
	inserted := false
	if err == nil {
		// 節目已存在，檢查是否屬於同一個頻道
		if existingChannelID == channelID {
			// 同一個頻道的同一個節目，跳過（可能是重複遷移）
			// 但仍需要確保 tags 正確
			if _, err := tx.ExecContext(ctx, "DELETE FROM program_tags WHERE program_id = ?", program.ID); err != nil {
				return false, err
			}
		} else {
			// 不同頻道的相同節目 ID，這不應該發生，但我們跳過以避免衝突
			// 記錄警告但繼續
			return false, fmt.Errorf("program %d already exists in channel %s, cannot add to channel %s", program.ID, existingChannelID, channelID)
		}
	} else if err != sql.ErrNoRows {
		return false, err
	} else {
		// 節目不存在，插入新節目（保留原有 ID）
		query := `INSERT INTO programs (id, channel_id, name, desc, duration, type, youtube_id, created, last_modified)
		          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

		result, err := tx.ExecContext(ctx, query,
			program.ID,
			channelID,
			program.Name,
			program.Desc,
			program.Duration,
			program.Type,
			program.YouTubeID,
			program.Created,
			program.LastModified,
		)
		if err != nil {
			return false, err
		}
		
		// 檢查是否實際插入了行
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return false, err
		}
		inserted = rowsAffected > 0
	}

	// 插入 tags
	if len(program.Tags) > 0 {
		if err := r.insertProgramTagsTx(ctx, tx, program.ID, program.Tags); err != nil {
			return false, err
		}
	}

	// 更新頻道的 last_modified
	if _, err := tx.ExecContext(ctx, "UPDATE channels SET last_modified = ? WHERE id = ?", time.Now(), channelID); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return inserted, nil
}

// UpdateProgram 更新節目
func (r *SQLiteProgramRepository) UpdateProgram(ctx context.Context, channelID string, programID int, update map[string]interface{}) error {
	db := r.getDB()

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 建立 UPDATE 語句
	setParts := []string{"last_modified = ?"}
	args := []interface{}{time.Now()}

	for key, value := range update {
		// 處理 contents.$ 前綴（SQLite 中不需要）
		key = strings.TrimPrefix(key, "contents.$.")

		if key == "tags" {
			// 刪除舊的 tags 並插入新的（在交易中）
			if _, err := tx.ExecContext(ctx, "DELETE FROM program_tags WHERE program_id = ?", programID); err != nil {
				return err
			}
			if tags, ok := value.([]int); ok {
				if err := r.insertProgramTagsTx(ctx, tx, programID, tags); err != nil {
					return err
				}
			}
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}
	}

	if len(setParts) > 1 {
		// 更新節目
		query := fmt.Sprintf("UPDATE programs SET %s WHERE id = ? AND channel_id = ?", strings.Join(setParts, ", "))
		args = append(args, programID, channelID)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	// 更新頻道的 last_modified
	if _, err := tx.ExecContext(ctx, "UPDATE channels SET last_modified = ? WHERE id = ?", time.Now(), channelID); err != nil {
		return err
	}

	return tx.Commit()
}

// DeletePrograms 刪除節目
func (r *SQLiteProgramRepository) DeletePrograms(ctx context.Context, channelID string, programIDs []int) error {
	if len(programIDs) == 0 {
		return nil
	}

	db := r.getDB()

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 建立 IN 查詢的佔位符
	placeholders := make([]string, len(programIDs))
	args := make([]interface{}, len(programIDs)+1)
	args[0] = channelID
	for i, id := range programIDs {
		placeholders[i] = "?"
		args[i+1] = id
	}

	// 刪除節目（外鍵約束會自動刪除關聯的 tags 和 order）
	query := fmt.Sprintf(`DELETE FROM programs WHERE channel_id = ? AND id IN (%s)`,
		strings.Join(placeholders, ","))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete programs: %w", err)
	}

	// 檢查是否有節目被刪除
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// 如果沒有刪除任何節目，可能是 ID 不存在或 channel_id 不匹配
	if rowsAffected == 0 {
		return fmt.Errorf("no programs deleted: channel_id=%s, program_ids=%v", channelID, programIDs)
	}

	// 更新頻道的 last_modified
	if _, err := tx.ExecContext(ctx, "UPDATE channels SET last_modified = ? WHERE id = ?", time.Now(), channelID); err != nil {
		return err
	}

	return tx.Commit()
}

// SetOrder 設定節目順序
func (r *SQLiteProgramRepository) SetOrder(ctx context.Context, channelID string, order []int) error {
	db := r.getDB()

	// 開始交易
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 刪除舊的順序
	if _, err := tx.ExecContext(ctx, "DELETE FROM channel_program_order WHERE channel_id = ?", channelID); err != nil {
		return err
	}

	// 插入新的順序（使用 INSERT OR IGNORE 處理重複）
	query := `INSERT OR IGNORE INTO channel_program_order (channel_id, program_id, order_index) VALUES (?, ?, ?)`
	for i, programID := range order {
		if _, err := tx.ExecContext(ctx, query, channelID, programID, i); err != nil {
			return err
		}
	}

	// 更新頻道的 last_modified
	if _, err := tx.ExecContext(ctx, "UPDATE channels SET last_modified = ? WHERE id = ?", time.Now(), channelID); err != nil {
		return err
	}

	return tx.Commit()
}

// 輔助方法：插入 program tags（交易版本）
func (r *SQLiteProgramRepository) insertProgramTagsTx(ctx context.Context, tx *sql.Tx, programID int, tags []int) error {
	if len(tags) == 0 {
		return nil
	}
	query := `INSERT OR IGNORE INTO program_tags (program_id, tag) VALUES (?, ?)`
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, query, programID, tag); err != nil {
			return err
		}
	}
	return nil
}
