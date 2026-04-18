package middleware

import (
	"hermes-ai/internal/application"
)

type Middlewares struct {
	AuthMiddleware        *AuthMiddleware
	DistributorMiddleware *DistributorMiddleware
}

// NewMiddlewares 创建middleware
func NewMiddlewares(userService *application.UserService, tokenService *application.TokenService,
	channelService *application.ChannelService) *Middlewares {
	return &Middlewares{
		AuthMiddleware:        NewAuthMiddleware(userService, tokenService),
		DistributorMiddleware: NewDistributorMiddleware(userService, channelService),
	}
}
