# Changelog

## [Unreleased]

### Added
- ✅ **多資料庫支援**：完整的 SQLite 支援，可透過配置切換 MongoDB 或 SQLite
- ✅ **資料庫抽象層**：統一的資料庫介面，易於擴展
- ✅ **SQLite Repository 實作**：完整的 User、Channel、Program Repository 實作
- ✅ **資料遷移工具**：MongoDB 到 SQLite 的完整遷移工具 (`cmd/migrate/migrate_mongodb_to_sqlite.go`)
- ✅ **資料庫檢查工具**：統一的資料庫檢查工具 (`cmd/check_database/check_database.go`)
- ✅ **測試隔離機制**：每個測試使用獨立的資料庫連線，避免資料殘留
- ✅ 健康檢查端點 (`/health`, `/ready`)
- ✅ Rate Limiting 中介層
- ✅ Request ID 追蹤
- ✅ 資料庫索引自動建立（支援 MongoDB 和 SQLite）
- ✅ 配置驗證
- ✅ 資料庫連線池配置
- ✅ JSONP callback 驗證
- ✅ 單元測試範例

### Changed
- ✅ **測試系統重構**：使用 `TestDBContext` 確保每個測試有獨立的資料庫連線
- ✅ **Repository 層重構**：使用資料庫抽象介面，支援多種資料庫
- ✅ **Service 層更新**：移除 MongoDB 特定類型（bson.M, bson.D），改用通用介面
- ✅ 改進錯誤處理機制
- ✅ 增強日誌記錄（包含 Request ID）
- ✅ 優化資料庫連線配置（支援 MongoDB 和 SQLite）

### Fixed
- ✅ 修正測試中的資料殘留問題（使用獨立的資料庫連線）
- ✅ 修正 `DeletePrograms` 在 MongoDB 中的實作錯誤
- ✅ 修正多個 linter 錯誤（未檢查的錯誤返回值、空分支等）
- ✅ 修正 `cmd` 目錄結構（多個 main 函數問題）

### Security
- ✅ 新增 JSONP callback 參數驗證
- ✅ 配置驗證確保 Session Secret 已設定

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

