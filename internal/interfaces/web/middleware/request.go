package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/utils"
)

// AccessLog 访问日志中间件
func AccessLog() func(c *gin.Context) {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = utils.UUID()
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxkey.RequestIdKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		ip := c.ClientIP()
		slog.With(
			"request_id", requestID,
			"request_method", c.Request.Method,
			"request_uri", c.Request.RequestURI,
			"ip", ip,
		).Info("exec begin")

		c.Set(ctxkey.RequestIdKey, requestID)
		c.Header("X-Request-Id", requestID)
		c.Next()

		latency := fmt.Sprintf("%.4f", time.Since(start).Seconds())
		statusCode := c.Writer.Status()
		slog.With(
			"request_id", requestID,
			"request_method", c.Request.Method,
			"request_uri", c.Request.RequestURI,
			"status_code", statusCode,
			"latency", latency,
		).Info("exec end")
	}
}
