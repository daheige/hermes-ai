package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
)

func SetDashboardRouter(router *gin.Engine, h *handlers.HandlerContainers, mw *middleware.Middlewares) {
	apiRouter := router.Group("/")
	apiRouter.Use(middleware.CORS())
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	apiRouter.Use(mw.AuthMiddleware.TokenAuth())
	{
		apiRouter.GET("/dashboard/billing/subscription", h.BillingHandler.GetSubscription)
		apiRouter.GET("/v1/dashboard/billing/subscription", h.BillingHandler.GetSubscription)
		apiRouter.GET("/dashboard/billing/usage", h.BillingHandler.GetUsage)
		apiRouter.GET("/v1/dashboard/billing/usage", h.BillingHandler.GetUsage)
	}
}
