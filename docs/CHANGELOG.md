# Changelog

## [Unreleased]

### Added
- 健康檢查端點 (`/health`, `/ready`)
- Rate Limiting 中介層
- Request ID 追蹤
- 資料庫索引自動建立
- 配置驗證
- 資料庫連線池配置
- JSONP callback 驗證
- 單元測試範例

### Changed
- 改進錯誤處理機制
- 增強日誌記錄（包含 Request ID）
- 優化 MongoDB 連線配置

### Security
- 新增 JSONP callback 參數驗證
- 配置驗證確保 Session Secret 已設定

## Phase 2 - 進階功能

### Added
- 錯誤處理機制
- 日誌記錄系統（使用 zap）
- CORS 支援
- 請求驗證
- Docker 支援
- Makefile 建置腳本
- 測試框架

## Phase 1 - 基礎架構

### Added
- 完整的專案結構
- 資料模型定義（User, Channel, Program）
- Repository 層實作
- Service 層實作
- Handler 層實作
- 路由設定
- Session 管理
- 統一回應格式

