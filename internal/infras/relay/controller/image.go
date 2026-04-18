package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/relay"
	"hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/billing/ratio"
	"hermes-ai/internal/infras/relay/channeltype"
	"hermes-ai/internal/infras/relay/meta"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/relay/services"
)

func getImageRequest(c *gin.Context, _ int) (*model2.ImageRequest, error) {
	imageRequest := &model2.ImageRequest{}
	err := ginzo.UnmarshalBodyReusable(c, imageRequest)
	if err != nil {
		return nil, err
	}
	if imageRequest.N == 0 {
		imageRequest.N = 1
	}
	if imageRequest.Size == "" {
		imageRequest.Size = "1024x1024"
	}
	if imageRequest.Model == "" {
		imageRequest.Model = "dall-e-2"
	}
	return imageRequest, nil
}

func isValidImageSize(model string, size string) bool {
	if model == "cogview-3" || ratio.ImageSizeRatios[model] == nil {
		return true
	}
	_, ok := ratio.ImageSizeRatios[model][size]
	return ok
}

func isValidImagePromptLength(model string, promptLength int) bool {
	maxPromptLength, ok := ratio.ImagePromptLengthLimitations[model]
	return !ok || promptLength <= maxPromptLength
}

func isWithinRange(element string, value int) bool {
	amounts, ok := ratio.ImageGenerationAmounts[element]
	return !ok || (value >= amounts[0] && value <= amounts[1])
}

func getImageSizeRatio(model string, size string) float64 {
	if r, ok := ratio.ImageSizeRatios[model][size]; ok {
		return r
	}

	return 1
}

func validateImageRequest(imageRequest *model2.ImageRequest, _ *meta.Meta) *model2.ErrorWithStatusCode {
	// check prompt length
	if imageRequest.Prompt == "" {
		return openai.ErrorWrapper(errors.New("prompt is required"), "prompt_missing", http.StatusBadRequest)
	}

	// model validation
	if !isValidImageSize(imageRequest.Model, imageRequest.Size) {
		return openai.ErrorWrapper(errors.New("size not supported for this image model"), "size_not_supported", http.StatusBadRequest)
	}

	if !isValidImagePromptLength(imageRequest.Model, len(imageRequest.Prompt)) {
		return openai.ErrorWrapper(errors.New("prompt is too long"), "prompt_too_long", http.StatusBadRequest)
	}

	// Number of generated images validation
	if !isWithinRange(imageRequest.Model, imageRequest.N) {
		return openai.ErrorWrapper(errors.New("invalid value of n"), "n_not_within_range", http.StatusBadRequest)
	}
	return nil
}

func getImageCostRatio(imageRequest *model2.ImageRequest) (float64, error) {
	if imageRequest == nil {
		return 0, errors.New("imageRequest is nil")
	}
	imageCostRatio := getImageSizeRatio(imageRequest.Model, imageRequest.Size)
	if imageRequest.Quality == "hd" && imageRequest.Model == "dall-e-3" {
		if imageRequest.Size == "1024x1024" {
			imageCostRatio *= 2
		} else {
			imageCostRatio *= 1.5
		}
	}
	return imageCostRatio, nil
}

