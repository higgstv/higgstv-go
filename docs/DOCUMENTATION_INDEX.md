# 文檔索引

## 概述
本文檔提供 HiggsTV Go API Server 專案文檔的完整索引和說明。

## 核心文檔

### API 符合度
- **API_COMPLIANCE_AUDIT.md** - API 實作符合度深度審計報告
  - 詳細的 API 端點符合度檢查
  - 所有問題的修正狀態
  - 符合度統計（100%）

## 測試文檔

### 測試總結
- **COMPLETE_TEST_SUMMARY.md** - 完整的測試總結報告
  - 測試覆蓋率統計（100%）
  - 測試檔案列表
  - 所有 API 端點測試狀態
  - 測試執行結果

### 測試結構
- **TEST_REORGANIZATION.md** - 測試檔案重組記錄
  - 重組原因和過程
  - 重組後的檔案結構
  - 檔案大小對比

### 測試指南
- **TESTING_GUIDE.md** - 完整的測試指南（推薦）
  - MongoDB 和 SQLite 測試配置
  - 測試架構說明
  - 測試輔助函數說明
  - 完整的測試範例
- **TESTING.md** - 快速測試指南
  - 快速測試步驟
  - 手動測試 curl 命令
  - 適合快速驗證功能

## 技術文檔

### 部署與環境
- **DEPLOYMENT.md** - 部署說明
  - Docker 部署
  - Docker Compose 配置
  - 部署腳本

- **ENVIRONMENT.md** - 環境變數說明
  - 所有環境變數列表
  - 配置說明
  - 預設值

### 資料庫
- **SQLITE_IMPLEMENTATION_COMPLETE.md** - SQLite 支援實作完成報告
  - 完整的實作清單
  - 技術架構說明
  - 使用方式和範例
- **SQLITE_MIGRATION_PLAN.md** - SQLite 支援遷移計劃（歷史記錄）
  - 原始遷移計劃
  - 架構設計說明
  - 技術挑戰和解決方案
- **MIGRATION.md** - 資料庫遷移系統說明
  - 遷移系統說明（支援 MongoDB 和 SQLite）
  - 如何新增遷移
  - 遷移執行流程
- **DATA_MIGRATION_GUIDE.md** - 資料遷移指南
  - MongoDB → SQLite 資料遷移
  - 測試環境說明
  - 資料驗證方法
- **MIGRATION_TOOL_USAGE.md** - 遷移工具使用指南
  - 遷移工具詳細說明
  - 使用範例和輸出
  - 故障排除指南

### API 功能
- **API_VERSIONING.md** - API 版本控制說明
  - 版本策略
  - 版本檢查
  - 向後相容性

- **VALIDATION.md** - 輸入驗證說明
  - 驗證層級
  - 驗證規則
  - 自訂驗證器

## 變更記錄

- **CHANGELOG.md** - 變更記錄
  - 版本變更歷史
  - 新增功能
  - 修正問題

## Swagger 文檔

- **swagger/** - Swagger/OpenAPI 文檔（自動生成）
  - `docs.go` - Swagger 定義（自動生成）
  - `swagger.json` - OpenAPI JSON 格式
  - `swagger.yaml` - OpenAPI YAML 格式

## 文檔結構

```
docs/
├── API_COMPLIANCE_AUDIT.md          # API 符合度審計報告
├── COMPLETE_TEST_SUMMARY.md         # 完整測試總結
├── TEST_REORGANIZATION.md           # 測試檔案重組記錄
├── TESTING_GUIDE.md                 # 完整測試指南（推薦）
├── TESTING.md                       # 快速測試指南
├── API_VERSIONING.md                # API 版本控制
├── VALIDATION.md                    # 輸入驗證說明
├── SQLITE_IMPLEMENTATION_COMPLETE.md # SQLite 實作完成報告
├── SQLITE_MIGRATION_PLAN.md         # SQLite 遷移計劃（歷史記錄）
├── MIGRATION.md                     # 資料庫遷移系統說明
├── DATA_MIGRATION_GUIDE.md         # 資料遷移指南
├── MIGRATION_TOOL_USAGE.md          # 遷移工具使用指南
├── DEPLOYMENT.md                    # 部署說明
├── ENVIRONMENT.md                   # 環境變數說明
├── CHANGELOG.md                     # 變更記錄
├── DOCUMENTATION_INDEX.md           # 文檔索引（本檔案）
└── swagger/                         # Swagger 文檔
    ├── docs.go
    ├── swagger.json
    └── swagger.yaml
```

## 文檔維護原則

1. **保持最新**: 所有文檔都應該反映當前實作狀態
2. **避免重複**: 不保留重複或過時的文檔
3. **清晰結構**: 按功能模組組織文檔
4. **定期更新**: 當實作有重大變更時，更新相關文檔

## 快速查找

### 我需要了解...
- **API 端點規範** → Swagger 文檔 (`swagger/swagger.yaml`) 或 `API_COMPLIANCE_AUDIT.md`
- **測試覆蓋率** → `COMPLETE_TEST_SUMMARY.md`
- **如何執行測試** → `TESTING_GUIDE.md`（完整指南）或 `TESTING.md`（快速指南）
- **如何部署** → `DEPLOYMENT.md`
- **環境變數設定** → `ENVIRONMENT.md`
- **資料庫遷移系統** → `MIGRATION.md`
- **MongoDB → SQLite 資料遷移** → `DATA_MIGRATION_GUIDE.md` 或 `MIGRATION_TOOL_USAGE.md`
- **SQLite 實作詳情** → `SQLITE_IMPLEMENTATION_COMPLETE.md`
- **API 版本控制** → `API_VERSIONING.md`
- **輸入驗證規則** → `VALIDATION.md`
- **API 符合度** → `API_COMPLIANCE_AUDIT.md`
- **變更歷史** → `CHANGELOG.md`

## 文檔更新記錄

- **2025-11-27**: 清理重複和過時的測試文檔
- **2025-11-27**: 更新測試文檔以反映新的測試檔案結構
- **2025-11-27**: 建立文檔索引
- **2025-11-30**: 移除過時的 SQLITE_IMPLEMENTATION_STATUS.md
- **2025-11-30**: 更新所有文件中的工具路徑（cmd/migrate/, cmd/check_database/）
- **2025-11-30**: 更新 MIGRATION.md 以支援 MongoDB 和 SQLite
- **2025-11-30**: 新增 SQLite 相關文件索引
- **2025-12-04**: 更新 COMPLETE_TEST_SUMMARY.md（測試數量從 21 更新為 24）
- **2025-12-04**: 更新 CHANGELOG.md（加入最新修復和變更）

