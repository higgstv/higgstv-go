# 文檔索引

## 概述
本文檔提供 HiggsTV Go API Server 專案文檔的完整索引和說明。

## 核心文檔

### API 文檔
- **API_REFERENCE.md** (位於專案根目錄 `docs/`) - 完整的 API 參考文件
- **GOLANG_IMPLEMENTATION_GUIDE.md** (位於專案根目錄 `docs/`) - Golang 實作指南

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
- **TESTING.md** - 測試指南和手動測試步驟
  - 快速測試指南
  - 手動測試步驟
  - 預期結果

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
- **MIGRATION.md** - 資料庫遷移說明
  - 遷移系統說明
  - 如何新增遷移
  - 遷移執行流程

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
├── API_COMPLIANCE_AUDIT.md      # API 符合度審計報告
├── COMPLETE_TEST_SUMMARY.md     # 完整測試總結
├── TEST_REORGANIZATION.md        # 測試檔案重組記錄
├── TESTING.md                    # 測試指南
├── API_VERSIONING.md             # API 版本控制
├── VALIDATION.md                 # 輸入驗證說明
├── MIGRATION.md                  # 資料庫遷移說明
├── DEPLOYMENT.md                 # 部署說明
├── ENVIRONMENT.md                # 環境變數說明
├── CHANGELOG.md                  # 變更記錄
├── DOCUMENTATION_INDEX.md        # 文檔索引（本檔案）
└── swagger/                      # Swagger 文檔
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
- **API 端點規範** → `API_REFERENCE.md` (專案根目錄)
- **如何實作新功能** → `GOLANG_IMPLEMENTATION_GUIDE.md` (專案根目錄)
- **測試覆蓋率** → `COMPLETE_TEST_SUMMARY.md`
- **如何執行測試** → `TESTING.md`
- **如何部署** → `DEPLOYMENT.md`
- **環境變數設定** → `ENVIRONMENT.md`
- **資料庫遷移** → `MIGRATION.md`
- **API 版本控制** → `API_VERSIONING.md`
- **輸入驗證規則** → `VALIDATION.md`
- **API 符合度** → `API_COMPLIANCE_AUDIT.md`
- **變更歷史** → `CHANGELOG.md`

## 文檔更新記錄

- **2025-11-27**: 清理重複和過時的測試文檔
- **2025-11-27**: 更新測試文檔以反映新的測試檔案結構
- **2025-11-27**: 建立文檔索引

