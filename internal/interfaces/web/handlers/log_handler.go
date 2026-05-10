package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/ctxkey"
)

// LogHandler 日志处理器
type LogHandler struct {
	service      *application.LogService
	itemsPerPage int
}

// NewLogHandler 创建日志处理器
func NewLogHandler(service *application.LogService, itemsPerPage int) *LogHandler {
	return &LogHandler{service: service, itemsPerPage: itemsPerPage}
}

// AllLogsRequest 所有日志请求结构体
type AllLogsRequest struct {
	Page           int    `json:"p" form:"p"`
	LogType        int    `json:"type" form:"type"`
	StartTimestamp int64  `json:"start_timestamp" form:"start_timestamp"`
	EndTimestamp   int64  `json:"end_timestamp" form:"end_timestamp"`
	Username       string `json:"username" form:"username"`
	TokenName      string `json:"token_name" form:"token_name"`
	ModelName      string `json:"model_name" form:"model_name"`
	Channel        int    `json:"channel" form:"channel"`
}

// GetAllLogs 获取所有日志
func (h *LogHandler) GetAllLogs(c *gin.Context) {
	req := &AllLogsRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := application.AllLogsParams{
		Limit:          h.itemsPerPage,
		Offset:         req.Page * h.itemsPerPage,
		LogType:        req.LogType,
		StartTimestamp: req.StartTimestamp,
		EndTimestamp:   req.EndTimestamp,
		Username:       req.Username,
		TokenName:      req.TokenName,
		ModelName:      req.ModelName,
		Channel:        req.Channel,
	}
	logs, err := h.service.GetAllLogs(p)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if len(logs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
			"data":    []*entity.Log{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data":    logs,
	})
}

// UserLogsRequest 用户日志request
type UserLogsRequest struct {
	Page           int    `json:"p" form:"p"`
	LogType        int    `json:"type" form:"type"`
	StartTimestamp int64  `json:"start_timestamp" form:"start_timestamp"`
	EndTimestamp   int64  `json:"end_timestamp" form:"end_timestamp"`
	TokenName      string `json:"token_name" form:"token_name"`
	ModelName      string `json:"model_name" form:"model_name"`
}

// GetUserLogs 获取用户日志
func (h *LogHandler) GetUserLogs(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	req := &UserLogsRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := application.UserLogsParams{
		UserId:         userId,
		Limit:          h.itemsPerPage,
		Offset:         req.Page * h.itemsPerPage,
		LogType:        req.LogType,
		StartTimestamp: req.StartTimestamp,
		EndTimestamp:   req.EndTimestamp,
		TokenName:      req.TokenName,
		ModelName:      req.ModelName,
	}

	logs, err := h.service.GetUserLogs(p)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data":    logs,
	})
}

// SearchAllLogs 搜索所有日志
func (h *LogHandler) SearchAllLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	logs, err := h.service.SearchAllLogs(keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data":    logs,
	})
}

// SearchUserLogs 搜索用户日志
func (h *LogHandler) SearchUserLogs(c *gin.Context) {
	keyword := c.DefaultQuery("keyword", "")
	userId := c.GetInt(ctxkey.Id)
	logs, err := h.service.SearchUserLogs(userId, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data":    logs,
	})
}

// LogsStatsRequest 日志统计参数
type LogsStatsRequest struct {
	LogType        int    `json:"type" form:"type"`
	StartTimestamp int64  `json:"start_timestamp" form:"start_timestamp"`
	EndTimestamp   int64  `json:"end_timestamp" form:"end_timestamp"`
	Username       string `json:"username" form:"username"`
	TokenName      string `json:"token_name" form:"token_name"`
	ModelName      string `json:"model_name" form:"model_name"`
	Channel        int    `json:"channel" form:"channel"`
}

// GetLogsStat 获取日志统计
func (h *LogHandler) GetLogsStat(c *gin.Context) {
	req := &LogsStatsRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := application.SumUsedQuotaParams{
		LogType:        req.LogType,
		StartTimestamp: req.StartTimestamp,
		EndTimestamp:   req.EndTimestamp,
		ModelName:      req.ModelName,
		Username:       req.Username,
		TokenName:      req.TokenName,
		Channel:        req.Channel,
	}

	quotaNum := h.service.SumUsedQuota(p)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data": gin.H{
			"quota": quotaNum,
		},
	})
}

// UserLogsStatsRequest 用户自己日志统计参数
type UserLogsStatsRequest struct {
	LogType        int    `json:"type" form:"type"`
	StartTimestamp int64  `json:"start_timestamp" form:"start_timestamp"`
	EndTimestamp   int64  `json:"end_timestamp" form:"end_timestamp"`
	Username       string `json:"username" form:"username"`
	TokenName      string `json:"token_name" form:"token_name"`
	ModelName      string `json:"model_name" form:"model_name"`
	Channel        int    `json:"channel" form:"channel"`
}

// GetLogsSelfStat 获取当前用户日志统计
func (h *LogHandler) GetLogsSelfStat(c *gin.Context) {
	username := c.GetString(ctxkey.Username)
	req := &LogsStatsRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 这里是当前用户
	req.Username = username

	p := application.SumUsedQuotaParams{
		LogType:        req.LogType,
		StartTimestamp: req.StartTimestamp,
		EndTimestamp:   req.EndTimestamp,
		ModelName:      req.ModelName,
		Username:       req.Username,
		TokenName:      req.TokenName,
		Channel:        req.Channel,
	}

	quotaNum := h.service.SumUsedQuota(p)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"data": gin.H{
			"quota": quotaNum,
		},
	})
}

// DeleteHistoryLogs 删除历史日志
func (h *LogHandler) DeleteHistoryLogs(c *gin.Context) {
	targetTimestamp, _ := strconv.ParseInt(c.Query("target_timestamp"), 10, 64)
	if targetTimestamp == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "target timestamp is required",
		})
		return
	}

	count, err := h.service.DeleteOldLog(targetTimestamp)
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
		"data":    count,
	})
}
