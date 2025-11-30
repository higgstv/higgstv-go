package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/database"
)

func main() {
	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Config loaded:\n")
	fmt.Printf("   Database Type: %s\n", cfg.Database.Type)
	fmt.Printf("   Database URI: %s\n", cfg.Database.URI)
	fmt.Printf("   Database Name: %s\n", cfg.Database.Database)

	// 解析資料庫類型
	dbType, err := database.ParseDatabaseType(cfg.Database.Type)
	if err != nil {
		fmt.Printf("❌ Invalid database type: %v\n", err)
		os.Exit(1)
	}

	// 連線到資料庫
	fmt.Printf("\n🔌 Connecting to database...\n")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.NewDatabase(ctx, database.DatabaseConfig{
		Type:     dbType,
		URI:      cfg.Database.URI,
		Database: cfg.Database.Database,
	})
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			fmt.Printf("⚠️  Failed to close connection: %v\n", err)
		}
	}()

	// Ping 資料庫
	fmt.Printf("🏓 Pinging database...\n")
	err = db.Ping(ctx)
	if err != nil {
		fmt.Printf("❌ Ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Database connection successful!\n")

	// 根據資料庫類型顯示不同資訊
	switch dbType {
	case database.DatabaseTypeMongoDB:
		checkMongoDB(ctx, db)
	case database.DatabaseTypeSQLite:
		checkSQLite(ctx, db)
	}

	fmt.Printf("\n✅ Database connection check completed!\n")
}

func checkMongoDB(ctx context.Context, db database.Database) {
	// MongoDB 特定檢查
	fmt.Printf("\n📚 MongoDB-specific checks:\n")
	fmt.Printf("   Database type: MongoDB\n")
	fmt.Printf("   Note: Use mongosh for detailed MongoDB inspection\n")
}

func checkSQLite(ctx context.Context, db database.Database) {
	// SQLite 特定檢查
	fmt.Printf("\n📚 SQLite-specific checks:\n")
	fmt.Printf("   Database type: SQLite\n")
	
	// 檢查 collections
	collections := []string{"users", "channels", "programs", "counters", "migrations"}
	for _, collName := range collections {
		coll := db.Collection(collName)
		count, err := coll.CountDocuments(ctx, database.Filter{})
		if err != nil {
			fmt.Printf("   ⚠️  %s: error checking (%v)\n", collName, err)
		} else {
			fmt.Printf("   ✅ %s: %d records\n", collName, count)
		}
	}
}

