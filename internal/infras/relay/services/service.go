package services

import (
	"hermes-ai/internal/application"
)

var (
	UserService    *application.UserService
	TokenService   *application.TokenService
	LogService     *application.LogService
	ChannelService *application.ChannelService
)

// Init 初始化服务
func Init(userService *application.UserService, tokenService *application.TokenService,
	logService *application.LogService, channelService *application.ChannelService) {
	UserService = userService
	TokenService = tokenService
	LogService = logService
	ChannelService = channelService
}
