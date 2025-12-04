package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDatabase SQLite 資料庫實作
type SQLiteDatabase struct {
	db *sql.DB
}

// GetDB 取得底層 SQL 資料庫連線（供 Repository 使用）
func (d *SQLiteDatabase) GetDB() *sql.DB {
	return d.db
}

// NewSQLiteDatabase 建立 SQLite 資料庫連線
func NewSQLiteDatabase(ctx context.Context, config DatabaseConfig) (*SQLiteDatabase, error) {
	// SQLite URI 格式: file:path?cache=shared&mode=rwc
	// 如果 URI 不是以 file: 開頭，則視為檔案路徑
	dsn := config.URI
	if !strings.HasPrefix(dsn, "file:") && !strings.HasPrefix(dsn, "sqlite:") {
		dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc", dsn)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// 設定連線池參數
	db.SetMaxOpenConns(1) // SQLite 建議使用單一連線
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// 測試連線
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// 啟用外鍵約束
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	sqliteDB := &SQLiteDatabase{db: db}

	// 確保資料庫 schema 已建立
	if err := sqliteDB.ensureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ensure schema: %w", err)
	}

	return sqliteDB, nil
}

// Type 回傳資料庫類型
func (d *SQLiteDatabase) Type() DatabaseType {
	return DatabaseTypeSQLite
}

// Collection 取得集合操作介面
func (d *SQLiteDatabase) Collection(name string) Collection {
	return &SQLiteCollection{
		db:   d.db,
		name: name,
	}
}

// Close 關閉資料庫連線
func (d *SQLiteDatabase) Close(ctx context.Context) error {
	return d.db.Close()
}

// Ping 測試連線
func (d *SQLiteDatabase) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// BeginTx 開始交易
func (d *SQLiteDatabase) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &SQLiteTx{tx: tx}, nil
}

// SQLiteTx SQLite 交易實作
type SQLiteTx struct {
	tx *sql.Tx
}

