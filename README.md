# HiggsTV Go API Server

HiggsTV API Server 的 Golang 實作版本，採用分層架構設計，提供完整的 RESTful API 服務。

## 專案結構

```
higgstv-go/
├── cmd/
│   ├── server/              # 主程式入口
│   ├── check_database/      # 資料庫連線檢查工具
│   ├── check_mongodb/       # MongoDB 檢查工具
│   └── migrate/             # MongoDB 到 SQLite 遷移工具
├── internal/
│   ├── api/                 # API 層
│   │   ├── handlers/        # 請求處理器
│   │   ├── middleware/      # 中介層（認證、CORS、日誌、錯誤處理）
│   │   ├── response/        # 統一回應格式
│   │   └── router.go        # 路由設定
│   ├── config/              # 配置管理
│   ├── database/            # 資料庫抽象層
│   │   ├── interface.go     # 資料庫介面定義
│   │   ├── factory.go       # 資料庫工廠
│   │   ├── mongodb.go       # MongoDB 實作
│   │   └── sqlite.go        # SQLite 實作
│   ├── models/              # 資料模型（User, Channel, Program）
│   ├── repository/          # 資料存取層（支援 MongoDB 和 SQLite）
│   └── service/             # 業務邏輯層
├── pkg/                     # 共用套件
│   ├── errors/              # 錯誤定義
│   ├── logger/              # 日誌記錄（使用 zap）
│   ├── mail/                # 郵件服務
│   ├── session/             # Session 管理
│   ├── uuidutil/            # UUID 工具
│   ├── validator/           # 請求驗證器
│   └── youtube/             # YouTube 相關工具
├── tests/                   # 測試檔案
├── config/                  # 配置檔案
│   ├── config.yaml          # 預設配置
│   └── config.example.yaml  # 配置範例
├── Dockerfile               # Docker 建置檔
├── docker-compose.yml       # Docker Compose 配置
├── Makefile                 # 建置腳本
└── .golangci.yml           # Linter 配置
```

## 功能特色

- ✅ **完整的 API 實作**：所有認證、頻道、節目相關 API
- ✅ **分層架構**：清晰的 Repository → Service → Handler 架構
- ✅ **多資料庫支援**：支援 MongoDB 和 SQLite，可透過配置切換
- ✅ **資料庫抽象層**：統一的資料庫介面，易於擴展
- ✅ **錯誤處理**：統一的錯誤處理機制
- ✅ **日誌記錄**：使用 zap 進行結構化日誌記錄
- ✅ **Session 管理**：Cookie-based session 認證
- ✅ **CORS 支援**：跨域請求支援
- ✅ **請求驗證**：使用 validator 進行請求參數驗證
- ✅ **Docker 支援**：完整的 Docker 和 Docker Compose 配置
- ✅ **測試框架**：整合測試範例，支援測試隔離
- ✅ **MongoDB 相容性**：支援 UUID binary 和字串兩種格式，可無縫讀取舊資料庫
- ✅ **資料遷移工具**：提供 MongoDB 到 SQLite 的遷移工具

## 快速開始

### 1. 安裝依賴

```bash
go mod download
```

### 2. 設定配置

複製並編輯配置檔案：

```bash
cp config/config.example.yaml config/config.yaml
# 編輯 config.yaml，設定資料庫類型、URI、Session Secret 等
```

**資料庫配置範例：**

**SQLite（預設）：**
```yaml
database:
  type: "sqlite"
  uri: "file:./data/higgstv.db?cache=shared&mode=rwc"
  database: "higgstv"
```

**MongoDB：**
```yaml
database:
  type: "mongodb"
  uri: "mongodb://localhost:27017"
  database: "higgstv"
```

### 3. 啟動資料庫（使用 Docker，僅 MongoDB 需要）

**SQLite（預設）：** 無需額外啟動，會自動建立資料庫檔案

**MongoDB：**
```bash
docker-compose up -d mongodb
```

### 4. 執行

**開發模式：**
```bash
make dev
# 或
go run cmd/server/main.go
```

**建置後執行：**
```bash
make build
./bin/higgstv-go
```

**使用 Docker Compose：**
```bash
# 使用 SQLite（預設，不需要 MongoDB）
docker-compose up

# 使用 MongoDB
docker-compose -f docker-compose.mongodb.yml up
```

## 環境變數

可以透過環境變數覆蓋配置：

- `HIGGSTV_SERVER_PORT`: 伺服器埠號（預設：8080）
- `HIGGSTV_SERVER_ENV`: 環境（development/production）
- `HIGGSTV_DATABASE_TYPE`: 資料庫類型（sqlite/mongodb，預設：sqlite）
- `HIGGSTV_DATABASE_URI`: 資料庫連線 URI
  - SQLite（預設）: `file:./data/higgstv.db?cache=shared&mode=rwc`
  - MongoDB: `mongodb://localhost:27017`
- `HIGGSTV_DATABASE_DATABASE`: 資料庫名稱
- `HIGGSTV_SESSION_SECRET`: Session 加密金鑰

## API 端點

### 認證相關
- `POST /apis/signin` - 登入
- `GET /apis/signout` - 登出
- `POST /apis/signup` - 註冊
- `POST /apis/change_password` - 變更密碼（需登入）
- `POST /apis/forget_password` - 忘記密碼
- `POST /apis/reset_password` - 重設密碼

