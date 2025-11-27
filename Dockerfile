# 建置階段
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 複製 go mod 檔案
COPY go.mod go.sum ./
RUN go mod download

# 複製原始碼
COPY . .

# 建置應用程式
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/higgstv-go cmd/server/main.go

# 執行階段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 複製建置好的二進位檔
COPY --from=builder /app/bin/higgstv-go .
COPY --from=builder /app/config/config.example.yaml ./config/config.yaml

# 暴露埠號
EXPOSE 8080

# 執行應用程式
CMD ["./higgstv-go"]