func RelayImageHelper(c *gin.Context, relayMode int) *model2.ErrorWithStatusCode {
	ctx := c.Request.Context()
	metaEntry := meta.GetByContext(c)
	imageRequest, err := getImageRequest(c, metaEntry.Mode)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("getImageRequest failed: %s", err.Error()))
		return openai.ErrorWrapper(err, "invalid_image_request", http.StatusBadRequest)
	}

	// map model name
	var isModelMapped bool
	metaEntry.OriginModelName = imageRequest.Model
	imageRequest.Model, isModelMapped = getMappedModelName(imageRequest.Model, metaEntry.ModelMapping)
	metaEntry.ActualModelName = imageRequest.Model

	// model validation
	bizErr := validateImageRequest(imageRequest, metaEntry)
	if bizErr != nil {
		return bizErr
	}

	imageCostRatio, err := getImageCostRatio(imageRequest)
	if err != nil {
		return openai.ErrorWrapper(err, "get_image_cost_ratio_failed", http.StatusInternalServerError)
	}

	imageModel := imageRequest.Model
	// Convert the original image model
	imageRequest.Model, _ = getMappedModelName(imageRequest.Model, ratio.ImageOriginModelName)
	c.Set(ctxkey.ResponseFormat, imageRequest.ResponseFormat)

	var requestBody io.Reader
	if isModelMapped || metaEntry.ChannelType == channeltype.Azure { // make Azure channel request body
		jsonStr, err := json.Marshal(imageRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}

	adaptor := relay.GetAdaptor(metaEntry.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", metaEntry.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(metaEntry)

	// these adaptors need to convert the request
	switch metaEntry.ChannelType {
	case channeltype.Zhipu,
		channeltype.Ali,
		channeltype.Replicate,
		channeltype.Baidu:
		finalRequest, err := adaptor.ConvertImageRequest(imageRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "convert_image_request_failed", http.StatusInternalServerError)
		}
		jsonStr, err := json.Marshal(finalRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_image_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	default:
	}

	modelRatio := ratio.GetModelRatio(imageModel, metaEntry.ChannelType)
	groupRatio := ratio.GetGroupRatio(metaEntry.Group)
	r := modelRatio * groupRatio
	userQuota, err := services.UserService.CacheGetUserQuota(c.Request.Context(), metaEntry.UserId)

	var quota int64
	switch metaEntry.ChannelType {
	case channeltype.Replicate:
		// replicate always return 1 image
		quota = int64(r * imageCostRatio * 1000)
	default:
		quota = int64(r*imageCostRatio*1000) * int64(imageRequest.N)
	}

	if userQuota-quota < 0 {
		return openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	// do request
	resp, err := adaptor.DoRequest(c, metaEntry, requestBody)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("DoRequest failed: %s", err.Error()))
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	defer func(ctx context.Context) {
		if resp != nil &&
			resp.StatusCode != http.StatusCreated && // replicate returns 201
			resp.StatusCode != http.StatusOK {
			return
		}

		err2 := services.TokenService.PostConsumeTokenQuota(metaEntry.TokenId, quota)
		if err2 != nil {
			slog.With("request_id", logger.GetRequestID(ctx)).
				Error(fmt.Sprintf("failed to handler postConsumeTokenQuota err: %s", err2.Error()))
		}

		err2 = services.UserService.CacheUpdateUserQuota(ctx, metaEntry.UserId)
		if err2 != nil {
			slog.With("request_id", logger.GetRequestID(ctx)).
				Error(fmt.Sprintf("failed to handler cacheUpdateUserQuota err: %s", err2.Error()))
		}

		if quota != 0 {
			logContent := fmt.Sprintf("倍率：%.2f × %.2f", modelRatio, groupRatio)
			logEntry := &entity.Log{
				UserId:           metaEntry.UserId,
				ChannelId:        metaEntry.ChannelId,
				PromptTokens:     0,
				CompletionTokens: 0,
				ModelName:        imageRequest.Model,
				TokenName:        metaEntry.TokenName,
				Quota:            int(quota),
				Content:          logContent,
			}

			services.LogService.RecordConsumeLog(ctx, logEntry)
			services.UserService.UpdateUserUsedQuotaAndRequestCount(metaEntry.UserId, quota)
			channelId := c.GetInt(ctxkey.ChannelId)
			services.ChannelService.UpdateChannelUsedQuota(channelId, quota)
		}
	}(c.Request.Context())

	// do response
	_, respErr := adaptor.DoResponse(c, resp, metaEntry)
	if respErr != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("respErr is not nil: %+v", respErr))
		return respErr
	}

	return nil
}
