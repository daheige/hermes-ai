package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/network"
	"hermes-ai/internal/infras/utils"
)

// TokenHandler 令牌处理器
type TokenHandler struct {
	service      *application.TokenService
	itemsPerPage int
}

// NewTokenHandler 创建令牌处理器
func NewTokenHandler(service *application.TokenService, itemsPerPage int) *TokenHandler {
	return &TokenHandler{service: service, itemsPerPage: itemsPerPage}
}

// GetAllTokens 获取用户所有令牌
func (h *TokenHandler) GetAllTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.Query("order")
	tokens, err := h.service.GetAllUserTokens(userId, p*h.itemsPerPage, h.itemsPerPage, order)

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
		"data":    tokens,
	})
}

// SearchTokens 搜索用户令牌
func (h *TokenHandler) SearchTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	keyword := c.Query("keyword")
	tokens, err := h.service.SearchUserTokens(userId, keyword)
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
		"data":    tokens,
	})
}

// GetToken 根据ID获取令牌
func (h *TokenHandler) GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	token, err := h.service.GetTokenByIds(id, userId)
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
		"data":    token,
	})
}

// GetTokenStatus 获取令牌状态
func (h *TokenHandler) GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt(ctxkey.TokenId)
	userId := c.GetInt(ctxkey.Id)
	token, err := h.service.GetTokenByIds(tokenId, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

// TokenCreateRequest 令牌创建请求
type TokenCreateRequest struct {
	Name           string  `json:"name" binding:"required,max=30"`
	ExpiredTime    int64   `json:"expired_time"`
	RemainQuota    int64   `json:"remain_quota"`
	UnlimitedQuota bool    `json:"unlimited_quota"`
	Models         *string `json:"models"`
	Subnet         *string `json:"subnet"`
}

func (h *TokenHandler) validateToken(c *gin.Context, req TokenCreateRequest) error {
	if len(req.Name) > 30 {
		return fmt.Errorf("令牌名称过长")
	}
	if req.Subnet != nil && *req.Subnet != "" {
		err := network.IsValidSubnets(*req.Subnet)
		if err != nil {
			return fmt.Errorf("无效的网段：%s", err.Error())
		}
	}
	return nil
}

// AddToken 添加令牌
func (h *TokenHandler) AddToken(c *gin.Context) {
	var req TokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	err := h.validateToken(c, req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}

	cleanToken := entity.Token{
		UserId:         c.GetInt(ctxkey.Id),
		Name:           req.Name,
		Key:            utils.GenerateKey(),
		CreatedTime:    utils.GetTimestamp(),
		AccessedTime:   utils.GetTimestamp(),
		ExpiredTime:    req.ExpiredTime,
		RemainQuota:    req.RemainQuota,
		UnlimitedQuota: req.UnlimitedQuota,
		Models:         req.Models,
		Subnet:         req.Subnet,
	}
	err = h.service.Insert(&cleanToken)
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
		"data":    cleanToken,
	})
}

// DeleteToken 删除令牌
func (h *TokenHandler) DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	err := h.service.Delete(id, userId)
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

// TokenUpdateRequest 令牌更新请求
type TokenUpdateRequest struct {
	ID             int     `json:"id" binding:"required"`
	Name           string  `json:"name"`
	ExpiredTime    int64   `json:"expired_time"`
	RemainQuota    int64   `json:"remain_quota"`
	UnlimitedQuota bool    `json:"unlimited_quota"`
	Models         *string `json:"models"`
	Subnet         *string `json:"subnet"`
	Status         int     `json:"status"`
}

// UpdateToken 更新令牌
func (h *TokenHandler) UpdateToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	statusOnly := c.Query("status_only")

	var req TokenUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	err := h.validateToken(c, TokenCreateRequest{Name: req.Name, Subnet: req.Subnet})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}

	cleanToken, err := h.service.GetTokenByIds(req.ID, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if req.Status == entity.TokenStatusEnabled {
		if cleanToken.Status == entity.TokenStatusExpired && cleanToken.ExpiredTime <= utils.GetTimestamp() && cleanToken.ExpiredTime != -1 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期",
			})
			return
		}
		if cleanToken.Status == entity.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度",
			})
			return
		}
	}

	if statusOnly != "" {
		cleanToken.Status = req.Status
	} else {
		cleanToken.Name = req.Name
		cleanToken.ExpiredTime = req.ExpiredTime
		cleanToken.RemainQuota = req.RemainQuota
		cleanToken.UnlimitedQuota = req.UnlimitedQuota
		cleanToken.Models = req.Models
		cleanToken.Subnet = req.Subnet
	}

	err = h.service.Update(cleanToken)
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
		"data":    cleanToken,
	})
}
