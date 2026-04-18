package handlers

import (
	"hermes-ai/internal/application"
	monitor2 "hermes-ai/internal/infras/monitor"
)

// HandlerContainers 处理器容器，集中管理所有handler实例
type HandlerContainers struct {
	TokenHandler          *TokenHandler
	UserHandler           *UserHandler
	ChannelHandler        *ChannelHandler
	LogHandler            *LogHandler
	OptionHandler         *OptionHandler
	RedemptionHandler     *RedemptionHandler
	MiscHandler           *MiscHandler
	AuthHandler           *AuthHandler
	ChannelTestHandler    *ChannelTestHandler
	ChannelBillingHandler *ChannelBillingHandler
	BillingHandler        *BillingHandler
	ModelHandler          *ModelHandler
	GroupHandler          *GroupHandler
	RelayHandler          *RelayHandler
	GitHubHandler         *GitHubHandler
	LarkUserHandler       *LarkUserHandler
	OidcUserHandler       *OidcUserHandler
	WeChatUserHandler     *WeChatUserHandler
}

// NewHandlerContainer 创建处理器容器
func NewHandlerContainer(
	services *application.Services,
	channelMonitor *monitor2.ChannelMonitor,
	larkUserConfig LarkUserConfig,
	githubUserConfig GitHubUserConfig,
	authConfig AuthConfig,
) *HandlerContainers {
	return &HandlerContainers{
		TokenHandler:      NewTokenHandler(services.TokenService),
		UserHandler:       NewUserHandler(services.UserService, services.LogService, services.RedemptionService),
		ChannelHandler:    NewChannelHandler(services.ChannelService),
		LogHandler:        NewLogHandler(services.LogService),
		OptionHandler:     NewOptionHandler(services.OptionService),
		RedemptionHandler: NewRedemptionHandler(services.RedemptionService),
		MiscHandler:       NewMiscHandler(services.UserService),
		AuthHandler:       NewAuthHandler(services.UserService, authConfig),
		ChannelTestHandler: NewChannelTestHandler(
			services.ChannelService,
			services.LogService,
			services.UserService,
			channelMonitor,
		),
		ChannelBillingHandler: NewChannelBillingHandler(services.ChannelService, channelMonitor),
		BillingHandler:        NewBillingHandler(services.UserService, services.TokenService),
		ModelHandler:          NewModelHandler(services.UserService, services.ChannelService),
		GroupHandler:          NewGroupHandler(),
		RelayHandler:          NewRelayHandler(services.ChannelService, channelMonitor),
		GitHubHandler:         NewGitHubHandler(services.UserService, githubUserConfig),
		LarkUserHandler:       NewLarkUserHandler(services.UserService, larkUserConfig),
		OidcUserHandler:       NewOidcUserHandler(services.UserService),
		WeChatUserHandler:     NewWechatLoginHandler(services.UserService),
	}
}
