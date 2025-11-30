package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/models"
	"github.com/higgstv/higgstv-go/internal/repository"
	"github.com/higgstv/higgstv-go/internal/service"
	"github.com/higgstv/higgstv-go/pkg/logger"
	"github.com/higgstv/higgstv-go/pkg/session"
)

// AddChannelRequest 新增頻道請求
type AddChannelRequest struct {
	Name string `json:"name" binding:"required" example:"我的頻道"` // 頻道名稱
	Tags []int  `json:"tags"` // 標籤列表
}

// AddChannel 新增頻道
// @Summary      新增頻道
// @Description  建立新頻道（需要登入）
// @Tags         頻道
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body AddChannelRequest true "新增頻道請求"
// @Success      200 {object} map[string]interface{} "成功回應"
// @Failure      200 {object} map[string]interface{} "需要登入" example({"state":1,"code":1})
// @Router       /apis/addchannel [post]
func AddChannel(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddChannelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		if req.Name == "" {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		if req.Tags == nil {
			req.Tags = []int{}
		}

		userID := session.GetUserID(c)
		username := session.GetUsername(c)
		if userID == "" || username == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		channelRepo := repository.NewChannelRepository(db)
		userRepo := repository.NewUserRepository(db)
		channelService := service.NewChannelService(channelRepo, userRepo)

		channel, err := channelService.AddChannel(c.Request.Context(), userID, username, req.Name, req.Tags)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{"channel": channel})
	}
}

// GetOwnChannels 取得自己的頻道列表
// @Summary      取得自己的頻道列表
// @Description  取得當前登入使用者擁有的所有頻道（需要登入），支援 q 和 types[] 過濾
// @Tags         頻道
// @Produce      json
// @Security     ApiAuth
// @Param        q query string false "關鍵字搜尋（頻道名稱）"
// @Param        types query []string false "頻道類型陣列（例如：default,unclassified）"
// @Success      200 {object} map[string]interface{} "成功回應"
// @Failure      200 {object} map[string]interface{} "需要登入" example({"state":1,"code":1})
// @Router       /apis/getownchannels [get]
func GetOwnChannels(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := session.GetUserID(c)
		if userID == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		// 建立過濾條件
		filter := database.Filter{"owners": userID}

		// 處理關鍵字搜尋（q 參數）
		if q := c.Query("q"); q != "" {
			filter["name"] = database.Filter{"$regex": q, "$options": "i"}
		}

		// 處理頻道類型過濾（types[] 參數）
		types := c.QueryArray("types")
		if len(types) > 0 {
			filter["type"] = database.Filter{"$in": types}
		}

		channelRepo := repository.NewChannelRepository(db)
		channels, err := channelRepo.ListChannels(
			c.Request.Context(),
			filter,
			database.Sort{{Field: "last_modified", Order: -1}},
			0,
			0,
		)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{"channels": channels})
	}
}

