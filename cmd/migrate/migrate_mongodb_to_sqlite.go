package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
	"github.com/higgstv/higgstv-go/internal/repository"
)

// MigrationStats 遷移統計
type MigrationStats struct {
	Users      int
	Channels   int
	Programs   int
	Counters   int
	Migrations int
	Errors     []string
}

func main() {
	fmt.Println("🔄 MongoDB 到 SQLite 資料遷移工具")
	fmt.Println("=====================================")
	fmt.Println()

	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 載入配置失敗: %v\n", err)
		os.Exit(1)
	}

	// 檢查 MongoDB 配置
	if cfg.Database.Type != "mongodb" {
		fmt.Printf("⚠️  警告: 當前配置的資料庫類型是 %s，不是 mongodb\n", cfg.Database.Type)
		fmt.Println("   請確認您要從 MongoDB 遷移資料")
		fmt.Println()
	}

	// 連線到 MongoDB（來源）
	fmt.Println("📥 連線到 MongoDB（來源資料庫）...")
	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(cfg.Database.URI))
	if err != nil {
		fmt.Printf("❌ MongoDB 連線失敗: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
	}()

	if err := mongoClient.Ping(mongoCtx, nil); err != nil {
		fmt.Printf("❌ MongoDB Ping 失敗: %v\n", err)
		os.Exit(1)
	}

	mongoDB := mongoClient.Database(cfg.Database.Database)
	fmt.Println("✅ MongoDB 連線成功")
	fmt.Println()

	// 顯示 MongoDB 統計
	if err := showMongoDBStats(mongoCtx, mongoDB); err != nil {
		fmt.Printf("⚠️  無法取得 MongoDB 統計: %v\n", err)
	}

	// 建立 SQLite 資料庫（目標）
	fmt.Println("\n📤 建立 SQLite 資料庫（目標資料庫）...")
	
	// 詢問 SQLite 檔案路徑
	sqlitePath := "./data/migrated_higgstv.db"
	if len(os.Args) > 1 {
		sqlitePath = os.Args[1]
	} else {
		fmt.Printf("   使用預設路徑: %s\n", sqlitePath)
		fmt.Println("   提示: 可以透過命令列參數指定路徑: go run cmd/migrate_mongodb_to_sqlite.go <path>")
	}

	// 檢查檔案是否已存在
	if _, err := os.Stat(sqlitePath); err == nil {
		fmt.Printf("⚠️  警告: 檔案 %s 已存在，將覆蓋現有資料\n", sqlitePath)
		fmt.Print("   是否繼續？(y/N): ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if strings.ToLower(answer) != "y" {
			fmt.Println("❌ 遷移已取消")
			os.Exit(0)
		}
		// 刪除舊檔案
		if err := os.Remove(sqlitePath); err != nil {
			fmt.Printf("⚠️  無法刪除舊檔案: %v\n", err)
		}
	}

	sqliteDB, err := database.NewSQLiteDatabase(context.Background(), database.DatabaseConfig{
		Type:     database.DatabaseTypeSQLite,
		URI:      fmt.Sprintf("file:%s?cache=shared&mode=rwc", sqlitePath),
		Database: cfg.Database.Database,
	})
	if err != nil {
		fmt.Printf("❌ SQLite 資料庫建立失敗: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = sqliteDB.Close(context.Background())
	}()

	fmt.Println("✅ SQLite 資料庫建立成功")
	fmt.Println()

	// 開始遷移
	fmt.Println("🚀 開始遷移資料...")
	fmt.Println()

	ctx := context.Background()
	stats := &MigrationStats{}

	// 1. 遷移使用者
	if err := migrateUsers(ctx, mongoDB, sqliteDB, stats); err != nil {
		fmt.Printf("❌ 遷移使用者失敗: %v\n", err)
		os.Exit(1)
	}

	// 2. 遷移頻道（不含節目）
	if err := migrateChannels(ctx, mongoDB, sqliteDB, stats); err != nil {
		fmt.Printf("❌ 遷移頻道失敗: %v\n", err)
		os.Exit(1)
	}

	// 3. 遷移節目（從頻道的 contents 中）
	if err := migratePrograms(ctx, mongoDB, sqliteDB, stats); err != nil {
		fmt.Printf("❌ 遷移節目失敗: %v\n", err)
		os.Exit(1)
	}

	// 4. 遷移計數器
	if err := migrateCounters(ctx, mongoDB, sqliteDB, stats); err != nil {
		fmt.Printf("❌ 遷移計數器失敗: %v\n", err)
		os.Exit(1)
	}

	// 5. 遷移遷移記錄
	if err := migrateMigrations(ctx, mongoDB, sqliteDB, stats); err != nil {
		fmt.Printf("⚠️  遷移記錄遷移失敗（可忽略）: %v\n", err)
	}

	// 顯示遷移統計
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📊 遷移統計")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("   使用者:     %d\n", stats.Users)
	fmt.Printf("   頻道:      %d\n", stats.Channels)
	fmt.Printf("   節目:      %d\n", stats.Programs)
	fmt.Printf("   計數器:    %d\n", stats.Counters)
	fmt.Printf("   遷移記錄:  %d\n", stats.Migrations)
	if len(stats.Errors) > 0 {
		fmt.Printf("\n⚠️  錯誤數量: %d\n", len(stats.Errors))
		for i, err := range stats.Errors {
			if i < 10 { // 只顯示前 10 個錯誤
				fmt.Printf("   - %s\n", err)
			}
		}
		if len(stats.Errors) > 10 {
			fmt.Printf("   ... 還有 %d 個錯誤未顯示\n", len(stats.Errors)-10)
		}
	}
	fmt.Println(strings.Repeat("=", 50))

	// 驗證資料完整性
	fmt.Println("\n🔍 驗證資料完整性...")
	if err := verifyMigration(ctx, mongoDB, sqliteDB); err != nil {
		fmt.Printf("⚠️  驗證失敗: %v\n", err)
	} else {
		fmt.Println("✅ 資料驗證通過")
	}

	fmt.Println("\n✅ 資料遷移完成！")
	fmt.Printf("   SQLite 資料庫位置: %s\n", sqlitePath)
	fmt.Println("\n💡 提示: 更新 config.yaml 使用 SQLite:")
	fmt.Printf("   database:\n")
	fmt.Printf("     type: \"sqlite\"\n")
	fmt.Printf("     uri: \"file:%s\"\n", sqlitePath)
}

// showMongoDBStats 顯示 MongoDB 統計資訊
func showMongoDBStats(ctx context.Context, mongoDB *mongo.Database) error {
	fmt.Println("📊 MongoDB 資料統計:")
	
	collections := []string{"users", "channels", "counters", "migrations"}
	for _, collName := range collections {
		coll := mongoDB.Collection(collName)
		count, err := coll.CountDocuments(ctx, bson.M{})
		if err != nil {
			continue
		}
		fmt.Printf("   %s: %d 筆\n", collName, count)
	}

	// 計算節目總數（從頻道的 contents 中）
	channelsColl := mongoDB.Collection("channels")
	cursor, err := channelsColl.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	totalPrograms := 0
	for cursor.Next(ctx) {
		var channel models.Channel
		if err := cursor.Decode(&channel); err == nil {
			totalPrograms += len(channel.Contents)
		}
	}
	fmt.Printf("   programs (總計): %d 筆\n", totalPrograms)

	return nil
}

// migrateUsers 遷移使用者資料
func migrateUsers(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats) error {
	fmt.Println("📋 遷移使用者資料...")

	mongoUsersColl := mongoDB.Collection("users")
	cursor, err := mongoUsersColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB users 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return fmt.Errorf("讀取 MongoDB users 失敗: %w", err)
	}

	fmt.Printf("   找到 %d 個使用者\n", len(users))

	userRepo := repository.NewUserRepository(sqliteDB)
	successCount := 0
	for i, user := range users {
		if err := userRepo.Create(ctx, &user); err != nil {
			// 如果使用者已存在，跳過
			if !isDuplicateError(err) {
				stats.Errors = append(stats.Errors, fmt.Sprintf("使用者 %s: %v", user.Username, err))
				fmt.Printf("   ❌ [%d/%d] 使用者 %s 失敗: %v\n", i+1, len(users), user.Username, err)
				continue
			}
			fmt.Printf("   ⚠️  [%d/%d] 使用者 %s 已存在，跳過\n", i+1, len(users), user.Username)
		} else {
			successCount++
			if (i+1)%10 == 0 || i == len(users)-1 {
				fmt.Printf("   ✅ [%d/%d] 使用者遷移中...\n", i+1, len(users))
			}
		}
	}

	stats.Users = successCount
	fmt.Printf("   ✅ 使用者遷移完成: %d/%d 成功\n\n", successCount, len(users))
	return nil
}

