package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetChannel 測試取得頻道（獨立測試）
func TestGetChannel(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Public Channel",
		"tags": []int{},
	}
	jsonData, _ := json.Marshal(channelPayload)
	req, _ := http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &channelResp)
	require.NoError(t, err)

	// 檢查 Data 欄位
	channelData, ok := channelResp["Data"]
	if !ok {
		channelData, ok = channelResp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, channelData, "data should not be nil")

	channelDataMap := channelData.(map[string]interface{})
	channelObj, okChannel := channelDataMap["channel"]
	require.True(t, okChannel, "data should have 'channel' field")
	require.NotNil(t, channelObj, "channel should not be nil")

	channel := channelObj.(map[string]interface{})
	channelID, ok := channel["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	// 測試 1: 無需登入即可取得頻道（公開頻道）
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查回應包含 channel 欄位
	responseData, ok := response["Data"]
	if !ok {
		responseData, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, responseData, "data should not be nil")

	responseDataMap := responseData.(map[string]interface{})
	retrievedChannelObj, okRetrievedChannel := responseDataMap["channel"]
	require.True(t, okRetrievedChannel, "data should have 'channel' field")
	require.NotNil(t, retrievedChannelObj, "channel should not be nil")

	retrievedChannel := retrievedChannelObj.(map[string]interface{})
	assert.Equal(t, channelID, retrievedChannel["_id"])
	assert.Equal(t, "Public Channel", retrievedChannel["name"])

	// 測試 2: 登入後取得頻道
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 測試 3: 取得不存在的頻道
	req, _ = http.NewRequest("GET", "/apis/getchannel/nonexistent", nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// 應該回傳錯誤或空資料
	assert.NotEqual(t, float64(0), response["state"])
}

// TestGetChannelInfo 測試取得頻道資訊（獨立測試）
func TestGetChannelInfo(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Test Channel Info",
		"tags": []int{},
	}
	jsonData, _ := json.Marshal(channelPayload)
	req, _ := http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &channelResp)
	require.NoError(t, err)

	// 檢查 Data 欄位
	channelData, ok := channelResp["Data"]
	if !ok {
		channelData, ok = channelResp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, channelData, "data should not be nil")

	channelDataMap := channelData.(map[string]interface{})
	channelObj, okChannel := channelDataMap["channel"]
	require.True(t, okChannel, "data should have 'channel' field")
	require.NotNil(t, channelObj, "channel should not be nil")

	channel := channelObj.(map[string]interface{})
	channelID, ok := channel["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	// 測試 1: 需要登入才能取得頻道資訊
	req, _ = http.NewRequest("GET", "/apis/getchannelinfo/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// 應該回傳需要登入的錯誤
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEqual(t, float64(0), response["state"])

	// 測試 2: 登入後取得頻道資訊
	req, _ = http.NewRequest("GET", "/apis/getchannelinfo/"+channelID, nil)
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查回應包含 channel 欄位和 owners_info
	responseData, ok := response["Data"]
	if !ok {
		responseData, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, responseData, "data should not be nil")

	responseDataMap := responseData.(map[string]interface{})
	retrievedChannelObj, okRetrievedChannel := responseDataMap["channel"]
	require.True(t, okRetrievedChannel, "data should have 'channel' field")
	require.NotNil(t, retrievedChannelObj, "channel should not be nil")

	retrievedChannel := retrievedChannelObj.(map[string]interface{})
	assert.Equal(t, channelID, retrievedChannel["_id"])
	assert.Equal(t, "Test Channel Info", retrievedChannel["name"])

	// 檢查 owners_info 欄位（如果存在）
	if ownersInfo, ok := responseDataMap["owners_info"]; ok {
		require.NotNil(t, ownersInfo, "owners_info should not be nil")
		ownersInfoArray, ok := ownersInfo.([]interface{})
		if ok {
			assert.GreaterOrEqual(t, len(ownersInfoArray), 1, "should have at least one owner")
		}
	}
}

// TestSaveChannel 測試儲存頻道功能
func TestSaveChannel(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Original Channel",
		"tags": []int{1, 2},
	}
	jsonData, _ := json.Marshal(channelPayload)
	req, _ := http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &channelResp)
	require.NoError(t, err)

	// 檢查 Data 欄位
	channelData, ok := channelResp["Data"]
	if !ok {
		channelData, ok = channelResp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, channelData, "data should not be nil")

	channelDataMap := channelData.(map[string]interface{})
	channelObj, okChannel := channelDataMap["channel"]
	require.True(t, okChannel, "data should have 'channel' field")
	require.NotNil(t, channelObj, "channel should not be nil")

	channel := channelObj.(map[string]interface{})
	channelID, ok := channel["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	// 測試 1: 更新頻道名稱、描述和標籤
	saveChannelPayload := map[string]interface{}{
		"id":   channelID,
		"name": "Updated Channel Name",
		"desc": "Updated channel description",
		"tags": []int{3, 4, 5},
	}
	jsonData, _ = json.Marshal(saveChannelPayload)
	req, _ = http.NewRequest("POST", "/apis/savechannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證頻道已更新
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &channelResponse)
	require.NoError(t, err)

	channelResponseData, ok := channelResponse["Data"]
	if !ok {
		channelResponseData, ok = channelResponse["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, channelResponseData, "data should not be nil")

	channelResponseDataMap := channelResponseData.(map[string]interface{})
	channelResponseObj, okChannelResp := channelResponseDataMap["channel"]
	require.True(t, okChannelResp, "data should have 'channel' field")
	require.NotNil(t, channelResponseObj, "channel should not be nil")

	updatedChannel := channelResponseObj.(map[string]interface{})
	assert.Equal(t, "Updated Channel Name", updatedChannel["name"])
	assert.Equal(t, "Updated channel description", updatedChannel["desc"])

	// 驗證標籤已更新
	if tags, ok := updatedChannel["tags"].([]interface{}); ok {
		assert.Equal(t, 3, len(tags), "should have 3 tags")
		assert.Contains(t, tags, float64(3))
		assert.Contains(t, tags, float64(4))
		assert.Contains(t, tags, float64(5))
	}

	// 測試 2: 只更新名稱（不更新描述和標籤）
	saveChannelPayload2 := map[string]interface{}{
		"id":   channelID,
		"name": "Final Channel Name",
	}
	jsonData, _ = json.Marshal(saveChannelPayload2)
	req, _ = http.NewRequest("POST", "/apis/savechannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 測試 3: 無權限更新（使用其他使用者的頻道）
	cookie2 := getAuthCookie(t, testRouter, "testuser2", "test2@example.com", "testpass123")

	// 建立第二個使用者的頻道
	channelPayload2 := map[string]interface{}{
		"name": "User2 Channel",
		"tags": []int{},
	}
	jsonData, _ = json.Marshal(channelPayload2)
	req, _ = http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie2)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResp2 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &channelResp2)
	require.NoError(t, err)

	channelData2, ok := channelResp2["Data"]
	if !ok {
		channelData2, ok = channelResp2["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	channelDataMap2 := channelData2.(map[string]interface{})
	channelObj2, okChannel2 := channelDataMap2["channel"]
	require.True(t, okChannel2, "data should have 'channel' field")
	channel2 := channelObj2.(map[string]interface{})
	channelID2, ok := channel2["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	// 嘗試用第一個使用者更新第二個使用者的頻道（應該失敗）
	saveChannelPayload3 := map[string]interface{}{
		"id":   channelID2,
		"name": "Unauthorized Update",
	}
	jsonData, _ = json.Marshal(saveChannelPayload3)
	req, _ = http.NewRequest("POST", "/apis/savechannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie) // 使用第一個使用者的 cookie
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// 應該回傳權限錯誤
	assert.NotEqual(t, float64(0), response["state"])

	// 測試 4: 缺少必填欄位
	saveChannelPayload4 := map[string]interface{}{
		"id": channelID,
		// name 欄位缺失
	}
	jsonData, _ = json.Marshal(saveChannelPayload4)
	req, _ = http.NewRequest("POST", "/apis/savechannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// 應該回傳缺少必填欄位錯誤
	assert.NotEqual(t, float64(0), response["state"])
}

// TestAddChannel 測試新增頻道（需要登入）
func TestAddChannel(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 先註冊並登入
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

	// 取得 Cookie
	cookie := w.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookie)

	// 新增頻道
	channelPayload := map[string]interface{}{
		"name": "Test Channel",
		"tags": []int{1, 2},
	}
	jsonData, _ = json.Marshal(channelPayload)
	req, _ = http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查 Data 欄位（注意：JSON 回應中是大寫 Data）
	data, ok := response["Data"]
	if !ok {
		data, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	assert.NotNil(t, data, "data should not be nil")
}

// TestGetOwnChannelsWithQueryParams 測試 GetOwnChannels 的 q 和 types[] 參數
func TestGetOwnChannelsWithQueryParams(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立多個頻道
	channels := []map[string]interface{}{
		{"name": "Music Channel", "tags": []int{}},
		{"name": "Tech Channel", "tags": []int{}},
		{"name": "Music Playlist", "tags": []int{}},
	}

	for _, ch := range channels {
		jsonData, _ := json.Marshal(ch)
		req, _ := http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	// 測試關鍵字搜尋（q 參數）
	req, _ := http.NewRequest("GET", "/apis/getownchannels?q=Music", nil)
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查 Data 欄位
	data, ok := response["Data"]
	if !ok {
		data, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, data, "data should not be nil")

	dataMap := data.(map[string]interface{})
	channelsObj, ok := dataMap["channels"]
	require.True(t, ok, "data should have 'channels' field")
	require.NotNil(t, channelsObj, "channels should not be nil")

	channelsData := channelsObj.([]interface{})
	assert.GreaterOrEqual(t, len(channelsData), 2) // 應該找到 "Music Channel" 和 "Music Playlist"

	// 驗證所有結果都包含 "Music"
	for _, ch := range channelsData {
		channel := ch.(map[string]interface{})
		name := channel["name"].(string)
		assert.Contains(t, name, "Music")
	}
}

// TestGetChannelsWithAllQueryParams 測試 GetChannels 的完整 query 參數
func TestGetChannelsWithAllQueryParams(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 建立使用者
	cookie1 := getAuthCookie(t, testRouter, "user1", "user1@example.com", "pass123")

	// user1 建立頻道
	channelPayload := map[string]interface{}{
		"name": "User1 Channel",
		"tags": []int{},
	}
	jsonData, _ := json.Marshal(channelPayload)
	req, _ := http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie1)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// 測試 user 參數
	req, _ = http.NewRequest("GET", "/apis/getchannels?user=user1", nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查 Data 欄位
	data, ok := response["Data"]
	if !ok {
		data, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, data, "data should not be nil")

	dataMap := data.(map[string]interface{})
	channelsObj, ok := dataMap["channels"]
	require.True(t, ok, "data should have 'channels' field")
	require.NotNil(t, channelsObj, "channels should not be nil")

	channelsData := channelsObj.([]interface{})
	assert.GreaterOrEqual(t, len(channelsData), 1)

	// 測試 q 參數（關鍵字搜尋）
	req, _ = http.NewRequest("GET", "/apis/getchannels?q=User1", nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 測試 has_contents 參數
	req, _ = http.NewRequest("GET", "/apis/getchannels?has_contents=1", nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 測試 ignore_types 參數
	req, _ = http.NewRequest("GET", "/apis/getchannels?ignore_types=unclassified", nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 測試 start 和 desc 參數
	req, _ = http.NewRequest("GET", "/apis/getchannels?start=0&desc=1&sort=last_modified", nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])
}

// TestSetChannelOwnerWithEmail 測試 SetChannelOwner 的 email 參數
func TestSetChannelOwnerWithEmail(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	// 建立兩個使用者
	cookie1 := getAuthCookie(t, testRouter, "owner1", "owner1@example.com", "pass123")
	_ = getAuthCookie(t, testRouter, "owner2", "owner2@example.com", "pass123") // 建立第二個使用者

	// owner1 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Shared Channel",
		"tags": []int{},
	}
	jsonData, _ := json.Marshal(channelPayload)
	req, _ := http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie1)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &channelResp)
	require.NoError(t, err)

	// 檢查 Data 欄位
	channelData, ok := channelResp["Data"]
	if !ok {
		channelData, ok = channelResp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, channelData, "data should not be nil")

	channelDataMap := channelData.(map[string]interface{})
	channelObj, ok := channelDataMap["channel"]
	require.True(t, ok, "data should have 'channel' field")
	require.NotNil(t, channelObj, "channel should not be nil")

	channel := channelObj.(map[string]interface{})
	channelID, ok := channel["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	// 使用 email 參數新增共用者
	setOwnerPayload := map[string]interface{}{
		"id":    channelID,
		"c":     channelID,
		"email": "owner2@example.com",
	}
	jsonData, _ = json.Marshal(setOwnerPayload)
	req, _ = http.NewRequest("POST", "/apis/setchannelowner", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie1)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證頻道擁有者已更新
	req, _ = http.NewRequest("GET", "/apis/getchannelinfo/"+channelID, nil)
	req.Header.Set("Cookie", cookie1)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 檢查 Data 欄位
	responseData, ok := response["Data"]
	if !ok {
		responseData, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, responseData, "data should not be nil")

	responseDataMap := responseData.(map[string]interface{})
	channelObj, okChannel := responseDataMap["channel"]
	require.True(t, okChannel, "data should have 'channel' field")
	require.NotNil(t, channelObj, "channel should not be nil")

	channelInfo := channelObj.(map[string]interface{})
	owners, okOwners := channelInfo["owners"].([]interface{})
	require.True(t, okOwners, "channel.owners should be an array")
	assert.GreaterOrEqual(t, len(owners), 2) // 應該有至少兩個擁有者
}

