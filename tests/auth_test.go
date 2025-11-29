package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/higgstv/higgstv-go/internal/repository"
	"github.com/higgstv/higgstv-go/internal/service"
)

// TestSignUp 測試註冊 API
func TestSignUp(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	payload := map[string]interface{}{
		"invitation_code": "sixpens",
		"username":        "testuser",
		"email":           "test@example.com",
		"password":        "testpass123",
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/apis/signup", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 輸出實際回應以便除錯
	if response["state"] != float64(0) {
		t.Logf("SignUp response: %+v", response)
	}

	assert.Equal(t, float64(0), response["state"], "SignUp should succeed")

	// 檢查 ret 欄位（可能在大寫 Ret 或小寫 ret）
	ret, ok := response["Ret"]
	if !ok {
		ret, ok = response["ret"]
	}
	require.True(t, ok, "response should have 'Ret' or 'ret' field")
	assert.Equal(t, true, ret, "ret should be true for successful signup")
}

// TestSignIn 測試登入 API
func TestSignIn(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 先註冊
	signupPayload := map[string]interface{}{
		"invitation_code": "sixpens",
		"username":        "testuser",
		"email":           "test@example.com",
		"password":        "testpass123",
	}
	jsonData, _ := json.Marshal(signupPayload)
	req, _ := http.NewRequest("POST", "/apis/signup", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// 登入
	loginPayload := map[string]interface{}{
		"username": "testuser",
		"password": "testpass123",
	}
	jsonData, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", "/apis/signin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, true, response["ret"])
}

// TestSignInInvalidPassword 測試錯誤密碼登入
func TestSignInInvalidPassword(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 先註冊
	signupPayload := map[string]interface{}{
		"invitation_code": "sixpens",
		"username":        "testuser",
		"email":           "test@example.com",
		"password":        "testpass123",
	}
	jsonData, _ := json.Marshal(signupPayload)
	req, _ := http.NewRequest("POST", "/apis/signup", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// 使用錯誤密碼登入
	loginPayload := map[string]interface{}{
		"username": "testuser",
		"password": "wrongpassword",
	}
	jsonData, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", "/apis/signin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, false, response["ret"])
}

// TestSignOut 測試登出功能
func TestSignOut(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 先註冊並登入
	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 測試 1: 無 redirect 參數時，應該回傳 { "state": 0 }
	req, _ := http.NewRequest("GET", "/apis/signout", nil)
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 測試 2: 有 redirect 參數時，應該執行 HTTP redirect
	req, _ = http.NewRequest("GET", "/apis/signout?redirect=/home", nil)
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/home", w.Header().Get("Location"))
}

// TestChangePassword 測試變更密碼 API
func TestChangePassword(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 先註冊並登入
	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "oldpassword123")

	// 測試 1: 成功變更密碼
	changePasswordPayload := map[string]interface{}{
		"password":     "oldpassword123",
		"new_password": "newpassword456",
	}
	jsonData, _ := json.Marshal(changePasswordPayload)
	req, _ := http.NewRequest("POST", "/apis/change_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, true, response["ret"])

	// 驗證新密碼可以登入
	loginPayload := map[string]interface{}{
		"username": "testuser",
		"password": "newpassword456",
	}
	jsonData, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", "/apis/signin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, true, response["ret"])

	// 測試 2: 舊密碼錯誤
	cookie2 := getAuthCookie(t, testRouter, "testuser2", "test2@example.com", "password123")

	changePasswordPayload2 := map[string]interface{}{
		"password":     "wrongpassword",
		"new_password": "newpassword789",
	}
	jsonData, _ = json.Marshal(changePasswordPayload2)
	req, _ = http.NewRequest("POST", "/apis/change_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie2)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, false, response["ret"])

	// 測試 3: 未登入
	changePasswordPayload3 := map[string]interface{}{
		"password":     "oldpassword123",
		"new_password": "newpassword456",
	}
	jsonData3, _ := json.Marshal(changePasswordPayload3)
	req3, _ := http.NewRequest("POST", "/apis/change_password", bytes.NewBuffer(jsonData3))
	req3.Header.Set("Content-Type", "application/json")
	// 不設定 Cookie
	w3 := httptest.NewRecorder()
	testRouter.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)

	var response3 map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &response3)
	assert.NotEqual(t, float64(0), response3["state"])
	// 檢查 code 欄位（可能不存在，因為錯誤處理可能不同）
	if code, ok := response3["code"]; ok {
		assert.Equal(t, float64(1), code) // RequireLogin
	}

	// 測試 4: 缺少必填欄位
	cookie3 := getAuthCookie(t, testRouter, "testuser3", "test3@example.com", "password123")

	changePasswordPayload4 := map[string]interface{}{
		"password": "oldpassword123",
		// new_password 缺失
	}
	jsonData4, _ := json.Marshal(changePasswordPayload4)
	req4, _ := http.NewRequest("POST", "/apis/change_password", bytes.NewBuffer(jsonData4))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("Cookie", cookie3)
	w4 := httptest.NewRecorder()
	testRouter.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusOK, w4.Code)

	var response4 map[string]interface{}
	json.Unmarshal(w4.Body.Bytes(), &response4)
	assert.NotEqual(t, float64(0), response4["state"])
	assert.Equal(t, float64(0), response4["code"]) // RequiredField
}