### 頻道相關
- `POST /apis/addchannel` - 新增頻道（需登入）
- `GET /apis/getownchannels` - 取得自己的頻道列表（需登入）
- `GET /apis/getchannels` - 取得頻道列表（支援過濾）
- `GET /apis/getchannel/:id` - 取得單一頻道
- `GET /apis/getchannelinfo/:id` - 取得頻道資訊（含擁有者）
- `POST /apis/savechannel` - 儲存頻道（需登入）
- `POST /apis/setchannelowner` - 設定頻道擁有者（需登入）

### 節目相關
- `POST /apis/addprog` - 新增節目（需登入）
- `POST /apis/saveprog` - 儲存節目（需登入）
- `POST /apis/delprog` - 刪除節目（需登入）
- `POST /apis/progmoveto` - 移動節目（需登入）
- `POST /apis/prog/saveorder` - 儲存節目順序（需登入）

### Pick API（Bookmarklet）
- `GET /apis/pickprog` - Pick 節目（支援 JSONP，需登入）

詳細的 API 文件請參考 `docs/API_REFERENCE.md`。

## 開發

### 執行測試

```bash
make test
# 或
go test ./...
```

### 測試覆蓋率

```bash
make test-coverage
```

### 程式碼檢查

```bash
make lint
# 或
golangci-lint run
```

### 格式化程式碼

```bash
make fmt
# 或
go fmt ./...
```

### 建置

```bash
make build
```

### 清理

```bash
make clean
```

## Docker 部署

### 建置映像檔

```bash
docker build -t higgstv-go .
```

### 使用 Docker Compose

**使用 SQLite（預設）：**

```bash
# 啟動服務（僅 API，不需要 MongoDB）
docker-compose up -d

# 查看日誌
docker-compose logs -f api

# 停止服務
docker-compose down
```

**使用 MongoDB：**

```bash
# 啟動所有服務（MongoDB + API）
docker-compose -f docker-compose.mongodb.yml up -d

# 查看日誌
docker-compose -f docker-compose.mongodb.yml logs -f api

# 停止服務
docker-compose -f docker-compose.mongodb.yml down
```

**注意事項：**
- SQLite（預設）資料庫會儲存在 `./data` 目錄中（會自動建立）
- MongoDB 版本需要 MongoDB 服務運行
- SQLite 版本不需要 MongoDB，適合輕量級部署和開發環境

## 專案狀態

### Phase 1：基礎架構 ✅
- [x] 專案結構建立
- [x] 資料模型定義
- [x] Repository 層實作
- [x] Service 層實作
- [x] Handler 層實作
- [x] 路由設定

### Phase 2：進階功能 ✅
- [x] 錯誤處理機制
- [x] 日誌記錄系統
- [x] CORS 支援
- [x] 請求驗證
- [x] 測試框架
- [x] Docker 支援

### Phase 3：生產就緒功能 ✅
- [x] 健康檢查端點
- [x] 資料庫索引自動建立
- [x] Rate Limiting
- [x] 配置驗證
- [x] 資料庫連線池配置
- [x] Request ID 追蹤
- [x] JSONP 安全驗證
- [x] MongoDB UUID 格式相容性（支援 binary 和字串格式）

### Phase 4：監控與部署 ✅
- [x] Prometheus 指標收集
- [x] CI/CD 配置（GitHub Actions）
- [x] 整合測試
- [x] 環境變數文件
- [x] 部署腳本

### Phase 5：多資料庫支援 ✅
- [x] 資料庫抽象層設計
- [x] SQLite 完整支援
- [x] MongoDB 和 SQLite 並存
- [x] 資料遷移工具
- [x] 測試隔離機制
- [x] 完整的文檔說明

## 監控與指標

### Prometheus 指標

應用程式提供 Prometheus 指標端點 `/metrics`，包含：

- `http_requests_total`: HTTP 請求總數（按方法、路徑、狀態碼）
- `http_request_duration_seconds`: HTTP 請求持續時間
- `db_operations_total`: 資料庫操作總數
- `db_operation_duration_seconds`: 資料庫操作持續時間

### 健康檢查

- `GET /health`: 基本健康檢查
- `GET /ready`: 就緒檢查（包含資料庫狀態）

## 技術棧

- **Web 框架**: Gin
- **資料庫**: 
  - MongoDB (官方 Go Driver)
  - SQLite (go-sqlite3)
  - 支援透過配置切換資料庫類型
- **Session**: gorilla/sessions
- **日誌**: zap
- **驗證**: go-playground/validator
- **配置**: Viper
- **測試**: testify
- **監控**: Prometheus
- **CI/CD**: GitHub Actions

## 資料庫支援

### MongoDB
- 完整的 MongoDB 支援
- 支援 UUID binary 和字串兩種格式
- 自動建立索引
- 交易支援

### SQLite
- 完整的 SQLite 支援
- 自動建立資料庫結構
- 外鍵約束支援
- 交易支援
- 適合開發和測試環境

### 資料遷移
提供 MongoDB 到 SQLite 的遷移工具：
```bash
go run cmd/migrate/migrate_mongodb_to_sqlite.go [sqlite_path]
```

詳細說明請參考 `docs/DATA_MIGRATION_GUIDE.md`。

## 資料庫支援

### MongoDB
- 完整的 MongoDB 支援
- 支援 UUID binary 和字串兩種格式
- 自動建立索引
- 交易支援

### SQLite
- 完整的 SQLite 支援
- 自動建立資料庫結構
- 外鍵約束支援
- 交易支援
- 適合開發和測試環境

### 資料遷移
提供 MongoDB 到 SQLite 的遷移工具：
```bash
go run cmd/migrate/migrate_mongodb_to_sqlite.go [sqlite_path]
```

詳細說明請參考 `docs/DATA_MIGRATION_GUIDE.md`。

## 授權

與 HiggsTV 專案相同。
