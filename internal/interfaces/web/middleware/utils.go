package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/utils"
)

func abortWithMessage(c *gin.Context, statusCode int, message string) {
	slog.With("request_id", logger.GetRequestID(c.Request.Context())).Error(message)

	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": utils.MessageWithRequestId(message, c.GetString(ctxkey.RequestIdKey)),
			"type":    "one_api_error",
		},
	})
	c.Abort()
}

func getRequestModel(c *gin.Context) (string, error) {
	var modelRequest ModelRequest
	err := ginzo.UnmarshalBodyReusable(c, &modelRequest)
	if err != nil {
		return "", fmt.Errorf("UnmarshalBodyReusable failed: %w", err)
	}

	if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	}
	if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "dall-e-2"
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") || strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "whisper-1"
		}
	}
	return modelRequest.Model, nil
}

func isModelInList(modelName string, models string) bool {
	modelList := strings.Split(models, ",")
	for _, model := range modelList {
		if modelName == model {
			return true
		}
	}
	return false
}
