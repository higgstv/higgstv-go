# 資料庫遷移指南

## 概述

HiggsTV Go API Server 使用遷移系統來管理資料庫結構變更。系統支援 MongoDB 和 SQLite 兩種資料庫。

## 遷移系統

遷移系統會自動執行所有未執行的遷移，確保資料庫結構與程式碼同步。系統會根據配置的資料庫類型自動選擇對應的遷移實作。

## 遷移記錄

遷移記錄儲存在 `migrations` collection/table 中，包含：
- `_id` / `id`: 遷移 ID
- `description`: 遷移描述
- `executed_at`: 執行時間

系統會根據資料庫類型自動選擇對應的儲存方式（MongoDB collection 或 SQLite table）。

## 自動執行

應用程式啟動時會自動執行所有未執行的遷移：

```go
migration.RunMigrationsWithTimeout(db)
```

## 手動執行遷移

如果需要手動執行遷移，應用程式啟動時會自動執行所有未執行的遷移。

如果需要手動執行資料遷移（MongoDB → SQLite），可以使用遷移工具：

```bash
go run cmd/migrate/migrate_mongodb_to_sqlite.go
```

## 建立新遷移

1. 在 `internal/migration/migration.go` 的 `migrations` slice 中添加新遷移：

```go
{
    ID:          "002_add_new_field",
    Description: "新增欄位到 User model",
    Up: func(ctx context.Context, db database.Database) error {
        // 執行遷移邏輯（使用抽象介面，支援兩種資料庫）
        usersColl := db.Collection("users")
        _, err := usersColl.UpdateMany(ctx, database.Filter{}, database.Update{
            Set: map[string]interface{}{"new_field": "default_value"},
        })
        return err
    },
    Down: func(ctx context.Context, db database.Database) error {
        // 向下遷移（可選）
        return nil
    },
},
```

2. 遷移會在下一次應用程式啟動時自動執行

## 遷移最佳實踐

1. **保持遷移簡單**：每個遷移應該只做一件事
2. **可逆遷移**：盡可能實作 Down 方法
3. **測試遷移**：在測試環境先測試遷移
4. **備份資料**：執行遷移前備份資料庫
5. **遷移順序**：使用時間戳或序號確保遷移順序

## 遷移範例

### 新增索引

索引管理已統一在 `internal/database/indexes_unified.go` 中處理，會根據資料庫類型自動建立對應的索引。如需新增索引，請更新該檔案。

### 新增欄位

```go
{
    ID:          "004_add_user_avatar",
    Description: "新增 avatar 欄位到 User",
    Up: func(ctx context.Context, db database.Database) error {
        usersColl := db.Collection("users")
        _, err := usersColl.UpdateMany(ctx, database.Filter{}, database.Update{
            Set: map[string]interface{}{"avatar": ""},
        })
        return err
    },
    Down: func(ctx context.Context, db database.Database) error {
        usersColl := db.Collection("users")
        // SQLite 不支援 $unset，需要手動處理
        if db.Type() == database.DatabaseTypeSQLite {
            // SQLite 處理邏輯
            return nil
        }
        _, err := usersColl.UpdateMany(ctx, database.Filter{}, database.Update{
            Unset: map[string]interface{}{"avatar": ""},
        })
        return err
    },
},
```

## 故障排除

### 遷移失敗

如果遷移失敗：
1. 檢查日誌了解失敗原因
2. 修復遷移邏輯
3. 手動移除失敗的遷移記錄（如果需要）
4. 重新執行遷移

### 檢查遷移狀態

查詢 `migrations` collection/table：

**MongoDB:**
```javascript
db.migrations.find().sort({_id: 1})
```

**SQLite:**
```sql
SELECT * FROM migrations ORDER BY id;
```

## 注意事項

1. 遷移應該在應用程式啟動前完成
2. 生產環境執行遷移前務必備份
3. 大型遷移考慮分批執行
4. 遷移應該盡可能快速執行