// Commit 提交交易
func (t *SQLiteTx) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback 回滾交易
func (t *SQLiteTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// SQLiteCollection SQLite 集合實作
type SQLiteCollection struct {
	db   *sql.DB
	name string
}

// ensureSchema 確保資料庫 schema 已建立
func (d *SQLiteDatabase) ensureSchema(ctx context.Context) error {
	schemas := []string{
		// users 表
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			access_key TEXT,
			unclassified_channel TEXT,
			created DATETIME NOT NULL,
			last_modified DATETIME NOT NULL
		)`,
		// user_channels 表（使用者擁有的頻道）
		`CREATE TABLE IF NOT EXISTS user_channels (
			user_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			PRIMARY KEY (user_id, channel_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// channels 表
		`CREATE TABLE IF NOT EXISTS channels (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			desc TEXT,
			contents_seq TEXT,
			cover_default TEXT,
			created DATETIME NOT NULL,
			last_modified DATETIME NOT NULL
		)`,
		// channel_tags 表
		`CREATE TABLE IF NOT EXISTS channel_tags (
			channel_id TEXT NOT NULL,
			tag INTEGER NOT NULL,
			PRIMARY KEY (channel_id, tag),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		)`,
		// channel_owners 表
		`CREATE TABLE IF NOT EXISTS channel_owners (
			channel_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (channel_id, user_id),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// channel_permissions 表
		`CREATE TABLE IF NOT EXISTS channel_permissions (
			channel_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			admin INTEGER NOT NULL DEFAULT 0,
			read INTEGER NOT NULL DEFAULT 0,
			write INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (channel_id, user_id),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		// programs 表
		`CREATE TABLE IF NOT EXISTS programs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			channel_id TEXT NOT NULL,
			name TEXT NOT NULL,
			desc TEXT,
			duration INTEGER,
			type TEXT NOT NULL,
			youtube_id TEXT,
			created DATETIME NOT NULL,
			last_modified DATETIME NOT NULL,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		)`,
		// program_tags 表
		`CREATE TABLE IF NOT EXISTS program_tags (
			program_id INTEGER NOT NULL,
			tag INTEGER NOT NULL,
			PRIMARY KEY (program_id, tag),
			FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE
		)`,
		// channel_program_order 表（節目順序）
		`CREATE TABLE IF NOT EXISTS channel_program_order (
			channel_id TEXT NOT NULL,
			program_id INTEGER NOT NULL,
			order_index INTEGER NOT NULL,
			PRIMARY KEY (channel_id, program_id),
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
			FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE
		)`,
		// counters 表
		`CREATE TABLE IF NOT EXISTS counters (
			id TEXT PRIMARY KEY,
			seq INTEGER NOT NULL DEFAULT 0
		)`,
		// migrations 表
		`CREATE TABLE IF NOT EXISTS migrations (
			id TEXT PRIMARY KEY,
			description TEXT,
			executed_at DATETIME NOT NULL
		)`,
	}

	// 建立索引
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_access_key ON users(access_key) WHERE access_key IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_channels_owners ON channel_owners(channel_id, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_last_modified ON channels(last_modified DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_name ON channels(name)`,
		`CREATE INDEX IF NOT EXISTS idx_programs_channel_id ON programs(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_program_tags_program_id ON program_tags(program_id)`,
	}

	for _, schema := range schemas {
		if _, err := d.db.ExecContext(ctx, schema); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	for _, index := range indexes {
		if _, err := d.db.ExecContext(ctx, index); err != nil {
			// 索引建立失敗不中斷，只記錄警告
			_ = err
		}
	}

	return nil
}

// FindOne 查詢單筆文件（需要根據 collection 名稱實作不同的查詢邏輯）
func (c *SQLiteCollection) FindOne(ctx context.Context, filter Filter, result interface{}) error {
	// 這個方法需要根據不同的 collection 實作
	// 暫時返回錯誤，需要在具體的 Repository 中實作
	return fmt.Errorf("FindOne not implemented for SQLite collection: %s", c.name)
}

// Find 查詢多筆文件
func (c *SQLiteCollection) Find(ctx context.Context, filter Filter, sort Sort, limit, skip int64, results interface{}) error {
	// 根據 collection 名稱實作不同的查詢邏輯
	switch c.name {
	case "migrations":
		return c.findMigrations(ctx, filter, sort, limit, skip, results)
	default:
		return fmt.Errorf("Find not implemented for SQLite collection: %s", c.name)
	}
}

// InsertOne 新增單筆文件
func (c *SQLiteCollection) InsertOne(ctx context.Context, document interface{}) error {
	// 根據 collection 名稱實作不同的插入邏輯
	switch c.name {
	case "migrations":
		return c.insertMigration(ctx, document)
	case "counters":
		return c.insertCounter(ctx, document)
	default:
		return fmt.Errorf("InsertOne not implemented for SQLite collection: %s", c.name)
	}
}

// UpdateOne 更新單筆文件
func (c *SQLiteCollection) UpdateOne(ctx context.Context, filter Filter, update Update) error {
	// 這個方法需要根據不同的 collection 實作
	return fmt.Errorf("UpdateOne not implemented for SQLite collection: %s", c.name)
}

// DeleteOne 刪除單筆文件
func (c *SQLiteCollection) DeleteOne(ctx context.Context, filter Filter) error {
	// 根據 collection 名稱實作不同的刪除邏輯
	switch c.name {
	case "counters":
		return c.deleteCounter(ctx, filter)
	default:
		return fmt.Errorf("DeleteOne not implemented for SQLite collection: %s", c.name)
	}
}

// CountDocuments 計算文件數量
func (c *SQLiteCollection) CountDocuments(ctx context.Context, filter Filter) (int64, error) {
	// 根據 collection 名稱實作不同的計數邏輯
	switch c.name {
	case "migrations":
		return c.countMigrations(ctx, filter)
	case "users":
		return c.countUsers(ctx, filter)
	default:
		return 0, fmt.Errorf("CountDocuments not implemented for SQLite collection: %s", c.name)
	}
}

// FindOneAndUpdate 查詢並更新單筆文件
func (c *SQLiteCollection) FindOneAndUpdate(ctx context.Context, filter Filter, update Update, returnAfter bool, result interface{}) error {
	// 這個方法需要根據不同的 collection 實作
	return fmt.Errorf("FindOneAndUpdate not implemented for SQLite collection: %s", c.name)
}

// CreateIndex 建立索引
func (c *SQLiteCollection) CreateIndex(ctx context.Context, keys map[string]interface{}, options IndexOptions) error {
	// SQLite 索引建立
	if len(keys) == 0 {
		return fmt.Errorf("index keys cannot be empty")
	}

	var keyParts []string
	for k, v := range keys {
		order := "ASC"
		if orderVal, ok := v.(int); ok && orderVal < 0 {
			order = "DESC"
		}
		keyParts = append(keyParts, fmt.Sprintf("%s %s", k, order))
	}

	indexName := options.Name
	if indexName == "" {
		indexName = fmt.Sprintf("idx_%s_%s", c.name, strings.Join(keyParts, "_"))
	}

	unique := ""
	if options.Unique {
		unique = "UNIQUE"
	}

	sql := fmt.Sprintf("CREATE %s INDEX IF NOT EXISTS %s ON %s (%s)",
		unique, indexName, c.name, strings.Join(keyParts, ", "))

	_, err := c.db.ExecContext(ctx, sql)
	return err
}

// ListIndexes 列出索引
func (c *SQLiteCollection) ListIndexes(ctx context.Context) ([]IndexInfo, error) {
	query := `
		SELECT name, sql 
		FROM sqlite_master 
		WHERE type='index' AND tbl_name=? AND name NOT LIKE 'sqlite_%'
	`

	rows, err := c.db.QueryContext(ctx, query, c.name)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var indexes []IndexInfo
	for rows.Next() {
		var name, sqlStr string
		if err := rows.Scan(&name, &sqlStr); err != nil {
			continue
		}

		// 解析 SQL 以提取 keys 和 unique
		keys := make(map[string]interface{})
		unique := strings.Contains(sqlStr, "UNIQUE")

		// 簡單解析（實際應該使用更複雜的 SQL 解析器）
		if strings.Contains(sqlStr, "(") && strings.Contains(sqlStr, ")") {
			start := strings.Index(sqlStr, "(")
			end := strings.LastIndex(sqlStr, ")")
			fields := sqlStr[start+1 : end]
			for _, field := range strings.Split(fields, ",") {
				field = strings.TrimSpace(field)
				parts := strings.Fields(field)
				if len(parts) > 0 {
					order := 1
					if len(parts) > 1 && strings.ToUpper(parts[1]) == "DESC" {
						order = -1
					}
					keys[parts[0]] = order
				}
			}
		}

		indexes = append(indexes, IndexInfo{
			Name:   name,
			Keys:   keys,
			Unique: unique,
		})
	}

	return indexes, nil
}

// findMigrations 查詢 migrations 表
func (c *SQLiteCollection) findMigrations(ctx context.Context, filter Filter, sort Sort, limit, skip int64, results interface{}) error {
	query := "SELECT id, description, executed_at FROM migrations"
	args := []interface{}{}

	// 處理過濾條件（目前 migrations 查詢通常不需要過濾）
	// 如果需要過濾，可以在這裡實作 WHERE 子句
	_ = filter

	// 處理排序
	if len(sort) > 0 {
		orderParts := []string{}
		for _, s := range sort {
			order := "ASC"
			if s.Order < 0 {
				order = "DESC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", s.Field, order))
		}
		if len(orderParts) > 0 {
			query += " ORDER BY " + strings.Join(orderParts, ", ")
		}
	}

	// 處理 LIMIT 和 OFFSET
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
		if skip > 0 {
			query += fmt.Sprintf(" OFFSET %d", skip)
		}
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	// 使用反射來填充結果
	resultsValue := reflect.ValueOf(results)
	if resultsValue.Kind() != reflect.Ptr || resultsValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("results must be a pointer to a slice")
	}

	sliceValue := resultsValue.Elem()
	elemType := sliceValue.Type().Elem()

	for rows.Next() {
		elemValue := reflect.New(elemType).Elem()

		// 根據結構體欄位掃描
		var id, description string
		var executedAt time.Time

		if err := rows.Scan(&id, &description, &executedAt); err != nil {
			return err
		}

		// 填充結構體欄位
		for i := 0; i < elemValue.NumField(); i++ {
			field := elemValue.Field(i)
			fieldType := elemType.Field(i)

			switch fieldType.Name {
			case "ID":
				field.SetString(id)
			case "Description":
				if field.Kind() == reflect.String {
					field.SetString(description)
				}
			case "ExecutedAt":
				if field.Type() == reflect.TypeOf(time.Time{}) {
					field.Set(reflect.ValueOf(executedAt))
				}
			}
		}

		sliceValue.Set(reflect.Append(sliceValue, elemValue))
	}

	return rows.Err()
}

// insertMigration 插入 migration 記錄
func (c *SQLiteCollection) insertMigration(ctx context.Context, document interface{}) error {
	var id, description string
	var executedAt time.Time

	// 先檢查是否為 map 類型
	if docMap, ok := document.(map[string]interface{}); ok {
		// 從 map 取得值
		if idVal, ok := docMap["_id"]; ok {
			if idStr, ok := idVal.(string); ok {
				id = idStr
			}
		}
		if descVal, ok := docMap["description"]; ok {
			if descStr, ok := descVal.(string); ok {
				description = descStr
			}
		}
		if execVal, ok := docMap["executed_at"]; ok {
			if execTime, ok := execVal.(time.Time); ok {
				executedAt = execTime
			}
		}
	} else {
		// 使用反射取得欄位值（struct 類型）
		docValue := reflect.ValueOf(document)
		if docValue.Kind() == reflect.Ptr {
			docValue = docValue.Elem()
		}

		// 只有當是 struct 類型時才使用反射
		if docValue.Kind() == reflect.Struct {
			docType := docValue.Type()
			for i := 0; i < docValue.NumField(); i++ {
				field := docValue.Field(i)
				fieldType := docType.Field(i)

				// 處理 bson tag 或 json tag
				bsonTag := fieldType.Tag.Get("bson")
				if bsonTag == "" {
					bsonTag = fieldType.Tag.Get("json")
				}

				switch bsonTag {
				case "_id", "id":
					if field.Kind() == reflect.String {
						id = field.String()
					} else if field.Kind() == reflect.Interface {
						if str, ok := field.Interface().(string); ok {
							id = str
						}
					}
				case "description":
					if field.Kind() == reflect.String {
						description = field.String()
					} else if field.Kind() == reflect.Interface {
						if str, ok := field.Interface().(string); ok {
							description = str
						}
					}
				case "executed_at":
					if field.Type() == reflect.TypeOf(time.Time{}) {
						executedAt = field.Interface().(time.Time)
					} else if field.Kind() == reflect.Interface {
						if t, ok := field.Interface().(time.Time); ok {
							executedAt = t
						}
					}
				}
			}
		}
	}

	if id == "" {
		return fmt.Errorf("migration document must have _id field")
	}

	if executedAt.IsZero() {
		executedAt = time.Now()
	}

	query := "INSERT INTO migrations (id, description, executed_at) VALUES (?, ?, ?)"
	_, err := c.db.ExecContext(ctx, query, id, description, executedAt)
	return err
}

// countMigrations 計算 migrations 數量
func (c *SQLiteCollection) countMigrations(ctx context.Context, filter Filter) (int64, error) {
	query := "SELECT COUNT(*) FROM migrations"
	args := []interface{}{}

	// 處理過濾條件（目前 migrations 計數通常不需要過濾）
	// 如果需要過濾，可以在這裡實作 WHERE 子句
	_ = filter

	var count int64
	err := c.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// countUsers 計算 users 數量
func (c *SQLiteCollection) countUsers(ctx context.Context, filter Filter) (int64, error) {
	query := "SELECT COUNT(*) FROM users"
	args := []interface{}{}

	// 處理過濾條件
	whereParts := []string{}
	for key, value := range filter {
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	if len(whereParts) > 0 {
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}

	var count int64
	err := c.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// insertCounter 插入計數器記錄
func (c *SQLiteCollection) insertCounter(ctx context.Context, document interface{}) error {
	// 使用反射或 map 取得欄位值
	var id string
	var seq int

	if docMap, ok := document.(map[string]interface{}); ok {
		if idVal, ok := docMap["_id"]; ok {
			if idStr, ok := idVal.(string); ok {
				id = idStr
			}
		}
		if seqVal, ok := docMap["seq"]; ok {
			switch v := seqVal.(type) {
			case int:
				seq = v
			case int64:
				seq = int(v)
			case float64:
				seq = int(v)
			}
		}
	} else {
		// 嘗試使用反射
		docValue := reflect.ValueOf(document)
		if docValue.Kind() == reflect.Ptr {
			docValue = docValue.Elem()
		}

		docType := docValue.Type()
		for i := 0; i < docValue.NumField(); i++ {
			field := docValue.Field(i)
			fieldType := docType.Field(i)

			bsonTag := fieldType.Tag.Get("bson")
			if bsonTag == "" {
				bsonTag = fieldType.Tag.Get("json")
			}

			switch bsonTag {
			case "_id", "id":
				if field.Kind() == reflect.String {
					id = field.String()
				}
			case "seq":
				if field.Kind() == reflect.Int || field.Kind() == reflect.Int64 {
					seq = int(field.Int())
				}
			}
		}
	}

	if id == "" {
		return fmt.Errorf("counter document must have _id field")
	}

	query := "INSERT OR REPLACE INTO counters (id, seq) VALUES (?, ?)"
	_, err := c.db.ExecContext(ctx, query, id, seq)
	return err
}

// deleteCounter 刪除計數器記錄
func (c *SQLiteCollection) deleteCounter(ctx context.Context, filter Filter) error {
	query := "DELETE FROM counters WHERE "
	args := []interface{}{}

	whereParts := []string{}
	for key, value := range filter {
		// 處理 _id 欄位
		if key == "_id" || key == "id" {
			whereParts = append(whereParts, "id = ?")
			args = append(args, value)
		} else {
			whereParts = append(whereParts, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}
	}

	if len(whereParts) == 0 {
		return fmt.Errorf("delete filter cannot be empty")
	}

	query += strings.Join(whereParts, " AND ")
	_, err := c.db.ExecContext(ctx, query, args...)
	return err
}
