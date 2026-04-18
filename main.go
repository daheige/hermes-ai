package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/crypto"
	"hermes-ai/internal/infras/env"
	"hermes-ai/internal/infras/httpclient"
	"hermes-ai/internal/infras/i18n"
	"hermes-ai/internal/infras/logger"
	monitor2 "hermes-ai/internal/infras/monitor"
	"hermes-ai/internal/infras/relay/adaptor/openai"
	relayServices "hermes-ai/internal/infras/relay/services"
	"hermes-ai/internal/infras/utils"
	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
	"hermes-ai/internal/interfaces/web/router"
	"hermes-ai/internal/providers"
)

//go:embed web/build/*
var buildFS embed.FS

func main() {
	// 初始化日志
	opts := []logger.Option{
		logger.WithAddSource(true),
		logger.WithEnableJSON(),
		logger.WithLevel(logger.GetLevel(env.String("LOG_LEVEL", "info"))),
	}

	if logDir := env.String("LOG_DIR", ""); logDir != "" {
		opts = append(opts, logger.WithLogDir(logDir), logger.WithOutputToFile(true))
	}

	logger.Default(opts...)

	if os.Getenv("GIN_MODE") != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	if config.DebugEnabled {
		slog.Info("running in debug mode")
	}

	// Initialize SQL Database
	db, logDB := config.InitDatabase()
	defer func() {
		err := config.CloseDB(db)
		if err != nil {
			log.Fatalln("failed to close database: " + err.Error())
		}

		if os.Getenv("LOG_SQL_DSN") != "" {
			err := config.CloseDB(logDB)
			if err != nil {
				log.Fatalln("failed to close database: " + err.Error())
			}
		}
	}()

	err := CreateRootAccountIfNeed(db)
	if err != nil {
		log.Fatalln("database init error: " + err.Error())
	}

	// Initialize Redis
	redisClient, err := config.InitRedisClient()
	if err != nil {
		log.Fatalln("failed to initialize Redis: " + err.Error())
	}

	// Initialize application config
	cfg := initAppConfig()
	cfg.CacheEnabled = true // 使用redis cache

	// init repos
	repos := providers.InitRepositories(db, logDB, redisClient)
	// Initialize application services
	services := providers.InitServices(repos, cfg)

	// Initialize options
	services.OptionService.InitOptionMap()
	slog.Info(fmt.Sprintf("using theme %s", cfg.Theme))

	// 内存缓存对于redis也是
	if cfg.CacheEnabled {
		slog.Info("memory cache enabled")
		slog.Info(fmt.Sprintf("sync frequency: %d seconds", cfg.SyncFrequency))
		services.ChannelService.InitChannelCache()
		go services.OptionService.SyncOptions(cfg.SyncFrequency)
		go services.ChannelService.SyncChannelCache(cfg.SyncFrequency)
	}

	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err != nil {
			log.Fatalln("failed to parse CHANNEL_TEST_FREQUENCY: " + err.Error())
		}
		go handlers.AutomaticallyTestChannels(frequency)
	}

	// 启动批量更新器
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		cfg.BatchUpdateEnabled = true
		slog.Info("batch update enabled with interval " + strconv.Itoa(cfg.BatchUpdateInterval) + "s")
		services.BatchUpdater.Start()
	}

	if config.EnableMetric {
		slog.Info("metric enabled, will disable channel if too much request failed")
	}

	openai.InitTokenEncoders()
	httpclient.Init(httpclient.ClientConfig{
		UserContentRequestProxy:   config.UserContentRequestProxy,
		UserContentRequestTimeout: config.UserContentRequestTimeout,
		RelayProxy:                config.RelayProxy,
		RelayTimeout:              config.RelayTimeout,
	})

	// Initialize i18n
	if err := i18n.Init(); err != nil {
		log.Fatalln("failed to initialize i18n: " + err.Error())
	}

	// Initialize HTTP server
	ginRouter := gin.New()
	ginRouter.Use(gin.Recovery())
	// This will cause SSE not to work!!!
	// ginRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	ginRouter.Use(middleware.AccessLog())
	ginRouter.Use(middleware.Language())

	// init channel monitor
	channelMonitor := monitor2.NewChannelMonitor(services.UserService, services.ChannelService)
	monitor2.InitMetric(channelMonitor)

	// Initialize handler container with services
	handlerContainer := handlers.NewHandlerContainer(
		services,
		channelMonitor,
		initHandlerParams(cfg),
	)

	// init relay services
	relayServices.Init(services.UserService, services.TokenService, services.LogService, services.ChannelService)

	// init middlewares
	rateLimitConf := middleware.RateLimitConfig{
		GlobalWebRateLimitNum:          config.GlobalWebRateLimitNum,
		GlobalWebRateLimitDuration:     config.GlobalWebRateLimitDuration,
		GlobalApiRateLimitNum:          config.GlobalApiRateLimitNum,
		GlobalApiRateLimitDuration:     config.GlobalApiRateLimitDuration,
		CriticalRateLimitNum:           config.CriticalRateLimitNum,
		CriticalRateLimitDuration:      config.CriticalRateLimitDuration,
		DownloadRateLimitNum:           config.DownloadRateLimitNum,
		DownloadRateLimitDuration:      config.DownloadRateLimitDuration,
		UploadRateLimitNum:             config.UploadRateLimitNum,
		UploadRateLimitDuration:        config.UploadRateLimitDuration,
		RateLimitKeyExpirationDuration: config.RateLimitKeyExpirationDuration,
		DebugEnabled:                   config.DebugEnabled,
		RedisEnabled:                   true, // 使用redis cache
	}
	middlewares := middleware.NewMiddlewares(
		services, redisClient, rateLimitConf, config.TurnstileCheckEnabled, config.TurnstileSecretKey,
	)

	// Create router with handlers
	router.SetRouter(ginRouter, buildFS, handlerContainer, middlewares, cfg.Theme)
	var port = env.Int("PORT", 1337)
	log.Printf("server started on http://localhost:%d", port)

	// 启动服务
	address := fmt.Sprintf("0.0.0.0:%d", port)
	server := &http.Server{
		Handler:           ginRouter,
		Addr:              address,
		ReadHeaderTimeout: 10 * time.Second, // read header timeout
		ReadTimeout:       10 * time.Second, // read request timeout
		WriteTimeout:      30 * time.Second, // write timeout
		IdleTimeout:       20 * time.Second, // tcp idle time
	}

	// 在独立携程中运行
	log.Printf("server listening on %s\n", address)
	go func() {
		if err2 := server.ListenAndServe(); err != nil {
			if !errors.Is(err2, http.ErrServerClosed) {
				log.Println("server close error", map[string]interface{}{
					"trace_error": err2.Error(),
				})
				return
			}

			log.Println("server will exit...")
		}
	}()

	// 等待平滑退出
	shutdown(server, time.Duration(env.Int("GRACEFUL_WAIT", 5))*time.Second)
}

