package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/logger"
	monitor2 "hermes-ai/internal/infras/monitor"
	controller2 "hermes-ai/internal/infras/relay/controller"
	"hermes-ai/internal/infras/relay/model"
	relaymode2 "hermes-ai/internal/infras/relay/relaymode"
	"hermes-ai/internal/infras/utils"
)

// RelayHandler 转发处理器
type RelayHandler struct {
	channelService *application.ChannelService
	channelMonitor *monitor2.ChannelMonitor
	debugEnabled   bool
	retryTimes     int
}

// NewRelayHandler 创建转发处理器
func NewRelayHandler(channelService *application.ChannelService, monitor *monitor2.ChannelMonitor, debugEnabled bool, retryTimes int) *RelayHandler {
	return &RelayHandler{channelService: channelService, channelMonitor: monitor, debugEnabled: debugEnabled, retryTimes: retryTimes}
}

func (h *RelayHandler) relayHelper(c *gin.Context, relayMode int) *model.ErrorWithStatusCode {
	var err *model.ErrorWithStatusCode
	switch relayMode {
	case relaymode2.ImagesGenerations:
		err = controller2.RelayImageHelper(c, relayMode)
	case relaymode2.AudioSpeech:
		fallthrough
	case relaymode2.AudioTranslation:
		fallthrough
	case relaymode2.AudioTranscription:
		err = controller2.RelayAudioHelper(c, relayMode)
	case relaymode2.Proxy:
		err = controller2.RelayProxyHelper(c, relayMode)
	case relaymode2.AnthropicMessages:
		err = controller2.RelayAnthropicMessagesHelper(c)
	default:
		err = controller2.RelayTextHelper(c)
	}
	return err
}

// Relay 转发请求
func (h *RelayHandler) Relay(c *gin.Context) {
	ctx := c.Request.Context()
	relayMode := relaymode2.GetByPath(c.Request.URL.Path)
	if h.debugEnabled {
		requestBody, _ := ginzo.GetRequestBody(c)
		slog.With("request_id", logger.GetRequestID(ctx)).
			Debug(fmt.Sprintf("request body: %s", string(requestBody)))
	}
	channelId := c.GetInt(ctxkey.ChannelId)
	userId := c.GetInt(ctxkey.Id)
	// log.Println("relayMode:", relayMode)
	bizErr := h.relayHelper(c, relayMode)
	if bizErr == nil {
		monitor2.Emit(channelId, true)
		return
	}
	lastFailedChannelId := channelId
	channelName := c.GetString(ctxkey.ChannelName)
	group := c.GetString(ctxkey.Group)
	originalModel := c.GetString(ctxkey.OriginalModel)
	go h.processChannelRelayError(ctx, userId, channelId, channelName, *bizErr)
	requestId := c.GetString(ctxkey.RequestIdKey)
	retryTimes := h.retryTimes
	if !h.shouldRetry(c, bizErr.StatusCode) {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error(fmt.Sprintf("relay error happen, status code is %d, won't retry in this case", bizErr.StatusCode))
		retryTimes = 0
	}
	for i := retryTimes; i > 0; i-- {
		channel, err := h.channelService.CacheGetRandomSatisfiedChannel(group, originalModel, i != retryTimes)
		if err != nil {
			slog.With("request_id", logger.GetRequestID(ctx)).
				Error(fmt.Sprintf("CacheGetRandomSatisfiedChannel failed: %+v", err))
			break
		}
		slog.With("request_id", logger.GetRequestID(ctx)).
			Info(fmt.Sprintf("using channel #%d to retry (remain times %d)", channel.Id, i))
		if channel.Id == lastFailedChannelId {
			continue
		}

		ginzo.SetupContextForSelectedChannel(c, channel, originalModel)
		requestBody, err := ginzo.GetRequestBody(c)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		bizErr = h.relayHelper(c, relayMode)
		if bizErr == nil {
			return
		}
		channelId := c.GetInt(ctxkey.ChannelId)
		lastFailedChannelId = channelId
		channelName := c.GetString(ctxkey.ChannelName)
		go h.processChannelRelayError(ctx, userId, channelId, channelName, *bizErr)
	}
	if bizErr != nil {
		if bizErr.StatusCode == http.StatusTooManyRequests {
			bizErr.Error.Message = "当前分组上游负载已饱和，请稍后再试"
		}

		bizErr.Error.Message = utils.MessageWithRequestId(bizErr.Error.Message, requestId)
		c.JSON(bizErr.StatusCode, gin.H{
			"error": bizErr.Error,
		})
	}
}

func (h *RelayHandler) shouldRetry(c *gin.Context, statusCode int) bool {
	if _, ok := c.Get(ctxkey.SpecificChannelId); ok {
		return false
	}
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode/100 == 5 {
		return true
	}
	if statusCode == http.StatusBadRequest {
		return false
	}
	if statusCode/100 == 2 {
		return false
	}
	return true
}

func (h *RelayHandler) processChannelRelayError(ctx context.Context, userId int, channelId int, channelName string, err model.ErrorWithStatusCode) {
	slog.With("request_id", logger.GetRequestID(ctx)).
		Error(fmt.Sprintf("relay error (channel id %d, user id: %d): %s", channelId, userId, err.Message))
	if monitor2.ShouldDisableChannel(&err.Error, err.StatusCode) {
		h.channelMonitor.DisableChannel(channelId, channelName, err.Message)
	} else {
		monitor2.Emit(channelId, false)
	}
}

// RelayNotImplemented 未实现接口
func (h *RelayHandler) RelayNotImplemented(c *gin.Context) {
	err := model.Error{
		Message: "API not implemented",
		Type:    "one_api_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

// RelayNotFound 接口不存在
func (h *RelayHandler) RelayNotFound(c *gin.Context) {
	err := model.Error{
		Message: "Invalid URL",
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}