// GetChannels 取得頻道列表（支援過濾）
// @Summary      取得頻道列表
// @Description  取得公開頻道列表，支援過濾和排序
// @Tags         頻道
// @Produce      json
// @Param        user query string false "只列出特定使用者的頻道（username）"
// @Param        q query string false "以名稱模糊搜尋"
// @Param        has_contents query string false "是否只顯示有節目的頻道（0/1）"
// @Param        ignore_types query []string false "要排除的頻道類型陣列"
// @Param        sort query string false "排序欄位（例如：last_modified, name）"
// @Param        desc query string false "是否遞減排序（0/1）"
// @Param        start query int false "分頁起始 index"
// @Param        limit query int false "限制筆數"
// @Success      200 {object} map[string]interface{} "成功回應"
// @Router       /apis/getchannels [get]
func GetChannels(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := database.Filter{}

		// 處理 user 參數（透過 username 查詢使用者 ID）
		if username := c.Query("user"); username != "" {
			userRepo := repository.NewUserRepository(db)
			user, err := userRepo.FindByUsername(c.Request.Context(), username)
			if err == nil && user != nil {
				filter["owners"] = user.ID
			} else {
				// 如果使用者不存在，回傳空結果
				response.Success(c, gin.H{"channels": []models.Channel{}})
				return
			}
		}

		// 處理關鍵字搜尋（q 參數，文件規範使用 q 而非 name）
		if q := c.Query("q"); q != "" {
			filter["name"] = database.Filter{"$regex": q, "$options": "i"}
		} else if name := c.Query("name"); name != "" {
			// 向後相容：也支援 name 參數
			filter["name"] = database.Filter{"$regex": name, "$options": "i"}
		}

		// 處理 has_contents 參數（是否只顯示有節目的頻道）
		if hasContentsStr := c.Query("has_contents"); hasContentsStr != "" {
			if hasContentsStr == "1" {
				filter["contents.0"] = database.Filter{"$exists": true}
			}
		}

		// 處理 ignore_types 參數（要排除的頻道類型）
		ignoreTypes := c.QueryArray("ignore_types")
		if len(ignoreTypes) > 0 {
			filter["type"] = database.Filter{"$nin": ignoreTypes}
		}

		// 排序
		sortField := c.DefaultQuery("sort", "last_modified")
		descStr := c.DefaultQuery("desc", "0")
		order := 1
		if descStr == "1" {
			order = -1
		}

		var sort database.Sort
		if sortField == "name" {
			sort = database.Sort{{Field: "name", Order: order}}
		} else {
			sort = database.Sort{{Field: "last_modified", Order: order}}
		}

		// 分頁（支援 start 和 skip，start 優先）
		limit := int64(0)
		skip := int64(0)
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
				limit = l
			}
		}
		// 優先使用 start 參數（文件規範）
		if startStr := c.Query("start"); startStr != "" {
			if s, err := strconv.ParseInt(startStr, 10, 64); err == nil {
				skip = s
			}
		} else if skipStr := c.Query("skip"); skipStr != "" {
			// 向後相容：也支援 skip 參數
			if s, err := strconv.ParseInt(skipStr, 10, 64); err == nil {
				skip = s
			}
		}

		channelRepo := repository.NewChannelRepository(db)
		channels, err := channelRepo.ListChannels(c.Request.Context(), filter, sort, limit, skip)
		if err != nil {
			// 記錄錯誤以便除錯
			if logger.Logger != nil {
				logger.Logger.Error("Failed to list channels",
					zap.Error(err),
					zap.Any("filter", filter),
				)
			}
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, gin.H{"channels": channels})
	}
}

