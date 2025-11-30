# MongoDB 到 SQLite 資料遷移工具使用指南

## 概述

這是一個深度且嚴謹的資料遷移工具，可以將 MongoDB 資料庫中的所有資料遷移到 SQLite，包括：

- ✅ 使用者（users）
- ✅ 頻道（channels）
- ✅ 節目（programs）- **保留原有 ID**
- ✅ 計數器（counters）
- ✅ 遷移記錄（migrations）
- ✅ 所有關聯資料（tags, owners, permissions, program_tags, contents_order）

## 功能特點

### 1. 完整的資料遷移
- 遷移所有集合的資料
- 保留資料的完整性和關聯性
- 保留節目的原有 ID（確保 contents_order 正確）

### 2. 資料驗證
- 遷移前顯示 MongoDB 統計資訊
- 遷移後驗證資料完整性
- 檢查使用者數量和頻道數量是否匹配

### 3. 錯誤處理
- 詳細的錯誤記錄
- 繼續執行（即使部分資料遷移失敗）
- 顯示遷移統計和錯誤摘要

### 4. 進度顯示
- 即時顯示遷移進度
- 每 10 筆資料顯示一次進度
- 顯示成功/失敗統計

## 使用方法

### 基本使用

```bash
# 使用預設路徑（./data/migrated_higgstv.db）
go run cmd/migrate_mongodb_to_sqlite.go

# 指定 SQLite 檔案路徑
go run cmd/migrate_mongodb_to_sqlite.go /path/to/output.db
```

### 編譯後使用

```bash
# 編譯
go build -o migrate_mongodb_to_sqlite cmd/migrate_mongodb_to_sqlite.go

# 執行
./migrate_mongodb_to_sqlite
./migrate_mongodb_to_sqlite /path/to/output.db
```

## 遷移流程

### 1. 連線檢查
- 連線到 MongoDB（來源資料庫）
- 顯示 MongoDB 資料統計
- 建立 SQLite 資料庫（目標資料庫）

### 2. 資料遷移順序
1. **使用者** - 遷移所有使用者資料（含 own_channels）
2. **頻道** - 遷移頻道基本資訊（不含節目）
3. **節目** - 遷移所有節目的完整資料（保留原有 ID）
4. **計數器** - 遷移計數器（確保 ID 生成正確）
5. **遷移記錄** - 遷移資料庫遷移記錄

### 3. 資料驗證
- 驗證使用者數量
- 驗證頻道數量
- 顯示驗證結果

## 注意事項

### 1. 資料庫連線
- 確保 MongoDB 正在運行且可連線
- 確保有足夠的磁碟空間存放 SQLite 檔案
- 如果 SQLite 檔案已存在，會詢問是否覆蓋

### 2. 節目 ID 保留
- 工具會自動找出所有節目的最大 ID
- 設定 counter 為該值，確保後續新增不會衝突
- 保留所有節目的原有 ID，確保 `contents_order` 正確

### 3. 交易處理
- 每個頻道的節目遷移都在單一交易中完成
- 如果某個節目遷移失敗，該頻道的所有節目都會回滾
- 確保資料一致性

### 4. 錯誤處理
- 如果某筆資料遷移失敗，會記錄錯誤但繼續執行
- 遷移完成後會顯示所有錯誤的摘要
- 可以根據錯誤訊息進行修復

## 遷移後配置

遷移完成後，更新 `config.yaml`：

```yaml
database:
  type: "sqlite"
  uri: "file:./data/migrated_higgstv.db?cache=shared&mode=rwc"
  database: "higgstv"
```

## 驗證遷移結果

### 1. 使用檢查工具

```bash
go run cmd/check_database.go
```

### 2. 手動驗證

```bash
# SQLite
sqlite3 ./data/migrated_higgstv.db "SELECT COUNT(*) FROM users;"
sqlite3 ./data/migrated_higgstv.db "SELECT COUNT(*) FROM channels;"
sqlite3 ./data/migrated_higgstv.db "SELECT COUNT(*) FROM programs;"

# MongoDB
mongosh --eval "db.users.countDocuments()"
mongosh --eval "db.channels.countDocuments()"
```

### 3. 檢查節目 ID

```bash
# 檢查最大節目 ID
sqlite3 ./data/migrated_higgstv.db "SELECT MAX(id) FROM programs;"

# 檢查 counter
sqlite3 ./data/migrated_higgstv.db "SELECT seq FROM counters WHERE id = 'program_id';"
```

## 故障排除

### 1. 連線失敗

**問題**: 無法連線到 MongoDB

