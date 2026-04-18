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

	"hermes-ai/internal/application"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/env"
	"hermes-ai/internal/infras/httpclient"
	"hermes-ai/internal/infras/i18n"
	"hermes-ai/internal/infras/logger"
	monitor2 "hermes-ai/internal/infras/monitor"
	"hermes-ai/internal/infras/rds"
	"hermes-ai/internal/infras/relay/adaptor/openai"
	relayServices "hermes-ai/internal/infras/relay/services"
	"hermes-ai/internal/interfaces/web/handlers"
	"hermes-ai/internal/interfaces/web/middleware"
	"hermes-ai/internal/interfaces/web/router"
	"hermes-ai/internal/providers/web"
)

//go:embed web/build/*
var buildFS embed.FS

func main() {
	web.Init()

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
	config.InitDB()
	config.InitLogDB()

	var err error
	err = web.CreateRootAccountIfNeed()
	if err != nil {
		log.Fatalln("database init error: " + err.Error())
	}
	defer func() {
		err := config.CloseDB()
		if err != nil {
			log.Fatalln("failed to close database: " + err.Error())
		}
	}()

	// Initialize Redis
	err = rds.InitRedisClient()
	if err != nil {
		log.Fatalln("failed to initialize Redis: " + err.Error())
	}

	// Initialize application services
	services := application.InitServices()

	// Initialize options
	services.OptionService.InitOptionMap()
	slog.Info(fmt.Sprintf("using theme %s", config.Theme))
	if rds.RedisEnabled {
		// for compatibility with old versions
		config.MemoryCacheEnabled = true
	}

	if config.MemoryCacheEnabled {
		slog.Info("memory cache enabled")
		slog.Info(fmt.Sprintf("sync frequency: %d seconds", config.SyncFrequency))
		services.ChannelService.InitChannelCache()
	}

	// 内存缓存对于redis也是
	if config.MemoryCacheEnabled {
		go services.OptionService.SyncOptions(config.SyncFrequency)
		go services.ChannelService.SyncChannelCache(config.SyncFrequency)
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
		config.BatchUpdateEnabled = true
		slog.Info("batch update enabled with interval " + strconv.Itoa(config.BatchUpdateInterval) + "s")
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
		services, channelMonitor,
	)

	// init relay services
	relayServices.Init(services.UserService, services.TokenService, services.LogService, services.ChannelService)

	// init middlewares
	middlewares := middleware.NewMiddlewares(services.UserService, services.TokenService, services.ChannelService)
	// Create router with handlers
	router.SetRouter(ginRouter, buildFS, handlerContainer, middlewares)
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
