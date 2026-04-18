package providers

import (
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/cache"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/persistence"
)

// Services 服务列表
type Services struct {
	UserService       *application.UserService
	TokenService      *application.TokenService
	ChannelService    *application.ChannelService
	LogService        *application.LogService
	OptionService     *application.OptionService
	RedemptionService *application.RedemptionService
	BatchUpdater      *application.BatchUpdater
}

// Repositories 资源列表
type Repositories struct {
	UserRepo       repo.UserRepository
	TokenRepo      repo.TokenRepository
	ChannelRepo    repo.ChannelRepository
	LogRepo        repo.LogRepository
	OptionRepo     repo.OptionRepository
	RedemptionRepo repo.RedemptionRepository
	CacheRepo      repo.CacheRepository
	AbilityRepo    repo.AbilityRepository
}

// InitRepositories 初始化资源池列表
func InitRepositories(db, logDB *gorm.DB, redisClient redis.UniversalClient) *Repositories {
	userRepo := persistence.NewUserRepo(db)
	tokenRepo := persistence.NewTokenRepo(db)
	channelRepo := persistence.NewChannelRepo(db)
	abilityRepo := persistence.NewAbilityRepo(db)
	logRepo := persistence.NewLogRepo(logDB)
	optionRepo := persistence.NewOptionRepo(db)
	redemptionRepo := persistence.NewRedemptionRepo(db)
	cacheRepo := cache.NewRedisCache(redisClient)

	repos := &Repositories{
		UserRepo:       userRepo,
		TokenRepo:      tokenRepo,
		ChannelRepo:    channelRepo,
		LogRepo:        logRepo,
		OptionRepo:     optionRepo,
		RedemptionRepo: redemptionRepo,
		CacheRepo:      cacheRepo,
		AbilityRepo:    abilityRepo,
	}

	return repos
}

// InitServices 初始化所有服务
func InitServices(repos *Repositories, cfg *config.AppConfig) *Services {
	// 初始化日志服务（无依赖）
	logService := application.NewLogService(repos.LogRepo, repos.UserRepo, cfg.LogConsumeEnabled, cfg.MaxRecentItems)

	// 初始化批量更新器
	batchUpdater := application.NewBatchUpdater(
		repos.UserRepo, repos.TokenRepo, repos.ChannelRepo, repos.UserRepo,
		time.Duration(cfg.BatchUpdateInterval)*time.Second, cfg.BatchUpdateEnabled,
	)

	// 初始化用户服务
	userService := application.NewUserService(
		repos.UserRepo, repos.TokenRepo, repos.CacheRepo, logService, batchUpdater,
		application.UserConfig{
			SyncFrequency:            cfg.SyncFrequency,
			QuotaForNewUser:          cfg.QuotaForNewUser,
			QuotaForInvitee:          cfg.QuotaForInvitee,
			QuotaForInviter:          cfg.QuotaForInviter,
			QuotaPerUnit:             cfg.QuotaPerUnit,
			DisplayInCurrencyEnabled: cfg.DisplayInCurrencyEnabled,
			PreConsumedQuota:         cfg.PreConsumedQuota,
			BatchUpdateEnabled:       cfg.BatchUpdateEnabled,
		},
	)

	// 初始化令牌服务
	tokenService := application.NewTokenService(
		repos.TokenRepo, repos.UserRepo, repos.CacheRepo, batchUpdater,
		application.TokenConfig{
			SyncFrequency:        cfg.SyncFrequency,
			BatchUpdateEnabled:   cfg.BatchUpdateEnabled,
			QuotaRemindThreshold: cfg.QuotaRemindThreshold,
			ServerAddress:        cfg.ServerAddress,
		},
	)

	// 初始化渠道服务
	channelService := application.NewChannelService(
		repos.ChannelRepo, repos.AbilityRepo, repos.CacheRepo, batchUpdater,
		cfg.BatchUpdateEnabled,
		cfg.SyncFrequency,
		cfg.CacheEnabled,
	)

	// 初始化选项服务
	optionService := application.NewOptionService(repos.OptionRepo, cfg)

	// 初始化兑换码服务
	redemptionService := application.NewRedemptionService(
		repos.RedemptionRepo, logService,
		cfg.QuotaPerUnit,
		cfg.DisplayInCurrencyEnabled,
	)

	return &Services{
		UserService:       userService,
		TokenService:      tokenService,
		ChannelService:    channelService,
		LogService:        logService,
		OptionService:     optionService,
		RedemptionService: redemptionService,
		BatchUpdater:      batchUpdater,
	}
}
