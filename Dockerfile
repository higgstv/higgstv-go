# 建置階段
FROM golang:1.24-alpine AS builder

# 安裝 SQLite 編譯依賴（需要 CGO）
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# 複製 go mod 檔案
COPY go.mod go.sum ./
RUN go mod download

# 複製原始碼
COPY . .

# 建置應用程式（啟用 CGO 以支援 SQLite）
RUN CGO_ENABLED=1 GOOS=linux go build -a -tags sqlite3 -o bin/higgstv-go cmd/server/main.go

# 執行階段
FROM alpine:latest

# 安裝 SQLite 執行時依賴和基本工具
RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /root/

# 建立資料目錄（用於 SQLite 資料庫）
RUN mkdir -p /root/data

# 複製建置好的二進位檔
COPY --from=builder /app/bin/higgstv-go .
COPY --from=builder /app/config/config.example.yaml ./config/config.yaml

# 暴露埠號
EXPOSE 8080

# 執行應用程式
CMD ["./higgstv-go"]

