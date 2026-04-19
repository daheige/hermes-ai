package persistence

import (
	"fmt"

	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
)

var _ repo.LogRepository = (*LogRepoImpl)(nil)

// LogRepoImpl 日志仓储实现
type LogRepoImpl struct {
	db *gorm.DB
}

// NewLogRepo 创建日志仓储
func NewLogRepo(db *gorm.DB) repo.LogRepository {
	return &LogRepoImpl{db: db}
}

// Create 创建日志
func (l *LogRepoImpl) Create(log *entity.Log) error {
	return l.db.Create(log).Error
}

// GetAllLogs 获取所有日志
func (l *LogRepoImpl) GetAllLogs(p entity.LogQueryParams) ([]*entity.Log, error) {
	tx := l.db.Model(&entity.Log{})
	if p.LogType != entity.LogTypeUnknown {
		tx = tx.Where("type = ?", p.LogType)
	}
	if p.ModelName != "" {
		tx = tx.Where("model_name = ?", p.ModelName)
	}
	if p.Username != "" {
		tx = tx.Where("username = ?", p.Username)
	}
	if p.TokenName != "" {
		tx = tx.Where("token_name = ?", p.TokenName)
	}
	if p.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", p.StartTimestamp)
	}
	if p.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", p.EndTimestamp)
	}
	if p.Channel != 0 {
		tx = tx.Where("channel_id = ?", p.Channel)
	}

	var logs []*entity.Log
	err := tx.Order("id desc").Limit(p.Limit).Offset(p.Offset).Find(&logs).Error
	return logs, err
}

// GetUserLogs 获取用户日志
func (l *LogRepoImpl) GetUserLogs(p entity.LogUserQueryParams) ([]*entity.Log, error) {
	tx := l.db.Model(&entity.Log{})
	if p.LogType == entity.LogTypeUnknown {
		tx = tx.Where("user_id = ?", p.UserId)
	} else {
		tx = tx.Where("user_id = ? and type = ?", p.UserId, p.LogType)
	}
	if p.ModelName != "" {
		tx = tx.Where("model_name = ?", p.ModelName)
	}
	if p.TokenName != "" {
		tx = tx.Where("token_name = ?", p.TokenName)
	}
	if p.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", p.StartTimestamp)
	}
	if p.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", p.EndTimestamp)
	}

	var logs []*entity.Log
	err := tx.Order("id desc").Limit(p.Limit).Offset(p.Offset).Omit("id").Find(&logs).Error
	return logs, err
}

// SearchAllLogs 搜索所有日志
func (l *LogRepoImpl) SearchAllLogs(keyword string, maxRecentItems int) ([]*entity.Log, error) {
	var logs []*entity.Log
	err := l.db.Where("type = ? or content LIKE ?", keyword, keyword+"%").
		Order("id desc").Limit(maxRecentItems).Find(&logs).Error
	return logs, err
}

// SearchUserLogs 搜索用户日志
func (l *LogRepoImpl) SearchUserLogs(userId int, keyword string, maxRecentItems int) ([]*entity.Log, error) {
	var logs []*entity.Log
	err := l.db.Where("user_id = ? and type = ?", userId, keyword).
		Order("id desc").Limit(maxRecentItems).Omit("id").Find(&logs).Error
	return logs, err
}

// SumUsedQuota 统计已用配额
func (l *LogRepoImpl) SumUsedQuota(p entity.LogUsedQuotaQueryParams) int64 {
	fn := ifnullFunc(l.db)
	tx := l.db.Table("logs").Select(fmt.Sprintf("%s(sum(quota),0)", fn))
	if p.Username != "" {
		tx = tx.Where("username = ?", p.Username)
	}
	if p.TokenName != "" {
		tx = tx.Where("token_name = ?", p.TokenName)
	}
	if p.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", p.StartTimestamp)
	}
	if p.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", p.EndTimestamp)
	}
	if p.ModelName != "" {
		tx = tx.Where("model_name = ?", p.ModelName)
	}
	if p.Channel != 0 {
		tx = tx.Where("channel_id = ?", p.Channel)
	}

	var quota int64
	tx.Where("type = ?", entity.LogTypeConsume).First(&quota)
	return quota
}

// SumUsedToken 统计已用token
func (l *LogRepoImpl) SumUsedToken(p entity.LogUsedTokenQueryParams) int {
	fn := ifnullFunc(l.db)
	tx := l.db.Table("logs").Select(fmt.Sprintf("%s(sum(prompt_tokens),0) + %s(sum(completion_tokens),0)", fn, fn))
	if p.Username != "" {
		tx = tx.Where("username = ?", p.Username)
	}
	if p.TokenName != "" {
		tx = tx.Where("token_name = ?", p.TokenName)
	}
	if p.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", p.StartTimestamp)
	}
	if p.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", p.EndTimestamp)
	}
	if p.ModelName != "" {
		tx = tx.Where("model_name = ?", p.ModelName)
	}

	var token int
	tx.Where("type = ?", entity.LogTypeConsume).First(&token)
	return token
}

// DeleteOldLog 删除旧日志
func (l *LogRepoImpl) DeleteOldLog(targetTimestamp int64) (int64, error) {
	result := l.db.Where("created_at < ?", targetTimestamp).Delete(&entity.Log{})
	return result.RowsAffected, result.Error
}

// SearchLogsByDayAndModel 按天和模型搜索日志
func (l *LogRepoImpl) SearchLogsByDayAndModel(userId int, start, end int) ([]*entity.LogStatistic, error) {
	groupSelect := dateFormatSQL(l.db)

	logStatistics := make([]*entity.LogStatistic, 0, 100)
	err := l.db.Raw(`
		SELECT `+groupSelect+`,
		model_name, count(1) as request_count,
		sum(quota) as quota,
		sum(prompt_tokens) as prompt_tokens,
		sum(completion_tokens) as completion_tokens
		FROM logs
		WHERE type=2
		AND user_id= ?
		AND created_at BETWEEN ? AND ?
		GROUP BY day, model_name
		ORDER BY day, model_name
	`, userId, start, end).Scan(&logStatistics).Error

	return logStatistics, err
}
