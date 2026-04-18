// Package handlers is a package for handling the relay handlers
package controller

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/relay"
	"hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/meta"
	relaymodel "hermes-ai/internal/infras/relay/model"
)

// RelayProxyHelper is a helper function to proxy the request to the upstream service
func RelayProxyHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	resp, err := adaptor.DoRequest(c, meta, c.Request.Body)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("DoRequest failed: %s", err.Error()))
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	// do response
	_, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("respErr is not nil: %+v", respErr))
		return respErr
	}

	return nil
}
