.PHONY: build run test clean lint fmt swagger ci check install-tools

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

# 測試（與 CI 一致）
test:
	@echo "Running tests..."
	@go test -v -coverprofile=coverage.out ./...

# 測試覆蓋率
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Lint（與 CI 一致）
lint:
	@echo "Running linter..."
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

# CI 檢查：運行所有 CI 相關檢查（測試、lint、構建）
ci: test lint build
	@echo ""
	@echo "✅ All CI checks passed!"

# 快速檢查：只運行測試和 lint（不構建）
check: test lint
	@echo ""
	@echo "✅ Quick check passed!"

# 安裝開發工具
install-tools:
	@echo "Installing golangci-lint v1 (latest stable)..."
	@which golangci-lint > /dev/null && golangci-lint --version | grep -q "version 1\." || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.64.8)
	@echo "✅ Tools installed!"
	@golangci-lint --version

# 執行前檢查
pre-run: fmt lint test

# 開發模式
dev: fmt
	@go run $(MAIN_PATH)

