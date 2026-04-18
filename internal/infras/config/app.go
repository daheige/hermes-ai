package config

import (
	"sync"
	"time"
)

// AppConfig 应用运行时配置
// OptionService 通过修改此结构的字段实现运行时配置变更
// 配置变更会通过 OptionService 实时同步到 AppConfig 中
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

	// OptionMap 配置键值对，由 OptionService 统一管理并保证并发安全
	// handlers 应通过 OptionService 读取，不直接持有此 map 的副本
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
