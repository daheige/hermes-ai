package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/network"
)

type AuthMiddleware struct {
	userService  *application.UserService
	tokenService *application.TokenService
}

func NewAuthMiddleware(userService *application.UserService, tokenService *application.TokenService) *AuthMiddleware {
	return &AuthMiddleware{userService: userService, tokenService: tokenService}
}

func (a *AuthMiddleware) authHelper(c *gin.Context, minRole int) {
	var (
		user *entity.User
		err  error
	)

	// 1. 优先从 Cookie 读取 access_token
	token, _ := c.Cookie("access_token")
	if token != "" {
		user, err = a.userService.ValidateAccessToken(token)
		if err != nil {
			slog.With("request_id", logger.GetRequestID(c.Request.Context())).Info("access token invalid")
		}
	}

	// 2. Cookie 中没有，fallback 到 Authorization header
	if user == nil || user.Username == "" {
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无权进行此操作，未登录且未提供 access token",
			})
			return
		}

		user, err = a.userService.ValidateAccessToken(accessToken)
		if err != nil {
			slog.With("request_id", logger.GetRequestID(c.Request.Context())).Info("access token invalid")
		}

		if user == nil || user.Username == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权进行此操作，access token 无效",
			})
			c.Abort()
			return
		}
	}

	ok, _ := a.userService.IsUserBanned(user.Id)
	if user.Status == entity.UserStatusDisabled || ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户已被封禁",
		})

		c.SetCookie("access_token", "", -1, "/", "", false, true)
		c.Abort()
		return
	}

	if user.Role < minRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权进行此操作，权限不足",
		})
		c.Abort()
		return
	}

	c.Set(ctxkey.Username, user.Username)
	c.Set(ctxkey.Role, user.Role)
	c.Set(ctxkey.Id, user.Id)
	c.Next()
}

func (a *AuthMiddleware) UserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		a.authHelper(c, entity.RoleCommonUser)
	}
}

func (a *AuthMiddleware) AdminAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		a.authHelper(c, entity.RoleAdminUser)
	}
}

func (a *AuthMiddleware) RootAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		a.authHelper(c, entity.RoleRootUser)
	}
}

func (a *AuthMiddleware) TokenAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := c.GetHeader("Authorization")
		key = strings.TrimPrefix(key, "Bearer ")
		if len(key) == 0 {
			key = c.GetHeader("x-api-key")
		}
		key = strings.TrimPrefix(key, "sk-")
		parts := strings.Split(key, "-")
		key = parts[0]
		if len(key) == 0 {
			abortWithMessage(c, http.StatusUnauthorized, "Authorization or x-api-key required")
			return
		}

		token, err := a.tokenService.ValidateUserToken(key)
		if err != nil {
			abortWithMessage(c, http.StatusUnauthorized, err.Error())
			return
		}
		if token.Subnet != nil && *token.Subnet != "" {
			if !network.IsIpInSubnets(ctx, c.ClientIP(), *token.Subnet) {
				abortWithMessage(c, http.StatusForbidden, fmt.Sprintf("该令牌只能在指定网段使用：%s，当前 ip：%s", *token.Subnet, c.ClientIP()))
				return
			}
		}

		userEnabled, _ := a.userService.CacheIsUserEnabled(token.UserId)
		ok, _ := a.userService.IsUserBanned(token.UserId)
		if !userEnabled || ok {
			abortWithMessage(c, http.StatusForbidden, "用户已被封禁")
			return
		}

		requestModel, err := getRequestModel(c)
		if err != nil && shouldCheckModel(c) {
			abortWithMessage(c, http.StatusBadRequest, err.Error())
			return
		}
		c.Set(ctxkey.RequestModel, requestModel)
		if token.Models != nil && *token.Models != "" {
			c.Set(ctxkey.AvailableModels, *token.Models)
			if requestModel != "" && !isModelInList(requestModel, *token.Models) {
				abortWithMessage(c, http.StatusForbidden, fmt.Sprintf("该令牌无权使用模型：%s", requestModel))
				return
			}
		}

		c.Set(ctxkey.Id, token.UserId)
		c.Set(ctxkey.TokenId, token.Id)
		c.Set(ctxkey.TokenName, token.Name)
		if len(parts) > 1 {
			if a.userService.IsAdmin(token.UserId) {
				c.Set(ctxkey.SpecificChannelId, parts[1])
			} else {
				abortWithMessage(c, http.StatusForbidden, "普通用户不支持指定渠道")
				return
			}
		}

		// set channel id for proxy relay
		if channelId := c.Param("channelid"); channelId != "" {
			c.Set(ctxkey.SpecificChannelId, channelId)
		}

		c.Next()
	}
}

func shouldCheckModel(c *gin.Context) bool {
	if strings.HasPrefix(c.Request.URL.Path, "/v1/completions") {
		return true
	}

	if strings.HasPrefix(c.Request.URL.Path, "/v1/chat/completions") {
		return true
	}

	if strings.HasPrefix(c.Request.URL.Path, "/v1/messages") {
		return true
	}

	if strings.HasPrefix(c.Request.URL.Path, "/v1/images") {
		return true
	}

	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		return true
	}

	return false
}
