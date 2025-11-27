package errors

// 錯誤碼定義（與原 API 保持一致）
const (
	// 系統錯誤
	CodeServerError = -1

	// 輸入錯誤
	CodeRequiredField = 0

	// 認證錯誤（與原 API 保持一致）
	CodeRequireLogin = 1  // 需要登入
	CodeAccessDenied = 2  // 權限不足

	// 擴展錯誤碼（用於內部處理，不會回傳給客戶端）
	CodeInvalidFormat     = 100
	CodeInvalidValue      = 101
	CodeValueTooLong      = 102
	CodeValueTooShort     = 103
	CodeSessionExpired    = 104
	CodeUserExists        = 105
	CodeUserNotFound      = 106
	CodeInvalidPassword   = 107
	CodeInvalidInvitation = 108
	CodeChannelNotFound   = 109
	CodeProgramNotFound   = 110
	CodeInvalidAccessKey  = 111
)

// ErrorMessage 錯誤訊息對應
var ErrorMessage = map[int]string{
	// 原 API 錯誤碼
	CodeServerError:   "伺服器內部錯誤",
	CodeRequiredField: "缺少必要欄位",
	CodeRequireLogin:  "需要登入",
	CodeAccessDenied:  "權限不足",

	// 擴展錯誤碼（內部使用）
	CodeInvalidFormat:     "格式不正確",
	CodeInvalidValue:      "值無效",
	CodeValueTooLong:      "值過長",
	CodeValueTooShort:     "值過短",
	CodeSessionExpired:    "Session 已過期",
	CodeUserExists:        "使用者已存在",
	CodeUserNotFound:      "使用者不存在",
	CodeInvalidPassword:   "密碼錯誤",
	CodeInvalidInvitation: "邀請碼錯誤",
	CodeChannelNotFound:   "頻道不存在",
	CodeProgramNotFound:   "節目不存在",
	CodeInvalidAccessKey:  "無效的存取金鑰",
}

// GetErrorMessage 取得錯誤訊息
func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessage[code]; ok {
		return msg
	}
	return "未知錯誤"
}

