package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	billingratio "hermes-ai/internal/infras/relay/billing/ratio"
)

// GroupHandler 分组处理器
type GroupHandler struct {
}

// NewGroupHandler 创建分组处理器
func NewGroupHandler() *GroupHandler {
	return &GroupHandler{}
}

// GetGroups 获取所有分组
func (h *GroupHandler) GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName := range billingratio.GroupRatio {
		groupNames = append(groupNames, groupName)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}
