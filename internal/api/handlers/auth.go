package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/higgstv/higgstv-go/internal/api/response"
	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/repository"
	"github.com/higgstv/higgstv-go/internal/service"
	"github.com/higgstv/higgstv-go/pkg/logger"
	"github.com/higgstv/higgstv-go/pkg/mail"
	"github.com/higgstv/higgstv-go/pkg/session"
	"go.uber.org/zap"
)

// SignInRequest 登入請求
type SignInRequest struct {
	Username string `json:"username" binding:"required" example:"testuser"` // 使用者名稱
	Password string `json:"password" binding:"required" example:"password123"` // 密碼
}

// SignIn 登入
// @Summary      登入
// @Description  使用者登入，成功後會設定 Session Cookie
// @Tags         認證
// @Accept       json
// @Produce      json
// @Param        request body SignInRequest true "登入請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0,"ret":true})
// @Failure      200 {object} map[string]interface{} "失敗回應" example({"state":0,"ret":false})
// @Router       /apis/signin [post]
func SignIn(db *mongo.Database, config interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SignInRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userRepo := repository.NewUserRepository(db)
		authService := service.NewAuthService(userRepo)

		user, err := authService.SignIn(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			response.SuccessWithRet(c, false)
			return
		}

		// 設定 Session
		unclassifiedChannel := ""
		if user.UnclassifiedChannel != nil {
			unclassifiedChannel = *user.UnclassifiedChannel
		}
		if err := session.SetLoggedIn(c, user.ID, user.Username, user.Email, unclassifiedChannel); err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.SuccessWithRet(c, true)
	}
}

// SignOut 登出
// @Summary      登出
// @Description  使用者登出，清除 Session。可選的 redirect 參數會執行 HTTP redirect
// @Tags         認證
// @Param        redirect query string false "重定向 URL（選填）"
// @Success      302 "重定向到指定 URL 或首頁"
// @Success      200 {object} map[string]interface{} "成功回應（無 redirect 時）" example({"state":0})
// @Router       /apis/signout [get]
func SignOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		session.Clear(c)

		// 檢查是否有 redirect 參數
		if redirect := c.Query("redirect"); redirect != "" {
			c.Redirect(http.StatusFound, redirect)
			return
		}

		// 若無 redirect 參數，回 { "state": 0 }
		response.Success(c, nil)
	}
}

// SignUpRequest 註冊請求
type SignUpRequest struct {
	InvitationCode string `json:"invitation_code" binding:"required" example:"sixpens"` // 邀請碼
	Username       string `json:"username" binding:"required" example:"testuser"` // 使用者名稱
	Email          string `json:"email" binding:"required,email" example:"test@example.com"` // 電子郵件
	Password       string `json:"password" binding:"required" example:"password123"` // 密碼
}

// SignUp 註冊
// @Summary      註冊
// @Description  新使用者註冊，需要有效的邀請碼
// @Tags         認證
// @Accept       json
// @Produce      json
// @Param        request body SignUpRequest true "註冊請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0,"ret":true})
// @Failure      200 {object} map[string]interface{} "失敗回應" example({"state":1,"code":2})
// @Router       /apis/signup [post]
func SignUp(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SignUpRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userRepo := repository.NewUserRepository(db)
		channelRepo := repository.NewChannelRepository(db)
		authService := service.NewAuthService(userRepo)
		channelService := service.NewChannelService(channelRepo, userRepo)

		// 註冊使用者
		user, err := authService.SignUp(c.Request.Context(), req.InvitationCode, req.Username, req.Email, req.Password)
		if err != nil {
			if err.Error() == "invalid invitation code" {
				response.Error(c, response.ErrorAccessDenied)
				return
			}
			if err.Error() == "user already exists" {
				response.SuccessWithRet(c, false)
				return
			}
			response.Error(c, response.ErrorServerError)
			return
		}

		// 建立未分類頻道
		unclassifiedChannel, err := channelService.CreateUnclassifiedChannel(c.Request.Context(), user.ID, user.Username)
		if err != nil {
			// 記錄錯誤以便除錯
			if logger.Logger != nil {
				logger.Logger.Error("Failed to create unclassified channel",
					zap.String("user_id", user.ID),
					zap.String("username", user.Username),
					zap.Error(err),
				)
			}
			response.Error(c, response.ErrorServerError)
			return
		}

		// 建立預設頻道
		_, err = channelService.CreateDefaultChannel(c.Request.Context(), user.ID, user.Username)
		if err != nil {
			// 記錄錯誤以便除錯
			if logger.Logger != nil {
				logger.Logger.Error("Failed to create default channel",
					zap.String("user_id", user.ID),
					zap.String("username", user.Username),
					zap.Error(err),
				)
			}
			response.Error(c, response.ErrorServerError)
			return
		}

		// 設定 Session
		if err := session.SetLoggedIn(c, user.ID, user.Username, user.Email, unclassifiedChannel.ID); err != nil {
			response.Error(c, response.ErrorServerError)
			return
		}

		response.SuccessWithRet(c, true)
	}
}

