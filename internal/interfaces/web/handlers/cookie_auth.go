package handlers

import (
	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
)

const authCookieName = "access_token"
const authCookieMaxAge = 30 * 24 * 60 * 60 // 30 days

// SetAuthCookie 设置认证 Cookie
func SetAuthCookie(c *gin.Context, token string) {
	c.SetCookie(authCookieName, token, authCookieMaxAge, "/", "", false, true)
}

// ClearAuthCookie 清除认证 Cookie
func ClearAuthCookie(c *gin.Context) {
	c.SetCookie(authCookieName, "", -1, "/", "", false, true)
}

// GetUserFromAuthCookie 从 Cookie 中读取 access_token 并验证用户
func GetUserFromAuthCookie(c *gin.Context, userService *application.UserService) *entity.User {
	token, err := c.Cookie(authCookieName)
	if err != nil || token == "" {
		return nil
	}

	return userService.ValidateAccessToken(token)
}
