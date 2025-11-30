# 測試指南

## 概述

本專案支援 MongoDB 和 SQLite 兩種資料庫的測試。測試系統已完全抽象化，可以根據配置自動選擇資料庫類型。

## 測試架構

### 測試檔案結構

```
tests/
├── test_helper.go        # 測試輔助函數（支援兩種資料庫）
├── auth_test.go          # 認證 API 測試
├── channel_test.go       # 頻道 API 測試
├── program_test.go       # 節目 API 測試
├── pick_test.go          # Pick API 測試
├── system_test.go        # 系統測試
└── benchmark_test.go     # 效能測試

internal/service/
└── auth_test.go          # Service 層測試
```

### 測試輔助函數

#### `SetupTestDB(t *testing.T)`
- 根據配置自動選擇資料庫類型（MongoDB 或 SQLite）
- SQLite 測試使用記憶體資料庫（`file::memory:`）
- MongoDB 測試使用 `{database}_test` 資料庫
- 自動建立索引和執行遷移

#### `CleanupTestDB(t *testing.T)`
- 清理測試資料庫
- SQLite 記憶體資料庫：關閉連線即可
- MongoDB：關閉連線（可選清理集合）

#### `getAuthCookie(t, router, username, email, password)`
- 註冊並登入使用者
- 返回 Cookie 用於後續 API 測試

## 配置測試環境

### 1. 使用 MongoDB 測試

```bash
# 設定環境變數
export HIGGSTV_DATABASE_TYPE=mongodb
export HIGGSTV_DATABASE_URI=mongodb://localhost:27017
export HIGGSTV_DATABASE_DATABASE=higgstv

# 執行測試
go test ./tests/...
go test ./internal/service/...
```

### 2. 使用 SQLite 測試

```bash
# 設定環境變數
export HIGGSTV_DATABASE_TYPE=sqlite
export HIGGSTV_DATABASE_URI=file::memory:?cache=shared
export HIGGSTV_DATABASE_DATABASE=higgstv

# 執行測試
go test ./tests/...
go test ./internal/service/...
```

### 3. 使用配置檔案

在 `config/config.yaml` 中設定：

```yaml
database:
  type: "sqlite"  # 或 "mongodb"
  uri: "file::memory:?cache=shared"  # SQLite 記憶體資料庫
  # uri: "mongodb://localhost:27017"  # MongoDB
  database: "higgstv"
```

## 執行測試

### 執行所有測試

```bash
# 執行所有測試
go test ./tests/... -v

# 執行特定測試套件
go test ./tests/... -run TestSignUp -v

# 執行 Service 層測試
go test ./internal/service/... -v
```

### 執行特定測試

```bash
# 認證測試
go test ./tests/... -run TestSignUp -v
go test ./tests/... -run TestSignIn -v

# 頻道測試
go test ./tests/... -run TestGetChannel -v
go test ./tests/... -run TestAddChannel -v

# 節目測試
go test ./tests/... -run TestSaveProgram -v
```

### 並行測試

```bash
# 並行執行測試（注意：SQLite 記憶體資料庫不支援並行）
go test ./tests/... -parallel 4 -v
```

## 測試資料庫設定

### SQLite 測試資料庫

- **類型**: 記憶體資料庫（`file::memory:`）
- **優點**: 
  - 快速執行
  - 自動清理（測試結束後自動刪除）
  - 不需要額外設定
- **限制**: 
  - 不支援並行測試
  - 測試間無法共享資料

### MongoDB 測試資料庫

- **類型**: `{database}_test` 資料庫
- **優點**: 
  - 支援並行測試
  - 可以手動檢查測試資料
  - 更接近生產環境
- **限制**: 
  - 需要 MongoDB 服務運行
  - 需要手動清理（可選）

## 測試最佳實踐

### 1. 測試隔離

每個測試都應該：
- 使用 `SetupTestDB(t)` 建立獨立的資料庫連線
- 使用 `defer CleanupTestDB(t)` 清理資源
- 不依賴其他測試的資料

### 2. 測試資料

- 測試應該自己建立所需的資料
- 使用 `getAuthCookie` 建立測試使用者
- 避免使用硬編碼的資料 ID

### 3. 錯誤處理

- 使用 `require` 進行關鍵斷言（會中斷測試）
- 使用 `assert` 進行一般斷言（不會中斷測試）
- 檢查錯誤訊息以確保正確性

### 4. 測試命名

- 使用描述性的測試名稱
- 使用 `TestFunctionName_Scenario` 格式
- 使用子測試（`t.Run`）組織相關測試

## 效能測試

### 執行效能測試

```bash
# 執行所有效能測試
go test ./tests/... -bench=. -benchmem

# 執行特定效能測試
go test ./tests/... -bench=BenchmarkSignIn -benchmem
```

### 效能測試配置

效能測試會根據配置自動選擇資料庫類型：
- MongoDB: 使用 `{database}_benchmark` 資料庫
- SQLite: 使用記憶體資料庫

## 常見問題

### 1. 測試失敗：無法連線資料庫

**問題**: `Failed to connect to database`

**解決方案**:
- 檢查 MongoDB 是否運行（如果使用 MongoDB）
- 檢查配置檔案或環境變數
- 檢查網路連線

### 2. 測試失敗：遷移記錄重複

**問題**: `duplicate key error collection: migrations`

**解決方案**:
- 這是正常的，遷移系統會自動處理重複記錄
- 測試會繼續執行，不會中斷

### 3. SQLite 測試失敗：並行執行

**問題**: SQLite 記憶體資料庫不支援並行測試

**解決方案**:
- 不要使用 `-parallel` 參數
- 或使用 MongoDB 進行並行測試

### 4. 測試資料殘留

**問題**: MongoDB 測試資料殘留

**解決方案**:
```bash
# 手動清理測試資料庫
mongosh --eval "db.getSiblingDB('higgstv_test').dropDatabase()"
```

## CI/CD 測試

### GitHub Actions

測試會在 CI/CD 中自動執行：

```yaml
# 測試 MongoDB
- name: Test with MongoDB
  env:
    HIGGSTV_DATABASE_TYPE: mongodb
    HIGGSTV_DATABASE_URI: mongodb://localhost:27017
  run: go test ./tests/... -v

# 測試 SQLite
- name: Test with SQLite
  env:
    HIGGSTV_DATABASE_TYPE: sqlite
    HIGGSTV_DATABASE_URI: file::memory:?cache=shared
  run: go test ./tests/... -v
```

## 測試覆蓋率

### 生成測試覆蓋率報告

```bash
# 生成覆蓋率報告
go test ./tests/... -coverprofile=coverage.out

# 查看覆蓋率報告
go tool cover -html=coverage.out
```

### 目標覆蓋率

- 單元測試：> 80%
- 整合測試：> 70%
- 整體覆蓋率：> 75%

## 測試資料庫遷移

測試會自動執行資料庫遷移：
- 建立必要的資料表/集合
- 建立索引
- 記錄遷移歷史

如果遷移失敗，測試會記錄警告但繼續執行（除非是關鍵錯誤）。

## 總結

測試系統已完全抽象化，支援：

- ✅ MongoDB 和 SQLite 兩種資料庫
- ✅ 自動選擇資料庫類型
- ✅ 自動建立和清理測試環境
- ✅ 統一的測試輔助函數
- ✅ 完整的錯誤處理

使用 `SetupTestDB` 和 `CleanupTestDB` 可以確保測試的一致性和可重複性。

