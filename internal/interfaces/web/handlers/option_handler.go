package handlers

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/i18n"
)

// OptionHandler 配置选项处理器
type OptionHandler struct {
	service          *application.OptionService
	optionMapRWMutex sync.RWMutex
	OptionConfig
}

type OptionConfig struct {
	// todo 这个 OptionMap 可能没有实时更新
	OptionMap            map[string]string
	ValidThemes          map[string]bool
	GithubClientId       string
	EmailDomainWhitelist []string
	WeChatServerAddress  string
	TurnstileSiteKey     string
}

// NewOptionHandler 创建配置选项处理器
func NewOptionHandler(service *application.OptionService, conf OptionConfig) *OptionHandler {
	return &OptionHandler{
		service:      service,
		OptionConfig: conf,
	}
}

// GetOptions 获取所有配置选项
func (h *OptionHandler) GetOptions(c *gin.Context) {
	var options []*entity.Option
	h.optionMapRWMutex.Lock()
	for k, v := range h.OptionMap {
		if strings.HasSuffix(k, "Token") || strings.HasSuffix(k, "Secret") {
			continue
		}

		options = append(options, &entity.Option{
			Key:   k,
			Value: v,
		})
	}

	h.optionMapRWMutex.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
}

// OptionUpdateRequest 配置更新请求
type OptionUpdateRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

// UpdateOption 更新配置选项
func (h *OptionHandler) UpdateOption(c *gin.Context) {
	var req OptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	switch req.Key {
	case "Theme":
		if !h.ValidThemes[req.Value] {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无效的主题",
			})
			return
		}
	case "GitHubOAuthEnabled":
		if req.Value == "true" && h.GithubClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 GitHub OAuth，请先填入 GitHub Client Id 以及 GitHub Client Secret！",
			})
			return
		}
	case "EmailDomainRestrictionEnabled":
		if req.Value == "true" && len(h.EmailDomainWhitelist) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用邮箱域名限制，请先填入限制的邮箱域名！",
			})
			return
		}
	case "WeChatAuthEnabled":
		if req.Value == "true" && h.WeChatServerAddress == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用微信登录，请先填入微信登录相关配置信息！",
			})
			return
		}
	case "TurnstileCheckEnabled":
		if req.Value == "true" && h.TurnstileSiteKey == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Turnstile 校验，请先填入 Turnstile 校验相关配置信息！",
			})
			return
		}
	}

	err := h.service.UpdateOption(req.Key, req.Value)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
