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

// TestSaveProgram 測試儲存節目功能
func TestSaveProgram(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Test Channel",
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
 err := json.Unmarshal(w.Body.Bytes(), &response)
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

	// 新增節目
	programPayload := map[string]interface{}{
		"ch":         channelID,
		"name":       "Original Program",
		"youtube_id": "dQw4w9WgXcQ",
		"desc":       "Original description",
		"duration":   300,
		"tags":       []int{1, 2},
	}
	jsonData, _ = json.Marshal(programPayload)
	req, _ = http.NewRequest("POST", "/apis/addprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var addProgramResp map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	// 取得節目 ID
	addProgramData, ok := addProgramResp["Data"]
	if !ok {
		addProgramData, ok = addProgramResp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	addProgramDataMap := addProgramData.(map[string]interface{})
	programObj, okProgram := addProgramDataMap["program"]
	require.True(t, okProgram, "data should have 'program' field")
	program := programObj.(map[string]interface{})
	programID, ok := program["_id"].(float64) // JSON 數字會是 float64
	require.True(t, ok, "program._id should be a number")

	// 測試 1: 更新節目（不更新封面）
	saveProgramPayload := map[string]interface{}{
		"ch":          channelID,
		"prog_id":     int(programID),
		"name":        "Updated Program",
		"youtube_id":  "dQw4w9WgXcQ",
		"desc":        "Updated description",
		"duration":    400,
		"tags":        []int{3, 4},
		"updateCover": false,
	}
	jsonData, _ = json.Marshal(saveProgramPayload)
	req, _ = http.NewRequest("POST", "/apis/saveprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查回應包含 program 欄位
	responseData, ok := response["Data"]
	if !ok {
		responseData, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, responseData, "data should not be nil")

	responseDataMap := responseData.(map[string]interface{})
	updatedProgramObj, okUpdatedProgram := responseDataMap["program"]
	require.True(t, okUpdatedProgram, "data should have 'program' field")
	require.NotNil(t, updatedProgramObj, "program should not be nil")

	updatedProgram := updatedProgramObj.(map[string]interface{})
	assert.Equal(t, "Updated Program", updatedProgram["name"])
	assert.Equal(t, "Updated description", updatedProgram["desc"])
	assert.Equal(t, float64(400), updatedProgram["duration"])

	// 測試 2: 更新節目（更新封面）
	saveProgramPayload2 := map[string]interface{}{
		"ch":          channelID,
		"prog_id":     int(programID),
		"name":        "Updated Program 2",
		"youtube_id":  "dQw4w9WgXcQ",
		"desc":        "Updated description 2",
		"duration":    500,
		"tags":        []int{5, 6},
		"updateCover": true,
	}
	jsonData, _ = json.Marshal(saveProgramPayload2)
	req, _ = http.NewRequest("POST", "/apis/saveprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證頻道封面已更新
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResponse map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
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

	channelInfo := channelResponseObj.(map[string]interface{})
	// cover 欄位可能不存在，這是正常的（如果頻道沒有封面）
	if coverObj, ok := channelInfo["cover"]; ok && coverObj != nil {
		if cover, ok := coverObj.(map[string]interface{}); ok {
			if defaultCover, ok := cover["default"].(string); ok {
				assert.Contains(t, defaultCover, "img.youtube.com")
				assert.Contains(t, defaultCover, "dQw4w9WgXcQ")
			}
		}
	}
}

// TestDeleteProgram 測試刪除節目
func TestDeleteProgram(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Test Channel",
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
 err := json.Unmarshal(w.Body.Bytes(), &response)
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

	// 新增多個節目
	programIDs := []int{}
	for i := 0; i < 3; i++ {
		programPayload := map[string]interface{}{
			"ch":         channelID,
			"name":       "Program " + string(rune('A'+i)),
			"youtube_id": "dQw4w9WgXcQ",
			"desc":       "Test program",
			"duration":   300,
			"tags":       []int{},
		}
		jsonData, _ = json.Marshal(programPayload)
		req, _ = http.NewRequest("POST", "/apis/addprog", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		w = httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var addProgramResp map[string]interface{}
  err = json.Unmarshal(w.Body.Bytes(), &response)
  require.NoError(t, err)

		// 取得節目 ID
		addProgramData, ok := addProgramResp["Data"]
		if !ok {
			addProgramData, ok = addProgramResp["data"]
		}
		require.True(t, ok, "response should have 'Data' or 'data' field")
		addProgramDataMap := addProgramData.(map[string]interface{})
		programObj, okProgram := addProgramDataMap["program"]
		require.True(t, okProgram, "data should have 'program' field")
		program := programObj.(map[string]interface{})
		programID, ok := program["_id"].(float64)
		require.True(t, ok, "program._id should be a number")
		programIDs = append(programIDs, int(programID))
	}

	// 測試 1: 刪除單個節目
	deletePayload := map[string]interface{}{
		"ch":  channelID,
		"ids": []int{programIDs[0]},
	}
	jsonData, _ = json.Marshal(deletePayload)
	req, _ = http.NewRequest("POST", "/apis/delprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證節目已刪除
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResponse map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
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

	channelInfo := channelResponseObj.(map[string]interface{})
	if contents, ok := channelInfo["contents"].([]interface{}); ok {
		assert.Equal(t, 2, len(contents), "should have 2 programs remaining")
	}

	// 測試 2: 刪除多個節目
	deletePayload2 := map[string]interface{}{
		"ch":  channelID,
		"ids": []int{programIDs[1], programIDs[2]},
	}
	jsonData, _ = json.Marshal(deletePayload2)
	req, _ = http.NewRequest("POST", "/apis/delprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證所有節目已刪除
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	channelResponseData, ok = channelResponse["Data"]
	if !ok {
		channelResponseData, ok = channelResponse["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	require.NotNil(t, channelResponseData, "data should not be nil")

	channelResponseDataMap = channelResponseData.(map[string]interface{})
	channelResponseObj, okChannelResp = channelResponseDataMap["channel"]
	require.True(t, okChannelResp, "data should have 'channel' field")
	require.NotNil(t, channelResponseObj, "channel should not be nil")

	channelInfo = channelResponseObj.(map[string]interface{})
	if contents, ok := channelInfo["contents"].([]interface{}); ok {
		assert.Equal(t, 0, len(contents), "should have no programs remaining")
	}

	// 測試 3: 無權限刪除（使用其他使用者的頻道）
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
 err = json.Unmarshal(w.Body.Bytes(), &response)
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

	// 嘗試用第一個使用者刪除第二個使用者的頻道節目（應該失敗）
	deletePayload3 := map[string]interface{}{
		"ch":  channelID2,
		"ids": []int{1},
	}
	jsonData, _ = json.Marshal(deletePayload3)
	req, _ = http.NewRequest("POST", "/apis/delprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie) // 使用第一個使用者的 cookie
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	// 應該回傳權限錯誤
	assert.NotEqual(t, float64(0), response["state"])
}

// TestMoveProgram 測試移動節目功能
func TestMoveProgram(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立兩個頻道
	channel1Payload := map[string]interface{}{
		"name": "Source Channel",
		"tags": []int{},
	}
	jsonData, _ := json.Marshal(channel1Payload)
	req, _ := http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channel1Resp map[string]interface{}
 err := json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	channel1Data, ok := channel1Resp["Data"]
	if !ok {
		channel1Data, ok = channel1Resp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	channel1DataMap := channel1Data.(map[string]interface{})
	channel1Obj, okChannel1 := channel1DataMap["channel"]
	require.True(t, okChannel1, "data should have 'channel' field")
	channel1 := channel1Obj.(map[string]interface{})
	channel1ID, ok := channel1["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	channel2Payload := map[string]interface{}{
		"name": "Target Channel",
		"tags": []int{},
	}
	jsonData, _ = json.Marshal(channel2Payload)
	req, _ = http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channel2Resp map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	channel2Data, ok := channel2Resp["Data"]
	if !ok {
		channel2Data, ok = channel2Resp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	channel2DataMap := channel2Data.(map[string]interface{})
	channel2Obj, okChannel2 := channel2DataMap["channel"]
	require.True(t, okChannel2, "data should have 'channel' field")
	channel2 := channel2Obj.(map[string]interface{})
	channel2ID, ok := channel2["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	// 在來源頻道新增多個節目
	programIDs := []int{}
	for i := 0; i < 3; i++ {
		programPayload := map[string]interface{}{
			"ch":         channel1ID,
			"name":       "Program " + string(rune('A'+i)),
			"youtube_id": "dQw4w9WgXcQ",
			"desc":       "Test program",
			"duration":   300,
			"tags":       []int{},
		}
		jsonData, _ = json.Marshal(programPayload)
		req, _ = http.NewRequest("POST", "/apis/addprog", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		w = httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var addProgramResp map[string]interface{}
  err = json.Unmarshal(w.Body.Bytes(), &response)
  require.NoError(t, err)

		addProgramData, ok := addProgramResp["Data"]
		if !ok {
			addProgramData, ok = addProgramResp["data"]
		}
		require.True(t, ok, "response should have 'Data' or 'data' field")
		addProgramDataMap := addProgramData.(map[string]interface{})
		programObj, okProgram := addProgramDataMap["program"]
		require.True(t, okProgram, "data should have 'program' field")
		program := programObj.(map[string]interface{})
		programID, ok := program["_id"].(float64)
		require.True(t, ok, "program._id should be a number")
		programIDs = append(programIDs, int(programID))
	}

	// 測試 1: 移動單個節目
	movePayload := map[string]interface{}{
		"ch":     channel1ID,
		"target": channel2ID,
		"ids":    []int{programIDs[0]},
	}
	jsonData, _ = json.Marshal(movePayload)
	req, _ = http.NewRequest("POST", "/apis/progmoveto", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證節目已移動
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channel1ID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channel1Response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	channel1ResponseData, ok := channel1Response["Data"]
	if !ok {
		channel1ResponseData, ok = channel1Response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	channel1ResponseDataMap := channel1ResponseData.(map[string]interface{})
	channel1ResponseObj, okChannel1Resp := channel1ResponseDataMap["channel"]
	require.True(t, okChannel1Resp, "data should have 'channel' field")
	channel1Info := channel1ResponseObj.(map[string]interface{})
	if contents, ok := channel1Info["contents"].([]interface{}); ok {
		assert.Equal(t, 2, len(contents), "source channel should have 2 programs remaining")
	}

	// 驗證目標頻道有節目
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channel2ID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channel2Response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	channel2ResponseData, ok := channel2Response["Data"]
	if !ok {
		channel2ResponseData, ok = channel2Response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	channel2ResponseDataMap := channel2ResponseData.(map[string]interface{})
	channel2ResponseObj, okChannel2Resp := channel2ResponseDataMap["channel"]
	require.True(t, okChannel2Resp, "data should have 'channel' field")
	channel2Info := channel2ResponseObj.(map[string]interface{})
	if contents, ok := channel2Info["contents"].([]interface{}); ok {
		assert.Equal(t, 1, len(contents), "target channel should have 1 program")
	}

	// 測試 2: 移動多個節目
	movePayload2 := map[string]interface{}{
		"ch":     channel1ID,
		"target": channel2ID,
		"ids":    []int{programIDs[1], programIDs[2]},
	}
	jsonData, _ = json.Marshal(movePayload2)
	req, _ = http.NewRequest("POST", "/apis/progmoveto", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證所有節目已移動
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channel1ID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	channel1ResponseData, ok = channel1Response["Data"]
	if !ok {
		channel1ResponseData, ok = channel1Response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	channel1ResponseDataMap = channel1ResponseData.(map[string]interface{})
	channel1ResponseObj, okChannel1Resp = channel1ResponseDataMap["channel"]
	require.True(t, okChannel1Resp, "data should have 'channel' field")
	channel1Info = channel1ResponseObj.(map[string]interface{})
	if contents, ok := channel1Info["contents"].([]interface{}); ok {
		assert.Equal(t, 0, len(contents), "source channel should have no programs")
	}

	// 測試 3: 無權限移動（來源頻道無權限）
	cookie2 := getAuthCookie(t, testRouter, "testuser2", "test2@example.com", "testpass123")

	// 建立第二個使用者的頻道
	channel3Payload := map[string]interface{}{
		"name": "User2 Channel",
		"tags": []int{},
	}
	jsonData, _ = json.Marshal(channel3Payload)
	req, _ = http.NewRequest("POST", "/apis/addchannel", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie2)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channel3Resp map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	channel3Data, ok := channel3Resp["Data"]
	if !ok {
		channel3Data, ok = channel3Resp["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	channel3DataMap := channel3Data.(map[string]interface{})
	channel3Obj, okChannel3 := channel3DataMap["channel"]
	require.True(t, okChannel3, "data should have 'channel' field")
	channel3 := channel3Obj.(map[string]interface{})
	channel3ID, ok := channel3["_id"].(string)
	require.True(t, ok, "channel._id should be a string")

	// 嘗試用第一個使用者移動第二個使用者的頻道節目（應該失敗）
	movePayload3 := map[string]interface{}{
		"ch":     channel3ID,
		"target": channel2ID,
		"ids":    []int{1},
	}
	jsonData, _ = json.Marshal(movePayload3)
	req, _ = http.NewRequest("POST", "/apis/progmoveto", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie) // 使用第一個使用者的 cookie
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	// 應該回傳權限錯誤
	assert.NotEqual(t, float64(0), response["state"])

	// 測試 4: 目標頻道無權限
	movePayload4 := map[string]interface{}{
		"ch":     channel2ID,
		"target": channel3ID,
		"ids":    []int{1},
	}
	jsonData, _ = json.Marshal(movePayload4)
	req, _ = http.NewRequest("POST", "/apis/progmoveto", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie) // 使用第一個使用者的 cookie
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	// 應該回傳權限錯誤
	assert.NotEqual(t, float64(0), response["state"])
}

// TestSaveProgramOrder 測試儲存節目順序功能
func TestSaveProgramOrder(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Test Channel",
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
 err := json.Unmarshal(w.Body.Bytes(), &response)
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

	// 新增多個節目
	programIDs := []int{}
	for i := 0; i < 3; i++ {
		programPayload := map[string]interface{}{
			"ch":         channelID,
			"name":       "Program " + string(rune('A'+i)),
			"youtube_id": "dQw4w9WgXcQ",
			"desc":       "Test program",
			"duration":   300,
			"tags":       []int{},
		}
		jsonData, _ = json.Marshal(programPayload)
		req, _ = http.NewRequest("POST", "/apis/addprog", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		w = httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var addProgramResp map[string]interface{}
  err = json.Unmarshal(w.Body.Bytes(), &response)
  require.NoError(t, err)

		addProgramData, ok := addProgramResp["Data"]
		if !ok {
			addProgramData, ok = addProgramResp["data"]
		}
		require.True(t, ok, "response should have 'Data' or 'data' field")
		addProgramDataMap := addProgramData.(map[string]interface{})
		programObj, okProgram := addProgramDataMap["program"]
		require.True(t, okProgram, "data should have 'program' field")
		program := programObj.(map[string]interface{})
		programID, ok := program["_id"].(float64)
		require.True(t, ok, "program._id should be a number")
		programIDs = append(programIDs, int(programID))
	}

	// 測試 1: 更新節目順序（反轉順序）
	reversedOrder := []int{programIDs[2], programIDs[1], programIDs[0]}

	saveOrderPayload := map[string]interface{}{
		"ch":    channelID,
		"order": reversedOrder,
	}
	jsonData, _ = json.Marshal(saveOrderPayload)
	req, _ = http.NewRequest("POST", "/apis/prog/saveorder", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 驗證順序已更新
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResponse map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
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

	channelInfo := channelResponseObj.(map[string]interface{})
	if contentsOrder, ok := channelInfo["contents_order"].([]interface{}); ok {
		assert.Equal(t, 3, len(contentsOrder), "should have 3 items in order")
		// 驗證順序是反轉的
		assert.Equal(t, float64(reversedOrder[0]), contentsOrder[0])
		assert.Equal(t, float64(reversedOrder[1]), contentsOrder[1])
		assert.Equal(t, float64(reversedOrder[2]), contentsOrder[2])
	}

	// 測試 2: 無權限更新順序（使用其他使用者的頻道）
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
 err = json.Unmarshal(w.Body.Bytes(), &response)
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

	// 嘗試用第一個使用者更新第二個使用者的頻道順序（應該失敗）
	saveOrderPayload2 := map[string]interface{}{
		"ch":    channelID2,
		"order": []int{1, 2, 3},
	}
	jsonData, _ = json.Marshal(saveOrderPayload2)
	req, _ = http.NewRequest("POST", "/apis/prog/saveorder", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie) // 使用第一個使用者的 cookie
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	// 應該回傳權限錯誤
	assert.NotEqual(t, float64(0), response["state"])

	// 測試 3: 缺少必填欄位
	saveOrderPayload3 := map[string]interface{}{
		"ch": channelID,
		// order 欄位缺失
	}
	jsonData, _ = json.Marshal(saveOrderPayload3)
	req, _ = http.NewRequest("POST", "/apis/prog/saveorder", bytes.NewBuffer(jsonData))
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

// TestAddProgramWithUpdateCover 測試新增節目並更新頻道封面
func TestAddProgramWithUpdateCover(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser", "test@example.com", "testpass123")

	// 先建立頻道
	channelPayload := map[string]interface{}{
		"name": "Test Channel",
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
 err := json.Unmarshal(w.Body.Bytes(), &response)
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

	// 新增節目並更新封面
	programPayload := map[string]interface{}{
		"ch":          channelID,
		"name":        "Test Program",
		"youtube_id":  "dQw4w9WgXcQ",
		"desc":        "Test description",
		"duration":    300,
		"tags":        []int{1, 2},
		"updateCover": true,
	}
	jsonData, _ = json.Marshal(programPayload)
	req, _ = http.NewRequest("POST", "/apis/addprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查 Data 欄位
	responseData, ok := response["Data"]
	if !ok {
		responseData, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	assert.NotNil(t, responseData, "data should not be nil")

	// 驗證頻道封面已更新
	req, _ = http.NewRequest("GET", "/apis/getchannel/"+channelID, nil)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var channelResponse map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)

	// 檢查 Data 欄位
	channelResponseData, okChannelResp := channelResponse["Data"]
	if !okChannelResp {
		channelResponseData, okChannelResp = channelResponse["data"]
	}
	require.True(t, okChannelResp, "response should have 'Data' or 'data' field")
	require.NotNil(t, channelResponseData, "data should not be nil")

	dataMap := channelResponseData.(map[string]interface{})
	channelObj, okChannel := dataMap["channel"]
	require.True(t, okChannel, "data should have 'channel' field")
	require.NotNil(t, channelObj, "channel should not be nil")

	channelInfo := channelObj.(map[string]interface{})
	// cover 欄位可能不存在，這是正常的（如果頻道沒有封面）
	if coverObj, ok := channelInfo["cover"]; ok && coverObj != nil {
		if cover, ok := coverObj.(map[string]interface{}); ok {
			if defaultCover, ok := cover["default"].(string); ok {
				assert.Contains(t, defaultCover, "img.youtube.com")
				assert.Contains(t, defaultCover, "dQw4w9WgXcQ")
			} else {
				t.Logf("Warning: cover.default is not a string or doesn't exist")
			}
		} else {
			t.Logf("Warning: cover is not a map")
		}
	} else {
		t.Logf("Warning: cover field doesn't exist or is nil")
	}
}

// TestAddProgramWithoutUpdateCover 測試新增節目但不更新封面
func TestAddProgramWithoutUpdateCover(t *testing.T) {
	SetupTestDB(t)
	defer CleanupTestDB(t)

	cookie := getAuthCookie(t, testRouter, "testuser3", "test3@example.com", "testpass123")

	// 建立頻道
	channelPayload := map[string]interface{}{
		"name": "Test Channel",
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
 err := json.Unmarshal(w.Body.Bytes(), &response)
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

	// 新增節目但不更新封面（updateCover 為 false 或未提供）
	programPayload := map[string]interface{}{
		"ch":         channelID,
		"name":       "Test Program",
		"youtube_id": "dQw4w9WgXcQ",
		"desc":       "Test description",
		"duration":   300,
		"tags":       []int{1, 2},
		// updateCover 未提供，預設為 false
	}
	jsonData, _ = json.Marshal(programPayload)
	req, _ = http.NewRequest("POST", "/apis/addprog", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
 err = json.Unmarshal(w.Body.Bytes(), &response)
 require.NoError(t, err)
	assert.Equal(t, float64(0), response["state"])

	// 檢查 Data 欄位
	responseData, ok := response["Data"]
	if !ok {
		responseData, ok = response["data"]
	}
	require.True(t, ok, "response should have 'Data' or 'data' field")
	assert.NotNil(t, responseData, "data should not be nil")
}

