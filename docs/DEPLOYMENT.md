# 部署指南

本文件說明如何部署 HiggsTV Go API Server。

## 前置需求

- Go 1.24+ 或 Docker
- MongoDB 7+
- 配置檔案或環境變數

## 部署方式

### 1. 直接部署（Go Binary）

#### 建置

```bash
# 下載依賴
go mod download

# 執行測試
go test ./...

# 建置
go build -o bin/higgstv-go cmd/server/main.go
```

#### 執行

```bash
# 設定環境變數
export HIGGSTV_SESSION_SECRET=$(openssl rand -base64 32)
export HIGGSTV_DATABASE_URI=mongodb://localhost:27017
export HIGGSTV_DATABASE_DATABASE=higgstv

# 執行
./bin/higgstv-go
```

### 2. Docker 部署

#### 建置映像檔

```bash
docker build -t higgstv-go .
```

#### 使用 Docker Compose

```bash
# 啟動所有服務
docker-compose up -d

# 查看日誌
docker-compose logs -f api

# 停止服務
docker-compose down
```

### 3. 使用部署腳本

```bash
# 執行部署腳本
./scripts/deploy.sh

# 執行建置的應用程式
./bin/higgstv-go
```

## 生產環境配置

### 1. 環境變數設定

建立 `.env` 檔案或設定環境變數：

```bash
HIGGSTV_SERVER_PORT=8080
HIGGSTV_SERVER_ENV=production
HIGGSTV_DATABASE_URI=mongodb://user:pass@mongodb.example.com:27017
HIGGSTV_DATABASE_DATABASE=higgstv
HIGGSTV_SESSION_SECRET=<強隨機字串>
HIGGSTV_MAIL_SMTP_HOST=smtp.example.com
HIGGSTV_MAIL_SMTP_PORT=587
HIGGSTV_MAIL_SMTP_USER=noreply@example.com
HIGGSTV_MAIL_SMTP_PASSWORD=<secure-password>
HIGGSTV_MAIL_FROM="HiggsTV <noreply@example.com>"
HIGGSTV_MAIL_BASE_URL=https://www.higgstv.com
```

### 2. 產生 Session Secret

```bash
openssl rand -base64 32
```

### 3. 資料庫準備

確保 MongoDB 已啟動並可連線：

```bash
# 測試連線
mongosh "mongodb://localhost:27017"
```

### 4. 反向代理設定（Nginx）

```nginx
server {
    listen 80;
    server_name api.higgstv.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 5. Systemd Service（Linux）

建立 `/etc/systemd/system/higgstv-go.service`:

```ini
[Unit]
Description=HiggsTV Go API Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/higgstv-go
ExecStart=/opt/higgstv-go/bin/higgstv-go
Restart=always
RestartSec=10
EnvironmentFile=/etc/higgstv-go/env

[Install]
WantedBy=multi-user.target
```

啟動服務：

```bash
sudo systemctl daemon-reload
sudo systemctl enable higgstv-go
sudo systemctl start higgstv-go
sudo systemctl status higgstv-go
```

## 監控設定

### Prometheus 設定

在 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'higgstv-go'
    static_configs:
      - targets: ['localhost:8080']
```

### Grafana Dashboard

可以建立 Grafana Dashboard 來視覺化指標：
- HTTP 請求速率
- 錯誤率
- 回應時間
- 資料庫操作統計

## 安全建議

1. **使用 HTTPS**: 透過反向代理（如 Nginx）設定 SSL/TLS
2. **防火牆**: 只開放必要的埠號
3. **Session Secret**: 使用強隨機字串
4. **資料庫認證**: 啟用 MongoDB 認證
5. **Rate Limiting**: 根據需求調整 Rate Limiting 設定
6. **日誌管理**: 設定日誌輪轉和歸檔

## 故障排除

### 無法連線資料庫

```bash
# 檢查 MongoDB 是否運行
systemctl status mongod

# 檢查連線
mongosh "mongodb://localhost:27017"
```

### 應用程式無法啟動

```bash
# 檢查日誌
journalctl -u higgstv-go -f

# 檢查配置
./bin/higgstv-go --help
```

### 效能問題

1. 檢查資料庫索引是否已建立
2. 檢查連線池設定
3. 查看 Prometheus 指標
4. 檢查 Rate Limiting 設定

## 更新部署

```bash
# 1. 停止服務
sudo systemctl stop higgstv-go

# 2. 備份資料
mongodump --out=/backup/$(date +%Y%m%d)

# 3. 更新應用程式
git pull
go build -o bin/higgstv-go cmd/server/main.go

# 4. 重啟服務
sudo systemctl start higgstv-go
```

## 回滾

如果新版本有問題，可以快速回滾：

```bash
# 1. 停止服務
sudo systemctl stop higgstv-go

# 2. 恢復舊版本
git checkout <previous-version>
go build -o bin/higgstv-go cmd/server/main.go

# 3. 重啟服務
sudo systemctl start higgstv-go
```

