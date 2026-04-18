package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/logger"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	billingratio "hermes-ai/internal/infras/relay/billing/ratio"
	"hermes-ai/internal/infras/relay/channeltype"
	"hermes-ai/internal/infras/relay/constant/role"
	"hermes-ai/internal/infras/relay/controller/validator"
	"hermes-ai/internal/infras/relay/meta"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/relay/relaymode"
	"hermes-ai/internal/infras/relay/services"

	"github.com/gin-gonic/gin"
)

func getAndValidateTextRequest(c *gin.Context, relayMode int) (*model2.GeneralOpenAIRequest, error) {
	textRequest := &model2.GeneralOpenAIRequest{}
	err := ginzo.UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}
	if relayMode == relaymode.Moderations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == relaymode.Embeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}
	err = validator.ValidateTextRequest(textRequest, relayMode)
	if err != nil {
		return nil, err
	}
	return textRequest, nil
}

func getPromptTokens(textRequest *model2.GeneralOpenAIRequest, relayMode int) int {
	switch relayMode {
	case relaymode.ChatCompletions:
		return openai2.CountTokenMessages(textRequest.Messages, textRequest.Model)
	case relaymode.Completions:
		return openai2.CountTokenInput(textRequest.Prompt, textRequest.Model)
	case relaymode.Moderations:
		return openai2.CountTokenInput(textRequest.Input, textRequest.Model)
	default:
	}

	return 0
}

func getPreConsumedQuota(textRequest *model2.GeneralOpenAIRequest, promptTokens int, ratio float64) int64 {
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens)
	if textRequest.MaxTokens != 0 {
		preConsumedTokens += int64(textRequest.MaxTokens)
	}
	return int64(float64(preConsumedTokens) * ratio)
}

func preConsumeQuota(ctx context.Context, textRequest *model2.GeneralOpenAIRequest, promptTokens int, ratio float64, meta *meta.Meta) (int64, *model2.ErrorWithStatusCode) {
	preConsumedQuota := getPreConsumedQuota(textRequest, promptTokens, ratio)

	userQuota, err := services.UserService.CacheGetUserQuota(ctx, meta.UserId)
	if err != nil {
		return preConsumedQuota, openai2.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-preConsumedQuota < 0 {
		return preConsumedQuota, openai2.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	err = services.UserService.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota)
	if err != nil {
		return preConsumedQuota, openai2.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
		slog.With("request_id", logger.GetRequestID(ctx)).
			Info(fmt.Sprintf("user %d has enough quota %d, trusted and no need to pre-consume", meta.UserId, userQuota))
	}

	if preConsumedQuota > 0 {
		err := services.TokenService.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota)
		if err != nil {
			return preConsumedQuota, openai2.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}

	return preConsumedQuota, nil
}

type postConsumeQuotaParams struct {
	usage            *model2.Usage
	meta             *meta.Meta
	textRequest      *model2.GeneralOpenAIRequest
	ratio            float64
	preConsumedQuota int64

	modelRatio        float64
	groupRatio        float64
	systemPromptReset bool
}

func postConsumeQuota(ctx context.Context, p postConsumeQuotaParams) {
	if p.usage == nil {
		slog.With("request_id", logger.GetRequestID(ctx)).Error("usage is nil, which is unexpected")
		return
	}

	var quota int64
	completionRatio := billingratio.GetCompletionRatio(p.textRequest.Model, p.meta.ChannelType)
	promptTokens := p.usage.PromptTokens
	completionTokens := p.usage.CompletionTokens
	quota = int64(math.Ceil((float64(promptTokens) + float64(completionTokens)*completionRatio) * p.ratio))
	if p.ratio != 0 && quota <= 0 {
		quota = 1
	}

	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
	}

	quotaDelta := quota - p.preConsumedQuota
	err := services.TokenService.PostConsumeTokenQuota(p.meta.TokenId, quotaDelta)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error("failed to handler postConsumeTokenQuota err: %s", err.Error())
	}

	err = services.UserService.CacheUpdateUserQuota(ctx, p.meta.UserId)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("failed to handler cacheUpdateUserQuota err: %s", err.Error()))
	}

	logContent := fmt.Sprintf("倍率：%.2f × %.2f × %.2f", p.modelRatio, p.groupRatio, completionRatio)
	logEntry := &entity.Log{
		UserId:            p.meta.UserId,
		ChannelId:         p.meta.ChannelId,
		PromptTokens:      promptTokens,
		CompletionTokens:  completionTokens,
		ModelName:         p.textRequest.Model,
		TokenName:         p.meta.TokenName,
		Quota:             int(quota),
		Content:           logContent,
		IsStream:          p.meta.IsStream,
		ElapsedTime:       time.Now().Sub(p.meta.StartTime).Milliseconds(),
		SystemPromptReset: p.systemPromptReset,
	}

	services.LogService.RecordConsumeLog(ctx, logEntry)
	services.UserService.UpdateUserUsedQuotaAndRequestCount(p.meta.UserId, quota)
	services.ChannelService.UpdateChannelUsedQuota(p.meta.ChannelId, quota)
}

func getMappedModelName(modelName string, mapping map[string]string) (string, bool) {
	if mapping == nil {
		return modelName, false
	}
	mappedModelName := mapping[modelName]
	if mappedModelName != "" {
		return mappedModelName, true
	}
	return modelName, false
}

func isErrorHappened(meta *meta.Meta, resp *http.Response) bool {
	if resp == nil {
		if meta.ChannelType == channeltype.AwsClaude {
			return false
		}
		return true
	}
	if resp.StatusCode != http.StatusOK &&
		// replicate return 201 to create a task
		resp.StatusCode != http.StatusCreated {
		return true
	}
	if meta.ChannelType == channeltype.DeepL {
		// skip stream check for deepl
		return false
	}

	if meta.IsStream && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") &&
		// Even if stream mode is enabled, replicate will first return a task info in JSON format,
		// requiring the client to request the stream endpoint in the task info
		meta.ChannelType != channeltype.Replicate {
		return true
	}
	return false
}

func setSystemPrompt(ctx context.Context, request *model2.GeneralOpenAIRequest, prompt string) (reset bool) {
	if prompt == "" {
		return false
	}
	if len(request.Messages) == 0 {
		return false
	}
	if request.Messages[0].Role == role.System {
		request.Messages[0].Content = prompt
		slog.With("request_id", logger.GetRequestID(ctx)).Info("rewrite system prompt")
		return true
	}
	request.Messages = append([]model2.Message{{
		Role:    role.System,
		Content: prompt,
	}}, request.Messages...)
	slog.With("request_id", logger.GetRequestID(ctx)).Info("add system prompt")
	return true
}