// GetChannel 取得單一頻道
// @Summary      取得單一頻道
// @Description  根據頻道 ID 取得頻道詳細資訊
// @Tags         頻道
// @Produce      json
// @Param        id path string true "頻道 ID"
// @Success      200 {object} map[string]interface{} "成功回應"
// @Failure      200 {object} map[string]interface{} "頻道不存在" example({"state":1,"code":2})
// @Router       /apis/getchannel/{id} [get]
func GetChannel(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("id")
		if channelID == "" {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		channelRepo := repository.NewChannelRepository(db)
		channel, err := channelRepo.FindByID(c.Request.Context(), channelID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if channel == nil {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		response.Success(c, gin.H{"channel": channel})
	}
}

// GetChannelInfo 取得頻道資訊（含擁有者資訊）
// @Summary      取得頻道資訊（含擁有者）
// @Description  取得頻道詳細資訊，包含擁有者基本資訊
// @Tags         頻道
// @Produce      json
// @Param        id path string true "頻道 ID"
// @Success      200 {object} map[string]interface{} "成功回應"
// @Failure      200 {object} map[string]interface{} "頻道不存在" example({"state":1,"code":2})
// @Router       /apis/getchannelinfo/{id} [get]
func GetChannelInfo(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID := c.Param("id")
		if channelID == "" {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		channelRepo := repository.NewChannelRepository(db)
		userRepo := repository.NewUserRepository(db)

		channel, err := channelRepo.FindByID(c.Request.Context(), channelID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if channel == nil {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		// 取得擁有者資訊
		ownersInfo, err := userRepo.GetUsersBasicInfo(c.Request.Context(), channel.Owners)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		result := models.ChannelWithOwnersInfo{
			Channel:    *channel,
			OwnersInfo: ownersInfo,
		}

		response.Success(c, gin.H{"channel": result})
	}
}

// SaveChannelRequest 儲存頻道請求
type SaveChannelRequest struct {
	ID    string `json:"id" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Desc  string `json:"desc"`
	Tags  []int  `json:"tags"`
}

// SaveChannel 儲存頻道
// @Summary      儲存頻道
// @Description  更新頻道名稱、描述和標籤（需登入，需有權限）
// @Tags         頻道
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body SaveChannelRequest true "儲存頻道請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0})
// @Failure      200 {object} map[string]interface{} "缺少必填欄位" example({"state":1,"code":0})
// @Failure      200 {object} map[string]interface{} "權限不足" example({"state":1,"code":2})
// @Router       /apis/savechannel [post]
func SaveChannel(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveChannelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		if req.ID == "" || req.Name == "" {
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

		channelRepo := repository.NewChannelRepository(db)
		channelService := service.NewChannelService(channelRepo, nil)

		// 檢查權限
		isAdmin, err := channelService.IsAdmin(c.Request.Context(), req.ID, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdmin {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		err = channelService.UpdateChannel(c.Request.Context(), req.ID, req.Name, req.Desc, req.Tags)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, nil)
	}
}

// SetChannelOwnerRequest 設定頻道擁有者請求
type SetChannelOwnerRequest struct {
	ID     string `json:"id" binding:"required"`     // 頻道 ID
	C      string `json:"c"`                          // 頻道 ID（歷史遺留參數，與 id 相同）
	Email  string `json:"email"`                      // 要新增為共用者的使用者 Email（優先使用）
	Owners []string `json:"owners"`                   // 要新增的使用者 ID 陣列（向後相容）
}

// SetChannelOwner 設定頻道擁有者
// @Summary      設定頻道擁有者
// @Description  新增共用者，呼叫者需為 admin（優先使用 email 參數）
// @Tags         頻道
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body SetChannelOwnerRequest true "設定頻道擁有者請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0})
// @Failure      200 {object} map[string]interface{} "缺少必填欄位" example({"state":1,"code":0})
// @Failure      200 {object} map[string]interface{} "權限不足" example({"state":1,"code":2})
// @Router       /apis/setchannelowner [post]
func SetChannelOwner(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SetChannelOwnerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		if req.ID == "" {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userID := session.GetUserID(c)
		if userID == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		channelRepo := repository.NewChannelRepository(db)
		userRepo := repository.NewUserRepository(db)
		channelService := service.NewChannelService(channelRepo, userRepo)

		// 檢查權限
		isAdmin, err := channelService.IsAdmin(c.Request.Context(), req.ID, userID)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}
		if !isAdmin {
			response.Error(c, response.ErrorAccessDenied)
			return
		}

		// 處理要新增的使用者 ID 列表
		var userIDsToAdd []string

		// 優先處理 email 參數（文件規範）
		if req.Email != "" {
			user, err := userRepo.FindByEmail(c.Request.Context(), req.Email)
			if err != nil || user == nil {
				response.Error(c, response.ErrorAccessDenied)
				return
			}
			userIDsToAdd = append(userIDsToAdd, user.ID)
		} else if len(req.Owners) > 0 {
			// 向後相容：支援 owners 陣列
			userIDsToAdd = req.Owners
		} else {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		if len(userIDsToAdd) == 0 {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		err = channelService.AddOwners(c.Request.Context(), req.ID, userIDsToAdd)
		if err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.Success(c, nil)
	}
}


