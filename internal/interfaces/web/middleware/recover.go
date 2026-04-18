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
				slog.With("request_id", requestID).
					Error(fmt.Sprintf(
						"request: %s path: %s panic detected: %v stack: %s",
						c.Request.Method, c.Request.URL.Path, err, string(debug.Stack())),
					)

				body, _ := ginzo.GetRequestBody(c)
				slog.With("request_id", requestID).
					Error(fmt.Sprintf(fmt.Sprintf("request body: %s", string(body))))
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
