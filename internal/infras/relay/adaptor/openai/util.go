package openai

import (
	"context"
	"fmt"
	"log/slog"

	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/relay/model"
)

func ErrorWrapper(err error, code string, statusCode int) *model.ErrorWithStatusCode {
	slog.With("request_id", logger.GetRequestID(context.Background())).Error(fmt.Sprintf("[%s]%+v", code, err))

	Error := model.Error{
		Message: err.Error(),
		Type:    "one_api_error",
		Code:    code,
	}
	return &model.ErrorWithStatusCode{
		Error:      Error,
		StatusCode: statusCode,
	}
}
