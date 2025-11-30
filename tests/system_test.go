package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCheck 測試健康檢查端點
func TestHealthCheck(t *testing.T) {
	ctx := SetupTestDB(t)
	defer CleanupTestDB(t, ctx)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 檢查 state 是否為 0（成功）
	state, ok := response["state"]
	require.True(t, ok, "response should have 'state' field")
	assert.Equal(t, float64(0), state, "state should be 0 for success")

	// 檢查 Data 欄位（注意：JSON 回應中是大寫 Data，因為 Response 結構體的欄位名稱）
	data, ok := response["Data"]
	if !ok {
		// 如果沒有 Data，嘗試小寫 data（向後相容）
		data, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field when state is 0")
	require.NotNil(t, data, "data should not be nil")

	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, "ok", dataMap["status"])
}

// TestReadinessCheck 測試就緒檢查端點
func TestReadinessCheck(t *testing.T) {
	ctx := SetupTestDB(t)
	defer CleanupTestDB(t, ctx)

	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 檢查 state 是否為 0（成功）
	state, ok := response["state"]
	require.True(t, ok, "response should have 'state' field")
	assert.Equal(t, float64(0), state, "state should be 0 for success")

	// 檢查 Data 欄位
	data, ok := response["Data"]
	if !ok {
		data, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field when state is 0")
	require.NotNil(t, data, "data should not be nil")

	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, "ready", dataMap["status"])
	assert.Equal(t, "higgstv-api", dataMap["service"])
	assert.Equal(t, "connected", dataMap["database"])
}

