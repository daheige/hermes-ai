package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/utils"
)

// RedemptionHandler 兑换码处理器
type RedemptionHandler struct {
	service      *application.RedemptionService
	itemsPerPage int
}

// NewRedemptionHandler 创建兑换码处理器
func NewRedemptionHandler(service *application.RedemptionService, itemsPerPage int) *RedemptionHandler {
	return &RedemptionHandler{service: service, itemsPerPage: itemsPerPage}
}

// GetAllRedemptions 获取所有兑换码
func (h *RedemptionHandler) GetAllRedemptions(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	redemptions, err := h.service.GetAllRedemptions(p*h.itemsPerPage, h.itemsPerPage)
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
		"data":    redemptions,
	})
}

// SearchRedemptions 搜索兑换码
func (h *RedemptionHandler) SearchRedemptions(c *gin.Context) {
	keyword := c.Query("keyword")
	redemptions, err := h.service.SearchRedemptions(keyword)
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
		"data":    redemptions,
	})
}

// GetRedemption 根据ID获取兑换码
func (h *RedemptionHandler) GetRedemption(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	redemption, err := h.service.GetRedemptionById(id)
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
		"data":    redemption,
	})
}

// RedemptionCreateRequest 兑换码创建请求
type RedemptionCreateRequest struct {
	Name  string `json:"name" binding:"required"`
	Count int    `json:"count" binding:"required,min=1,max=100"`
	Quota int64  `json:"quota" binding:"required"`
}

// AddRedemption 添加兑换码
func (h *RedemptionHandler) AddRedemption(c *gin.Context) {
	var req RedemptionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if len(req.Name) == 0 || len(req.Name) > 20 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "兑换码名称长度必须在1-20之间",
		})
		return
	}

	var keys []string
	for i := 0; i < req.Count; i++ {
		key := utils.UUID()
		cleanRedemption := entity.Redemption{
			UserId:      c.GetInt(ctxkey.Id),
			Name:        req.Name,
			Key:         key,
			CreatedTime: utils.GetTimestamp(),
			Quota:       req.Quota,
		}
		err := h.service.Insert(&cleanRedemption)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
				"data":    keys,
			})
			return
		}
		keys = append(keys, key)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    keys,
	})
}

// DeleteRedemption 删除兑换码
func (h *RedemptionHandler) DeleteRedemption(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.service.Delete(id)
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

// RedemptionUpdateRequest 兑换码更新请求
type RedemptionUpdateRequest struct {
	ID     int    `json:"id" binding:"required"`
	Name   string `json:"name"`
	Quota  int64  `json:"quota"`
	Status int    `json:"status"`
}

// UpdateRedemption 更新兑换码
func (h *RedemptionHandler) UpdateRedemption(c *gin.Context) {
	statusOnly := c.Query("status_only")

	var req RedemptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	cleanRedemption, err := h.service.GetRedemptionById(req.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if statusOnly != "" {
		cleanRedemption.Status = req.Status
	} else {
		cleanRedemption.Name = req.Name
		cleanRedemption.Quota = req.Quota
	}

	err = h.service.Update(cleanRedemption)
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
		"data":    cleanRedemption,
	})
}