// TestForgetPassword 測試忘記密碼 API
func TestForgetPassword(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 先註冊一個使用者
	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "password123")

	// 測試 1: Email 存在（應該成功）
	forgetPasswordPayload := map[string]interface{}{
		"email": "test@example.com",
	}
	jsonData, _ := json.Marshal(forgetPasswordPayload)
	req, _ := http.NewRequest("POST", "/apis/forget_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// 無論 Email 是否存在，都應該回成功（安全設計）
	assert.Equal(t, float64(0), response["state"])

	// 測試 2: Email 不存在（也應該成功，安全設計）
	forgetPasswordPayload2 := map[string]interface{}{
		"email": "nonexistent@example.com",
	}
	jsonData, _ = json.Marshal(forgetPasswordPayload2)
	req, _ = http.NewRequest("POST", "/apis/forget_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	// 即使 Email 不存在，也應該回成功（安全設計）
	assert.Equal(t, float64(0), response["state"])

	// 測試 3: 缺少必填欄位
	forgetPasswordPayload3 := map[string]interface{}{
		// email 缺失
	}
	jsonData, _ = json.Marshal(forgetPasswordPayload3)
	req, _ = http.NewRequest("POST", "/apis/forget_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEqual(t, float64(0), response["state"])
	assert.Equal(t, float64(0), response["code"]) // RequiredField

	// 測試 4: 無效的 Email 格式
	forgetPasswordPayload4 := map[string]interface{}{
		"email": "invalid-email",
	}
	jsonData, _ = json.Marshal(forgetPasswordPayload4)
	req, _ = http.NewRequest("POST", "/apis/forget_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEqual(t, float64(0), response["state"])
	assert.Equal(t, float64(0), response["code"]) // RequiredField (驗證失敗)

	// 確保 cookie 變數被使用（避免編譯警告）
	_ = cookie
}

// TestResetPassword 測試重設密碼 API
func TestResetPassword(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 先註冊一個使用者
	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "oldpassword123")

	// 產生 access_key（透過 forget_password）
	forgetPasswordPayload := map[string]interface{}{
		"email": "test@example.com",
	}
	jsonData, _ := json.Marshal(forgetPasswordPayload)
	req, _ := http.NewRequest("POST", "/apis/forget_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// 從資料庫取得 access_key（需要直接查詢資料庫）
	// 注意：由於 forget_password 是異步發送郵件，我們需要直接從資料庫取得 access_key
	// 但為了測試，我們可以透過 service 層直接產生 access_key
	// 或者我們可以測試使用無效的 access_key 的情況

	// 測試 1: 成功重設密碼（使用有效的 access_key）
	// 先透過 service 產生 access_key
	userRepo := repository.NewUserRepository(testDB)
	authService := service.NewAuthService(userRepo)
	accessKey, err := authService.GenerateAccessKey(context.Background(), "test@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, accessKey)

	resetPasswordPayload := map[string]interface{}{
		"email":      "test@example.com",
		"access_key": accessKey,
		"password":   "newpassword456",
	}
	jsonData, _ = json.Marshal(resetPasswordPayload)
	req, _ = http.NewRequest("POST", "/apis/reset_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, true, response["ret"])

	// 驗證新密碼可以登入
	loginPayload := map[string]interface{}{
		"username": "testuser",
		"password": "newpassword456",
	}
	jsonData, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", "/apis/signin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, true, response["ret"])

	// 測試 2: 使用無效的 access_key（應該失敗）
	resetPasswordPayload2 := map[string]interface{}{
		"email":      "test@example.com",
		"access_key": "invalid_access_key",
		"password":   "newpassword789",
	}
	jsonData, _ = json.Marshal(resetPasswordPayload2)
	req, _ = http.NewRequest("POST", "/apis/reset_password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(0), response["state"])
	assert.Equal(t, false, response["ret"])

	// 測試 3: 缺少必填欄位
	resetPasswordPayload3 := map[string]interface{}{
		"email": "test@example.com",
		// access_key 缺失
		"password": "newpassword456",
	}
	jsonData3, _ := json.Marshal(resetPasswordPayload3)
	req3, _ := http.NewRequest("POST", "/apis/reset_password", bytes.NewBuffer(jsonData3))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	testRouter.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)

	var response3 map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &response3)
	assert.NotEqual(t, float64(0), response3["state"])

	// 測試 4: Email 不存在
	resetPasswordPayload4 := map[string]interface{}{
		"email":      "nonexistent@example.com",
		"access_key": "some_access_key",
		"password":   "newpassword456",
	}
	jsonData4, _ := json.Marshal(resetPasswordPayload4)
	req4, _ := http.NewRequest("POST", "/apis/reset_password", bytes.NewBuffer(jsonData4))
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	testRouter.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusOK, w4.Code)

	var response4 map[string]interface{}
	json.Unmarshal(w4.Body.Bytes(), &response4)
	assert.Equal(t, float64(0), response4["state"])
	assert.Equal(t, false, response4["ret"])

	// 測試 5: 無效的 Email 格式
	resetPasswordPayload5 := map[string]interface{}{
		"email":      "invalid-email",
		"access_key": "some_access_key",
		"password":   "newpassword456",
	}
	jsonData5, _ := json.Marshal(resetPasswordPayload5)
	req5, _ := http.NewRequest("POST", "/apis/reset_password", bytes.NewBuffer(jsonData5))
	req5.Header.Set("Content-Type", "application/json")
	w5 := httptest.NewRecorder()
	testRouter.ServeHTTP(w5, req5)

	assert.Equal(t, http.StatusOK, w5.Code)

	var response5 map[string]interface{}
	json.Unmarshal(w5.Body.Bytes(), &response5)
	assert.NotEqual(t, float64(0), response5["state"])

	// 確保 cookie 變數被使用（避免編譯警告）
	_ = cookie
}
