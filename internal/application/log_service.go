package application

import (
	"context"
	"log/slog"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/utils"
)

// LogService 日志服务
type LogService struct {
	logRepo           repo.LogRepository
	userRepo          repo.UserRepository
	logConsumeEnabled bool
	maxRecentItems    int
}

// NewLogService 创建日志服务
func NewLogService(logRepo repo.LogRepository, userRepo repo.UserRepository, logConsumeEnabled bool, maxRecentItems int) *LogService {
	return &LogService{logRepo: logRepo, userRepo: userRepo, logConsumeEnabled: logConsumeEnabled, maxRecentItems: maxRecentItems}
}

func (s *LogService) getRequestID(ctx context.Context) string {
	rawRequestId := ctx.Value(ctxkey.RequestIdKey)
	if rawRequestId == nil {
		return ""
	}
	return rawRequestId.(string)
}

func (s *LogService) recordLogHelper(ctx context.Context, log *entity.Log) {
	log.RequestId = s.getRequestID(ctx)
	err := s.logRepo.Create(log)
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).Error("failed to record log: " + err.Error())
		return
	}

	slog.With("request_id", logger.GetRequestID(ctx)).Info("record log: %+v", log)
}

// RecordLog 记录日志
func (s *LogService) RecordLog(ctx context.Context, userId int, logType int, content string) {
	if logType == entity.LogTypeConsume && !s.logConsumeEnabled {
		return
	}
	log := &entity.Log{
		UserId:    userId,
		Username:  s.userRepo.GetUsernameById(userId),
		CreatedAt: utils.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	s.recordLogHelper(ctx, log)
}

// RecordTopupLog 记录充值日志
func (s *LogService) RecordTopupLog(ctx context.Context, userId int, content string, quota int) {
	log := &entity.Log{
		UserId:    userId,
		Username:  s.userRepo.GetUsernameById(userId),
		CreatedAt: utils.GetTimestamp(),
		Type:      entity.LogTypeTopup,
		Content:   content,
		Quota:     quota,
	}
	s.recordLogHelper(ctx, log)
}

// RecordConsumeLog 记录消费日志
func (s *LogService) RecordConsumeLog(ctx context.Context, log *entity.Log) {
	if !s.logConsumeEnabled {
		return
	}

	log.Username = s.userRepo.GetUsernameById(log.UserId)
	log.CreatedAt = utils.GetTimestamp()
	log.Type = entity.LogTypeConsume
	s.recordLogHelper(ctx, log)
}

// RecordTestLog 记录测试日志
func (s *LogService) RecordTestLog(ctx context.Context, log *entity.Log) {
	log.CreatedAt = utils.GetTimestamp()
	log.Type = entity.LogTypeTest
	s.recordLogHelper(ctx, log)
}

// AllLogsParams 日志参数
type AllLogsParams struct {
	Limit          int
	Offset         int
	LogType        int
	StartTimestamp int64
	EndTimestamp   int64
	Username       string
	TokenName      string
	ModelName      string
	Channel        int
}

// GetAllLogs 获取所有日志
func (s *LogService) GetAllLogs(p AllLogsParams) ([]*entity.Log, error) {
	return s.logRepo.GetAllLogs(entity.LogQueryParams{
		LogType:        p.LogType,
		StartTimestamp: p.StartTimestamp,
		EndTimestamp:   p.EndTimestamp,
		ModelName:      p.ModelName,
		Username:       p.Username,
		TokenName:      p.TokenName,
		Offset:         p.Offset,
		Limit:          p.Limit,
		Channel:        p.Channel,
	})
}

// UserLogsParams 用户日志参数
type UserLogsParams struct {
	UserId         int
	Limit          int
	Offset         int
	LogType        int
	StartTimestamp int64
	EndTimestamp   int64
	TokenName      string
	ModelName      string
}

// GetUserLogs 获取用户日志
func (s *LogService) GetUserLogs(p UserLogsParams) ([]*entity.Log, error) {
	return s.logRepo.GetUserLogs(entity.LogUserQueryParams{
		UserId:         p.UserId,
		LogType:        p.LogType,
		StartTimestamp: p.StartTimestamp,
		EndTimestamp:   p.EndTimestamp,
		ModelName:      p.ModelName,
		TokenName:      p.TokenName,
		Offset:         p.Offset,
		Limit:          p.Limit,
	})
}

// SearchAllLogs 搜索所有日志
func (s *LogService) SearchAllLogs(keyword string) ([]*entity.Log, error) {
	return s.logRepo.SearchAllLogs(keyword, s.maxRecentItems)
}

// SearchUserLogs 搜索用户日志
func (s *LogService) SearchUserLogs(userId int, keyword string) ([]*entity.Log, error) {
	return s.logRepo.SearchUserLogs(userId, keyword, s.maxRecentItems)
}

// SumUsedQuotaParams 统计参数
type SumUsedQuotaParams struct {
	LogType        int
	StartTimestamp int64
	EndTimestamp   int64
	ModelName      string
	Username       string
	TokenName      string
	Channel        int
}

// SumUsedQuota 统计已用配额
func (s *LogService) SumUsedQuota(p SumUsedQuotaParams) int64 {
	return s.logRepo.SumUsedQuota(entity.LogUsedQuotaQueryParams{
		LogType:        p.LogType,
		StartTimestamp: p.StartTimestamp,
		EndTimestamp:   p.EndTimestamp,
		ModelName:      p.ModelName,
		Username:       p.Username,
		TokenName:      p.TokenName,
		Channel:        p.Channel,
	})
}

// SumUsedToken 统计已用token
func (s *LogService) SumUsedToken(logType int, startTimestamp int64, endTimestamp int64,
	modelName string, username string, tokenName string) int {
	return s.logRepo.SumUsedToken(entity.LogUsedTokenQueryParams{
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		Username:       username,
		TokenName:      tokenName,
	})
}

// DeleteOldLog 删除旧日志
func (s *LogService) DeleteOldLog(targetTimestamp int64) (int64, error) {
	return s.logRepo.DeleteOldLog(targetTimestamp)
}

// SearchLogsByDayAndModel 按天和模型搜索日志
func (s *LogService) SearchLogsByDayAndModel(userId, start, end int) ([]*entity.LogStatistic, error) {
	return s.logRepo.SearchLogsByDayAndModel(userId, start, end)
}
