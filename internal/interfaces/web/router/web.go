package router

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	embed2 "hermes-ai/internal/infras/ginzo/embed"
	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
)

func SetWebRouter(router *gin.Engine, buildFS embed.FS, hc *handlers.HandlerContainers,
	mw *middleware.Middlewares, theme string) {
	indexPageData, _ := buildFS.ReadFile(fmt.Sprintf("web/build/%s/index.html", theme))
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(mw.RateLimitMiddleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", embed2.EmbedFolder(buildFS, fmt.Sprintf("web/build/%s", theme))))
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") {
			hc.RelayHandler.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageData)
	})
}
