package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/relay"
	"hermes-ai/internal/infras/relay/adaptor/anthropic"
	"hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/billing"
	ratio2 "hermes-ai/internal/infras/relay/billing/ratio"
	"hermes-ai/internal/infras/relay/channeltype"
	"hermes-ai/internal/infras/relay/meta"
	model2 "hermes-ai/internal/infras/relay/model"
)

func RelayAnthropicMessagesHelper(c *gin.Context) *model2.ErrorWithStatusCode {
	ctx := c.Request.Context()
	metaInfo := meta.GetByContext(c)

	var anthropicRequest anthropic.Request
	err := ginzo.UnmarshalBodyReusable(c, &anthropicRequest)
	if err != nil {
		return openai.ErrorWrapper(err, "invalid_anthropic_request", http.StatusBadRequest)
	}

	if anthropicRequest.Model == "" {
		return openai.ErrorWrapper(errors.New("model is required"), "invalid_anthropic_request", http.StatusBadRequest)
	}

	metaInfo.OriginModelName = anthropicRequest.Model
	anthropicRequest.Model, _ = getMappedModelName(anthropicRequest.Model, metaInfo.ModelMapping)
	metaInfo.ActualModelName = anthropicRequest.Model
	metaInfo.IsStream = anthropicRequest.Stream

	modelRatio := ratio2.GetModelRatio(metaInfo.ActualModelName, metaInfo.ChannelType)
	groupRatio := ratio2.GetGroupRatio(metaInfo.Group)
	ratio := modelRatio * groupRatio

	promptTokens := countAnthropicPromptTokens(&anthropicRequest, metaInfo.ActualModelName)
	metaInfo.PromptTokens = promptTokens

	maxTokens := anthropicRequest.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	textRequest := &model2.GeneralOpenAIRequest{
		Model:     metaInfo.ActualModelName,
		MaxTokens: maxTokens,
		Stream:    anthropicRequest.Stream,
	}
	preConsumedQuota, bizErr := preConsumeQuota(ctx, textRequest, promptTokens, ratio, metaInfo)
	if bizErr != nil {
		return bizErr
	}

	adaptor := relay.GetAdaptor(metaInfo.APIType)
	if adaptor == nil {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, metaInfo.TokenId)
		return openai.ErrorWrapper(errors.New("invalid api type"), "invalid_api_type", http.StatusBadRequest)
	}

	// DeepSeek supports Anthropic Messages API via /anthropic base path
	if metaInfo.ChannelType == channeltype.DeepSeek {
		baseURL := strings.TrimSuffix(metaInfo.BaseURL, "/")
		if !strings.HasSuffix(baseURL, "/anthropic") {
			baseURL += "/anthropic"
		}
		metaInfo.BaseURL = baseURL
		adaptor = &anthropic.Adaptor{}
	}

	adaptor.Init(metaInfo)

	requestBody, _ := ginzo.GetRequestBody(c)
	resp, err := adaptor.DoRequest(c, metaInfo, bytes.NewReader(requestBody))
	if err != nil {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, metaInfo.TokenId)
		slog.With("request_id", logger.GetRequestID(ctx)).Error(fmt.Sprintf("DoRequest failed: %s", err.Error()))
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	if isErrorHappened(metaInfo, resp) {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, metaInfo.TokenId)
		return RelayErrorHandler(resp)
	}

	var usage *model2.Usage
	if !metaInfo.IsStream {
		usage = passthroughAnthropicResponse(c, resp)
	} else {
		usage = passthroughAnthropicStreamResponse(c, resp)
	}

	go postConsumeQuota(ctx, postConsumeQuotaParams{
		usage:             usage,
		meta:              metaInfo,
		textRequest:       textRequest,
		ratio:             ratio,
		preConsumedQuota:  preConsumedQuota,
		modelRatio:        modelRatio,
		groupRatio:        groupRatio,
		systemPromptReset: false,
	})

	return nil
}

func countAnthropicPromptTokens(request *anthropic.Request, modelName string) int {
	tokenNum := 0
	for _, message := range request.Messages {
		tokenNum += openai.CountTokenText(message.Role, modelName)
		contents, err := message.GetContentItems()
		if err != nil {
			continue
		}
		for _, content := range contents {
			if content.Type == "text" {
				tokenNum += openai.CountTokenText(content.Text, modelName)
			}
		}
	}
	if !request.System.IsEmpty() {
		tokenNum += openai.CountTokenText(request.System.String(), modelName)
	}
	return tokenNum
}

func passthroughAnthropicResponse(c *gin.Context, resp *http.Response) *model2.Usage {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(c.Request.Context())).
			Error(fmt.Sprintf("read response body failed: %s", err.Error()))
	}

	_ = resp.Body.Close()

	var anthropicResponse anthropic.Response
	_ = json.Unmarshal(responseBody, &anthropicResponse)

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(responseBody)

	if anthropicResponse.Usage.InputTokens > 0 || anthropicResponse.Usage.OutputTokens > 0 {
		return &model2.Usage{
			PromptTokens:     anthropicResponse.Usage.InputTokens,
			CompletionTokens: anthropicResponse.Usage.OutputTokens,
			TotalTokens:      anthropicResponse.Usage.InputTokens + anthropicResponse.Usage.OutputTokens,
		}
	}

	return nil
}

func passthroughAnthropicStreamResponse(c *gin.Context, resp *http.Response) *model2.Usage {
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)

	var usage model2.Usage
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// 原样写回客户端
		_, _ = c.Writer.WriteString(line + "\n")
		c.Writer.Flush()

		// 解析 usage
		if len(line) < 6 || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)
		if data == "" {
			continue
		}

		var streamResp anthropic.StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		switch streamResp.Type {
		case "message_start":
			if streamResp.Message != nil {
				usage.PromptTokens += streamResp.Message.Usage.InputTokens
				usage.CompletionTokens += streamResp.Message.Usage.OutputTokens
			}
		case "message_delta":
			if streamResp.Usage != nil {
				usage.CompletionTokens += streamResp.Usage.OutputTokens
			}
		}
	}

	if err := scanner.Err(); err != nil {
		slog.With("request_id", logger.GetRequestID(c.Request.Context())).
			Error(fmt.Sprintf("read stream response failed: %s", err.Error()))
	}

	_ = resp.Body.Close()

	if usage.PromptTokens > 0 || usage.CompletionTokens > 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
		return &usage
	}
	return nil
}
