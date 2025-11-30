# SQLite 支援遷移計劃

## 概述

本文件說明如何為 HiggsTV Go 專案新增 SQLite 資料庫支援，同時保持與現有 MongoDB 的相容性。

## 架構設計

### 1. 資料庫抽象層

建立統一的資料庫介面，讓 Repository 層可以無縫切換不同的資料庫實作。

```
internal/database/
├── interface.go      # 資料庫抽象介面
├── mongodb.go        # MongoDB 實作
├── sqlite.go         # SQLite 實作
└── factory.go        # 資料庫工廠（根據配置選擇實作）
```

### 2. Repository 層重構

將 Repository 層改為使用抽象介面，而非直接依賴 MongoDB driver。

### 3. Service 層調整

移除 MongoDB 特定類型（bson.M, bson.D），改用通用的 map 和 slice。

### 4. 配置系統擴展

新增資料庫類型選擇（mongodb/sqlite）和對應的連線配置。

## 實作步驟

### 階段 1: 建立抽象層

1. 定義 `Database` 介面
2. 定義 `Collection` 介面（對應 MongoDB Collection 或 SQLite Table）
3. 定義查詢過濾器介面（Filter, Sort, Update）

### 階段 2: 實作 SQLite 驅動

1. 實作 SQLite 連線管理
2. 實作 SQLite Collection（Table）操作
3. 實作 SQLite 查詢轉換（Filter -> SQL WHERE, Sort -> SQL ORDER BY）

### 階段 3: 重構現有程式碼

1. 重構 Repository 層
2. 重構 Service 層
3. 更新 Handlers
4. 更新配置和驗證

### 階段 4: 遷移系統

1. 實作 SQLite 遷移系統
2. 建立 SQLite Schema 定義
3. 建立資料遷移工具（MongoDB -> SQLite）

### 階段 5: 測試與文件

1. 更新測試檔案
2. 更新文件
3. 建立使用範例

## 資料庫 Schema 設計（SQLite）

### users 表
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    access_key TEXT,
    unclassified_channel TEXT,
    created DATETIME NOT NULL,
    last_modified DATETIME NOT NULL
);