// migrateChannels 遷移頻道資料（不含節目）
func migrateChannels(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats) error {
	fmt.Println("📋 遷移頻道資料...")

	mongoChannelsColl := mongoDB.Collection("channels")
	cursor, err := mongoChannelsColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB channels 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var channels []models.Channel
	if err := cursor.All(ctx, &channels); err != nil {
		return fmt.Errorf("讀取 MongoDB channels 失敗: %w", err)
	}

	fmt.Printf("   找到 %d 個頻道\n", len(channels))

	channelRepo := repository.NewChannelRepository(sqliteDB)
	successCount := 0
	
	// 儲存頻道和節目的對應關係（用於後續遷移節目）
	channelProgramsMap := make(map[string][]models.Program)
	
	for i, channel := range channels {
		// 儲存節目的引用（稍後遷移）
		if len(channel.Contents) > 0 {
			channelProgramsMap[channel.ID] = channel.Contents
		}
		
		// 暫時清空 Contents，因為 ChannelRepository.Create 不會處理它們
		originalContents := channel.Contents
		channel.Contents = nil
		
		if err := channelRepo.Create(ctx, &channel); err != nil {
			if !isDuplicateError(err) {
				stats.Errors = append(stats.Errors, fmt.Sprintf("頻道 %s: %v", channel.ID, err))
				fmt.Printf("   ❌ [%d/%d] 頻道 %s 失敗: %v\n", i+1, len(channels), channel.ID, err)
				continue
			}
			fmt.Printf("   ⚠️  [%d/%d] 頻道 %s 已存在，跳過\n", i+1, len(channels), channel.ID)
		} else {
			successCount++
			if (i+1)%10 == 0 || i == len(channels)-1 {
				fmt.Printf("   ✅ [%d/%d] 頻道遷移中...\n", i+1, len(channels))
			}
		}
		
		// 恢復 Contents（用於後續遷移）
		channel.Contents = originalContents
	}

	stats.Channels = successCount
	fmt.Printf("   ✅ 頻道遷移完成: %d/%d 成功\n\n", successCount, len(channels))
	
	// 儲存 channelProgramsMap 到 context 或全域變數（這裡簡化處理，直接傳遞）
	// 注意：這裡需要將 channelProgramsMap 傳遞給 migratePrograms
	// 為了簡化，我們在 migratePrograms 中重新讀取
	
	return nil
}

