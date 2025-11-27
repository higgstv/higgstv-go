# API 版本控制

## 概述

HiggsTV Go API Server 支援 API 版本控制，確保向後相容性和平滑升級。

## 版本標識

當前 API 版本：**v1**

版本資訊會自動添加到所有回應的 HTTP Header 中：

```
X-API-Version: v1
```

## 版本策略

### 當前版本（v1）

所有現有 API 端點都屬於 v1 版本，無需在 URL 中指定版本號：

- `/apis/signin`
- `/apis/signup`
- `/apis/getchannels`
- 等等...

### 未來版本（v2+）

當需要進行不向後相容的變更時，可以引入新版本：

```
/api/v2/signin
/api/v2/signup
```

## 版本檢查

客戶端可以通過檢查 `X-API-Version` Header 來確認 API 版本：

```javascript
fetch('/apis/signin', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  }
})
.then(response => {
  const apiVersion = response.headers.get('X-API-Version');
  console.log('API Version:', apiVersion);
  return response.json();
});
```

## 向後相容性

- v1 API 將持續支援，不會被移除
- 新功能優先添加到 v1（如果向後相容）
- 不向後相容的變更會引入新版本

## 遷移指南

當新版本發布時，會提供詳細的遷移指南，說明：
- 變更的 API 端點
- 新的請求/回應格式
- 遷移步驟
- 範例程式碼

## 版本號規則

遵循語義化版本控制（Semantic Versioning）：
- **主版本號（Major）**：不向後相容的變更
- **次版本號（Minor）**：向後相容的新功能
- **修訂版本號（Patch）**：向後相容的問題修復

