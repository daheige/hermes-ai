package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/relay"
	"hermes-ai/internal/infras/relay/adaptor"
	"hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/apitype"
	"hermes-ai/internal/infras/relay/billing"
	ratio2 "hermes-ai/internal/infras/relay/billing/ratio"
	"hermes-ai/internal/infras/relay/channeltype"
	"hermes-ai/internal/infras/relay/meta"
	model2 "hermes-ai/internal/infras/relay/model"
)

func RelayTextHelper(c *gin.Context) *model2.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)
	// get & validate textRequest
	textRequest, err := getAndValidateTextRequest(c, meta.Mode)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("getAndValidateTextRequest failed: %s", err.Error()))
		return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
	}

	meta.IsStream = textRequest.Stream

	// map model name
	meta.OriginModelName = textRequest.Model
	textRequest.Model, _ = getMappedModelName(textRequest.Model, meta.ModelMapping)
	meta.ActualModelName = textRequest.Model
	// set system prompt if not empty
	systemPromptReset := setSystemPrompt(ctx, textRequest, meta.ForcedSystemPrompt)
	// get model ratio & group ratio
	modelRatio := ratio2.GetModelRatio(textRequest.Model, meta.ChannelType)
	groupRatio := ratio2.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio
	// pre-consume quota
	promptTokens := getPromptTokens(textRequest, meta.Mode)
	meta.PromptTokens = promptTokens
	preConsumedQuota, bizErr := preConsumeQuota(ctx, textRequest, promptTokens, ratio, meta)
	if bizErr != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).Error(fmt.Sprintf("preConsumeQuota failed: %+v", *bizErr))
		return bizErr
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	// get request body
	requestBody, err := getRequestBody(c, meta, textRequest, adaptor)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
	}

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("DoRequest failed: %s", err.Error()))
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if isErrorHappened(meta, resp) {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return RelayErrorHandler(resp)
	}

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("respErr is not nil: %+v", respErr))
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return respErr
	}

	// post-consume quota
	go postConsumeQuota(ctx, postConsumeQuotaParams{
		usage:             usage,
		meta:              meta,
		textRequest:       textRequest,
		ratio:             ratio,
		preConsumedQuota:  preConsumedQuota,
		modelRatio:        modelRatio,
		groupRatio:        groupRatio,
		systemPromptReset: systemPromptReset,
	})

	return nil
}

func getRequestBody(c *gin.Context, meta *meta.Meta, textRequest *model2.GeneralOpenAIRequest, adaptor adaptor.Adaptor) (io.Reader, error) {
	if !config.EnforceIncludeUsage &&
		meta.APIType == apitype.OpenAI &&
		meta.OriginModelName == meta.ActualModelName &&
		meta.ChannelType != channeltype.Baichuan &&
		meta.ForcedSystemPrompt == "" {
		// no need to convert request for openai
		return c.Request.Body, nil
	}

	// get request body
	var requestBody io.Reader
	convertedRequest, err := adaptor.ConvertRequest(c, meta.Mode, textRequest)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(c.Request.Context())).
			Debug(fmt.Sprintf("converted request failed: %s\n", err.Error()))

		return nil, err
	}
	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(c.Request.Context())).
			Debug(fmt.Sprintf("converted request json_marshal_failed: %s\n", err.Error()))

		return nil, err
	}
	slog.With("request_id", logger.GetRequestID(c.Request.Context())).
		Debug(fmt.Sprintf("converted request: \n%s", string(jsonData)))
	requestBody = bytes.NewBuffer(jsonData)
	return requestBody, nil
}