**解決方案**:
- 檢查 MongoDB 是否正在運行
- 檢查 `config.yaml` 中的 URI 是否正確
- 檢查網路連線和防火牆設定

### 2. 檔案權限錯誤

**問題**: 無法建立 SQLite 檔案

**解決方案**:
- 檢查目錄權限
- 確保有寫入權限
- 使用絕對路徑

### 3. 資料不匹配

**問題**: 遷移後資料數量不匹配

**解決方案**:
- 檢查錯誤日誌
- 確認是否有重複資料
- 檢查外鍵約束

### 4. 節目 ID 衝突

**問題**: 節目 ID 衝突

**解決方案**:
- 工具會自動處理，設定 counter 為最大 ID
- 如果仍有問題，檢查 counter 表

## 技術細節

### 1. 節目 ID 保留機制

```go
// 1. 找出所有節目的最大 ID
maxProgramID := 0
for _, programs := range allPrograms {
    for _, program := range programs {
        if program.ID > maxProgramID {
            maxProgramID = program.ID
        }
    }
}

// 2. 設定 counter
countersColl.InsertOne(ctx, map[string]interface{}{
    "_id": "program_id",
    "seq": maxProgramID,
})

// 3. 使用 MigrateProgram 方法保留原有 ID
programRepo.MigrateProgram(ctx, channelID, &program)
```

### 2. 交易處理

- 每個頻道的節目遷移都在單一交易中完成
- 確保原子性：要麼全部成功，要麼全部回滾
- 避免部分遷移導致的資料不一致

### 3. 錯誤收集

- 使用 `MigrationStats` 結構收集所有錯誤
- 遷移完成後統一顯示
- 不中斷遷移流程（除非關鍵錯誤）

## 效能考量

### 1. 批次處理
- 每 10 筆資料顯示一次進度
- 減少輸出，提高效能

### 2. 交易優化
- 每個頻道的節目在單一交易中處理
- 減少交易開銷

### 3. 記憶體使用
- 分批讀取資料，避免記憶體溢出
- 使用 cursor 逐筆處理

## 最佳實踐

1. **備份資料**: 遷移前備份 MongoDB 資料
2. **測試遷移**: 先在測試環境測試遷移流程
3. **驗證結果**: 遷移後仔細驗證資料完整性
4. **監控錯誤**: 檢查錯誤日誌，修復問題
5. **更新配置**: 遷移完成後更新配置檔案

## 範例輸出

```
🔄 MongoDB 到 SQLite 資料遷移工具
=====================================

📥 連線到 MongoDB（來源資料庫）...
✅ MongoDB 連線成功

📊 MongoDB 資料統計:
   users: 10 筆
   channels: 5 筆
   counters: 2 筆
   migrations: 1 筆
   programs (總計): 25 筆

📤 建立 SQLite 資料庫（目標資料庫）...
   使用預設路徑: ./data/migrated_higgstv.db
✅ SQLite 資料庫建立成功

🚀 開始遷移資料...

📋 遷移使用者資料...
   找到 10 個使用者
   ✅ [10/10] 使用者遷移中...
   ✅ 使用者遷移完成: 10/10 成功

📋 遷移頻道資料...
   找到 5 個頻道
   ✅ [5/5] 頻道遷移中...
   ✅ 頻道遷移完成: 5/5 成功

📋 遷移節目資料...
   ✅ 節目遷移完成: 25/25 成功

📋 遷移計數器資料...
   找到 2 個計數器
   ✅ [1/2] 計數器: program_id = 25
   ✅ [2/2] 計數器: channel_id = 5
   ✅ 計數器遷移完成: 2/2 成功

📋 遷移遷移記錄...
   找到 1 個遷移記錄
   ✅ 遷移記錄遷移完成: 1/1 成功

==================================================
📊 遷移統計
==================================================
   使用者:     10
   頻道:      5
   節目:      25
   計數器:    2
   遷移記錄:  1
==================================================

🔍 驗證資料完整性...
   ✅ 使用者數量: 10
   ✅ 頻道數量: 5
✅ 資料驗證通過

✅ 資料遷移完成！
   SQLite 資料庫位置: ./data/migrated_higgstv.db
```

## 總結

這個遷移工具提供了完整、嚴謹的資料遷移功能，確保：

- ✅ 所有資料都被正確遷移
- ✅ 資料關聯性保持完整
- ✅ 節目 ID 被正確保留
- ✅ 資料完整性得到驗證
- ✅ 錯誤被詳細記錄和報告

使用這個工具，您可以安全地將 MongoDB 資料遷移到 SQLite，並確保資料的完整性和一致性。

