package middleware

import (
	"github.com/redis/go-redis/v9"

	"hermes-ai/internal/providers"
)

type Middlewares struct {
	AuthMiddleware        *AuthMiddleware
	DistributorMiddleware *DistributorMiddleware
	RateLimitMiddleware   *RateLimitMiddleware
	TurnstileMiddleware   *TurnstileMiddleware
}

// NewMiddlewares 创建middleware
func NewMiddlewares(
	services *providers.Services,
	rdb redis.UniversalClient,
	conf RateLimitConfig,
	turnstileCheckEnabled bool,
	turnstileSecretKey string,
) *Middlewares {
	return &Middlewares{
		AuthMiddleware:        NewAuthMiddleware(services.UserService, services.TokenService),
		DistributorMiddleware: NewDistributorMiddleware(services.UserService, services.ChannelService),
		RateLimitMiddleware:   NewRateLimitMiddleware(rdb, conf),
		TurnstileMiddleware:   NewTurnstileMiddleware(turnstileCheckEnabled, turnstileSecretKey),
	}
}
