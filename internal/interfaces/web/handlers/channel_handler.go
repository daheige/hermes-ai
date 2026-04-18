package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/utils"
)

// ChannelHandler 渠道处理器
type ChannelHandler struct {
	service *application.ChannelService
}

// NewChannelHandler 创建渠道处理器
func NewChannelHandler(service *application.ChannelService) *ChannelHandler {
	return &ChannelHandler{service: service}
}

// GetAllChannels 获取所有渠道
func (h *ChannelHandler) GetAllChannels(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	channels, err := h.service.GetAllChannels(p*config.ItemsPerPage, config.ItemsPerPage, "limited")
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
		"data":    channels,
	})
}

// SearchChannels 搜索渠道
func (h *ChannelHandler) SearchChannels(c *gin.Context) {
	keyword := c.Query("keyword")
	channels, err := h.service.SearchChannels(keyword)
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
		"data":    channels,
	})
}

// GetChannel 根据ID获取渠道
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	channel, err := h.service.GetChannelById(id, false)
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
		"data":    channel,
	})
}

// ChannelCreateRequest 渠道创建请求
type ChannelCreateRequest struct {
	Type         int     `json:"type" binding:"required"`
	Key          string  `json:"key" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	BaseURL      string  `json:"base_url"`
	Models       string  `json:"models"`
	Group        string  `json:"group"`
	ModelMapping *string `json:"model_mapping"`
	Status       int     `json:"status"`
}

// AddChannel 添加渠道
func (h *ChannelHandler) AddChannel(c *gin.Context) {
	var req ChannelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	keys := strings.Split(req.Key, "\n")
	channels := make([]entity.Channel, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		channel := entity.Channel{
			Type:         req.Type,
			Key:          key,
			Name:         req.Name,
			Models:       req.Models,
			Group:        req.Group,
			ModelMapping: req.ModelMapping,
			Status:       req.Status,
			CreatedTime:  utils.GetTimestamp(),
		}
		if req.BaseURL != "" {
			channel.BaseURL = &req.BaseURL
		}
		channels = append(channels, channel)
	}
	err := h.service.BatchInsertChannels(channels)
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

// DeleteChannel 删除渠道
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
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

// DeleteDisabledChannel 删除禁用渠道
func (h *ChannelHandler) DeleteDisabledChannel(c *gin.Context) {
	rows, err := h.service.DeleteDisabledChannel()
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
		"data":    rows,
	})
}

// ChannelUpdateRequest 渠道更新请求
type ChannelUpdateRequest struct {
	ID           int     `json:"id" binding:"required"`
	Type         int     `json:"type"`
	Key          string  `json:"key"`
	Name         string  `json:"name"`
	BaseURL      string  `json:"base_url"`
	Models       string  `json:"models"`
	Group        string  `json:"group"`
	ModelMapping *string `json:"model_mapping"`
	Status       int     `json:"status"`
}

// UpdateChannel 更新渠道
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	var req ChannelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	channel := &entity.Channel{
		Id:           req.ID,
		Type:         req.Type,
		Key:          req.Key,
		Name:         req.Name,
		Models:       req.Models,
		Group:        req.Group,
		ModelMapping: req.ModelMapping,
		Status:       req.Status,
	}
	if req.BaseURL != "" {
		channel.BaseURL = &req.BaseURL
	}

	err := h.service.Update(channel)
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
		"data":    channel,
	})
}
