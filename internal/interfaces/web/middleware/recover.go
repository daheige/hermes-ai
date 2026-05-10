package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/logger"
)

func RelayPanicRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := c.Request.Context()
				requestID := logger.GetRequestID(ctx)
				body, _ := ginzo.GetRequestBody(c)
				slog.With("request_id", requestID).
					Error("reply exec panic",
						slog.String("error", fmt.Sprintf("%v", err)),
						slog.String("stack", string(debug.Stack())),
						slog.String("request_method", c.Request.Method),
						slog.String("request_uri", c.Request.RequestURI),
						slog.String("request_body", string(body)),
					)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf("Panic detected, error: %v", err),
						"type":    "one_api_panic",
					},
				})
			}
		}()

		c.Next()
	}
}
