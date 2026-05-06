package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"hermes-ai/internal/infras/env"
)

// SystemConfig 系统启动配置
// 所有环境变量相关的配置统一在此结构体中定义
// 通过 InitSystemConfig() 函数显式初始化
type SystemConfig struct {
	SystemName               string
	ServerAddress            string
	Footer                   string
	Logo                     string
	TopUpLink                string
	ChatLink                 string
	QuotaPerUnit             float64
	DisplayInCurrencyEnabled bool
	DisplayTokenStatEnabled  bool

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

	DebugEnabled       bool
	DebugSQLEnabled    bool
	MemoryCacheEnabled bool

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

	RelayTimeout int

	Theme       string
	ValidThemes map[string]bool

	GlobalApiRateLimitNum          int
	GlobalApiRateLimitDuration     int64
	GlobalWebRateLimitNum          int
	GlobalWebRateLimitDuration     int64
	UploadRateLimitNum             int
	UploadRateLimitDuration        int64
	DownloadRateLimitNum           int
	DownloadRateLimitDuration      int64
	CriticalRateLimitNum           int
	CriticalRateLimitDuration      int64
	RateLimitKeyExpirationDuration time.Duration

	EnableMetric               bool
	MetricQueueSize            int
	MetricSuccessRateThreshold float64
	MetricSuccessChanSize      int
	MetricFailChanSize         int

	InitialRootToken       string
	InitialRootAccessToken string

	RelayProxy                string
	UserContentRequestProxy   string
	UserContentRequestTimeout int

	EnforceIncludeUsage bool
	TestPrompt          string

	AESSecretKey string

	GinMode              string
	LogSQLDSN            string
	ChannelTestFrequency int
	FrontendBaseURL      string

	SQLDSN          string
	SQLMaxIdleConns int
	SQLMaxOpenConns int
	SQLMaxLifetime  int

	RedisConnString    string
	RedisEnableCluster bool
	RedisPassword      string
	RedisUsername      string

	Port         int
	GracefulWait int
	LogLevel     string
	LogDir       string
}

