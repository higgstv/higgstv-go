# 資料遷移指南

## 概述

本指南說明如何在 MongoDB 和 SQLite 之間遷移資料，以及測試環境的資料處理方式。

## 重要說明

### 測試環境不需要資料遷移

**測試環境會自動建立空資料庫並建立測試資料**，不需要手動遷移資料：

1. **SQLite 測試**：使用記憶體資料庫（`file::memory:`），每次測試都是全新的空資料庫
2. **MongoDB 測試**：使用 `{database}_test` 資料庫，測試會自動建立資料
3. **測試資料**：測試會透過 API 自動建立（註冊使用者、建立頻道等）

### 生產環境才需要資料遷移

只有在以下情況才需要資料遷移工具：
- 將生產環境的 MongoDB 資料遷移到 SQLite
- 將生產環境的 SQLite 資料遷移到 MongoDB
- 備份和還原資料

## MongoDB → SQLite 資料遷移

### 使用遷移工具

我們提供了一個命令列工具來遷移 MongoDB 資料到 SQLite：

```bash
# 使用預設路徑（./data/migrated_higgstv.db）
go run cmd/migrate/migrate_mongodb_to_sqlite.go

# 指定 SQLite 檔案路徑
go run cmd/migrate/migrate_mongodb_to_sqlite.go /path/to/output.db
```

### 遷移流程

1. **連線到 MongoDB**（來源資料庫）
   - 使用 `config.yaml` 中的 MongoDB 配置
   - 讀取所有集合的資料

2. **建立 SQLite 資料庫**（目標資料庫）
   - 自動建立資料庫結構和索引
   - 如果檔案已存在，會覆蓋或更新

3. **遷移資料**：
   - ✅ 使用者（users）
   - ✅ 頻道（channels）
   - ✅ 計數器（counters）
   - ✅ 遷移記錄（migrations）

### 遷移後的配置

遷移完成後，更新 `config.yaml`：

```yaml
database:
  type: "sqlite"
  uri: "file:./data/migrated_higgstv.db?cache=shared&mode=rwc"
  database: "higgstv"
```

## SQLite → MongoDB 資料遷移

目前尚未實作 SQLite 到 MongoDB 的遷移工具。如果需要，可以：

1. 使用 MongoDB 的 `mongoimport` 工具
2. 或實作類似的遷移工具（參考 `cmd/migrate/migrate_mongodb_to_sqlite.go`）

## 測試環境設定

### SQLite 測試

測試會自動使用記憶體資料庫，不需要任何設定：

```bash
# 設定環境變數（可選）
export HIGGSTV_DATABASE_TYPE=sqlite

# 執行測試
go test ./tests/...
```

### MongoDB 測試

測試會使用 `{database}_test` 資料庫：

```bash
# 設定環境變數（可選）
export HIGGSTV_DATABASE_TYPE=mongodb

# 執行測試
go test ./tests/...
```

## 資料驗證

遷移完成後，建議驗證資料：

### 檢查資料庫連線

```bash
# 使用檢查工具
go run cmd/check_database/check_database.go
```

### 手動驗證

```bash
# SQLite
sqlite3 ./data/migrated_higgstv.db "SELECT COUNT(*) FROM users;"
sqlite3 ./data/migrated_higgstv.db "SELECT COUNT(*) FROM channels;"

# MongoDB
mongosh --eval "db.users.countDocuments()"
mongosh --eval "db.channels.countDocuments()"
```

## 注意事項

1. **資料一致性**：
   - 遷移過程中請確保應用程式已停止
   - 遷移完成後驗證資料完整性

2. **外鍵約束**：
   - SQLite 會自動處理外鍵約束
   - 確保遷移順序正確（先使用者，後頻道）

3. **重複資料**：
   - 遷移工具會自動處理重複資料（使用 `INSERT OR REPLACE`）
   - 如果資料已存在，會更新而非報錯

4. **資料類型轉換**：
   - MongoDB 的 ObjectID 會轉換為字串
   - 日期時間會自動轉換
   - 陣列會正規化為關聯表

## 故障排除

### 遷移失敗

如果遷移失敗，檢查：

1. **MongoDB 連線**：
   ```bash
   mongosh "mongodb://localhost:27017"
   ```

2. **SQLite 檔案權限**：
   ```bash
   ls -la ./data/migrated_higgstv.db
   ```

3. **日誌輸出**：
   遷移工具會顯示詳細的錯誤訊息

### 測試失敗

如果測試失敗，檢查：

1. **資料庫類型配置**：
   ```bash
   echo $HIGGSTV_DATABASE_TYPE
   ```

2. **資料庫連線**：
   ```bash
   go run cmd/check_database/check_database.go
   ```

3. **測試日誌**：
   ```bash
   go test -v ./tests/...
   ```

## 總結

- ✅ **測試環境**：不需要資料遷移，測試會自動建立資料
- ✅ **生產環境**：使用 `cmd/migrate/migrate_mongodb_to_sqlite.go` 工具遷移資料
- ✅ **驗證**：使用 `cmd/check_database/check_database.go` 檢查資料庫狀態

