package router

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
)

func SetRouter(router *gin.Engine, buildFS embed.FS, hc *handlers.HandlerContainers, mw *middleware.Middlewares) {
	SetApiRouter(router, hc, mw)
	SetDashboardRouter(router, hc, mw)
	SetRelayRouter(router, hc, mw)
	frontendBaseUrl := os.Getenv("FRONTEND_BASE_URL")
	if config.IsMasterNode && frontendBaseUrl != "" {
		frontendBaseUrl = ""
		slog.Info("FRONTEND_BASE_URL is ignored on master node")
	}

	if frontendBaseUrl == "" {
		SetWebRouter(router, buildFS, hc, mw)
	} else {
		frontendBaseUrl = strings.TrimSuffix(frontendBaseUrl, "/")
		router.NoRoute(func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s%s", frontendBaseUrl, c.Request.RequestURI))
		})
	}
}