// InitSystemConfig 从环境变量初始化系统配置
func InitSystemConfig() *SystemConfig {
	requestInterval, _ := strconv.Atoi(os.Getenv("POLLING_INTERVAL"))

	channelTestFrequency := 0
	if v := os.Getenv("CHANNEL_TEST_FREQUENCY"); v != "" {
		channelTestFrequency, _ = strconv.Atoi(v)
	}

	return &SystemConfig{
		SystemName:               "AI Gateway",
		ServerAddress:            "http://localhost:1337",
		Footer:                   "",
		Logo:                     "",
		TopUpLink:                "",
		ChatLink:                 "",
		QuotaPerUnit:             500 * 1000.0,
		DisplayInCurrencyEnabled: true,
		DisplayTokenStatEnabled:  true,

		ItemsPerPage:   10,
		MaxRecentItems: 100,

		PasswordLoginEnabled:     true,
		PasswordRegisterEnabled:  true,
		EmailVerificationEnabled: false,
		GitHubOAuthEnabled:       false,
		OidcEnabled:              false,
		WeChatAuthEnabled:        false,
		TurnstileCheckEnabled:    false,
		RegisterEnabled:          true,

		EmailDomainRestrictionEnabled: false,
		EmailDomainWhitelist: []string{
			"gmail.com",
			"163.com",
			"126.com",
			"qq.com",
			"outlook.com",
			"hotmail.com",
			"icloud.com",
			"yahoo.com",
			"foxmail.com",
		},

		DebugEnabled:       strings.ToLower(os.Getenv("DEBUG")) == "true",
		DebugSQLEnabled:    strings.ToLower(os.Getenv("DEBUG_SQL")) == "true",
		MemoryCacheEnabled: strings.ToLower(os.Getenv("MEMORY_CACHE_ENABLED")) == "true",

		LogConsumeEnabled: true,

		SMTPServer:  "",
		SMTPPort:    587,
		SMTPAccount: "",
		SMTPFrom:    "",
		SMTPToken:   "",

		GitHubClientId:     "",
		GitHubClientSecret: "",

		LarkClientId:     "",
		LarkClientSecret: "",

		OidcClientId:              "",
		OidcClientSecret:          "",
		OidcWellKnown:             "",
		OidcAuthorizationEndpoint: "",
		OidcTokenEndpoint:         "",
		OidcUserinfoEndpoint:      "",

		WeChatServerAddress:         "",
		WeChatServerToken:           "",
		WeChatAccountQRCodeImageURL: "",

		MessagePusherAddress: "",
		MessagePusherToken:   "",

		TurnstileSiteKey:   "",
		TurnstileSecretKey: "",

		QuotaForNewUser:         0,
		QuotaForInviter:         0,
		QuotaForInvitee:         0,
		ChannelDisableThreshold: 5.0,

		AutomaticDisableChannelEnabled: false,
		AutomaticEnableChannelEnabled:  false,
		QuotaRemindThreshold:           1000,
		PreConsumedQuota:               500,
		ApproximateTokenEnabled:        false,
		RetryTimes:                     0,

		RootUserEmail: "",

		RequestInterval: time.Duration(requestInterval) * time.Second,
		SyncFrequency:   env.Int("SYNC_FREQUENCY", 10*60),

		BatchUpdateEnabled:  env.Bool("BATCH_UPDATE_ENABLED", false),
		BatchUpdateInterval: env.Int("BATCH_UPDATE_INTERVAL", 5),

		RelayTimeout: env.Int("RELAY_TIMEOUT", 0),

		Theme: env.String("THEME", "default"),
		ValidThemes: map[string]bool{
			"default": true,
			"berry":   true,
			"air":     true,
		},

		GlobalApiRateLimitNum:      env.Int("GLOBAL_API_RATE_LIMIT", 480),
		GlobalApiRateLimitDuration: 3 * 60,

		GlobalWebRateLimitNum:      env.Int("GLOBAL_WEB_RATE_LIMIT", 240),
		GlobalWebRateLimitDuration: 3 * 60,

		UploadRateLimitNum:      10,
		UploadRateLimitDuration: 60,

		DownloadRateLimitNum:      10,
		DownloadRateLimitDuration: 60,

		CriticalRateLimitNum:      20,
		CriticalRateLimitDuration: 20 * 60,

		RateLimitKeyExpirationDuration: 20 * time.Minute,

		EnableMetric:               env.Bool("ENABLE_METRIC", false),
		MetricQueueSize:            env.Int("METRIC_QUEUE_SIZE", 10),
		MetricSuccessRateThreshold: env.Float64("METRIC_SUCCESS_RATE_THRESHOLD", 0.8),
		MetricSuccessChanSize:      env.Int("METRIC_SUCCESS_CHAN_SIZE", 1024),
		MetricFailChanSize:         env.Int("METRIC_FAIL_CHAN_SIZE", 128),

		InitialRootToken:       os.Getenv("INITIAL_ROOT_TOKEN"),
		InitialRootAccessToken: os.Getenv("INITIAL_ROOT_ACCESS_TOKEN"),

		RelayProxy:                env.String("RELAY_PROXY", ""),
		UserContentRequestProxy:   env.String("USER_CONTENT_REQUEST_PROXY", ""),
		UserContentRequestTimeout: env.Int("USER_CONTENT_REQUEST_TIMEOUT", 30),

		EnforceIncludeUsage: env.Bool("ENFORCE_INCLUDE_USAGE", false),
		TestPrompt:          env.String("TEST_PROMPT", "Output only your specific model name with no additional text."),

		AESSecretKey: os.Getenv("AES_SECRET_KEY"),

		GinMode:   os.Getenv("GIN_MODE"),
		LogSQLDSN: os.Getenv("LOG_SQL_DSN"),

		ChannelTestFrequency: channelTestFrequency,

		FrontendBaseURL: os.Getenv("FRONTEND_BASE_URL"),

		SQLDSN:          os.Getenv("SQL_DSN"),
		SQLMaxIdleConns: env.Int("SQL_MAX_IDLE_CONNS", 100),
		SQLMaxOpenConns: env.Int("SQL_MAX_OPEN_CONNS", 1000),
		SQLMaxLifetime:  env.Int("SQL_MAX_LIFETIME", 60),

		RedisConnString:    os.Getenv("REDIS_CONN_STRING"),
		RedisEnableCluster: env.Bool("REDIS_ENABLE_CLUSTER", false),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisUsername:      os.Getenv("REDIS_USERNAME"),

		Port:         env.Int("PORT", 1337),
		GracefulWait: env.Int("GRACEFUL_WAIT", 5),
		LogLevel:     env.String("LOG_LEVEL", "info"),
		LogDir:       env.String("LOG_DIR", ""),
	}
}
