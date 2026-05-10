package router

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
)

// RouterConfig 路由相关配置
type RouterConfig struct {
	BuildFS         embed.FS
	Hc              *handlers.HandlerContainers
	Middlewares     *middleware.Middlewares
	Theme           string
	FrontendBaseUrl string
}

// SetRouter 设置路由
func SetRouter(router *gin.Engine, conf *RouterConfig) {
	setApiRouter(router, conf.Hc, conf.Middlewares)
	setDashboardRouter(router, conf.Hc, conf.Middlewares)
	setRelayRouter(router, conf.Hc, conf.Middlewares)
	if conf.FrontendBaseUrl == "" {
		setWebRouter(router, conf.BuildFS, conf.Hc, conf.Middlewares, conf.Theme)
	} else {
		frontendBaseUrl := strings.TrimSuffix(conf.FrontendBaseUrl, "/")
		router.NoRoute(func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s%s", frontendBaseUrl, c.Request.RequestURI))
		})
	}
}
