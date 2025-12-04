package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	
	// 暫時禁用外鍵約束以允許遷移（遷移完成後會重新啟用）
	db := sqliteDB.GetDB()
	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		fmt.Printf("⚠️  無法禁用外鍵約束: %v\n", err)
	} else {
		fmt.Println("ℹ️  已暫時禁用外鍵約束以進行遷移")
	}

	fmt.Println("✅ SQLite 資料庫建立成功")
	fmt.Println()

	// 開始遷移
	fmt.Println("🚀 開始遷移資料...")
	fmt.Println()

	ctx := context.Background()
	stats := &MigrationStats{}

	// 建立 UUID 映射表（用於將 MongoDB UUID 映射到 SQLite ID）
	uuidMapping := make(map[string]string)

	// 1. 遷移使用者（並建立 UUID 映射）
	if err := migrateUsers(ctx, mongoDB, sqliteDB, stats, uuidMapping); err != nil {
		fmt.Printf("❌ 遷移使用者失敗: %v\n", err)
		os.Exit(1)
	}

	// 2. 遷移頻道（不含節目，使用 UUID 映射）
	if err := migrateChannels(ctx, mongoDB, sqliteDB, stats, uuidMapping); err != nil {
		fmt.Printf("❌ 遷移頻道失敗: %v\n", err)
		os.Exit(1)
	}

	// 3. 遷移節目（從頻道的 contents 中，使用 UUID 映射）
	if err := migratePrograms(ctx, mongoDB, sqliteDB, stats, uuidMapping); err != nil {
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

	// 重新啟用外鍵約束
	db = sqliteDB.GetDB()
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		fmt.Printf("⚠️  無法重新啟用外鍵約束: %v\n", err)
	} else {
		fmt.Println("ℹ️  已重新啟用外鍵約束")
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

// convertUUIDToID 將 UUID 轉換為 ID（支援多種格式）
func convertUUIDToID(uuidVal interface{}) string {
	if idStr, ok := uuidVal.(string); ok {
		return idStr
	} else if uuidBinary, ok := uuidVal.(primitive.Binary); ok {
		// UUID binary (subtype 4)
		return strings.ToUpper(hex.EncodeToString(uuidBinary.Data))
	} else {
		return fmt.Sprintf("%v", uuidVal)
	}
}

// migrateUsers 遷移使用者資料
func migrateUsers(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats, uuidMapping map[string]string) error {
	fmt.Println("📋 遷移使用者資料...")

	mongoUsersColl := mongoDB.Collection("users")
	cursor, err := mongoUsersColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB users 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	// 先讀取為 bson.M 以處理 UUID 類型的 _id
	var rawUsers []bson.M
	if err := cursor.All(ctx, &rawUsers); err != nil {
		return fmt.Errorf("讀取 MongoDB users 失敗: %w", err)
	}

	// 轉換為 models.User，處理 UUID 類型的 _id
	var users []models.User
	for _, rawUser := range rawUsers {
		var user models.User
		// 處理 _id（可能是 UUID 或 string）
		var mongoID string
		if idVal, ok := rawUser["_id"]; ok {
			mongoID = convertUUIDToID(idVal)
			user.ID = mongoID
		}
		// 處理其他欄位
		if username, ok := rawUser["username"].(string); ok {
			user.Username = username
		}
		if email, ok := rawUser["email"].(string); ok {
			user.Email = email
		}
		if password, ok := rawUser["password"].(string); ok {
			user.Password = password
		}
		if accessKey, ok := rawUser["access_key"].(string); ok {
			user.AccessKey = &accessKey
		}
		if ownChannels, ok := rawUser["own_channels"].(bson.A); ok {
			for _, ch := range ownChannels {
				if chStr, ok := ch.(string); ok {
					user.OwnChannels = append(user.OwnChannels, chStr)
				}
			}
		}
		if created, ok := rawUser["created"].(primitive.DateTime); ok {
			user.Created = created.Time()
		}
		if lastModified, ok := rawUser["last_modified"].(primitive.DateTime); ok {
			user.LastModified = lastModified.Time()
		}
		users = append(users, user)
	}

	fmt.Printf("   找到 %d 個使用者\n", len(users))

	userRepo := repository.NewUserRepository(sqliteDB)
	successCount := 0
	for i, user := range users {
		originalID := user.ID
		if err := userRepo.Create(ctx, &user); err != nil {
			// 如果使用者已存在，查詢現有的 ID
			if isDuplicateError(err) {
				// 嘗試從 SQLite 查詢現有使用者的 ID
				sqliteDBImpl, ok := sqliteDB.(*database.SQLiteDatabase)
				if ok {
					db := sqliteDBImpl.GetDB()
					var existingID string
					err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ?", user.Username).Scan(&existingID)
					if err == nil {
						user.ID = existingID
					}
				}
				fmt.Printf("   ⚠️  [%d/%d] 使用者 %s 已存在，跳過\n", i+1, len(users), user.Username)
			} else {
				stats.Errors = append(stats.Errors, fmt.Sprintf("使用者 %s: %v", user.Username, err))
				fmt.Printf("   ❌ [%d/%d] 使用者 %s 失敗: %v\n", i+1, len(users), user.Username, err)
				continue
			}
		} else {
			successCount++
			if (i+1)%10 == 0 || i == len(users)-1 {
				fmt.Printf("   ✅ [%d/%d] 使用者遷移中...\n", i+1, len(users))
			}
		}
		// 建立 UUID 映射（支援多種格式）
		if originalID != "" && user.ID != "" {
			uuidMapping[originalID] = user.ID
			// 如果原始 ID 是 base64 格式，也建立映射
			if strings.Contains(originalID, "==") || strings.Contains(originalID, "=") {
				uuidMapping[originalID] = user.ID
			}
		}
	}

	stats.Users = successCount
	fmt.Printf("   ✅ 使用者遷移完成: %d/%d 成功\n\n", successCount, len(users))
	return nil
}

// migrateChannels 遷移頻道資料（不含節目）
func migrateChannels(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats, uuidMapping map[string]string) error {
	fmt.Println("📋 遷移頻道資料...")

	mongoChannelsColl := mongoDB.Collection("channels")
	cursor, err := mongoChannelsColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB channels 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	// 先讀取為 bson.M 以處理 UUID 類型的 _id
	var rawChannels []bson.M
	if err := cursor.All(ctx, &rawChannels); err != nil {
		return fmt.Errorf("讀取 MongoDB channels 失敗: %w", err)
	}

	// 轉換為 models.Channel，處理 UUID 類型的 _id 和其他 UUID 欄位
	var channels []models.Channel
	for _, rawChannel := range rawChannels {
		// 處理 _id（可能是 UUID 或 string），統一轉換為無連字符的大寫 32 字符格式
		if idVal, ok := rawChannel["_id"]; ok {
			var idStr string
			if idStrVal, ok := idVal.(string); ok {
				// 如果是字串，移除連字符並轉大寫
				idStr = strings.ToUpper(strings.ReplaceAll(idStrVal, "-", ""))
			} else if uuidVal, ok := idVal.(primitive.Binary); ok {
				// UUID 類型（subtype 4）
				idStr = strings.ToUpper(hex.EncodeToString(uuidVal.Data))
			} else {
				// 嘗試轉換為字串，然後移除連字符並轉大寫
				idStr = strings.ToUpper(strings.ReplaceAll(fmt.Sprintf("%v", idVal), "-", ""))
			}
			rawChannel["_id"] = idStr
		}
		
		// 處理 owners 陣列中的 UUID（使用映射表）
		if owners, ok := rawChannel["owners"].(bson.A); ok {
			var ownerStrs []string
			for _, owner := range owners {
				ownerID := convertUUIDToID(owner)
				// 嘗試從映射表找到對應的 SQLite ID
				if mappedID, found := uuidMapping[ownerID]; found {
					ownerStrs = append(ownerStrs, mappedID)
				} else {
					// 如果找不到映射，嘗試直接使用（可能是已經正確的格式）
					ownerStrs = append(ownerStrs, ownerID)
				}
			}
			rawChannel["owners"] = ownerStrs
		}
		
		// 處理 permission 陣列中的 user_id UUID（使用映射表）
		if permissions, ok := rawChannel["permission"].(bson.A); ok {
			var permList []bson.M
			for _, perm := range permissions {
				if permMap, ok := perm.(bson.M); ok {
					if userID, ok := permMap["user_id"]; ok {
						userIDStr := convertUUIDToID(userID)
						// 嘗試從映射表找到對應的 SQLite ID
						if mappedID, found := uuidMapping[userIDStr]; found {
							permMap["user_id"] = mappedID
						} else {
							permMap["user_id"] = userIDStr
						}
					}
					permList = append(permList, permMap)
				}
			}
			rawChannel["permission"] = permList
		}
		
		// 處理 contents_seq（可能是 int 或 string）
		if contentsSeq, ok := rawChannel["contents_seq"]; ok {
			if contentsSeqStr, ok := contentsSeq.(string); ok {
				rawChannel["contents_seq"] = contentsSeqStr
			} else if contentsSeqInt, ok := contentsSeq.(int32); ok {
				rawChannel["contents_seq"] = fmt.Sprintf("%d", contentsSeqInt)
			} else if contentsSeqInt64, ok := contentsSeq.(int64); ok {
				rawChannel["contents_seq"] = fmt.Sprintf("%d", contentsSeqInt64)
			} else {
				rawChannel["contents_seq"] = fmt.Sprintf("%v", contentsSeq)
			}
		}
		
		// 處理頻道的 tags（確保是 int 陣列）
		if tags, ok := rawChannel["tags"].(bson.A); ok {
			var tagInts []int
			for _, tag := range tags {
				if tagInt, ok := tag.(int32); ok {
					tagInts = append(tagInts, int(tagInt))
				} else if tagInt64, ok := tag.(int64); ok {
					tagInts = append(tagInts, int(tagInt64))
				} else if tagInt, ok := tag.(int); ok {
					tagInts = append(tagInts, tagInt)
				} else if tagStr, ok := tag.(string); ok {
					// 嘗試將字串轉換為 int
					var tagInt int
					if _, err := fmt.Sscanf(tagStr, "%d", &tagInt); err == nil {
						tagInts = append(tagInts, tagInt)
					}
				}
			}
			rawChannel["tags"] = tagInts
		}
		
		// 處理 contents 中的 tags 和 duration（確保是正確類型）
		if contents, ok := rawChannel["contents"].(bson.A); ok {
			var contentsList []bson.M
			for _, content := range contents {
				if contentMap, ok := content.(bson.M); ok {
					// 處理 tags
					if tags, ok := contentMap["tags"].(bson.A); ok {
						var tagInts []int
						for _, tag := range tags {
							if tagInt, ok := tag.(int32); ok {
								tagInts = append(tagInts, int(tagInt))
							} else if tagInt64, ok := tag.(int64); ok {
								tagInts = append(tagInts, int(tagInt64))
							} else if tagInt, ok := tag.(int); ok {
								tagInts = append(tagInts, tagInt)
							} else if tagStr, ok := tag.(string); ok {
								// 嘗試將字串轉換為 int
								var tagInt int
								if _, err := fmt.Sscanf(tagStr, "%d", &tagInt); err == nil {
									tagInts = append(tagInts, tagInt)
								}
							}
						}
						contentMap["tags"] = tagInts
					}
					// 處理 duration
					if duration, ok := contentMap["duration"]; ok {
						var durationInt int
						converted := false
						
						// 嘗試各種類型的轉換
						if d, ok := duration.(int32); ok {
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(int64); ok {
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(int); ok {
							durationInt = d
							converted = true
						} else if d, ok := duration.(float64); ok {
							// 處理浮點數（可能是從 JSON 或其他來源轉換而來）
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(float32); ok {
							durationInt = int(d)
							converted = true
						} else if durationStr, ok := duration.(string); ok {
							// 嘗試將字串轉換為 int
							// 首先嘗試直接解析整數
							if d, err := strconv.Atoi(strings.TrimSpace(durationStr)); err == nil {
								durationInt = d
								converted = true
							} else {
								// 如果失敗，嘗試提取數字部分（例如 "123秒" -> 123）
								// 使用 fmt.Sscanf 來提取字串開頭的數字
								var extractedInt int
								if n, err := fmt.Sscanf(durationStr, "%d", &extractedInt); err == nil && n == 1 {
									durationInt = extractedInt
									converted = true
								}
							}
						}
						
						// 如果成功轉換，設置 duration；否則設置為 0（預設值）
						if converted {
							contentMap["duration"] = durationInt
						} else {
							// 無法轉換的 duration，設置為 0 以避免解碼錯誤
							contentMap["duration"] = 0
						}
					}
					contentsList = append(contentsList, contentMap)
				}
			}
			rawChannel["contents"] = contentsList
		}
		
		// 使用 bson.Unmarshal 處理複雜結構
		channelBytes, _ := bson.Marshal(rawChannel)
		var channel models.Channel
		if err := bson.Unmarshal(channelBytes, &channel); err == nil {
			// 確保 ID 格式一致（無連字符，大寫）
			channel.ID = strings.ToUpper(strings.ReplaceAll(channel.ID, "-", ""))
			channels = append(channels, channel)
		} else {
			// 如果解碼失敗，記錄錯誤但繼續
			stats.Errors = append(stats.Errors, fmt.Sprintf("頻道解碼失敗: %v", err))
		}
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
func migratePrograms(ctx context.Context, mongoDB *mongo.Database, sqliteDB database.Database, stats *MigrationStats, uuidMapping map[string]string) error {
	fmt.Println("📋 遷移節目資料...")

	mongoChannelsColl := mongoDB.Collection("channels")
	cursor, err := mongoChannelsColl.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查詢 MongoDB channels 失敗: %w", err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	// 先讀取為 bson.M 以處理類型轉換
	var rawChannels []bson.M
	if err := cursor.All(ctx, &rawChannels); err != nil {
		return fmt.Errorf("讀取 MongoDB channels 失敗: %w", err)
	}

	// 先找出所有節目的最大 ID，並設定 counter
	maxProgramID := 0
	allPrograms := make(map[string][]models.Program) // channelID -> programs
	
	for _, rawChannel := range rawChannels {
		// 處理頻道的 _id，統一轉換為無連字符的大寫 32 字符格式
		if idVal, ok := rawChannel["_id"]; ok {
			var idStr string
			if idStrVal, ok := idVal.(string); ok {
				// 如果是字串，移除連字符並轉大寫
				idStr = strings.ToUpper(strings.ReplaceAll(idStrVal, "-", ""))
			} else if uuidVal, ok := idVal.(primitive.Binary); ok {
				idStr = strings.ToUpper(hex.EncodeToString(uuidVal.Data))
			} else {
				// 嘗試轉換為字串，然後移除連字符並轉大寫
				idStr = strings.ToUpper(strings.ReplaceAll(fmt.Sprintf("%v", idVal), "-", ""))
			}
			rawChannel["_id"] = idStr
		}
		
		// 處理 owners 陣列中的 UUID
		if owners, ok := rawChannel["owners"].(bson.A); ok {
			var ownerStrs []string
			for _, owner := range owners {
				ownerID := convertUUIDToID(owner)
				if mappedID, found := uuidMapping[ownerID]; found {
					ownerStrs = append(ownerStrs, mappedID)
				} else {
					ownerStrs = append(ownerStrs, ownerID)
				}
			}
			rawChannel["owners"] = ownerStrs
		}
		
		// 處理 permission 陣列中的 user_id UUID
		if permissions, ok := rawChannel["permission"].(bson.A); ok {
			var permList []bson.M
			for _, perm := range permissions {
				if permMap, ok := perm.(bson.M); ok {
					if userID, ok := permMap["user_id"]; ok {
						userIDStr := convertUUIDToID(userID)
						if mappedID, found := uuidMapping[userIDStr]; found {
							permMap["user_id"] = mappedID
						} else {
							permMap["user_id"] = userIDStr
						}
					}
					permList = append(permList, permMap)
				}
			}
			rawChannel["permission"] = permList
		}
		
		// 處理 contents_seq（可能是 int 或 string）
		if contentsSeq, ok := rawChannel["contents_seq"]; ok {
			if contentsSeqStr, ok := contentsSeq.(string); ok {
				rawChannel["contents_seq"] = contentsSeqStr
			} else if contentsSeqInt, ok := contentsSeq.(int32); ok {
				rawChannel["contents_seq"] = fmt.Sprintf("%d", contentsSeqInt)
			} else if contentsSeqInt64, ok := contentsSeq.(int64); ok {
				rawChannel["contents_seq"] = fmt.Sprintf("%d", contentsSeqInt64)
			} else {
				rawChannel["contents_seq"] = fmt.Sprintf("%v", contentsSeq)
			}
		}
		
		// 處理頻道的 tags（確保是 int 陣列）
		if tags, ok := rawChannel["tags"].(bson.A); ok {
			var tagInts []int
			for _, tag := range tags {
				if tagInt, ok := tag.(int32); ok {
					tagInts = append(tagInts, int(tagInt))
				} else if tagInt64, ok := tag.(int64); ok {
					tagInts = append(tagInts, int(tagInt64))
				} else if tagInt, ok := tag.(int); ok {
					tagInts = append(tagInts, tagInt)
				} else if tagStr, ok := tag.(string); ok {
					var tagInt int
					if _, err := fmt.Sscanf(tagStr, "%d", &tagInt); err == nil {
						tagInts = append(tagInts, tagInt)
					}
				}
			}
			rawChannel["tags"] = tagInts
		}
		
		// 處理 contents 中的 tags 和 duration（確保是正確類型）
		if contents, ok := rawChannel["contents"].(bson.A); ok {
			var contentsList []bson.M
			for _, content := range contents {
				if contentMap, ok := content.(bson.M); ok {
					// 處理 tags
					if tags, ok := contentMap["tags"].(bson.A); ok {
						var tagInts []int
						for _, tag := range tags {
							if tagInt, ok := tag.(int32); ok {
								tagInts = append(tagInts, int(tagInt))
							} else if tagInt64, ok := tag.(int64); ok {
								tagInts = append(tagInts, int(tagInt64))
							} else if tagInt, ok := tag.(int); ok {
								tagInts = append(tagInts, tagInt)
							} else if tagStr, ok := tag.(string); ok {
								var tagInt int
								if _, err := fmt.Sscanf(tagStr, "%d", &tagInt); err == nil {
									tagInts = append(tagInts, tagInt)
								}
							}
						}
						contentMap["tags"] = tagInts
					}
					// 處理 duration
					if duration, ok := contentMap["duration"]; ok {
						var durationInt int
						converted := false
						
						// 嘗試各種類型的轉換
						if d, ok := duration.(int32); ok {
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(int64); ok {
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(int); ok {
							durationInt = d
							converted = true
						} else if d, ok := duration.(float64); ok {
							// 處理浮點數（可能是從 JSON 或其他來源轉換而來）
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(float32); ok {
							durationInt = int(d)
							converted = true
						} else if durationStr, ok := duration.(string); ok {
							// 嘗試將字串轉換為 int
							// 首先嘗試直接解析整數
							if d, err := strconv.Atoi(strings.TrimSpace(durationStr)); err == nil {
								durationInt = d
								converted = true
							} else {
								// 如果失敗，嘗試提取數字部分（例如 "123秒" -> 123）
								// 使用 fmt.Sscanf 來提取字串開頭的數字
								var extractedInt int
								if n, err := fmt.Sscanf(durationStr, "%d", &extractedInt); err == nil && n == 1 {
									durationInt = extractedInt
									converted = true
								}
							}
						}
						
						// 如果成功轉換，設置 duration；否則設置為 0（預設值）
						if converted {
							contentMap["duration"] = durationInt
						} else {
							// 無法轉換的 duration，設置為 0 以避免解碼錯誤
							contentMap["duration"] = 0
						}
					}
					contentsList = append(contentsList, contentMap)
				}
			}
			rawChannel["contents"] = contentsList
		}
		
		// 使用 bson.Unmarshal 處理複雜結構
		channelBytes, _ := bson.Marshal(rawChannel)
		var channel models.Channel
		if err := bson.Unmarshal(channelBytes, &channel); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("讀取頻道失敗: %v", err))
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

	// 使用已處理的頻道資料遷移節目
	programRepo := repository.NewProgramRepository(sqliteDB)
	totalPrograms := 0
	successCount := 0
	
	// 重新處理 rawChannels 以遷移節目（需要完整的類型轉換）
	for _, rawChannel := range rawChannels {
		// 處理頻道的 _id，統一轉換為無連字符的大寫 32 字符格式
		if idVal, ok := rawChannel["_id"]; ok {
			var idStr string
			if idStrVal, ok := idVal.(string); ok {
				// 如果是字串，移除連字符並轉大寫
				idStr = strings.ToUpper(strings.ReplaceAll(idStrVal, "-", ""))
			} else if uuidVal, ok := idVal.(primitive.Binary); ok {
				idStr = strings.ToUpper(hex.EncodeToString(uuidVal.Data))
			} else {
				// 嘗試轉換為字串，然後移除連字符並轉大寫
				idStr = strings.ToUpper(strings.ReplaceAll(fmt.Sprintf("%v", idVal), "-", ""))
			}
			rawChannel["_id"] = idStr
		}
		
		// 處理 owners 陣列中的 UUID（使用映射表）
		if owners, ok := rawChannel["owners"].(bson.A); ok {
			var ownerStrs []string
			for _, owner := range owners {
				ownerID := convertUUIDToID(owner)
				if mappedID, found := uuidMapping[ownerID]; found {
					ownerStrs = append(ownerStrs, mappedID)
				} else {
					ownerStrs = append(ownerStrs, ownerID)
				}
			}
			rawChannel["owners"] = ownerStrs
		}
		
		// 處理 permission 陣列中的 user_id UUID（使用映射表）
		if permissions, ok := rawChannel["permission"].(bson.A); ok {
			var permList []bson.M
			for _, perm := range permissions {
				if permMap, ok := perm.(bson.M); ok {
					if userID, ok := permMap["user_id"]; ok {
						userIDStr := convertUUIDToID(userID)
						if mappedID, found := uuidMapping[userIDStr]; found {
							permMap["user_id"] = mappedID
						} else {
							permMap["user_id"] = userIDStr
						}
					}
					permList = append(permList, permMap)
				}
			}
			rawChannel["permission"] = permList
		}
		
		// 處理 contents_seq（可能是 int 或 string）
		if contentsSeq, ok := rawChannel["contents_seq"]; ok {
			if contentsSeqStr, ok := contentsSeq.(string); ok {
				rawChannel["contents_seq"] = contentsSeqStr
			} else if contentsSeqInt, ok := contentsSeq.(int32); ok {
				rawChannel["contents_seq"] = fmt.Sprintf("%d", contentsSeqInt)
			} else if contentsSeqInt64, ok := contentsSeq.(int64); ok {
				rawChannel["contents_seq"] = fmt.Sprintf("%d", contentsSeqInt64)
			} else {
				rawChannel["contents_seq"] = fmt.Sprintf("%v", contentsSeq)
			}
		}
		
		// 處理頻道的 tags（確保是 int 陣列）
		if tags, ok := rawChannel["tags"].(bson.A); ok {
			var tagInts []int
			for _, tag := range tags {
				if tagInt, ok := tag.(int32); ok {
					tagInts = append(tagInts, int(tagInt))
				} else if tagInt64, ok := tag.(int64); ok {
					tagInts = append(tagInts, int(tagInt64))
				} else if tagInt, ok := tag.(int); ok {
					tagInts = append(tagInts, tagInt)
				} else if tagStr, ok := tag.(string); ok {
					var tagInt int
					if _, err := fmt.Sscanf(tagStr, "%d", &tagInt); err == nil {
						tagInts = append(tagInts, tagInt)
					}
				}
			}
			rawChannel["tags"] = tagInts
		}
		
		// 處理 contents 中的 tags 和 duration
		if contents, ok := rawChannel["contents"].(bson.A); ok {
			var contentsList []bson.M
			for _, content := range contents {
				if contentMap, ok := content.(bson.M); ok {
					// 處理 tags
					if tags, ok := contentMap["tags"].(bson.A); ok {
						var tagInts []int
						for _, tag := range tags {
							if tagInt, ok := tag.(int32); ok {
								tagInts = append(tagInts, int(tagInt))
							} else if tagInt64, ok := tag.(int64); ok {
								tagInts = append(tagInts, int(tagInt64))
							} else if tagInt, ok := tag.(int); ok {
								tagInts = append(tagInts, tagInt)
							} else if tagStr, ok := tag.(string); ok {
								var tagInt int
								if _, err := fmt.Sscanf(tagStr, "%d", &tagInt); err == nil {
									tagInts = append(tagInts, tagInt)
								}
							}
						}
						contentMap["tags"] = tagInts
					}
					// 處理 duration
					if duration, ok := contentMap["duration"]; ok {
						var durationInt int
						converted := false
						
						// 嘗試各種類型的轉換
						if d, ok := duration.(int32); ok {
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(int64); ok {
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(int); ok {
							durationInt = d
							converted = true
						} else if d, ok := duration.(float64); ok {
							// 處理浮點數（可能是從 JSON 或其他來源轉換而來）
							durationInt = int(d)
							converted = true
						} else if d, ok := duration.(float32); ok {
							durationInt = int(d)
							converted = true
						} else if durationStr, ok := duration.(string); ok {
							// 嘗試將字串轉換為 int
							// 首先嘗試直接解析整數
							if d, err := strconv.Atoi(strings.TrimSpace(durationStr)); err == nil {
								durationInt = d
								converted = true
							} else {
								// 如果失敗，嘗試提取數字部分（例如 "123秒" -> 123）
								// 使用 fmt.Sscanf 來提取字串開頭的數字
								var extractedInt int
								if n, err := fmt.Sscanf(durationStr, "%d", &extractedInt); err == nil && n == 1 {
									durationInt = extractedInt
									converted = true
								}
							}
						}
						
						// 如果成功轉換，設置 duration；否則設置為 0（預設值）
						if converted {
							contentMap["duration"] = durationInt
						} else {
							// 無法轉換的 duration，設置為 0 以避免解碼錯誤
							contentMap["duration"] = 0
						}
					}
					contentsList = append(contentsList, contentMap)
				}
			}
			rawChannel["contents"] = contentsList
		}
		
		// 使用 bson.Unmarshal 處理複雜結構
		channelBytes, _ := bson.Marshal(rawChannel)
		var channel models.Channel
		if err := bson.Unmarshal(channelBytes, &channel); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("讀取頻道失敗: %v", err))
			continue
		}

		if len(channel.Contents) == 0 {
			continue
		}

		// 檢查頻道是否在 SQLite 中存在（如果不存在，可能是重複頻道被跳過了）
		channelRepo := repository.NewChannelRepository(sqliteDB)
		existingChannel, err := channelRepo.FindByID(ctx, channel.ID)
		if err != nil {
			// 查詢錯誤，記錄但繼續處理節目
			stats.Errors = append(stats.Errors, fmt.Sprintf("查詢頻道 %s 失敗: %v", channel.ID, err))
		}
		
		// 如果頻道不存在，跳過節目遷移和順序設定（這是重複頻道）
		if existingChannel == nil {
			// 頻道不存在，可能是重複頻道被跳過了，不遷移其節目和順序
			continue
		}

		// 遷移該頻道的所有節目（保留原有 ID）
		for _, program := range channel.Contents {
			totalPrograms++
			programCopy := program
			
			// 使用 MigrateProgram 方法（保留原有 ID）
			inserted, err := programRepo.(*repository.SQLiteProgramRepository).MigrateProgram(ctx, channel.ID, &programCopy)
			if err != nil {
				// 如果是 UNIQUE constraint 錯誤或節目已存在錯誤，跳過
				if strings.Contains(err.Error(), "UNIQUE constraint") || 
				   strings.Contains(err.Error(), "already exists") {
					// 節目已存在，跳過（不計入錯誤）
					continue
				}
				stats.Errors = append(stats.Errors, fmt.Sprintf("節目 %d (頻道 %s): %v", program.ID, channel.ID, err))
				if totalPrograms%10000 == 0 {
					fmt.Printf("   ⚠️  節目遷移中... (%d 已處理, %d 成功插入)\n", totalPrograms, successCount)
				}
				continue
			}
			
			// 只有實際插入新節目時才計入成功
			if inserted {
			successCount++
				if successCount%10000 == 0 {
					fmt.Printf("   ✅ 節目遷移中... (%d/%d 成功插入)\n", successCount, totalPrograms)
				}
			}
		}
		
		// 如果有 contents_order，需要設定順序（只有在頻道存在時才設定）
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
