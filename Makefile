.PHONY: build run test clean lint fmt swagger

# 變數
BINARY_NAME=higgstv-go
MAIN_PATH=cmd/server/main.go
BUILD_DIR=bin

# 建置
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# 執行
run:
	@go run $(MAIN_PATH)

# 測試
test:
	@go test -v ./...

# 測試覆蓋率
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Lint
lint:
	@golangci-lint run

# 格式化
fmt:
	@go fmt ./...

# 生成 Swagger 文檔
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go -o docs/swagger
	@echo "Swagger documentation generated successfully!"

# 安裝依賴
deps:
	@go mod download
	@go mod tidy

# 執行前檢查
pre-run: fmt lint test

# 開發模式
dev: fmt
	@go run $(MAIN_PATH)

