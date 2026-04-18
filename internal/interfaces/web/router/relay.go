package router

import (
	"github.com/gin-gonic/gin"

	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
)

// SetRelayRouter 设置Relay路由
func SetRelayRouter(router *gin.Engine, hc *handlers.HandlerContainers, mw *middleware.Middlewares) {
	router.Use(middleware.CORS())
	router.Use(middleware.GzipDecodeMiddleware())

	// https://platform.openai.com/docs/api-reference/introduction
	modelsRouter := router.Group("/v1/models", mw.AuthMiddleware.TokenAuth())
	modelsRouter.GET("", hc.ModelHandler.ListModels)
	modelsRouter.GET("/:model", hc.ModelHandler.RetrieveModel)

	relayV1Router := router.Group("/v1")
	relayV1Router.Use(middleware.RelayPanicRecover(), mw.AuthMiddleware.TokenAuth(), mw.DistributorMiddleware.Distribute())

	relayV1Router.Any("/oneapi/proxy/:channelid/*target", hc.RelayHandler.Relay)
	relayV1Router.POST("/completions", hc.RelayHandler.Relay)
	relayV1Router.POST("/chat/completions", hc.RelayHandler.Relay)
	// claude code messages接口
	relayV1Router.POST("/messages", hc.RelayHandler.Relay)
	relayV1Router.POST("/edits", hc.RelayHandler.Relay)
	relayV1Router.POST("/images/generations", hc.RelayHandler.Relay)
	relayV1Router.POST("/images/edits", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/images/variations", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/embeddings", hc.RelayHandler.Relay)
	relayV1Router.POST("/engines/:model/embeddings", hc.RelayHandler.Relay)
	relayV1Router.POST("/audio/transcriptions", hc.RelayHandler.Relay)
	relayV1Router.POST("/audio/translations", hc.RelayHandler.Relay)
	relayV1Router.POST("/audio/speech", hc.RelayHandler.Relay)
	relayV1Router.GET("/files", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/files", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.DELETE("/files/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/files/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/files/:id/content", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/fine_tuning/jobs", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/fine_tuning/jobs", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/fine_tuning/jobs/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/fine_tuning/jobs/:id/cancel", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/fine_tuning/jobs/:id/events", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.DELETE("/models/:model", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/moderations", hc.RelayHandler.Relay)
	relayV1Router.POST("/assistants", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/assistants/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/assistants/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.DELETE("/assistants/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/assistants", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/assistants/:id/files", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/assistants/:id/files/:fileId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.DELETE("/assistants/:id/files/:fileId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/assistants/:id/files", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.DELETE("/threads/:id", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads/:id/messages", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id/messages/:messageId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads/:id/messages/:messageId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id/messages/:messageId/files/:filesId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id/messages/:messageId/files", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads/:id/runs", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id/runs/:runsId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads/:id/runs/:runsId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id/runs", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads/:id/runs/:runsId/submit_tool_outputs", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.POST("/threads/:id/runs/:runsId/cancel", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id/runs/:runsId/steps/:stepId", hc.RelayHandler.RelayNotImplemented)
	relayV1Router.GET("/threads/:id/runs/:runsId/steps", hc.RelayHandler.RelayNotImplemented)
}
