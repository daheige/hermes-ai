package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/logger"
)

type DistributorMiddleware struct {
	userService    *application.UserService
	channelService *application.ChannelService
}

func NewDistributorMiddleware(userService *application.UserService, channelService *application.ChannelService) *DistributorMiddleware {
	return &DistributorMiddleware{userService: userService, channelService: channelService}
}

type ModelRequest struct {
	Model string `json:"model" form:"model"`
}

func (d *DistributorMiddleware) Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt(ctxkey.Id)
		userGroup, _ := d.userService.CacheGetUserGroup(userId)
		c.Set(ctxkey.Group, userGroup)
		var requestModel string
		var channel *entity.Channel
		channelId, ok := c.Get(ctxkey.SpecificChannelId)
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = d.channelService.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != entity.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
		} else {
			requestModel = c.GetString(ctxkey.RequestModel)
			var err error
			channel, err = d.channelService.CacheGetRandomSatisfiedChannel(userGroup, requestModel, false)
			// log.Println("channel", channel, "err:", err)
			if err != nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, requestModel)
				if channel != nil {
					slog.Error(fmt.Sprintf("渠道不存在：%d", channel.Id))
					message = "数据库一致性已被破坏，请联系管理员"
				}
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
		}

		slog.With("request_id", logger.GetRequestID(ctx)).
			Debug(fmt.Sprintf("user id %d, user group: %s, request model: %s, using channel #%d",
				userId, userGroup, requestModel, channel.Id))
		d.SetupContextForSelectedChannel(c, channel, requestModel)
		c.Next()
	}
}

func (d *DistributorMiddleware) SetupContextForSelectedChannel(c *gin.Context, channel *entity.Channel, modelName string) {
	ginzo.SetupContextForSelectedChannel(c, channel, modelName)
}
