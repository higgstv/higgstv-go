# 快速測試指南

> **注意**：這是快速測試指南，適合快速驗證功能。如需完整的測試說明（包含 MongoDB 和 SQLite 配置、測試架構等），請參考 [TESTING_GUIDE.md](./TESTING_GUIDE.md)。

## 快速測試

### 1. 啟動伺服器

```bash
# 確保資料庫正在運行（MongoDB 或 SQLite）
# MongoDB: docker ps | grep mongo
# SQLite: 無需額外啟動

# 啟動伺服器
go run cmd/server/main.go
```

### 2. 測試健康檢查

```bash
# 基本健康檢查
curl http://localhost:8080/health

# 就緒檢查
curl http://localhost:8080/ready

# Prometheus 指標
curl http://localhost:8080/metrics
```

### 3. 測試註冊

```bash
curl -X POST http://localhost:8080/apis/signup \
  -H "Content-Type: application/json" \
  -d '{
    "invitation_code": "sixpens",
    "username": "testuser",
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

預期回應：
```json
{
  "state": 0,
  "ret": true
}
```

### 4. 測試登入

```bash
curl -X POST http://localhost:8080/apis/signin \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }' \
  -c cookies.txt
```

預期回應：
```json
{
  "state": 0,
  "ret": true
}
```

### 5. 測試需要認證的 API

```bash
# 取得自己的頻道列表
curl http://localhost:8080/apis/getownchannels -b cookies.txt

# 新增頻道
curl -X POST http://localhost:8080/apis/addchannel \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "My Channel",
    "tags": [1, 2]
  }'
```

## 錯誤測試

### 測試錯誤的邀請碼

```bash
curl -X POST http://localhost:8080/apis/signup \
  -H "Content-Type: application/json" \
  -d '{
    "invitation_code": "wrong",
    "username": "testuser",
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

預期回應：
```json
{
  "state": 1,
  "code": 2
}
```

### 測試缺少欄位

```bash
curl -X POST http://localhost:8080/apis/signin \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser"
  }'
```

預期回應：
```json
{
  "state": 1,
  "code": 0
}
```

### 測試錯誤密碼

```bash
curl -X POST http://localhost:8080/apis/signin \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "wrongpassword"
  }'
```

預期回應：
```json
{
  "state": 0,
  "ret": false
}
```

## 自動化測試

### 執行單元測試

```bash
go test ./internal/service/...
```

### 執行整合測試

```bash
go test ./tests/...
```

### 執行效能測試

```bash
go test -bench=. ./tests/...
```

## 測試腳本

可以使用 `scripts/test-api.sh` 進行完整的 API 測試：

```bash
./scripts/test-api.sh
```

## 測試檢查清單

- [ ] 健康檢查端點正常
- [ ] 註冊功能正常
- [ ] 登入功能正常
- [ ] Session 管理正常
- [ ] 需要認證的 API 正常
- [ ] 錯誤處理正確
- [ ] 輸入驗證正常
- [ ] Rate Limiting 正常
- [ ] CORS 正常
- [ ] 日誌記錄正常