CREATE TABLE user_channels (
    user_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    PRIMARY KEY (user_id, channel_id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### channels 表
```sql
CREATE TABLE channels (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    desc TEXT,
    contents_seq TEXT,
    cover_default TEXT,
    created DATETIME NOT NULL,
    last_modified DATETIME NOT NULL
);

CREATE TABLE channel_tags (
    channel_id TEXT NOT NULL,
    tag INTEGER NOT NULL,
    PRIMARY KEY (channel_id, tag),
    FOREIGN KEY (channel_id) REFERENCES channels(id)
);

CREATE TABLE channel_owners (
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (channel_id, user_id),
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE channel_permissions (
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    admin INTEGER NOT NULL DEFAULT 0,
    read INTEGER NOT NULL DEFAULT 0,
    write INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (channel_id, user_id),
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### programs 表
```sql
CREATE TABLE programs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id TEXT NOT NULL,
    name TEXT NOT NULL,
    desc TEXT,
    duration INTEGER,
    type TEXT NOT NULL,
    youtube_id TEXT,
    created DATETIME NOT NULL,
    last_modified DATETIME NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES channels(id)
);

CREATE TABLE program_tags (
    program_id INTEGER NOT NULL,
    tag INTEGER NOT NULL,
    PRIMARY KEY (program_id, tag),
    FOREIGN KEY (program_id) REFERENCES programs(id)
);

CREATE TABLE channel_program_order (
    channel_id TEXT NOT NULL,
    program_id INTEGER NOT NULL,
    order_index INTEGER NOT NULL,
    PRIMARY KEY (channel_id, program_id),
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (program_id) REFERENCES programs(id)
);
```

### counters 表
```sql
CREATE TABLE counters (
    id TEXT PRIMARY KEY,
    seq INTEGER NOT NULL DEFAULT 0
);
```

### migrations 表
```sql
CREATE TABLE migrations (
    id TEXT PRIMARY KEY,
    description TEXT,
    executed_at DATETIME NOT NULL
);
```

## 索引設計

```sql
-- users 索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_access_key ON users(access_key) WHERE access_key IS NOT NULL;

-- channels 索引
CREATE INDEX idx_channels_owners ON channel_owners(channel_id, user_id);
CREATE INDEX idx_channels_last_modified ON channels(last_modified DESC);
CREATE INDEX idx_channels_name ON channels(name);

-- programs 索引
CREATE INDEX idx_programs_channel_id ON programs(channel_id);
CREATE INDEX idx_program_tags_program_id ON program_tags(program_id);
```

## 查詢轉換規則

### Filter 轉換

MongoDB Filter -> SQL WHERE:
- `{"username": "test"}` -> `WHERE username = ?`
- `{"$or": [{"username": "test"}, {"email": "test@example.com"}]}` -> `WHERE username = ? OR email = ?`
- `{"name": {"$regex": "test", "$options": "i"}}` -> `WHERE name LIKE ?` (使用 `%test%`)
- `{"owners": "user123"}` -> `WHERE EXISTS (SELECT 1 FROM channel_owners WHERE channel_id = channels.id AND user_id = ?)`

### Sort 轉換

MongoDB Sort -> SQL ORDER BY:
- `bson.D{{"last_modified", -1}}` -> `ORDER BY last_modified DESC`
- `bson.D{{"name", 1}}` -> `ORDER BY name ASC`

### Update 轉換

MongoDB Update -> SQL UPDATE:
- `{"$set": {"name": "new name"}}` -> `UPDATE channels SET name = ? WHERE id = ?`
- `{"$addToSet": {"owners": "user123"}}` -> `INSERT INTO channel_owners (channel_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING`

## 配置範例

### config.yaml
```yaml
database:
  type: "sqlite"  # 或 "mongodb"
  uri: "file:./data/higgstv.db?cache=shared&mode=rwc"  # SQLite
  # uri: "mongodb://localhost:27017"  # MongoDB
  database: "higgstv"
```

### 環境變數
```bash
HIGGSTV_DATABASE_TYPE=sqlite
HIGGSTV_DATABASE_URI=file:./data/higgstv.db
HIGGSTV_DATABASE_DATABASE=higgstv
```

## 遷移注意事項

1. **UUID 格式**: MongoDB 使用 Base64 UUID，SQLite 使用字串 UUID，需要保持一致
2. **時間格式**: MongoDB 使用 `time.Time`，SQLite 使用 `DATETIME`，需要正確轉換
3. **陣列欄位**: MongoDB 的陣列在 SQLite 中需要正規化為關聯表
4. **內嵌文件**: MongoDB 的內嵌文件（如 Channel.Contents）在 SQLite 中需要分離為獨立表
5. **交易支援**: SQLite 支援交易，需要確保操作的一致性

## 測試策略

1. **單元測試**: 為每個 Repository 方法建立測試
2. **整合測試**: 測試完整的 API 流程
3. **效能測試**: 比較 MongoDB 和 SQLite 的效能
4. **資料一致性測試**: 確保兩種資料庫的資料結構一致

## 向後相容性

- 保持現有 MongoDB 支援
- 透過配置選擇資料庫類型
- API 介面保持不變
- 測試可以同時支援兩種資料庫

## 風險評估

1. **資料遷移風險**: MongoDB 到 SQLite 的資料遷移可能會有資料損失
2. **效能風險**: SQLite 在高併發情況下效能可能不如 MongoDB
3. **功能限制**: SQLite 不支援某些 MongoDB 進階功能（如 Aggregation Pipeline）

## 時間估算

- 階段 1: 2-3 天
- 階段 2: 3-4 天
- 階段 3: 2-3 天
- 階段 4: 2-3 天
- 階段 5: 1-2 天

總計: 10-15 個工作天

