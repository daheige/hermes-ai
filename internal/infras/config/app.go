package config

import (
	"sync"
	"time"
)

// AppConfig 应用运行时配置
// OptionService 通过修改此结构的字段实现运行时配置变更
type AppConfig struct {
	SystemName               string
	ServerAddress            string
	Footer                   string
	Logo                     string
	TopUpLink                string
	ChatLink                 string
	QuotaPerUnit             float64
	DisplayInCurrencyEnabled bool
	DisplayTokenStatEnabled  bool

	// OptionMap todo 这个如果配置改变，对于handlers目录中的misc_handler.go和option_handler.go改变后，需要实时更新
	OptionMap        map[string]string
	OptionMapRWMutex sync.RWMutex

	ItemsPerPage   int
	MaxRecentItems int

	PasswordLoginEnabled     bool
	PasswordRegisterEnabled  bool
	EmailVerificationEnabled bool
	GitHubOAuthEnabled       bool
	OidcEnabled              bool
	WeChatAuthEnabled        bool
	TurnstileCheckEnabled    bool
	RegisterEnabled          bool

	EmailDomainRestrictionEnabled bool
	EmailDomainWhitelist          []string

	DebugEnabled    bool
	DebugSQLEnabled bool
	CacheEnabled    bool

	LogConsumeEnabled bool

	SMTPServer  string
	SMTPPort    int
	SMTPAccount string
	SMTPFrom    string
	SMTPToken   string

	GitHubClientId     string
	GitHubClientSecret string

	LarkClientId     string
	LarkClientSecret string

	OidcClientId              string
	OidcClientSecret          string
	OidcWellKnown             string
	OidcAuthorizationEndpoint string
	OidcTokenEndpoint         string
	OidcUserinfoEndpoint      string

	WeChatServerAddress         string
	WeChatServerToken           string
	WeChatAccountQRCodeImageURL string

	MessagePusherAddress string
	MessagePusherToken   string

	TurnstileSiteKey   string
	TurnstileSecretKey string

	QuotaForNewUser                int64
	QuotaForInviter                int64
	QuotaForInvitee                int64
	ChannelDisableThreshold        float64
	AutomaticDisableChannelEnabled bool
	AutomaticEnableChannelEnabled  bool
	QuotaRemindThreshold           int64
	PreConsumedQuota               int64
	ApproximateTokenEnabled        bool
	RetryTimes                     int

	RootUserEmail string

	RequestInterval time.Duration
	SyncFrequency   int

	BatchUpdateEnabled  bool
	BatchUpdateInterval int

	Theme       string
	ValidThemes map[string]bool

	GlobalWebRateLimitNum          int
	GlobalWebRateLimitDuration     int64
	GlobalApiRateLimitNum          int
	GlobalApiRateLimitDuration     int64
	CriticalRateLimitNum           int
	CriticalRateLimitDuration      int64
	DownloadRateLimitNum           int
	DownloadRateLimitDuration      int64
	UploadRateLimitNum             int
	UploadRateLimitDuration        int64
	RateLimitKeyExpirationDuration time.Duration

	EnableMetric               bool
	MetricQueueSize            int
	MetricSuccessRateThreshold float64
	MetricSuccessChanSize      int
	MetricFailChanSize         int

	RelayTimeout              int
	UserContentRequestProxy   string
	UserContentRequestTimeout int
	RelayProxy                string

	EnforceIncludeUsage bool
	TestPrompt          string

	InitialRootToken       string
	InitialRootAccessToken string

	GeminiVersion       string
	GeminiSafetySetting string
}
