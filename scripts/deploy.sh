#!/bin/bash

# HiggsTV Go API Server 部署腳本

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
APP_NAME="higgstv-go"
BUILD_DIR="./bin"
BINARY_NAME="$APP_NAME"
MAIN_PATH="cmd/server/main.go"

echo -e "${GREEN}開始部署 $APP_NAME...${NC}"

# 檢查 Go 環境
if ! command -v go &> /dev/null; then
    echo -e "${RED}錯誤: 未找到 Go 環境${NC}"
    exit 1
fi

# 清理舊的建置
echo -e "${YELLOW}清理舊的建置...${NC}"
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# 下載依賴
echo -e "${YELLOW}下載依賴...${NC}"
go mod download
go mod tidy

# 執行測試
echo -e "${YELLOW}執行測試...${NC}"
go test -v ./...

# 建置
echo -e "${YELLOW}建置應用程式...${NC}"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $BUILD_DIR/$BINARY_NAME $MAIN_PATH

if [ $? -eq 0 ]; then
    echo -e "${GREEN}建置成功！${NC}"
    ls -lh $BUILD_DIR/$BINARY_NAME
else
    echo -e "${RED}建置失敗！${NC}"
    exit 1
fi

# 檢查配置檔案
if [ ! -f "config/config.yaml" ]; then
    echo -e "${YELLOW}警告: 未找到 config/config.yaml，請確保配置檔案存在${NC}"
fi

echo -e "${GREEN}部署準備完成！${NC}"
echo -e "${YELLOW}執行方式:${NC}"
echo "  ./$BUILD_DIR/$BINARY_NAME"

