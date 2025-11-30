package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPickProgramWithYouTubeID 測試 PickProgram 的 youtube_id 參數
func TestPickProgramWithYouTubeID(t *testing.T) {
	ctx := SetupTestDB(t)
	defer CleanupTestDB(t, ctx)

	cookie := getAuthCookie(t, ctx, "testuser", "test@example.com", "testpass123")

	// 測試使用 youtube_id 參數
	req, _ := http.NewRequest("GET", "/apis/pickprog?callback=testCallback&name=Test+Video&youtube_id=dQw4w9WgXcQ&desc=Test+description&duration=300&tags=1,2", nil)
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testCallback")
	assert.Contains(t, w.Body.String(), "state")
	assert.Contains(t, w.Body.String(), "program")
}

// TestPickProgramWithURL 測試 PickProgram 的 url 參數（向後相容）
func TestPickProgramWithURL(t *testing.T) {
	ctx := SetupTestDB(t)
	defer CleanupTestDB(t, ctx)

	cookie := getAuthCookie(t, ctx, "testuser2", "test2@example.com", "testpass123")

	// 測試使用 url 參數（向後相容）
	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	encodedURL := url.QueryEscape(testURL)
	req, _ := http.NewRequest("GET", "/apis/pickprog?callback=testCallback&name=Test+Video&url="+encodedURL, nil)
	req.Header.Set("Cookie", cookie)
	w := httptest.NewRecorder()
	ctx.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testCallback")
	assert.Contains(t, w.Body.String(), "state")
	assert.Contains(t, w.Body.String(), "program")
}

