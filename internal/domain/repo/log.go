package repo

import (
	"hermes-ai/internal/domain/entity"
)

// LogRepository 日志仓储接口
type LogRepository interface {
	// Create 创建日志
	Create(log *entity.Log) error
	// GetAllLogs 获取所有日志
	GetAllLogs(p entity.LogQueryParams) ([]*entity.Log, error)
	// GetUserLogs 获取用户日志
	GetUserLogs(p entity.LogUserQueryParams) ([]*entity.Log, error)
	// SearchAllLogs 搜索所有日志
	SearchAllLogs(keyword string, maxRecentItems int) ([]*entity.Log, error)
	// SearchUserLogs 搜索用户日志
	SearchUserLogs(userId int, keyword string, maxRecentItems int) ([]*entity.Log, error)
	// SumUsedQuota 统计已用配额
	SumUsedQuota(p entity.LogUsedQuotaQueryParams) int64
	// SumUsedToken 统计已用token
	SumUsedToken(p entity.LogUsedTokenQueryParams) int
	// DeleteOldLog 删除旧日志
	DeleteOldLog(targetTimestamp int64) (int64, error)
	// SearchLogsByDayAndModel 按天和模型搜索日志
	SearchLogsByDayAndModel(userId int, start, end int) ([]*entity.LogStatistic, error)
}
