package application

import (
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/cache"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/persistence"
	"hermes-ai/internal/infras/rds"
)

type Services struct {
	UserService       *UserService
	TokenService      *TokenService
	ChannelService    *ChannelService
	LogService        *LogService
	OptionService     *OptionService
	RedemptionService *RedemptionService
	BatchUpdater      *BatchUpdater
}

// InitServices 初始化所有服务
func InitServices() *Services {
	// 初始化仓储
	userRepo := persistence.NewUserRepo(config.DB)
	tokenRepo := persistence.NewTokenRepo(config.DB)
	channelRepo := persistence.NewChannelRepo(config.DB)
	abilityRepo := persistence.NewAbilityRepo(config.DB)
	logRepo := persistence.NewLogRepo(config.DB)
	optionRepo := persistence.NewOptionRepo(config.DB)
	redemptionRepo := persistence.NewRedemptionRepo(config.DB)

	var cacheRepo repo.CacheRepository
	if rds.RedisEnabled {
		cacheRepo = cache.NewRedisCache()
	} else {
		cacheRepo = cache.NewNoCache()
	}

	// 初始化日志服务（无依赖）
	logService := NewLogService(logRepo, userRepo)

	// 初始化批量更新器
	batchUpdater := NewBatchUpdater(userRepo, tokenRepo, channelRepo, userRepo)

	// 初始化用户服务
	userService := NewUserService(userRepo, tokenRepo, cacheRepo, logService, batchUpdater)

	// 初始化令牌服务
	tokenService := NewTokenService(tokenRepo, userRepo, cacheRepo, batchUpdater)

	// 初始化渠道服务
	channelService := NewChannelService(channelRepo, abilityRepo, cacheRepo, batchUpdater)

	// 初始化选项服务
	optionService := NewOptionService(optionRepo)

	// 初始化兑换码服务
	redemptionService := NewRedemptionService(redemptionRepo, logService)

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
