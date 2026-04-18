package handlers

import (
	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/ctxkey"
	relaymodel "hermes-ai/internal/infras/relay/model"
)

// BillingHandler 账单处理器
type BillingHandler struct {
	userService  *application.UserService
	tokenService *application.TokenService
}

// NewBillingHandler 创建账单处理器
func NewBillingHandler(userService *application.UserService, tokenService *application.TokenService) *BillingHandler {
	return &BillingHandler{
		userService:  userService,
		tokenService: tokenService,
	}
}

// GetSubscription 获取订阅信息
func (h *BillingHandler) GetSubscription(c *gin.Context) {
	var remainQuota int64
	var usedQuota int64
	var err error
	var token *entity.Token
	var expiredTime int64
	if config.DisplayTokenStatEnabled {
		tokenId := c.GetInt(ctxkey.TokenId)
		token, err = h.tokenService.GetTokenById(tokenId)
		if err == nil {
			expiredTime = token.ExpiredTime
			remainQuota = token.RemainQuota
			usedQuota = token.UsedQuota
		}
	} else {
		userId := c.GetInt(ctxkey.Id)
		remainQuota, err = h.userService.GetUserQuota(userId)
		if err != nil {
			usedQuota, err = h.userService.GetUserUsedQuota(userId)
		}
	}
	if expiredTime <= 0 {
		expiredTime = 0
	}
	if err != nil {
		Error := relaymodel.Error{
			Message: err.Error(),
			Type:    "upstream_error",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
		return
	}
	quota := remainQuota + usedQuota
	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}
	if token != nil && token.UnlimitedQuota {
		amount = 100000000
	}
	subscription := OpenAISubscriptionResponse{
		Object:             "billing_subscription",
		HasPaymentMethod:   true,
		SoftLimitUSD:       amount,
		HardLimitUSD:       amount,
		SystemHardLimitUSD: amount,
		AccessUntil:        expiredTime,
	}
	c.JSON(200, subscription)
}

// GetUsage 获取使用情况
func (h *BillingHandler) GetUsage(c *gin.Context) {
	var quota int64
	var err error
	var token *entity.Token
	if config.DisplayTokenStatEnabled {
		tokenId := c.GetInt(ctxkey.TokenId)
		token, err = h.tokenService.GetTokenById(tokenId)
		quota = token.UsedQuota
	} else {
		userId := c.GetInt(ctxkey.Id)
		quota, err = h.userService.GetUserUsedQuota(userId)
	}
	if err != nil {
		Error := relaymodel.Error{
			Message: err.Error(),
			Type:    "one_api_error",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
		return
	}
	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}
	usage := OpenAIUsageResponse{
		Object:     "list",
		TotalUsage: amount * 100,
	}
	c.JSON(200, usage)
}