func shutdown(server *http.Server, gracefulWait time.Duration) {
	// server平滑重启
	ch := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// receive signal to exit main goroutine
	// window signal
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// linux signal,please use this in production.
	// signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2, syscall.SIGHUP)

	// Block until we receive our signal.
	sig := <-ch

	log.Println("exit signal: ", sig.String())
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), gracefulWait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// if your application should wait for other services
	// to finalize based on context cancellation.
	done := make(chan struct{}, 1)
	go func() {
		defer close(done)

		err := server.Shutdown(ctx)
		if err != nil {
			log.Println("server shutdown error:", err)
		}
	}()

	select {
	case <-done:
		log.Println("server shutting down")
	case <-ctx.Done():
		log.Println("server ctx timeout")
	}
}

func initAppConfig() *config.AppConfig {
	return &config.AppConfig{
		SystemName:                     config.SystemName,
		ServerAddress:                  config.ServerAddress,
		Footer:                         config.Footer,
		Logo:                           config.Logo,
		TopUpLink:                      config.TopUpLink,
		ChatLink:                       config.ChatLink,
		QuotaPerUnit:                   config.QuotaPerUnit,
		DisplayInCurrencyEnabled:       config.DisplayInCurrencyEnabled,
		DisplayTokenStatEnabled:        config.DisplayTokenStatEnabled,
		ItemsPerPage:                   config.ItemsPerPage,
		MaxRecentItems:                 config.MaxRecentItems,
		PasswordLoginEnabled:           config.PasswordLoginEnabled,
		PasswordRegisterEnabled:        config.PasswordRegisterEnabled,
		EmailVerificationEnabled:       config.EmailVerificationEnabled,
		GitHubOAuthEnabled:             config.GitHubOAuthEnabled,
		OidcEnabled:                    config.OidcEnabled,
		WeChatAuthEnabled:              config.WeChatAuthEnabled,
		TurnstileCheckEnabled:          config.TurnstileCheckEnabled,
		RegisterEnabled:                config.RegisterEnabled,
		EmailDomainRestrictionEnabled:  config.EmailDomainRestrictionEnabled,
		EmailDomainWhitelist:           config.EmailDomainWhitelist,
		DebugEnabled:                   config.DebugEnabled,
		DebugSQLEnabled:                config.DebugSQLEnabled,
		CacheEnabled:                   config.MemoryCacheEnabled,
		LogConsumeEnabled:              config.LogConsumeEnabled,
		SMTPServer:                     config.SMTPServer,
		SMTPPort:                       config.SMTPPort,
		SMTPAccount:                    config.SMTPAccount,
		SMTPFrom:                       config.SMTPFrom,
		SMTPToken:                      config.SMTPToken,
		GitHubClientId:                 config.GitHubClientId,
		GitHubClientSecret:             config.GitHubClientSecret,
		LarkClientId:                   config.LarkClientId,
		LarkClientSecret:               config.LarkClientSecret,
		OidcClientId:                   config.OidcClientId,
		OidcClientSecret:               config.OidcClientSecret,
		OidcWellKnown:                  config.OidcWellKnown,
		OidcAuthorizationEndpoint:      config.OidcAuthorizationEndpoint,
		OidcTokenEndpoint:              config.OidcTokenEndpoint,
		OidcUserinfoEndpoint:           config.OidcUserinfoEndpoint,
		WeChatServerAddress:            config.WeChatServerAddress,
		WeChatServerToken:              config.WeChatServerToken,
		WeChatAccountQRCodeImageURL:    config.WeChatAccountQRCodeImageURL,
		MessagePusherAddress:           config.MessagePusherAddress,
		MessagePusherToken:             config.MessagePusherToken,
		TurnstileSiteKey:               config.TurnstileSiteKey,
		TurnstileSecretKey:             config.TurnstileSecretKey,
		QuotaForNewUser:                config.QuotaForNewUser,
		QuotaForInviter:                config.QuotaForInviter,
		QuotaForInvitee:                config.QuotaForInvitee,
		ChannelDisableThreshold:        config.ChannelDisableThreshold,
		AutomaticDisableChannelEnabled: config.AutomaticDisableChannelEnabled,
		AutomaticEnableChannelEnabled:  config.AutomaticEnableChannelEnabled,
		QuotaRemindThreshold:           config.QuotaRemindThreshold,
		PreConsumedQuota:               config.PreConsumedQuota,
		ApproximateTokenEnabled:        config.ApproximateTokenEnabled,
		RetryTimes:                     config.RetryTimes,
		RootUserEmail:                  config.RootUserEmail,
		RequestInterval:                config.RequestInterval,
		SyncFrequency:                  config.SyncFrequency,
		BatchUpdateEnabled:             config.BatchUpdateEnabled,
		BatchUpdateInterval:            config.BatchUpdateInterval,
		Theme:                          config.Theme,
		ValidThemes:                    config.ValidThemes,
		GlobalWebRateLimitNum:          config.GlobalWebRateLimitNum,
		GlobalWebRateLimitDuration:     config.GlobalWebRateLimitDuration,
		GlobalApiRateLimitNum:          config.GlobalApiRateLimitNum,
		GlobalApiRateLimitDuration:     config.GlobalApiRateLimitDuration,
		CriticalRateLimitNum:           config.CriticalRateLimitNum,
		CriticalRateLimitDuration:      config.CriticalRateLimitDuration,
		DownloadRateLimitNum:           config.DownloadRateLimitNum,
		DownloadRateLimitDuration:      config.DownloadRateLimitDuration,
		UploadRateLimitNum:             config.UploadRateLimitNum,
		UploadRateLimitDuration:        config.UploadRateLimitDuration,
		RateLimitKeyExpirationDuration: config.RateLimitKeyExpirationDuration,
		EnableMetric:                   config.EnableMetric,
		MetricQueueSize:                config.MetricQueueSize,
		MetricSuccessRateThreshold:     config.MetricSuccessRateThreshold,
		MetricSuccessChanSize:          config.MetricSuccessChanSize,
		MetricFailChanSize:             config.MetricFailChanSize,
		RelayTimeout:                   config.RelayTimeout,
		UserContentRequestProxy:        config.UserContentRequestProxy,
		UserContentRequestTimeout:      config.UserContentRequestTimeout,
		RelayProxy:                     config.RelayProxy,
		EnforceIncludeUsage:            config.EnforceIncludeUsage,
		TestPrompt:                     config.TestPrompt,
		InitialRootToken:               config.InitialRootToken,
		InitialRootAccessToken:         config.InitialRootAccessToken,
		GeminiVersion:                  config.GeminiVersion,
		GeminiSafetySetting:            config.GeminiSafetySetting,
	}
}

