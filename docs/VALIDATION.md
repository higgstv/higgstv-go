# 輸入驗證說明

## 概述

HiggsTV Go API Server 使用多層驗證機制確保輸入資料的安全性和正確性。

## 驗證層級

### 1. 結構體驗證（Struct Validation）

使用 `go-playground/validator` 進行結構體驗證：

```go
type SignUpRequest struct {
    InvitationCode string `json:"invitation_code" binding:"required"`
    Username       string `json:"username" binding:"required,username"`
    Email          string `json:"email" binding:"required,email"`
    Password       string `json:"password" binding:"required,password"`
}
```

### 2. 自訂驗證器

#### YouTube URL 驗證

```go
type AddProgramRequest struct {
    YouTubeID string `json:"youtube_id" binding:"required,youtube_id"`
}
```

支援的格式：
- `https://www.youtube.com/watch?v=VIDEO_ID`
- `https://youtu.be/VIDEO_ID`
- `https://www.youtube.com/embed/VIDEO_ID`

#### 使用者名稱驗證

- 長度：3-20 個字元
- 只允許：字母、數字、底線
- 驗證標籤：`username`

#### 密碼驗證

- 最小長度：6 個字元
- 驗證標籤：`password`

### 3. 業務邏輯驗證

在 Service 層進行業務邏輯驗證：

```go
// 檢查使用者是否存在
exists, err := userRepo.Exists(ctx, username, email)
if exists {
    return errors.ErrUserExists
}
```

## 驗證規則

### 必填欄位

使用 `binding:"required"` 標籤：

```go
Name string `json:"name" binding:"required"`
```

### 格式驗證

- `email`: Email 格式
- `youtube_id`: YouTube ID 格式
- `youtube_url`: YouTube URL 格式

### 長度驗證

- `min=3`: 最小長度
- `max=100`: 最大長度

### 自訂驗證

- `username`: 使用者名稱格式
- `password`: 密碼強度

## 錯誤回應

驗證失敗時，API 會回傳：

```json
{
  "state": 1,
  "code": 0
}
```

其中 `code: 0` 表示 `RequiredField`（缺少必要欄位或格式不正確）。

## 輸入清理

所有字串輸入都會自動清理：
- 移除前後空白
- 移除控制字元（\x00, \r, \n）

## 安全建議

1. **永遠驗證輸入**：不要信任客戶端輸入
2. **使用白名單**：只允許預期的字元
3. **限制長度**：防止過長輸入
4. **清理輸出**：輸出時進行適當的轉義

## 範例

### 註冊請求驗證

```go
type SignUpRequest struct {
    InvitationCode string `json:"invitation_code" binding:"required"`
    Username       string `json:"username" binding:"required,username,min=3,max=20"`
    Email          string `json:"email" binding:"required,email"`
    Password       string `json:"password" binding:"required,password,min=6"`
}
```

### 新增節目驗證

```go
type AddProgramRequest struct {
    Ch        string `json:"ch" binding:"required"`
    Name      string `json:"name" binding:"required,min=1,max=200"`
    YouTubeID string `json:"youtube_id" binding:"required,youtube_id"`
    Desc      string `json:"desc" binding:"max=1000"`
    Duration  int    `json:"duration" binding:"min=0"`
    Tags      []int  `json:"tags"`
}
```

