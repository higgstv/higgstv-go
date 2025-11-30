# SQLite 支援實作進度

## 已完成項目 ✅

### 1. 資料庫抽象層設計 ✅
- ✅ 建立 `internal/database/interface.go` - 定義統一的資料庫介面
- ✅ 建立 `internal/database/factory.go` - 資料庫工廠模式
- ✅ 定義 `Database`、`Collection`、`Filter`、`Sort`、`Update` 等抽象介面

### 2. MongoDB 包裝器 ✅
- ✅ 建立 `internal/database/mongodb.go` - 將 MongoDB driver 包裝為抽象介面
- ✅ 實作所有 Collection 方法
- ✅ 實作交易支援

### 3. SQLite 基礎架構 ✅
- ✅ 建立 `internal/database/sqlite.go` - SQLite 基礎實作
- ✅ 實作資料庫連線管理
- ✅ 實作 Schema 建立（CREATE TABLE）
- ✅ 實作索引建立和管理
- ✅ 實作交易支援

### 4. 配置系統更新 ✅
- ✅ 更新 `internal/config/config.go` - 新增 `database.type` 欄位
- ✅ 更新 `internal/config/validator.go` - 驗證資料庫類型
- ✅ 支援環境變數 `HIGGSTV_DATABASE_TYPE`

### 5. 索引管理系統 ✅
- ✅ 建立 `internal/database/indexes_unified.go` - 統一的索引管理
- ✅ 支援 MongoDB 和 SQLite 兩種資料庫

## 進行中項目 🚧

### SQLite Collection 實作
SQLite 的 Collection 方法需要根據不同的 collection 名稱實作不同的查詢邏輯，因為：
- MongoDB 使用文件結構（內嵌陣列）
- SQLite 使用關聯式結構（正規化表）

需要實作的 Collection：
- `users` - 使用者表
- `channels` - 頻道表
- `programs` - 節目表
- `counters` - 計數器表

## 待完成項目 📋

### 1. SQLite Repository 實作
需要為每個 Repository 建立 SQLite 專屬實作：

#### UserRepository (SQLite)
- `FindByUsername` - 查詢使用者
- `FindByEmail` - 查詢 Email
- `Exists` - 檢查使用者是否存在
- `Create` - 建立使用者
- `UpdatePassword` - 更新密碼
- `SetAccessKey` - 設定 access_key
- `ChangePasswordWithAccessKey` - 使用 access_key 重設密碼
- `AddChannel` - 新增頻道到使用者（需要操作 `user_channels` 表）
- `SetUnclassifiedChannel` - 設定未分類頻道
- `GetUsersBasicInfo` - 取得使用者基本資訊

#### ChannelRepository (SQLite)
- `FindByID` - 查詢頻道（需要 JOIN 查詢 tags、owners、permissions、programs）
- `Create` - 建立頻道（需要插入到多個表）
- `Update` - 更新頻道
- `ListChannels` - 列出頻道（需要複雜的 JOIN 查詢）
- `IsAdmin` - 檢查是否為管理員
- `AddOwners` - 新增擁有者

#### ProgramRepository (SQLite)
- `GetNextProgramID` - 取得下一個節目 ID（使用 counters 表）
- `AddProgram` - 新增節目（插入到 programs 表）
- `UpdateProgram` - 更新節目
- `DeletePrograms` - 刪除節目
- `SetOrder` - 設定節目順序（使用 channel_program_order 表）

### 2. Repository 層重構
需要建立 Repository 工廠，根據資料庫類型選擇對應的實作：

```go
// internal/repository/factory.go
func NewUserRepository(db Database) UserRepository {
    switch db.Type() {
    case DatabaseTypeMongoDB:
        return NewMongoDBUserRepository(db)
    case DatabaseTypeSQLite:
        return NewSQLiteUserRepository(db)
    }
}
```

### 3. Service 層調整
移除 MongoDB 特定類型（`bson.M`、`bson.D`），改用通用的 `map[string]interface{}` 和 `Sort`。

需要修改的檔案：
- `internal/service/channel.go` - 移除 `bson.M`、`bson.D`
- `internal/service/program.go` - 移除 `bson.M`

