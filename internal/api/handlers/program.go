package handlers

import (

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/internal/repository"
	"github.com/higgstv/higgstv-go/internal/service"
	"github.com/higgstv/higgstv-go/pkg/session"
)

// AddProgramRequest 新增節目請求
type AddProgramRequest struct {
	Ch         string `json:"ch" binding:"required" example:"channel_id"` // 頻道 ID
	Name       string `json:"name" binding:"required" example:"節目名稱"` // 節目名稱
	YouTubeID  string `json:"youtube_id" binding:"required" example:"dQw4w9WgXcQ"` // YouTube 影片 ID
	Desc       string `json:"desc" example:"節目描述"` // 節目描述
	Duration   int    `json:"duration" example:"300"` // 時長（秒）
	Tags       []int  `json:"tags"` // 標籤列表
	UpdateCover bool  `json:"updateCover" example:"false"` // 是否更新頻道封面
}

// AddProgram 新增節目
// @Summary      新增節目
// @Description  在頻道中新增節目（需要登入且為頻道管理員）
// @Tags         節目
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body AddProgramRequest true "新增節目請求"
// @Success      200 {object} map[string]interface{} "成功回應"
// @Failure      200 {object} map[string]interface{} "需要登入或權限不足" example({"state":1,"code":2})
// @Router       /apis/addprog [post]
func AddProgram(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		if req.Tags == nil {
			req.Tags = []int{}
		}

		userID := session.GetUserID(c)
		if userID == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		programRepo := repository.NewProgramRepository(db)
		channelRepo := repository.NewChannelRepository(db)
		programService := service.NewProgramService(programRepo, channelRepo)

		// 檢查權限
		channelService := service.NewChannelService(channelRepo, nil)
		isAdmin, err := channelService.IsAdmin(c.Request.Context(), req.Ch, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdmin {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		program, err := programService.AddProgram(
			c.Request.Context(),
			req.Ch,
			req.Name,
			req.YouTubeID,
			req.Desc,
			req.Duration,
			req.Tags,
			req.UpdateCover,
		)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{"program": program})
	}
}

// SaveProgramRequest 儲存節目請求
type SaveProgramRequest struct {
	Ch         string `json:"ch" binding:"required"`
	ProgID     int    `json:"prog_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	YouTubeID  string `json:"youtube_id" binding:"required"`
	Desc       string `json:"desc"`
	Duration   *int   `json:"duration"`
	Tags       []int  `json:"tags"`
	UpdateCover bool  `json:"updateCover"`
}

// SaveProgram 儲存節目
// @Summary      儲存節目
// @Description  更新節目內容（需登入，需有權限）
// @Tags         節目
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body SaveProgramRequest true "儲存節目請求"
// @Success      200 {object} map[string]interface{} "成功回應"
// @Failure      200 {object} map[string]interface{} "缺少必填欄位" example({"state":1,"code":0})
// @Failure      200 {object} map[string]interface{} "權限不足" example({"state":1,"code":2})
// @Router       /apis/saveprog [post]
func SaveProgram(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userID := session.GetUserID(c)
		if userID == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		programRepo := repository.NewProgramRepository(db)
		channelRepo := repository.NewChannelRepository(db)
		programService := service.NewProgramService(programRepo, channelRepo)

		// 檢查權限
		channelService := service.NewChannelService(channelRepo, nil)
		isAdmin, err := channelService.IsAdmin(c.Request.Context(), req.Ch, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdmin {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		program, err := programService.UpdateProgram(
			c.Request.Context(),
			req.Ch,
			req.ProgID,
			req.Name,
			req.YouTubeID,
			req.Desc,
			req.Duration,
			req.Tags,
			req.UpdateCover,
		)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{"program": program})
	}
}

// DeleteProgramRequest 刪除節目請求
type DeleteProgramRequest struct {
	Ch string `json:"ch" binding:"required"`
	IDs []int  `json:"ids" binding:"required"`
}

// DeleteProgram 刪除節目
// @Summary      刪除節目
// @Description  刪除節目（可一次刪除多個，需登入，需有權限）
// @Tags         節目
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body DeleteProgramRequest true "刪除節目請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0})
// @Failure      200 {object} map[string]interface{} "缺少必填欄位" example({"state":1,"code":0})
// @Failure      200 {object} map[string]interface{} "權限不足" example({"state":1,"code":2})
// @Router       /apis/delprog [post]
func DeleteProgram(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userID := session.GetUserID(c)
		if userID == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		programRepo := repository.NewProgramRepository(db)
		channelRepo := repository.NewChannelRepository(db)
		programService := service.NewProgramService(programRepo, channelRepo)

		// 檢查權限
		channelService := service.NewChannelService(channelRepo, nil)
		isAdmin, err := channelService.IsAdmin(c.Request.Context(), req.Ch, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdmin {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		err = programService.DeletePrograms(c.Request.Context(), req.Ch, req.IDs)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, nil)
	}
}

// MoveProgramRequest 移動節目請求
type MoveProgramRequest struct {
	Ch     string `json:"ch" binding:"required"`
	Target string `json:"target" binding:"required"`
	IDs    []int  `json:"ids" binding:"required"`
}

// MoveProgram 移動節目
// @Summary      移動節目
// @Description  將節目從一個頻道搬到另一個頻道（需登入，需有來源和目標頻道權限）
// @Tags         節目
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body MoveProgramRequest true "移動節目請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0})
// @Failure      200 {object} map[string]interface{} "缺少必填欄位" example({"state":1,"code":0})
// @Failure      200 {object} map[string]interface{} "權限不足" example({"state":1,"code":2})
// @Router       /apis/progmoveto [post]
func MoveProgram(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MoveProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userID := session.GetUserID(c)
		if userID == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		programRepo := repository.NewProgramRepository(db)
		channelRepo := repository.NewChannelRepository(db)
		programService := service.NewProgramService(programRepo, channelRepo)

		// 檢查來源頻道權限
		channelService := service.NewChannelService(channelRepo, nil)
		isAdmin, err := channelService.IsAdmin(c.Request.Context(), req.Ch, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdmin {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		// 檢查目標頻道權限
		isAdminTarget, err := channelService.IsAdmin(c.Request.Context(), req.Target, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdminTarget {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		err = programService.MoveProgram(c.Request.Context(), req.Ch, req.Target, req.IDs)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, nil)
	}
}

// SaveProgramOrderRequest 儲存節目順序請求
type SaveProgramOrderRequest struct {
	Ch    string `json:"ch" binding:"required"`
	Order []int  `json:"order" binding:"required"`
}

// SaveProgramOrder 儲存節目順序
// @Summary      儲存節目順序
// @Description  儲存節目在頻道中的自訂排序（需登入，需有權限）
// @Tags         節目
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body SaveProgramOrderRequest true "儲存節目順序請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0})
// @Failure      200 {object} map[string]interface{} "缺少必填欄位" example({"state":1,"code":0})
// @Failure      200 {object} map[string]interface{} "權限不足" example({"state":1,"code":2})
// @Router       /apis/prog/saveorder [post]
func SaveProgramOrder(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveProgramOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userID := session.GetUserID(c)
		if userID == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		programRepo := repository.NewProgramRepository(db)
		channelRepo := repository.NewChannelRepository(db)
		programService := service.NewProgramService(programRepo, channelRepo)

		// 檢查權限
		channelService := service.NewChannelService(channelRepo, nil)
		isAdmin, err := channelService.IsAdmin(c.Request.Context(), req.Ch, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdmin {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		err = programService.SetOrder(c.Request.Context(), req.Ch, req.Order)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, nil)
	}
}