// migratePrograms 遷移節目資料（保留原有 ID）
func migratePrograms(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats) error {
	fmt.Println("📋 遷移節目資料...")

	mongoChannelsColl := mongoDB.Collection("channels")
	cursor, err := mongoChannelsColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB channels 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	// 先找出所有節目的最大 ID，並設定 counter
	maxProgramID := 0
	allPrograms := make(map[string][]models.Program) // channelID -> programs
	
	for cursor.Next(ctx) {
		var channel models.Channel
		if err := cursor.Decode(&channel); err != nil {
			continue
		}

		if len(channel.Contents) > 0 {
			allPrograms[channel.ID] = channel.Contents
			for _, program := range channel.Contents {
				if program.ID > maxProgramID {
					maxProgramID = program.ID
				}
			}
		}
	}

	// 設定 counter 為最大 ID（確保後續新增不會衝突）
	if maxProgramID > 0 {
		countersColl := sqliteDB.Collection("counters")
		counterDoc := map[string]interface{}{
			"_id": "program_id",
			"seq": maxProgramID,
		}
		_ = countersColl.DeleteOne(ctx, database.Filter{"_id": "program_id"})
		if err := countersColl.InsertOne(ctx, counterDoc); err != nil {
			fmt.Printf("   ⚠️  設定 program_id counter 失敗: %v\n", err)
		}
	}

	// 重新查詢頻道（因為 cursor 已經遍歷完）
	_ = cursor.Close(ctx)
	cursor, err = mongoChannelsColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("重新查詢 MongoDB channels 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	programRepo := repository.NewProgramRepository(sqliteDB)
	totalPrograms := 0
	successCount := 0
	
	for cursor.Next(ctx) {
		var channel models.Channel
		if err := cursor.Decode(&channel); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("讀取頻道失敗: %v", err))
			continue
		}

		if len(channel.Contents) == 0 {
			continue
		}

		// 遷移該頻道的所有節目（保留原有 ID）
		for _, program := range channel.Contents {
			totalPrograms++
			programCopy := program
			
			// 使用 MigrateProgram 方法（保留原有 ID）
			if err := programRepo.(*repository.SQLiteProgramRepository).MigrateProgram(ctx, channel.ID, &programCopy); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("節目 %d (頻道 %s): %v", program.ID, channel.ID, err))
				fmt.Printf("   ❌ 節目 %d (頻道 %s) 失敗: %v\n", program.ID, channel.ID, err)
				continue
			}
			successCount++
		}
		
		// 如果有 contents_order，需要設定順序
		if len(channel.ContentsOrder) > 0 {
			if err := programRepo.SetOrder(ctx, channel.ID, channel.ContentsOrder); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("設定頻道 %s 節目順序失敗: %v", channel.ID, err))
			}
		}
	}

	stats.Programs = successCount
	fmt.Printf("   ✅ 節目遷移完成: %d/%d 成功\n\n", successCount, totalPrograms)
	return nil
}

