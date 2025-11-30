package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/repository"
	"github.com/higgstv/higgstv-go/internal/service"
	"github.com/higgstv/higgstv-go/pkg/session"
	"github.com/higgstv/higgstv-go/pkg/utils"
)

// PickProgramRequest Pick 節目請求（Query 參數）
type PickProgramRequest struct {
	Callback  string `form:"callback" example:"callback"` // JSONP callback 函數名
	Name      string `form:"name" example:"影片名稱"` // 影片名稱
	YouTubeID string `form:"youtube_id" example:"dQw4w9WgXcQ"` // YouTube 影片 ID（文件規範）
	URL       string `form:"url" example:"https://www.youtube.com/watch?v=dQw4w9WgXcQ"` // YouTube URL（向後相容）
	Desc      string `form:"desc" example:"影片描述"` // 描述（選填）
	Duration  string `form:"duration" example:"300"` // 時長（秒，選填）
	Tags      string `form:"tags" example:"1,2,3"` // 標籤（選填，逗號分隔）
}

// PickProgram Pick 節目（Bookmarklet API，支援 JSONP）
// @Summary      Pick 節目（Bookmarklet）
// @Description  透過 Bookmarklet 新增 YouTube 影片到未分類頻道（支援 JSONP），同時支援 youtube_id 和 url 參數
// @Tags         節目
// @Produce      json
// @Security     ApiAuth
// @Param        callback query string false "JSONP callback 函數名"
// @Param        name query string true "影片名稱"
// @Param        youtube_id query string false "YouTube 影片 ID（文件規範）"
// @Param        url query string false "YouTube URL（向後相容，與 youtube_id 二選一）"
// @Param        desc query string false "描述"
// @Param        duration query int false "時長（秒）"
// @Param        tags query string false "標籤（逗號分隔的數字）"
// @Success      200 "成功回應（JSONP 格式）"
// @Failure      200 "需要登入或參數無效"
// @Router       /apis/pickprog [get]
func PickProgram(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PickProgramRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			callback := c.DefaultQuery("callback", "callback")
			response.JSONPError(c, callback, response.ErrorRequiredField)
			return
		}

		// 驗證 JSONP callback
		if req.Callback != "" {
			if !utils.ValidateJSONPCallback(req.Callback) {
				response.JSONPError(c, req.Callback, response.ErrorRequiredField)
				return
			}
		}

		// 驗證必填欄位
		if req.Name == "" {
			response.JSONPError(c, req.Callback, response.ErrorRequiredField)
			return
		}

		userID := session.GetUserID(c)
		username := session.GetUsername(c)
		unclassifiedChannelID := session.GetUnclassifiedChannel(c)

		if userID == "" || username == "" {
			response.JSONPError(c, req.Callback, response.ErrorRequireLogin)
			return
		}

		// 如果沒有未分類頻道，建立一個
		if unclassifiedChannelID == "" {
			channelRepo := repository.NewChannelRepository(db)
			userRepo := repository.NewUserRepository(db)
			channelService := service.NewChannelService(channelRepo, userRepo)

			unclassifiedChannel, err := channelService.CreateUnclassifiedChannel(c.Request.Context(), userID, username)
			if err != nil {
				response.JSONPError(c, req.Callback, response.ErrorServerError)
				return
			}
			unclassifiedChannelID = unclassifiedChannel.ID

			// 更新 session
			if err := session.SetUnclassifiedChannel(c, unclassifiedChannelID); err != nil {
				response.JSONPError(c, req.Callback, response.ErrorServerError)
				return
			}
		}

		// 取得 YouTube ID（優先使用 youtube_id，否則從 url 提取）
		var youtubeID string
		if req.YouTubeID != "" {
			youtubeID = req.YouTubeID
		} else if req.URL != "" {
			youtubeID = extractYouTubeID(req.URL)
		}

		if youtubeID == "" {
			response.JSONPError(c, req.Callback, response.ErrorRequiredField)
			return
		}

		// 處理 duration
		duration := 0
		if req.Duration != "" {
			if d, err := strconv.Atoi(req.Duration); err == nil {
				duration = d
			}
		}

		// 處理 tags（從逗號分隔的字串轉換為整數陣列）
		var tags []int
		if req.Tags != "" {
			tagStrs := strings.Split(req.Tags, ",")
			for _, tagStr := range tagStrs {
				tagStr = strings.TrimSpace(tagStr)
				if tag, err := strconv.Atoi(tagStr); err == nil {
					tags = append(tags, tag)
				}
			}
		}

		// 新增節目到未分類頻道
		programRepo := repository.NewProgramRepository(db)
		channelRepo := repository.NewChannelRepository(db)
		programService := service.NewProgramService(programRepo, channelRepo)

		program, err := programService.AddProgram(
			c.Request.Context(),
			unclassifiedChannelID,
			req.Name,
			youtubeID,
			req.Desc,
			duration,
			tags,
			false, // updateCover 設為 false（pickprog 不需要更新封面）
		)
		if err != nil {
			response.JSONPError(c, req.Callback, response.ErrorServerError)
			return
		}

		// JSONP 回應
		respData := gin.H{
			"state": response.StateSuccess,
			"program": program,
		}
		response.JSONPSuccess(c, req.Callback, respData)
	}
}

// extractYouTubeID 從 URL 提取 YouTube ID
func extractYouTubeID(url string) string {
	if url == "" {
		return ""
	}

	// 處理 youtu.be 短網址
	if len(url) > 17 && url[:17] == "https://youtu.be/" {
		id := url[17:]
		// 移除可能的查詢參數
		if idx := len(id); idx > 0 {
			for i, c := range id {
				if c == '?' || c == '&' {
					return id[:i]
				}
			}
			return id
		}
	}

	// 處理標準 YouTube URL (watch?v=)
	if len(url) > 32 && url[:32] == "https://www.youtube.com/watch?v=" {
		id := url[32:]
		// 移除可能的查詢參數
		for i, c := range id {
			if c == '&' {
				return id[:i]
			}
		}
		return id
	}

	// 處理 embed URL
	if len(url) > 31 && url[:31] == "https://www.youtube.com/embed/" {
		id := url[31:]
		// 移除可能的查詢參數
		for i, c := range id {
			if c == '?' || c == '&' {
				return id[:i]
			}
		}
		return id
	}

	return ""
}