### 4. Handlers 和 Router 更新
更新所有 Handlers 和 Router 使用新的抽象層：

- `internal/api/router.go` - 改用 `Database` 介面
- `internal/api/handlers/*.go` - 改用 `Database` 介面
- `cmd/server/main.go` - 使用 `NewDatabase` 建立資料庫連線

### 5. 遷移系統
- 更新 `internal/migration/migration.go` 支援兩種資料庫
- 建立 SQLite 專屬的遷移腳本

### 6. 測試更新
- 更新 `tests/test_helper.go` 支援兩種資料庫
- 為 SQLite 建立專屬測試

## 技術挑戰

### 1. 資料結構差異
MongoDB 和 SQLite 的資料結構差異很大：

**MongoDB (文件結構)**:
```json
{
  "_id": "channel123",
  "name": "我的頻道",
  "tags": [1, 2, 3],
  "owners": ["user1", "user2"],
  "contents": [
    {"_id": 1, "name": "節目1"},
    {"_id": 2, "name": "節目2"}
  ]
}
```

**SQLite (關聯式結構)**:
```sql
channels: id, name, ...
channel_tags: channel_id, tag
channel_owners: channel_id, user_id
programs: id, channel_id, name, ...
```

### 2. 查詢轉換
需要將 MongoDB 的查詢轉換為 SQL：

- `{"owners": "user1"}` → `EXISTS (SELECT 1 FROM channel_owners WHERE channel_id = channels.id AND user_id = ?)`
- `{"name": {"$regex": "test", "$options": "i"}}` → `name LIKE '%test%'`
- `{"$or": [{"name": "test"}, {"desc": "test"}]}` → `name = ? OR desc = ?`

### 3. 更新操作轉換
MongoDB 的更新操作需要轉換為 SQL：

- `{"$set": {"name": "new"}}` → `UPDATE channels SET name = ? WHERE id = ?`
- `{"$addToSet": {"owners": "user1"}}` → `INSERT INTO channel_owners (channel_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING`
- `{"$pull": {"owners": "user1"}}` → `DELETE FROM channel_owners WHERE channel_id = ? AND user_id = ?`

## 建議的實作順序

1. **完成 SQLite Repository 實作**（最關鍵）
   - 先實作 UserRepository (SQLite)
   - 再實作 ChannelRepository (SQLite)
   - 最後實作 ProgramRepository (SQLite)

2. **建立 Repository 工廠**
   - 統一 Repository 建立方式

3. **更新 Service 層**
   - 移除 MongoDB 特定類型

4. **更新 Handlers 和 Router**
   - 使用新的抽象層

5. **更新主程式**
   - 使用 `NewDatabase` 建立資料庫連線

6. **測試和文件**
   - 更新測試檔案
   - 更新文件

## 使用範例

### 配置檔案 (config.yaml)
```yaml
database:
  type: "sqlite"  # 或 "mongodb"
  uri: "file:./data/higgstv.db?cache=shared&mode=rwc"
  database: "higgstv"
```

### 環境變數
```bash
export HIGGSTV_DATABASE_TYPE=sqlite
export HIGGSTV_DATABASE_URI=file:./data/higgstv.db
export HIGGSTV_DATABASE_DATABASE=higgstv
```

### 程式碼使用
```go
import "github.com/higgstv/higgstv-go/internal/database"

// 建立資料庫連線
db, err := database.NewDatabase(ctx, database.DatabaseConfig{
    Type:     database.DatabaseTypeSQLite,
    URI:      "file:./data/higgstv.db",
    Database: "higgstv",
})

// 使用 Repository
userRepo := repository.NewUserRepository(db)
```

## 注意事項

1. **效能考量**: SQLite 不適合高併發場景，建議用於開發、測試或小型部署
2. **功能限制**: SQLite 不支援某些 MongoDB 進階功能（如 Aggregation Pipeline）
3. **資料遷移**: MongoDB 到 SQLite 的資料遷移需要額外工具
4. **交易支援**: SQLite 支援交易，但寫入操作會鎖定整個資料庫