// migrateCounters 遷移計數器資料
func migrateCounters(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats) error {
	fmt.Println("📋 遷移計數器資料...")

	mongoCountersColl := mongoDB.Collection("counters")
	cursor, err := mongoCountersColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB counters 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	type Counter struct {
		ID  string `bson:"_id"`
		Seq int    `bson:"seq"`
	}

	var counters []Counter
	if err := cursor.All(ctx, &counters); err != nil {
		return fmt.Errorf("讀取 MongoDB counters 失敗: %w", err)
	}

	fmt.Printf("   找到 %d 個計數器\n", len(counters))

	// 使用 Collection 操作計數器
	countersColl := sqliteDB.Collection("counters")
	successCount := 0
	for i, counter := range counters {
		counterDoc := map[string]interface{}{
			"_id": counter.ID,
			"seq": counter.Seq,
		}
		
		// 先嘗試刪除舊的（如果存在）
		_ = countersColl.DeleteOne(ctx, database.Filter{"_id": counter.ID})
		
		// 插入新的
		if err := countersColl.InsertOne(ctx, counterDoc); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("計數器 %s: %v", counter.ID, err))
			fmt.Printf("   ❌ 計數器 %s 失敗: %v\n", counter.ID, err)
			continue
		}
		successCount++
		fmt.Printf("   ✅ [%d/%d] 計數器: %s = %d\n", i+1, len(counters), counter.ID, counter.Seq)
	}

	stats.Counters = successCount
	fmt.Printf("   ✅ 計數器遷移完成: %d/%d 成功\n\n", successCount, len(counters))
	return nil
}

// migrateMigrations 遷移遷移記錄
func migrateMigrations(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats) error {
	fmt.Println("📋 遷移遷移記錄...")

	mongoMigrationsColl := mongoDB.Collection("migrations")
	cursor, err := mongoMigrationsColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB migrations 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	type Migration struct {
		ID          string    `bson:"_id"`
		Description string    `bson:"description"`
		ExecutedAt  time.Time `bson:"executed_at"`
	}

	var migrations []Migration
	if err := cursor.All(ctx, &migrations); err != nil {
		return fmt.Errorf("讀取 MongoDB migrations 失敗: %w", err)
	}

	fmt.Printf("   找到 %d 個遷移記錄\n", len(migrations))

	migrationsColl := sqliteDB.Collection("migrations")
	successCount := 0
	for _, migration := range migrations {
		migrationDoc := map[string]interface{}{
			"_id":         migration.ID,
			"description": migration.Description,
			"executed_at": migration.ExecutedAt,
		}
		if err := migrationsColl.InsertOne(ctx, migrationDoc); err != nil {
			if !isDuplicateError(err) {
				stats.Errors = append(stats.Errors, fmt.Sprintf("遷移記錄 %s: %v", migration.ID, err))
				continue
			}
		}
		successCount++
	}

	stats.Migrations = successCount
	fmt.Printf("   ✅ 遷移記錄遷移完成: %d/%d 成功\n\n", successCount, len(migrations))
	return nil
}

// verifyMigration 驗證遷移結果
func verifyMigration(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database) error {
	// 驗證使用者數量
	mongoUsersCount, err := mongoDB.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("無法取得 MongoDB 使用者數量: %w", err)
	}

	sqliteUsersColl := sqliteDB.Collection("users")
	sqliteUsersCount, err := sqliteUsersColl.CountDocuments(ctx, database.Filter{})
	if err != nil {
		return fmt.Errorf("無法取得 SQLite 使用者數量: %w", err)
	}

	if mongoUsersCount != sqliteUsersCount {
		return fmt.Errorf("使用者數量不匹配: MongoDB=%d, SQLite=%d", mongoUsersCount, sqliteUsersCount)
	}

	// 驗證頻道數量
	mongoChannelsCount, err := mongoDB.Collection("channels").CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("無法取得 MongoDB 頻道數量: %w", err)
	}

	sqliteChannelsColl := sqliteDB.Collection("channels")
	sqliteChannelsCount, err := sqliteChannelsColl.CountDocuments(ctx, database.Filter{})
	if err != nil {
		return fmt.Errorf("無法取得 SQLite 頻道數量: %w", err)
	}

	if mongoChannelsCount != sqliteChannelsCount {
		return fmt.Errorf("頻道數量不匹配: MongoDB=%d, SQLite=%d", mongoChannelsCount, sqliteChannelsCount)
	}

	fmt.Printf("   ✅ 使用者數量: %d\n", sqliteUsersCount)
	fmt.Printf("   ✅ 頻道數量: %d\n", sqliteChannelsCount)

	return nil
}

// isDuplicateError 檢查是否為重複鍵錯誤
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "UNIQUE constraint") ||
		contains(errStr, "duplicate") ||
		contains(errStr, "already exists")
}

// contains 檢查字串是否包含子字串（不區分大小寫）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
