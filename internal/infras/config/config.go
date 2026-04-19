package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"hermes-ai/internal/infras/env"
)

var (
	SystemName               = "AI Gateway"
	ServerAddress            = "http://localhost:1337"
	Footer                   = ""
	Logo                     = ""
	TopUpLink                = ""
	ChatLink                 = ""
	QuotaPerUnit             = 500 * 1000.0 // $0.002 / 1K tokens
	DisplayInCurrencyEnabled = true
	DisplayTokenStatEnabled  = true

	ItemsPerPage   = 10
	MaxRecentItems = 100

	PasswordLoginEnabled     = true
	PasswordRegisterEnabled  = true
	EmailVerificationEnabled = false
	GitHubOAuthEnabled       = false
	OidcEnabled              = false
	WeChatAuthEnabled        = false
	TurnstileCheckEnabled    = false
	RegisterEnabled          = true

	EmailDomainRestrictionEnabled = false
	EmailDomainWhitelist          = []string{
		"gmail.com",
		"163.com",
		"126.com",
		"qq.com",
		"outlook.com",
		"hotmail.com",
		"icloud.com",
		"yahoo.com",
		"foxmail.com",
	}

	DebugEnabled       = strings.ToLower(os.Getenv("DEBUG")) == "true"
	DebugSQLEnabled    = strings.ToLower(os.Getenv("DEBUG_SQL")) == "true"
	MemoryCacheEnabled = strings.ToLower(os.Getenv("MEMORY_CACHE_ENABLED")) == "true"

	LogConsumeEnabled = true

	SMTPServer  = ""
	SMTPPort    = 587
	SMTPAccount = ""
	SMTPFrom    = ""
	SMTPToken   = ""

	GitHubClientId     = ""
	GitHubClientSecret = ""

	LarkClientId     = ""
	LarkClientSecret = ""

	OidcClientId              = ""
	OidcClientSecret          = ""
	OidcWellKnown             = ""
	OidcAuthorizationEndpoint = ""
	OidcTokenEndpoint         = ""
	OidcUserinfoEndpoint      = ""

	WeChatServerAddress         = ""
	WeChatServerToken           = ""
	WeChatAccountQRCodeImageURL = ""

	MessagePusherAddress = ""
	MessagePusherToken   = ""

	TurnstileSiteKey   = ""
	TurnstileSecretKey = ""

	QuotaForNewUser                int64 = 0
	QuotaForInviter                int64 = 0
	QuotaForInvitee                int64 = 0
	ChannelDisableThreshold              = 5.0
	AutomaticDisableChannelEnabled       = false
	AutomaticEnableChannelEnabled        = false
	QuotaRemindThreshold           int64 = 1000
	PreConsumedQuota               int64 = 500
	ApproximateTokenEnabled              = false
	RetryTimes                           = 0

	RootUserEmail = ""

	requestInterval, _ = strconv.Atoi(os.Getenv("POLLING_INTERVAL"))
	RequestInterval    = time.Duration(requestInterval) * time.Second

	SyncFrequency = env.Int("SYNC_FREQUENCY", 10*60) // unit is second

	BatchUpdateEnabled = env.Bool("BATCH_UPDATE_ENABLED", false)

	BatchUpdateInterval = env.Int("BATCH_UPDATE_INTERVAL", 5)

	RelayTimeout = env.Int("RELAY_TIMEOUT", 0) // unit is second

	GeminiSafetySetting = env.String("GEMINI_SAFETY_SETTING", "BLOCK_NONE")

	Theme       = env.String("THEME", "default")
	ValidThemes = map[string]bool{
		"default": true,
		"berry":   true,
		"air":     true,
	}

	// All duration's unit is seconds
	// Shouldn't larger then RateLimitKeyExpirationDuration

	GlobalApiRateLimitNum            = env.Int("GLOBAL_API_RATE_LIMIT", 480)
	GlobalApiRateLimitDuration int64 = 3 * 60

	GlobalWebRateLimitNum            = env.Int("GLOBAL_WEB_RATE_LIMIT", 240)
	GlobalWebRateLimitDuration int64 = 3 * 60

	UploadRateLimitNum            = 10
	UploadRateLimitDuration int64 = 60

	DownloadRateLimitNum            = 10
	DownloadRateLimitDuration int64 = 60

	CriticalRateLimitNum            = 20
	CriticalRateLimitDuration int64 = 20 * 60

	RateLimitKeyExpirationDuration = 20 * time.Minute

	EnableMetric               = env.Bool("ENABLE_METRIC", false)
	MetricQueueSize            = env.Int("METRIC_QUEUE_SIZE", 10)
	MetricSuccessRateThreshold = env.Float64("METRIC_SUCCESS_RATE_THRESHOLD", 0.8)
	MetricSuccessChanSize      = env.Int("METRIC_SUCCESS_CHAN_SIZE", 1024)
	MetricFailChanSize         = env.Int("METRIC_FAIL_CHAN_SIZE", 128)

	InitialRootToken = os.Getenv("INITIAL_ROOT_TOKEN")

	InitialRootAccessToken = os.Getenv("INITIAL_ROOT_ACCESS_TOKEN")

	GeminiVersion = env.String("GEMINI_VERSION", "v1")

	RelayProxy                = env.String("RELAY_PROXY", "")
	UserContentRequestProxy   = env.String("USER_CONTENT_REQUEST_PROXY", "")
	UserContentRequestTimeout = env.Int("USER_CONTENT_REQUEST_TIMEOUT", 30)

	EnforceIncludeUsage = env.Bool("ENFORCE_INCLUDE_USAGE", false)
	TestPrompt          = env.String("TEST_PROMPT", "Output only your specific model name with no additional text.")
)
