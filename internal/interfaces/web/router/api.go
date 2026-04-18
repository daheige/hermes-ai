package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
)

// SetApiRouter 设置API路由
func SetApiRouter(router *gin.Engine, h *handlers.HandlerContainers, mw *middleware.Middlewares) {
	apiRouter := router.Group("/api")
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())

	apiRouter.GET("/status", h.MiscHandler.GetStatus)
	apiRouter.GET("/models", mw.AuthMiddleware.UserAuth(), h.ModelHandler.DashboardListModels)
	apiRouter.GET("/notice", h.MiscHandler.GetNotice)
	apiRouter.GET("/about", h.MiscHandler.GetAbout)
	apiRouter.GET("/home_page_content", h.MiscHandler.GetHomePageContent)
	apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), h.MiscHandler.SendEmailVerification)
	apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), h.MiscHandler.SendPasswordResetEmail)
	apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), h.MiscHandler.ResetPassword)
	apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), h.GitHubHandler.GitHubOAuth)
	apiRouter.GET("/oauth/oidc", middleware.CriticalRateLimit(), h.OidcUserHandler.OidcAuth)
	apiRouter.GET("/oauth/lark", middleware.CriticalRateLimit(), h.LarkUserHandler.LarkOAuth)
	apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), h.GitHubHandler.GenerateOAuthCode)
	apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), h.WeChatUserHandler.WeChatAuth)
	apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), mw.AuthMiddleware.UserAuth(), h.WeChatUserHandler.WeChatBind)
	apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), mw.AuthMiddleware.UserAuth(), h.UserHandler.EmailBind)
	apiRouter.POST("/topup", mw.AuthMiddleware.AdminAuth(), h.UserHandler.AdminTopUp)

	userRoute := apiRouter.Group("/user")
	{
		userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), h.AuthHandler.Register)
		userRoute.POST("/login", middleware.CriticalRateLimit(), h.AuthHandler.Login)
		userRoute.GET("/logout", h.AuthHandler.Logout)

		selfRoute := userRoute.Group("/")
		selfRoute.Use(mw.AuthMiddleware.UserAuth())
		{
			selfRoute.GET("/dashboard", h.UserHandler.GetUserDashboard)
			selfRoute.GET("/self", h.UserHandler.GetSelf)
			selfRoute.PUT("/self", h.UserHandler.UpdateSelf)
			selfRoute.DELETE("/self", h.UserHandler.DeleteSelf)
			selfRoute.GET("/token", h.UserHandler.GenerateAccessToken)
			selfRoute.GET("/aff", h.UserHandler.GetAffCode)
			selfRoute.POST("/topup", h.UserHandler.TopUp)
			selfRoute.GET("/available_models", h.ModelHandler.GetUserAvailableModels)
		}

		adminRoute := userRoute.Group("/")
		adminRoute.Use(mw.AuthMiddleware.AdminAuth())
		{
			adminRoute.GET("/", h.UserHandler.GetAllUsers)
			adminRoute.GET("/search", h.UserHandler.SearchUsers)
			adminRoute.GET("/:id", h.UserHandler.GetUser)
			adminRoute.POST("/", h.UserHandler.CreateUser)
			adminRoute.POST("/manage", h.UserHandler.ManageUser)
			adminRoute.PUT("/", h.UserHandler.UpdateUser)
			adminRoute.DELETE("/:id", h.UserHandler.DeleteUser)
		}
	}
	optionRoute := apiRouter.Group("/option", mw.AuthMiddleware.RootAuth())
	optionRoute.GET("/", h.OptionHandler.GetOptions)
	optionRoute.PUT("/", h.OptionHandler.UpdateOption)

	channelRoute := apiRouter.Group("/channel", mw.AuthMiddleware.AdminAuth())
	channelRoute.GET("/", h.ChannelHandler.GetAllChannels)
	channelRoute.GET("/search", h.ChannelHandler.SearchChannels)
	channelRoute.GET("/models", h.ModelHandler.ListAllModels)
	channelRoute.GET("/:id", h.ChannelHandler.GetChannel)
	channelRoute.GET("/test", h.ChannelTestHandler.TestChannels)
	channelRoute.GET("/test/:id", h.ChannelTestHandler.TestChannel)
	channelRoute.GET("/update_balance", h.ChannelBillingHandler.UpdateAllChannelsBalance)
	channelRoute.GET("/update_balance/:id", h.ChannelBillingHandler.UpdateChannelBalance)
	channelRoute.POST("/", h.ChannelHandler.AddChannel)
	channelRoute.PUT("/", h.ChannelHandler.UpdateChannel)
	channelRoute.DELETE("/disabled", h.ChannelHandler.DeleteDisabledChannel)
	channelRoute.DELETE("/:id", h.ChannelHandler.DeleteChannel)

	tokenRoute := apiRouter.Group("/token", mw.AuthMiddleware.UserAuth())
	tokenRoute.GET("/", h.TokenHandler.GetAllTokens)
	tokenRoute.GET("/search", h.TokenHandler.SearchTokens)
	tokenRoute.GET("/:id", h.TokenHandler.GetToken)
	tokenRoute.POST("/", h.TokenHandler.AddToken)
	tokenRoute.PUT("/", h.TokenHandler.UpdateToken)
	tokenRoute.DELETE("/:id", h.TokenHandler.DeleteToken)

	redemptionRoute := apiRouter.Group("/redemption", mw.AuthMiddleware.AdminAuth())
	redemptionRoute.GET("/", h.RedemptionHandler.GetAllRedemptions)
	redemptionRoute.GET("/search", h.RedemptionHandler.SearchRedemptions)
	redemptionRoute.GET("/:id", h.RedemptionHandler.GetRedemption)
	redemptionRoute.POST("/", h.RedemptionHandler.AddRedemption)
	redemptionRoute.PUT("/", h.RedemptionHandler.UpdateRedemption)
	redemptionRoute.DELETE("/:id", h.RedemptionHandler.DeleteRedemption)

	logRoute := apiRouter.Group("/log")
	logRoute.GET("/", mw.AuthMiddleware.AdminAuth(), h.LogHandler.GetAllLogs)
	logRoute.DELETE("/", mw.AuthMiddleware.AdminAuth(), h.LogHandler.DeleteHistoryLogs)
	logRoute.GET("/stat", mw.AuthMiddleware.AdminAuth(), h.LogHandler.GetLogsStat)
	logRoute.GET("/self/stat", mw.AuthMiddleware.UserAuth(), h.LogHandler.GetLogsSelfStat)
	logRoute.GET("/search", mw.AuthMiddleware.AdminAuth(), h.LogHandler.SearchAllLogs)
	logRoute.GET("/self", mw.AuthMiddleware.UserAuth(), h.LogHandler.GetUserLogs)
	logRoute.GET("/self/search", mw.AuthMiddleware.UserAuth(), h.LogHandler.SearchUserLogs)

	groupRoute := apiRouter.Group("/group", mw.AuthMiddleware.AdminAuth())
	groupRoute.GET("/", h.GroupHandler.GetGroups)
}
