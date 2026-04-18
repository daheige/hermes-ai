package handlers

import (
	"time"

	monitor2 "hermes-ai/internal/infras/monitor"
	"hermes-ai/internal/providers"
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

type HandlerParams struct {
	LarkUserConfig   LarkUserConfig
	GithubUserConfig GitHubUserConfig
	AuthConfig       AuthConfig
	WeChatUserConfig WeChatUserConfig
	OidcUserConfig   OidcUserConfig
	MiscConfig       MiscConfig

	ItemsPerPage                   int
	QuotaPerUnit                   float64
	DisplayInCurrencyEnabled       bool
	RootUserEmail                  *string
	TestPrompt                     string
	ChannelDisableThreshold        float64
	AutomaticDisableChannelEnabled bool
	RequestInterval                time.Duration
	DisplayTokenStatEnabled        bool
	DebugEnabled                   bool
	RetryTimes                     int
	OptionMap                      map[string]string
	ValidThemes                    map[string]bool
	GithubClientId                 string
	EmailDomainWhitelist           []string
	WeChatServerAddress            string
	TurnstileSiteKey               string
}

// NewHandlerContainer 创建处理器容器
func NewHandlerContainer(
	services *providers.Services,
	channelMonitor *monitor2.ChannelMonitor,
	p *HandlerParams,
) *HandlerContainers {
	return &HandlerContainers{
		TokenHandler: NewTokenHandler(services.TokenService, p.ItemsPerPage),
		UserHandler: NewUserHandler(
			services.UserService, services.LogService, services.RedemptionService,
			&UserHandlerParams{
				itemsPerPage:             p.ItemsPerPage,
				quotaPerUnit:             p.QuotaPerUnit,
				displayInCurrencyEnabled: p.DisplayInCurrencyEnabled,
				rootUserEmail:            p.RootUserEmail,
			},
		),
		ChannelHandler: NewChannelHandler(services.ChannelService, p.ItemsPerPage),
		LogHandler:     NewLogHandler(services.LogService, p.ItemsPerPage),
		OptionHandler: NewOptionHandler(services.OptionService, OptionConfig{
			OptionMap:            p.OptionMap,
			ValidThemes:          p.ValidThemes,
			GithubClientId:       p.GithubClientId,
			EmailDomainWhitelist: p.EmailDomainWhitelist,
			WeChatServerAddress:  p.WeChatServerAddress,
			TurnstileSiteKey:     p.TurnstileSiteKey,
		}),
		RedemptionHandler: NewRedemptionHandler(services.RedemptionService, p.ItemsPerPage),
		MiscHandler:       NewMiscHandler(services.UserService, p.MiscConfig),
		AuthHandler:       NewAuthHandler(services.UserService, p.AuthConfig),
		ChannelTestHandler: NewChannelTestHandler(
			services.ChannelService,
			services.LogService,
			services.UserService,
			channelMonitor,
			p.TestPrompt,
			p.ChannelDisableThreshold,
			p.AutomaticDisableChannelEnabled,
			p.RequestInterval,
		),
		ChannelBillingHandler: NewChannelBillingHandler(services.ChannelService, channelMonitor, p.RequestInterval),
		BillingHandler: NewBillingHandler(
			services.UserService, services.TokenService,
			p.DisplayTokenStatEnabled, p.DisplayInCurrencyEnabled, p.QuotaPerUnit,
		),
		ModelHandler:      NewModelHandler(services.UserService, services.ChannelService),
		GroupHandler:      NewGroupHandler(),
		RelayHandler:      NewRelayHandler(services.ChannelService, channelMonitor, p.DebugEnabled, p.RetryTimes),
		GitHubHandler:     NewGitHubHandler(services.UserService, p.GithubUserConfig),
		LarkUserHandler:   NewLarkUserHandler(services.UserService, p.LarkUserConfig),
		OidcUserHandler:   NewOidcUserHandler(services.UserService, p.OidcUserConfig),
		WeChatUserHandler: NewWechatLoginHandler(services.UserService, p.WeChatUserConfig),
	}
}
