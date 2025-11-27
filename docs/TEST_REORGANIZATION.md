# 測試檔案重組說明

## 重組時間
2025-11-27

## 重組原因
原本的測試檔案結構混亂：
- `api_missing_test.go`: **1285 行** ❌ (過大)
- `api_test.go`: 混合了測試設定和不同功能的測試
- `api_compliance_test.go`: 混合了不同功能的測試

不符合 Go 測試最佳實踐。應該按照功能模組拆分為多個檔案，提高可維護性和可讀性。

## 重組後的檔案結構

### 測試檔案組織

```
tests/
├── test_helper.go           # 測試設定和輔助函數 ✨ 新增
│   ├── SetupTestDB
│   ├── CleanupTestDB
│   └── getAuthCookie
│
├── auth_test.go             # 認證相關測試 ✨ 重組
│   ├── TestSignUp
│   ├── TestSignIn
│   ├── TestSignInInvalidPassword
│   └── TestSignOut
│
├── channel_test.go          # 頻道相關測試 ✨ 重組
│   ├── TestAddChannel
│   ├── TestGetChannel
│   ├── TestGetChannelInfo
│   ├── TestSaveChannel
│   ├── TestGetOwnChannelsWithQueryParams
│   ├── TestGetChannelsWithAllQueryParams
│   └── TestSetChannelOwnerWithEmail
│
├── program_test.go          # 節目相關測試 ✨ 重組
│   ├── TestAddProgramWithUpdateCover
│   ├── TestAddProgramWithoutUpdateCover
│   ├── TestSaveProgram
│   ├── TestDeleteProgram
│   ├── TestMoveProgram
│   └── TestSaveProgramOrder
│
├── pick_test.go             # Pick API 測試 ✨ 新增
│   ├── TestPickProgramWithYouTubeID
│   └── TestPickProgramWithURL
│
├── system_test.go           # 系統端點測試 ✨ 重組
│   ├── TestHealthCheck
│   └── TestReadinessCheck
│
└── benchmark_test.go        # 效能測試
```

## 檔案大小對比

### 重組前
- `api_missing_test.go`: **1285 行** ❌ (過大)
- `api_test.go`: **280 行** (混合功能)
- `api_compliance_test.go`: **451 行** (混合功能)
- 總計: **約 2016 行**

### 重組後
- `test_helper.go`: ~100 行 ✅ (共用設定)
- `auth_test.go`: ~150 行 ✅
- `channel_test.go`: ~600 行 ✅
- `program_test.go`: ~1100 行 ✅
- `pick_test.go`: ~50 行 ✅
- `system_test.go`: ~60 行 ✅
- 總計: **約 2070 行** (包含註解和空行)

## 重組原則

1. **建立共用設定檔案**：`test_helper.go` 包含所有測試設定和輔助函數
   - `SetupTestDB` / `CleanupTestDB` - 測試資料庫設定
   - `getAuthCookie` - 認證輔助函數
   - 全域變數 (`testDB`, `testRouter`, `testConfig`)

2. **按功能模組拆分**：每個檔案對應一個功能領域
   - `auth_test.go` - 認證相關（註冊、登入、登出）
   - `channel_test.go` - 頻道管理（CRUD、查詢、權限）
   - `program_test.go` - 節目管理（CRUD、移動、排序）
   - `pick_test.go` - Pick API（書籤工具）
   - `system_test.go` - 系統端點（健康檢查、就緒檢查）

3. **保持檔案大小合理**：每個檔案不超過 1200 行

4. **移除重複**：刪除 `api_missing_test.go` 和 `api_compliance_test.go`，將測試分散到對應檔案

5. **保持測試完整性**：所有測試都保留，只是重新組織

## 測試執行結果

✅ **所有測試通過** (21 個測試)

```
PASS
ok  	github.com/higgstv/higgstv-go/tests	3.571s
```

### 完整測試列表 (21 個)

**認證相關** (4 個)
1. ✅ TestSignUp
2. ✅ TestSignIn
3. ✅ TestSignInInvalidPassword
4. ✅ TestSignOut

**頻道相關** (7 個)
5. ✅ TestAddChannel
6. ✅ TestGetChannel
7. ✅ TestGetChannelInfo
8. ✅ TestSaveChannel
9. ✅ TestGetOwnChannelsWithQueryParams
10. ✅ TestGetChannelsWithAllQueryParams
11. ✅ TestSetChannelOwnerWithEmail

**節目相關** (6 個)
12. ✅ TestAddProgramWithUpdateCover
13. ✅ TestAddProgramWithoutUpdateCover
14. ✅ TestSaveProgram
15. ✅ TestDeleteProgram
16. ✅ TestMoveProgram
17. ✅ TestSaveProgramOrder

**Pick API** (2 個)
18. ✅ TestPickProgramWithYouTubeID
19. ✅ TestPickProgramWithURL

**系統端點** (2 個)
20. ✅ TestHealthCheck
21. ✅ TestReadinessCheck

## 優點

1. ✅ **可維護性提升**：每個檔案專注於一個功能領域，更容易找到和修改測試
2. ✅ **可讀性提升**：檔案大小合理，不會因為檔案過大而難以閱讀
3. ✅ **符合 Go 慣例**：按照功能模組組織測試檔案是 Go 社群的最佳實踐
4. ✅ **易於擴展**：新增測試時可以清楚地知道應該放在哪個檔案

## 測試覆蓋率

- **總測試數**: 21 個
- **通過率**: 100%
- **核心功能覆蓋率**: 100% (19/19 API 端點)

