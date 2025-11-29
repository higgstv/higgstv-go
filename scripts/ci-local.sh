#!/bin/bash
# 本地 CI 檢查腳本
# 這個腳本會運行與 GitHub Actions CI 相同的檢查

set -e  # 遇到錯誤立即退出

echo "🔍 Running local CI checks..."
echo ""

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 檢查必要的工具
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}❌ $1 is not installed${NC}"
        echo "Run 'make install-tools' to install required tools"
        exit 1
    fi
}

# 檢查 Go
check_tool go

# 檢查 golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}⚠️  golangci-lint not found, installing...${NC}"
    make install-tools
fi

# 檢查 MongoDB 是否運行
if ! docker ps | grep -q mongo; then
    echo -e "${YELLOW}⚠️  MongoDB container not running${NC}"
    echo "Starting MongoDB..."
    docker-compose up -d mongodb
    echo "Waiting for MongoDB to be ready..."
    sleep 5
fi

echo -e "${GREEN}✅ Prerequisites check passed${NC}"
echo ""

# 1. 運行測試（與 CI 一致）
echo "📝 Running tests..."
if go test -v -coverprofile=coverage.out ./...; then
    echo -e "${GREEN}✅ Tests passed${NC}"
else
    echo -e "${RED}❌ Tests failed${NC}"
    exit 1
fi
echo ""

# 2. 運行 Lint
echo "🔍 Running linter..."
if golangci-lint run; then
    echo -e "${GREEN}✅ Lint passed${NC}"
else
    echo -e "${RED}❌ Lint failed${NC}"
    exit 1
fi
echo ""

# 3. 構建
echo "🔨 Building..."
if go build -o bin/higgstv-go cmd/server/main.go; then
    echo -e "${GREEN}✅ Build passed${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo ""

echo -e "${GREEN}🎉 All CI checks passed!${NC}"
echo "You can safely push to GitHub."

