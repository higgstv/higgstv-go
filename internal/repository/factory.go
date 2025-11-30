package repository

import (
	"github.com/higgstv/higgstv-go/internal/database"
)

// NewUserRepository 建立使用者 Repository（根據資料庫類型）
func NewUserRepository(db database.Database) database.UserRepository {
	switch db.Type() {
	case database.DatabaseTypeMongoDB:
		return NewMongoDBUserRepository(db)
	case database.DatabaseTypeSQLite:
		return NewSQLiteUserRepository(db)
	default:
		panic("unsupported database type")
	}
}

// NewChannelRepository 建立頻道 Repository（根據資料庫類型）
func NewChannelRepository(db database.Database) database.ChannelRepository {
	switch db.Type() {
	case database.DatabaseTypeMongoDB:
		return NewMongoDBChannelRepository(db)
	case database.DatabaseTypeSQLite:
		return NewSQLiteChannelRepository(db)
	default:
		panic("unsupported database type")
	}
}

// NewProgramRepository 建立節目 Repository（根據資料庫類型）
func NewProgramRepository(db database.Database) database.ProgramRepository {
	switch db.Type() {
	case database.DatabaseTypeMongoDB:
		return NewMongoDBProgramRepository(db)
	case database.DatabaseTypeSQLite:
		return NewSQLiteProgramRepository(db)
	default:
		panic("unsupported database type")
	}
}