// ChangePasswordRequest 變更密碼請求
type ChangePasswordRequest struct {
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword 變更密碼
// @Summary      變更密碼
// @Description  修改密碼（需登入）
// @Tags         認證
// @Accept       json
// @Produce      json
// @Security     ApiAuth
// @Param        request body ChangePasswordRequest true "變更密碼請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0,"ret":true})
// @Failure      200 {object} map[string]interface{} "舊密碼錯誤" example({"state":0,"ret":false})
// @Failure      200 {object} map[string]interface{} "缺少欄位" example({"state":1,"code":0})
// @Router       /apis/change_password [post]
func ChangePassword(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		username := session.GetUsername(c)
		if username == "" {
			response.Error(c, response.ErrorRequireLogin)
			return
		}

		userRepo := repository.NewUserRepository(db)
		authService := service.NewAuthService(userRepo)

		err := authService.ChangePassword(c.Request.Context(), username, req.Password, req.NewPassword)
		if err != nil {
			if err.Error() == "invalid old password" || err.Error() == "user not found" {
				response.SuccessWithRet(c, false)
				return
			}
			response.Error(c, response.ErrorServerError)
			return
		}

		response.SuccessWithRet(c, true)
	}
}

// ForgetPasswordRequest 忘記密碼請求
type ForgetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgetPassword 忘記密碼
// @Summary      忘記密碼
// @Description  請求寄送重設密碼郵件。無論 Email 是否存在，都回成功（安全設計）
// @Tags         認證
// @Accept       json
// @Produce      json
// @Param        request body ForgetPasswordRequest true "忘記密碼請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0})
// @Failure      200 {object} map[string]interface{} "缺少欄位" example({"state":1,"code":0})
// @Router       /apis/forget_password [post]
func ForgetPassword(db *mongo.Database, mailConfig interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ForgetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userRepo := repository.NewUserRepository(db)
		authService := service.NewAuthService(userRepo)

		// 檢查 Email 是否存在
		user, err := userRepo.FindByEmail(c.Request.Context(), req.Email)
		if err != nil {
			// 無論是否存在，都回成功（安全設計）
			response.Success(c, nil)
			return
		}

		// 如果使用者存在，產生 access_key 並寄送郵件
		if user != nil {
			accessKey, err := authService.GenerateAccessKey(c.Request.Context(), req.Email)
			if err != nil {
				response.Success(c, nil) // 仍然回成功
				return
			}

			// 發送郵件（非同步，不阻塞回應）
			go func() {
				if cfg, ok := mailConfig.(*config.Config); ok && cfg != nil {
					mailService := mail.NewMailService(mail.MailConfig{
						SMTPHost:     cfg.Mail.SMTPHost,
						SMTPPort:     cfg.Mail.SMTPPort,
						SMTPUser:     cfg.Mail.SMTPUser,
						SMTPPassword: cfg.Mail.SMTPPassword,
						From:         cfg.Mail.From,
					})
					mailService.SendPasswordReset(req.Email, accessKey, cfg.Mail.BaseURL)
				}
			}()
		}

		response.Success(c, nil)
	}
}

// ResetPasswordRequest 重設密碼請求
type ResetPasswordRequest struct {
	Email     string `json:"email" binding:"required,email"`
	AccessKey string `json:"access_key" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// ResetPassword 重設密碼
// @Summary      重設密碼
// @Description  使用 access_key 重設密碼（不需要登入）
// @Tags         認證
// @Accept       json
// @Produce      json
// @Param        request body ResetPasswordRequest true "重設密碼請求"
// @Success      200 {object} map[string]interface{} "成功回應" example({"state":0,"ret":true})
// @Failure      200 {object} map[string]interface{} "access_key 無效或過期" example({"state":0,"ret":false})
// @Failure      200 {object} map[string]interface{} "缺少欄位" example({"state":1,"code":0})
// @Router       /apis/reset_password [post]
func ResetPassword(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, response.ErrorRequiredField)
			return
		}

		userRepo := repository.NewUserRepository(db)
		authService := service.NewAuthService(userRepo)

		err := authService.ResetPassword(c.Request.Context(), req.Email, req.AccessKey, req.Password)
		if err != nil {
			response.SuccessWithRet(c, false)
			return
		}

		response.SuccessWithRet(c, true)
	}
}

