package handlers

import (
	"time"

	monitor2 "hermes-ai/internal/infras/monitor"
	"hermes-ai/internal/infras/relay"
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

	ItemsPerPage             int
	QuotaPerUnit             float64
	DisplayInCurrencyEnabled bool

	TestPrompt                     string
	ChannelDisableThreshold        float64
	AutomaticDisableChannelEnabled bool
	RequestInterval                time.Duration
	DisplayTokenStatEnabled        bool
	RetryTimes                     int
	PreConsumedQuota               int64
	EnforceIncludeUsage            bool
	EnableMetric                   bool
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
	adaptorFactory *relay.AdaptorFactory,
	metricCollector *monitor2.MetricCollector,
	p *HandlerParams,
) *HandlerContainers {
	billParams := &BillingHandlerParams{
		userService:              services.UserService,
		tokenService:             services.TokenService,
		displayTokenStatEnabled:  p.DisplayTokenStatEnabled,
		displayInCurrencyEnabled: p.DisplayInCurrencyEnabled,
		quotaPerUnit:             p.QuotaPerUnit,
	}

	userHandlerParams := &UserHandlerParams{
		itemsPerPage:             p.ItemsPerPage,
		quotaPerUnit:             p.QuotaPerUnit,
		displayInCurrencyEnabled: p.DisplayInCurrencyEnabled,
	}

	hc := &HandlerContainers{
		TokenHandler:   NewTokenHandler(services.TokenService, p.ItemsPerPage),
		UserHandler:    NewUserHandler(userHandlerParams),
		ChannelHandler: NewChannelHandler(services.ChannelService, p.ItemsPerPage),
		LogHandler:     NewLogHandler(services.LogService, p.ItemsPerPage),
		OptionHandler: NewOptionHandler(services.OptionService, OptionConfig{
			ValidThemes:          p.ValidThemes,
			GithubClientId:       p.GithubClientId,
			EmailDomainWhitelist: p.EmailDomainWhitelist,
			WeChatServerAddress:  p.WeChatServerAddress,
			TurnstileSiteKey:     p.TurnstileSiteKey,
		}),
		RedemptionHandler: NewRedemptionHandler(services.RedemptionService, p.ItemsPerPage),
		MiscHandler:       NewMiscHandler(services.UserService, services.OptionService, p.MiscConfig),
		AuthHandler:       NewAuthHandler(services.UserService, p.AuthConfig),
		ChannelTestHandler: NewChannelTestHandler(ChannelTestHandlerDeps{
			Service:        services.ChannelService,
			LogService:     services.LogService,
			UserService:    services.UserService,
			ChannelMonitor: channelMonitor,
			AdaptorFactory: adaptorFactory,
		}, ChannelTestHandlerConfig{
			TestPrompt:                     p.TestPrompt,
			ChannelDisableThreshold:        p.ChannelDisableThreshold,
			AutomaticDisableChannelEnabled: p.AutomaticDisableChannelEnabled,
			RequestInterval:                p.RequestInterval,
		}),
		ChannelBillingHandler: NewChannelBillingHandler(services.ChannelService, channelMonitor, p.RequestInterval),
		BillingHandler:        NewBillingHandler(billParams),
		ModelHandler:          NewModelHandler(services.UserService, services.ChannelService),
		GroupHandler:          NewGroupHandler(),
		RelayHandler: NewRelayHandler(RelayHandlerDeps{
			ChannelService:  services.ChannelService,
			ChannelMonitor:  channelMonitor,
			AdaptorFactory:  adaptorFactory,
			MetricCollector: metricCollector,
		}, RelayHandlerConfig{
			RetryTimes:                     p.RetryTimes,
			PreConsumedQuota:               p.PreConsumedQuota,
			EnforceIncludeUsage:            p.EnforceIncludeUsage,
			EnableMetric:                   p.EnableMetric,
			AutomaticDisableChannelEnabled: p.AutomaticDisableChannelEnabled,
		}),
		GitHubHandler:     NewGitHubHandler(services.UserService, p.GithubUserConfig),
		LarkUserHandler:   NewLarkUserHandler(services.UserService, p.LarkUserConfig),
		OidcUserHandler:   NewOidcUserHandler(services.UserService, p.OidcUserConfig),
		WeChatUserHandler: NewWechatLoginHandler(services.UserService, p.WeChatUserConfig),
	}

	return hc
}
