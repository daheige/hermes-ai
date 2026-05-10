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
	"hermes-ai/internal/infras/httpclient"
	"hermes-ai/internal/infras/i18n"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/message"
	monitor2 "hermes-ai/internal/infras/monitor"
	"hermes-ai/internal/infras/relay"
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
	// 显式初始化系统配置
	sysCfg := config.InitSystemConfig()

	// 初始化日志
	opts := []logger.Option{
		logger.WithAddSource(true),
		logger.WithEnableJSON(),
		logger.WithLevel(logger.GetLevel(sysCfg.LogLevel)),
	}

	if sysCfg.LogDir != "" {
		opts = append(opts, logger.WithLogDir(sysCfg.LogDir), logger.WithOutputToFile(true))
	}

	logger.Default(opts...)

	// 初始化加密的aes key
	crypto.InitAesKey(sysCfg.AESSecretKey)

	if sysCfg.GinMode != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	slog.Debug("running in debug mode")

	// Initialize SQL Database
	dbConf := config.DBConfig{
		DSN:          sysCfg.SQLDSN,
		LogDSN:       sysCfg.LogSQLDSN,
		DebugSQL:     sysCfg.DebugSQLEnabled,
		MaxIdleConns: sysCfg.SQLMaxIdleConns,
		MaxOpenConns: sysCfg.SQLMaxOpenConns,
		MaxLifetime:  sysCfg.SQLMaxLifetime,
	}
	db, logDB := config.InitDatabase(dbConf)
	defer func() {
		err := config.CloseDB(db)
		if err != nil {
			log.Fatalln("failed to close database: " + err.Error())
		}

		if sysCfg.LogSQLDSN != "" {
			err := config.CloseDB(logDB)
			if err != nil {
				log.Fatalln("failed to close database: " + err.Error())
			}
		}
	}()

	err := CreateRootAccountIfNeed(db, sysCfg)
	if err != nil {
		log.Fatalln("database init error: " + err.Error())
	}

	// Initialize Redis
	redisClient, err := config.InitRedisClient(config.RedisConfig{
		ConnString:    sysCfg.RedisConnString,
		EnableCluster: sysCfg.RedisEnableCluster,
		Password:      sysCfg.RedisPassword,
		Username:      sysCfg.RedisUsername,
	})
	if err != nil {
		log.Fatalln("failed to initialize Redis: " + err.Error())
	}

	// 使用 redis cache
	sysCfg.CacheEnabled = true

	// init repos
	repos := providers.InitRepositories(db, logDB, redisClient, providers.BatchUpdaterConfig{
		BatchInterval:      time.Duration(sysCfg.BatchUpdateInterval) * time.Second,
		BatchUpdateEnabled: sysCfg.BatchUpdateEnabled,
	})
	// Initialize application services
	services := providers.InitServices(repos, sysCfg)

	// Initialize options
	services.OptionService.InitOptionMap()
	sysCfg.RootUserEmail = services.UserService.GetRootUserEmail()
	slog.Info(fmt.Sprintf("using theme %s", sysCfg.Theme))

	// 内存缓存对于redis也是
	if sysCfg.CacheEnabled {
		slog.Info("sync option and channel from database")
		defer services.ChannelService.Stop()
		defer services.OptionService.Stop()

		go services.OptionService.SyncOptions(10 * time.Second)
		go services.ChannelService.SyncChannelCache(10 * time.Second)
	}

	if sysCfg.ChannelTestFrequency != 0 {
		go handlers.AutomaticallyTestChannels(sysCfg.ChannelTestFrequency)
	}

	// 启动批量更新
	if sysCfg.BatchUpdateEnabled {
		slog.Info("batch update enabled with interval " + strconv.Itoa(sysCfg.BatchUpdateInterval) + "s")
		repos.BatchUpdater.Start()
		defer repos.BatchUpdater.Stop()
	}

	if sysCfg.EnableMetric {
		slog.Info("metric enabled, will disable channel if too much request failed")
	}

	tokenCounter := openai.NewTokenCounter(sysCfg.ApproximateTokenEnabled)
	adaptorFactory := relay.NewAdaptorFactory(tokenCounter)
	httpclient.Init(httpclient.ClientConfig{
		UserContentRequestProxy:   sysCfg.UserContentRequestProxy,
		UserContentRequestTimeout: sysCfg.UserContentRequestTimeout,
		RelayProxy:                sysCfg.RelayProxy,
		RelayTimeout:              sysCfg.RelayTimeout,
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
	channelMonitor := monitor2.NewChannelMonitor(monitor2.ChannelMonitorDeps{
		UserService:    services.UserService,
		ChannelService: services.ChannelService,
	}, monitor2.ChannelMonitorConfig{
		SmtpCfg: message.SMTPConfig{
			Server:  sysCfg.SMTPServer,
			Port:    sysCfg.SMTPPort,
			Account: sysCfg.SMTPAccount,
			From:    sysCfg.SMTPFrom,
			Token:   sysCfg.SMTPToken,
		},
		PusherCfg: message.MessagePusherConfig{
			Address: sysCfg.MessagePusherAddress,
			Token:   sysCfg.MessagePusherToken,
		},
		SystemName: sysCfg.SystemName,
		RootEmail:  sysCfg.RootUserEmail,
	})
	var metricCollector *monitor2.MetricCollector
	if sysCfg.EnableMetric {
		metricCollector = monitor2.NewMetricCollector(channelMonitor, sysCfg.MetricQueueSize, sysCfg.MetricSuccessRateThreshold, sysCfg.MetricSuccessChanSize, sysCfg.MetricFailChanSize)
	}

	// Initialize handler container with services
	handlerContainer := handlers.NewHandlerContainer(
		services,
		channelMonitor,
		adaptorFactory,
		metricCollector,
		initHandlerParams(sysCfg),
	)

	// init relay services
	relayServices.Init(services.UserService, services.TokenService, services.LogService, services.ChannelService)

	// init middlewares
	rateLimitConf := middleware.RateLimitConfig{
		GlobalWebRateLimitNum:          sysCfg.GlobalWebRateLimitNum,
		GlobalWebRateLimitDuration:     sysCfg.GlobalWebRateLimitDuration,
		GlobalApiRateLimitNum:          sysCfg.GlobalApiRateLimitNum,
		GlobalApiRateLimitDuration:     sysCfg.GlobalApiRateLimitDuration,
		CriticalRateLimitNum:           sysCfg.CriticalRateLimitNum,
		CriticalRateLimitDuration:      sysCfg.CriticalRateLimitDuration,
		DownloadRateLimitNum:           sysCfg.DownloadRateLimitNum,
		DownloadRateLimitDuration:      sysCfg.DownloadRateLimitDuration,
		UploadRateLimitNum:             sysCfg.UploadRateLimitNum,
		UploadRateLimitDuration:        sysCfg.UploadRateLimitDuration,
		RateLimitKeyExpirationDuration: sysCfg.RateLimitKeyExpirationDuration,
		DebugEnabled:                   sysCfg.DebugEnabled,
		RedisEnabled:                   true, // 使用redis cache
	}
	middlewares := middleware.NewMiddlewares(
		services, redisClient, rateLimitConf, sysCfg.TurnstileCheckEnabled, sysCfg.TurnstileSecretKey,
	)

	// Create router with handlers
	routerConfig := &router.RouterConfig{
		BuildFS:         buildFS,
		Hc:              handlerContainer,
		Middlewares:     middlewares,
		Theme:           sysCfg.Theme,
		FrontendBaseUrl: sysCfg.FrontendBaseURL,
	}
	router.SetRouter(ginRouter, routerConfig)
	log.Printf("server started on http://localhost:%d", sysCfg.Port)

	// 启动服务
	address := fmt.Sprintf("0.0.0.0:%d", sysCfg.Port)
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
	shutdown(server, time.Duration(sysCfg.GracefulWait)*time.Second)
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

func initHandlerParams(cfg *config.SystemConfig) *handlers.HandlerParams {
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
			SMTPConfig: message.SMTPConfig{
				Server:  cfg.SMTPServer,
				Port:    cfg.SMTPPort,
				Account: cfg.SMTPAccount,
				From:    cfg.SMTPFrom,
				Token:   cfg.SMTPToken,
			},
			MessagePusherConfig: message.MessagePusherConfig{
				Address: cfg.MessagePusherAddress,
				Token:   cfg.MessagePusherToken,
			},
		},
		ItemsPerPage:                   cfg.ItemsPerPage,
		QuotaPerUnit:                   cfg.QuotaPerUnit,
		DisplayInCurrencyEnabled:       cfg.DisplayInCurrencyEnabled,
		TestPrompt:                     cfg.TestPrompt,
		ChannelDisableThreshold:        cfg.ChannelDisableThreshold,
		AutomaticDisableChannelEnabled: cfg.AutomaticDisableChannelEnabled,
		RequestInterval:                cfg.RequestInterval,
		DisplayTokenStatEnabled:        cfg.DisplayTokenStatEnabled,
		RetryTimes:                     cfg.RetryTimes,
		PreConsumedQuota:               cfg.PreConsumedQuota,
		EnforceIncludeUsage:            cfg.EnforceIncludeUsage,
		EnableMetric:                   cfg.EnableMetric,
		ValidThemes:                    cfg.ValidThemes,
		GithubClientId:                 cfg.GitHubClientId,
		EmailDomainWhitelist:           cfg.EmailDomainWhitelist,
		WeChatServerAddress:            cfg.WeChatServerAddress,
		TurnstileSiteKey:               cfg.TurnstileSiteKey,
	}

	return handlerParams
}

func CreateRootAccountIfNeed(db *gorm.DB, sysCfg *config.SystemConfig) error {
	var user entity.User
	// if user.Status != util.UserStatusEnabled {
	if err := db.First(&user).Error; err != nil {
		slog.Info("no user exists, creating a root user for you: username is root, password is 123456")
		hashedPassword, err := crypto.Password2Hash("123456")
		if err != nil {
			return err
		}

		accessToken := utils.UUID()
		if sysCfg.InitialRootAccessToken != "" {
			accessToken = sysCfg.InitialRootAccessToken
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
		if sysCfg.InitialRootToken != "" {
			slog.Info("creating initial root token as requested")
			encryptedKey, err := crypto.Encrypt(sysCfg.InitialRootToken)
			if err != nil {
				return fmt.Errorf("failed to encrypt initial root token: %w", err)
			}
			token := entity.Token{
				Id:             1,
				UserId:         rootUser.Id,
				Key:            encryptedKey,
				KeyHash:        crypto.KeyHash(sysCfg.InitialRootToken),
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
