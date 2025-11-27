# 資料庫遷移指南

## 概述

HiggsTV Go API Server 使用遷移系統來管理資料庫結構變更。

## 遷移系統

遷移系統會自動執行所有未執行的遷移，確保資料庫結構與程式碼同步。

## 遷移記錄

遷移記錄儲存在 `migrations` collection 中，包含：
- `_id`: 遷移 ID
- `description`: 遷移描述
- `executed_at`: 執行時間

## 自動執行

應用程式啟動時會自動執行所有未執行的遷移：

```go
migration.RunMigrationsWithTimeout(db)
```

## 手動執行遷移

如果需要手動執行遷移，可以使用遷移工具：

```bash
go run cmd/migrate/main.go
```

## 建立新遷移

1. 在 `internal/migration/migration.go` 的 `migrations` slice 中添加新遷移：

```go
{
    ID:          "002_add_new_field",
    Description: "新增欄位到 User model",
    Up: func(ctx context.Context, db *mongo.Database) error {
        // 執行遷移邏輯
        usersColl := db.Collection("users")
        _, err := usersColl.UpdateMany(ctx, bson.M{}, bson.M{
            "$set": bson.M{"new_field": "default_value"},
        })
        return err
    },
    Down: func(ctx context.Context, db *mongo.Database) error {
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

```go
{
    ID:          "003_add_user_email_index",
    Description: "為 User email 欄位新增索引",
    Up: func(ctx context.Context, db *mongo.Database) error {
        usersColl := db.Collection("users")
        _, err := usersColl.Indexes().CreateOne(ctx, mongo.IndexModel{
            Keys: bson.M{"email": 1},
            Options: options.Index().SetUnique(true),
        })
        return err
    },
    Down: func(ctx context.Context, db *mongo.Database) error {
        usersColl := db.Collection("users")
        _, err := usersColl.Indexes().DropOne(ctx, "email_1")
        return err
    },
},
```

### 新增欄位

```go
{
    ID:          "004_add_user_avatar",
    Description: "新增 avatar 欄位到 User",
    Up: func(ctx context.Context, db *mongo.Database) error {
        usersColl := db.Collection("users")
        _, err := usersColl.UpdateMany(ctx, bson.M{}, bson.M{
            "$set": bson.M{"avatar": ""},
        })
        return err
    },
    Down: func(ctx context.Context, db *mongo.Database) error {
        usersColl := db.Collection("users")
        _, err := usersColl.UpdateMany(ctx, bson.M{}, bson.M{
            "$unset": bson.M{"avatar": ""},
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

查詢 `migrations` collection：

```javascript
db.migrations.find().sort({_id: 1})
```

## 注意事項

1. 遷移應該在應用程式啟動前完成
2. 生產環境執行遷移前務必備份
3. 大型遷移考慮分批執行
4. 遷移應該盡可能快速執行