func initHandlerParams(cfg *config.AppConfig) *handlers.HandlerParams {
	handlerParams := &handlers.HandlerParams{
		LarkUserConfig: handlers.LarkUserConfig{
			LarkClientId:     cfg.LarkClientId,
			LarkClientSecret: cfg.LarkClientSecret,
			ServerAddress:    cfg.ServerAddress,
			RegisterEnabled:  cfg.RegisterEnabled,
		},
		GithubUserConfig: handlers.GitHubUserConfig{
			GitHubClientId:     cfg.GitHubClientId,
			GitHubClientSecret: cfg.GitHubClientSecret,
			GitHubOAuthEnabled: cfg.GitHubOAuthEnabled,
			RegisterEnabled:    cfg.RegisterEnabled,
		},
		AuthConfig: handlers.AuthConfig{
			PasswordLoginEnabled:     cfg.PasswordLoginEnabled,
			PasswordRegisterEnabled:  cfg.PasswordRegisterEnabled,
			RegisterEnabled:          cfg.RegisterEnabled,
			EmailVerificationEnabled: cfg.EmailVerificationEnabled,
		},
		WeChatUserConfig: handlers.WeChatUserConfig{
			WeChatServerAddress: cfg.WeChatServerAddress,
			WeChatServerToken:   cfg.WeChatServerToken,
			WeChatAuthEnabled:   cfg.WeChatAuthEnabled,
			RegisterEnabled:     cfg.RegisterEnabled,
		},
		OidcUserConfig: handlers.OidcUserConfig{
			OidcClientId:         cfg.OidcClientId,
			OidcClientSecret:     cfg.OidcClientSecret,
			ServerAddress:        cfg.ServerAddress,
			OidcTokenEndpoint:    cfg.OidcTokenEndpoint,
			OidcUserinfoEndpoint: cfg.OidcUserinfoEndpoint,
			OidcEnabled:          cfg.OidcEnabled,
			RegisterEnabled:      cfg.RegisterEnabled,
		},
		MiscConfig: handlers.MiscConfig{
			EmailVerificationEnabled:      cfg.EmailVerificationEnabled,
			GitHubOAuthEnabled:            cfg.GitHubOAuthEnabled,
			GitHubClientId:                cfg.GitHubClientId,
			LarkClientId:                  cfg.LarkClientId,
			SystemName:                    cfg.SystemName,
			Logo:                          cfg.Logo,
			Footer:                        cfg.Footer,
			WeChatAccountQRCodeImageURL:   cfg.WeChatAccountQRCodeImageURL,
			WeChatAuthEnabled:             cfg.WeChatAuthEnabled,
			ServerAddress:                 cfg.ServerAddress,
			TurnstileCheckEnabled:         cfg.TurnstileCheckEnabled,
			TurnstileSiteKey:              cfg.TurnstileSiteKey,
			TopUpLink:                     cfg.TopUpLink,
			ChatLink:                      cfg.ChatLink,
			QuotaPerUnit:                  cfg.QuotaPerUnit,
			DisplayInCurrencyEnabled:      cfg.DisplayInCurrencyEnabled,
			OidcEnabled:                   cfg.OidcEnabled,
			OidcClientId:                  cfg.OidcClientId,
			OidcWellKnown:                 cfg.OidcWellKnown,
			OidcAuthorizationEndpoint:     cfg.OidcAuthorizationEndpoint,
			OidcTokenEndpoint:             cfg.OidcTokenEndpoint,
			OidcUserinfoEndpoint:          cfg.OidcUserinfoEndpoint,
			EmailDomainRestrictionEnabled: cfg.EmailDomainRestrictionEnabled,
			EmailDomainWhitelist:          cfg.EmailDomainWhitelist,
			OptionMap:                     cfg.OptionMap,
		},
		ItemsPerPage:                   cfg.ItemsPerPage,
		QuotaPerUnit:                   cfg.QuotaPerUnit,
		DisplayInCurrencyEnabled:       cfg.DisplayInCurrencyEnabled,
		RootUserEmail:                  &cfg.RootUserEmail,
		TestPrompt:                     cfg.TestPrompt,
		ChannelDisableThreshold:        cfg.ChannelDisableThreshold,
		AutomaticDisableChannelEnabled: cfg.AutomaticDisableChannelEnabled,
		RequestInterval:                cfg.RequestInterval,
		DisplayTokenStatEnabled:        cfg.DisplayTokenStatEnabled,
		DebugEnabled:                   cfg.DebugEnabled,
		RetryTimes:                     cfg.RetryTimes,
		OptionMap:                      cfg.OptionMap,
		ValidThemes:                    cfg.ValidThemes,
		GithubClientId:                 cfg.GitHubClientId,
		EmailDomainWhitelist:           cfg.EmailDomainWhitelist,
		WeChatServerAddress:            cfg.WeChatServerAddress,
		TurnstileSiteKey:               cfg.TurnstileSiteKey,
	}

	return handlerParams
}

func CreateRootAccountIfNeed(db *gorm.DB) error {
	var user entity.User
	// if user.Status != util.UserStatusEnabled {
	if err := db.First(&user).Error; err != nil {
		slog.Info("no user exists, creating a root user for you: username is root, password is 123456")
		hashedPassword, err := crypto.Password2Hash("123456")
		if err != nil {
			return err
		}

		accessToken := utils.UUID()
		if config.InitialRootAccessToken != "" {
			accessToken = config.InitialRootAccessToken
		}
		rootUser := entity.User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        entity.RoleRootUser,
			Status:      entity.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: accessToken,
			Quota:       500000000000000,
		}
		db.Create(&rootUser)
		if config.InitialRootToken != "" {
			slog.Info("creating initial root token as requested")
			token := entity.Token{
				Id:             1,
				UserId:         rootUser.Id,
				Key:            config.InitialRootToken,
				Status:         entity.TokenStatusEnabled,
				Name:           "Initial Root Token",
				CreatedTime:    utils.GetTimestamp(),
				AccessedTime:   utils.GetTimestamp(),
				ExpiredTime:    -1,
				RemainQuota:    500000000000000,
				UnlimitedQuota: true,
			}
			db.Create(&token)
		}
	}

	return nil
}
