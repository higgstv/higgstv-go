package api

import (
	"github.com/gin-gonic/gin"

	"github.com/higgstv/higgstv-go/internal/api/handlers"
	"github.com/higgstv/higgstv-go/internal/api/middleware"
	"github.com/higgstv/higgstv-go/internal/database"
)

// SetupRoutes 設定路由
func SetupRoutes(router *gin.Engine, db database.Database, config interface{}) {
	// 404 處理
	router.NoRoute(middleware.NotFoundHandler())
	// 認證相關 API
	router.POST("/apis/signin", handlers.SignIn(db, config))
	router.GET("/apis/signout", handlers.SignOut())
	router.POST("/apis/signup", handlers.SignUp(db))
	router.POST("/apis/change_password", middleware.RequireAuth(), handlers.ChangePassword(db))
	router.POST("/apis/forget_password", handlers.ForgetPassword(db, config))
	router.POST("/apis/reset_password", handlers.ResetPassword(db))

	// 頻道相關 API
	router.POST("/apis/addchannel", middleware.RequireAuth(), handlers.AddChannel(db))
	router.GET("/apis/getownchannels", middleware.RequireAuth(), handlers.GetOwnChannels(db))
	router.GET("/apis/getchannels", handlers.GetChannels(db))
	router.GET("/apis/getchannel/:id", handlers.GetChannel(db))
	router.GET("/apis/getchannelinfo/:id", middleware.RequireAuth(), handlers.GetChannelInfo(db))
	router.POST("/apis/savechannel", middleware.RequireAuth(), handlers.SaveChannel(db))
	router.POST("/apis/setchannelowner", middleware.RequireAuth(), handlers.SetChannelOwner(db))

	// 節目相關 API
	router.POST("/apis/addprog", middleware.RequireAuth(), handlers.AddProgram(db))
	router.POST("/apis/saveprog", middleware.RequireAuth(), handlers.SaveProgram(db))
	router.POST("/apis/delprog", middleware.RequireAuth(), handlers.DeleteProgram(db))
	router.POST("/apis/progmoveto", middleware.RequireAuth(), handlers.MoveProgram(db))
	router.POST("/apis/prog/saveorder", middleware.RequireAuth(), handlers.SaveProgramOrder(db))

	// Pick API (Bookmarklet)
	router.GET("/apis/pickprog", middleware.RequireAuth(), handlers.PickProgram(db))
}

