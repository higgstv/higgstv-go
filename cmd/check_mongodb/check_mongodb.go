package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/config"
)

func main() {
	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Config loaded:\n")
	fmt.Printf("   Database URI: %s\n", cfg.Database.URI)
	fmt.Printf("   Database Name: %s\n", cfg.Database.Database)

	// 連線到 MongoDB
	fmt.Printf("\n🔌 Connecting to MongoDB...\n")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.URI))
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			fmt.Printf("⚠️  Failed to disconnect: %v\n", err)
		}
	}()

	// Ping MongoDB
	fmt.Printf("🏓 Pinging MongoDB...\n")
	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Printf("❌ Ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ MongoDB connection successful!\n")

	// 列出資料庫
	fmt.Printf("\n📚 Listing databases...\n")
	databases, err := client.ListDatabaseNames(ctx, nil)
	if err != nil {
		fmt.Printf("⚠️  Failed to list databases: %v\n", err)
	} else {
		fmt.Printf("   Found %d databases:\n", len(databases))
		for _, db := range databases {
			marker := "  "
			if db == cfg.Database.Database {
				marker = "👉"
			}
			fmt.Printf("   %s %s\n", marker, db)
		}
	}

	// 檢查測試資料庫
	testDBName := cfg.Database.Database + "_test"
	fmt.Printf("\n🧪 Checking test database: %s\n", testDBName)
	testDB := client.Database(testDBName)
	collections, err := testDB.ListCollectionNames(ctx, nil)
	if err != nil {
		fmt.Printf("⚠️  Test database may not exist or is empty: %v\n", err)
	} else {
		fmt.Printf("   Found %d collections:\n", len(collections))
		for _, coll := range collections {
			count, _ := testDB.Collection(coll).CountDocuments(ctx, nil)
			fmt.Printf("   - %s (%d documents)\n", coll, count)
		}
	}

	// 檢查主資料庫
	fmt.Printf("\n📊 Checking main database: %s\n", cfg.Database.Database)
	mainDB := client.Database(cfg.Database.Database)
	collections, err = mainDB.ListCollectionNames(ctx, nil)
	if err != nil {
		fmt.Printf("⚠️  Failed to list collections: %v\n", err)
	} else {
		fmt.Printf("   Found %d collections:\n", len(collections))
		for _, coll := range collections {
			count, _ := mainDB.Collection(coll).CountDocuments(ctx, nil)
			fmt.Printf("   - %s (%d documents)\n", coll, count)
		}
	}

	fmt.Printf("\n✅ MongoDB connection check completed!\n")
}

