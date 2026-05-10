package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/logger"
)

const (
	authCookieName   = "access_token"
	authCookieMaxAge = 30 * 24 * 60 * 60 // 30 days
)

// setAuthCookie 设置认证 Cookie
func setAuthCookie(c *gin.Context, token string) {
	c.SetCookie(authCookieName, token, authCookieMaxAge, "/", "", false, true)
}

// clearAuthCookie 清除认证 Cookie
func clearAuthCookie(c *gin.Context) {
	c.SetCookie(authCookieName, "", -1, "/", "", false, true)
}

// GetUserFromAuthCookie 从 Cookie 中读取 access_token 并验证用户
func GetUserFromAuthCookie(c *gin.Context, userService *application.UserService) *entity.User {
	token, err := c.Cookie(authCookieName)
	if err != nil || token == "" {
		return nil
	}

	user, err := userService.ValidateAccessToken(token)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(c.Request.Context())).
			Error("failed to validate access token", "error", err)
		return nil
	}

	return user
}
