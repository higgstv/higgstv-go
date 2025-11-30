package database

import (
	"context"
	"fmt"
)

// NewDatabase 根據配置建立資料庫連線
func NewDatabase(ctx context.Context, config DatabaseConfig) (Database, error) {
	switch config.Type {
	case DatabaseTypeMongoDB:
		return NewMongoDBDatabase(ctx, config)
	case DatabaseTypeSQLite:
		return NewSQLiteDatabase(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// ParseDatabaseType 解析資料庫類型字串
func ParseDatabaseType(s string) (DatabaseType, error) {
	switch s {
	case "mongodb", "mongo":
		return DatabaseTypeMongoDB, nil
	case "sqlite", "sqlite3":
		return DatabaseTypeSQLite, nil
	default:
		return "", fmt.Errorf("unknown database type: %s", s)
	}
}

