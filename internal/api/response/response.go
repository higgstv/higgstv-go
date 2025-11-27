package response

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StateCode 狀態碼
const (
	StateSuccess = 0
	StateFailed  = 1
)

// ErrorCode 錯誤碼
const (
	ErrorServerError   = -1
	ErrorRequiredField = 0
	ErrorRequireLogin  = 1
	ErrorAccessDenied  = 2
)

// Response 統一 API 回應格式
type Response struct {
	State int         `json:"state"`
	Code  *int        `json:"code,omitempty"`
	Ret   *bool       `json:"ret,omitempty"`
	Data  interface{} `json:",omitempty"`
}

// Success 成功回應
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		State: StateSuccess,
		Data:  data,
	})
}

// SuccessWithRet 成功回應（帶 ret 欄位）
func SuccessWithRet(c *gin.Context, ret bool) {
	c.JSON(http.StatusOK, Response{
		State: StateSuccess,
		Ret:   &ret,
	})
}

// Error 錯誤回應
func Error(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{
		State: StateFailed,
		Code:  &code,
	})
}

// ErrorWithRet 錯誤回應（帶 ret 欄位）
func ErrorWithRet(c *gin.Context, code int, ret bool) {
	c.JSON(http.StatusOK, Response{
		State: StateFailed,
		Code:  &code,
		Ret:   &ret,
	})
}

// JSONPSuccess JSONP 成功回應
func JSONPSuccess(c *gin.Context, callback string, data interface{}) {
	c.Header("Content-Type", "application/javascript")
	jsonData, err := json.Marshal(data)
	if err != nil {
		// 如果 JSON 序列化失敗，回傳錯誤
		c.String(http.StatusOK, "%s({\"state\":1,\"code\":-1});", callback)
		return
	}
	c.String(http.StatusOK, "%s(%s);", callback, string(jsonData))
}

// JSONPError JSONP 錯誤回應
func JSONPError(c *gin.Context, callback string, code int) {
	data := Response{
		State: StateFailed,
		Code:  &code,
	}
	jsonData, _ := json.Marshal(data)
	c.Header("Content-Type", "application/javascript")
	c.String(http.StatusOK, "%s(%s);", callback, string(jsonData))
}

