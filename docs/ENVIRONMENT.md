# 環境變數說明

本文件說明所有可用的環境變數及其用途。

## 伺服器配置

### HIGGSTV_SERVER_PORT
- **說明**: HTTP 伺服器監聽埠號
- **預設值**: `8080`
- **範例**: `HIGGSTV_SERVER_PORT=3000`

### HIGGSTV_SERVER_ENV
- **說明**: 執行環境
- **預設值**: `development`
- **可選值**: `development`, `production`
- **範例**: `HIGGSTV_SERVER_ENV=production`

## 資料庫配置

### HIGGSTV_DATABASE_URI
- **說明**: MongoDB 連線 URI
- **預設值**: `mongodb://localhost:27017`
- **範例**: `HIGGSTV_DATABASE_URI=mongodb://user:pass@localhost:27017`

### HIGGSTV_DATABASE_DATABASE
- **說明**: MongoDB 資料庫名稱
- **預設值**: `higgstv`
- **範例**: `HIGGSTV_DATABASE_DATABASE=higgstv_prod`

## Session 配置

### HIGGSTV_SESSION_SECRET
- **說明**: Session 加密金鑰（必須設定）
- **預設值**: `change-me-in-production`（會導致啟動失敗）
- **範例**: `HIGGSTV_SESSION_SECRET=your-secret-key-here`
- **注意**: 生產環境必須使用強隨機字串

## 郵件配置（可選）

### HIGGSTV_MAIL_SMTP_HOST
- **說明**: SMTP 伺服器主機
- **預設值**: `smtp.gmail.com`
- **範例**: `HIGGSTV_MAIL_SMTP_HOST=smtp.example.com`

### HIGGSTV_MAIL_SMTP_PORT
- **說明**: SMTP 伺服器埠號
- **預設值**: `587`
- **範例**: `HIGGSTV_MAIL_SMTP_PORT=465`

### HIGGSTV_MAIL_SMTP_USER
- **說明**: SMTP 使用者名稱
- **預設值**: 空字串
- **範例**: `HIGGSTV_MAIL_SMTP_USER=your-email@gmail.com`

### HIGGSTV_MAIL_SMTP_PASSWORD
- **說明**: SMTP 密碼或應用程式密碼
- **預設值**: 空字串
- **範例**: `HIGGSTV_MAIL_SMTP_PASSWORD=your-app-password`

### HIGGSTV_MAIL_FROM
- **說明**: 發送郵件的寄件人地址
- **預設值**: `HiggsTV <no-reply@higgstv.com>`
- **範例**: `HIGGSTV_MAIL_FROM=HiggsTV <noreply@example.com>`

### HIGGSTV_MAIL_BASE_URL
- **說明**: 應用程式基礎 URL（用於重設密碼連結）
- **預設值**: `http://localhost:8080`
- **範例**: `HIGGSTV_MAIL_BASE_URL=https://www.higgstv.com`

## 使用範例

### 開發環境
```bash
export HIGGSTV_SERVER_PORT=8080
export HIGGSTV_SERVER_ENV=development
export HIGGSTV_DATABASE_URI=mongodb://localhost:27017
export HIGGSTV_DATABASE_DATABASE=higgstv_dev
export HIGGSTV_SESSION_SECRET=dev-secret-key
```

### 生產環境
```bash
export HIGGSTV_SERVER_PORT=8080
export HIGGSTV_SERVER_ENV=production
export HIGGSTV_DATABASE_URI=mongodb://user:pass@mongodb.example.com:27017
export HIGGSTV_DATABASE_DATABASE=higgstv
export HIGGSTV_SESSION_SECRET=$(openssl rand -base64 32)
export HIGGSTV_MAIL_SMTP_HOST=smtp.example.com
export HIGGSTV_MAIL_SMTP_PORT=587
export HIGGSTV_MAIL_SMTP_USER=noreply@example.com
export HIGGSTV_MAIL_SMTP_PASSWORD=secure-password
export HIGGSTV_MAIL_FROM="HiggsTV <noreply@example.com>"
export HIGGSTV_MAIL_BASE_URL=https://www.higgstv.com
```

### Docker Compose
```yaml
environment:
  HIGGSTV_SERVER_PORT: "8080"
  HIGGSTV_SERVER_ENV: "production"
  HIGGSTV_DATABASE_URI: "mongodb://mongodb:27017"
  HIGGSTV_DATABASE_DATABASE: "higgstv"
  HIGGSTV_SESSION_SECRET: "${SESSION_SECRET}"
```

## 注意事項

1. **Session Secret**: 生產環境必須使用強隨機字串，可以使用以下命令產生：
   ```bash
   openssl rand -base64 32
   ```

2. **資料庫 URI**: 如果使用認證，URI 格式為：
   ```
   mongodb://username:password@host:port/database
   ```

3. **環境變數優先級**: 環境變數會覆蓋配置檔案中的設定

4. **敏感資訊**: 建議使用 secrets 管理工具（如 Kubernetes Secrets、Docker Secrets）來管理敏感資訊

