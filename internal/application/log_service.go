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
	logRepo          repo.LogRepository
	userRepo         repo.UserRepository
	logConsumeEnabled bool
	maxRecentItems   int
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

// GetAllLogs 获取所有日志
func (s *LogService) GetAllLogs(logType int, startTimestamp int64, endTimestamp int64, modelName string,
	username string, tokenName string, startIdx int, num int, channel int) ([]*entity.Log, error) {
	return s.logRepo.GetAllLogs(entity.LogQueryParams{
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		Username:       username,
		TokenName:      tokenName,
		Offset:         startIdx,
		Limit:          num,
		Channel:        channel,
	})
}

// GetUserLogs 获取用户日志
func (s *LogService) GetUserLogs(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string,
	tokenName string, startIdx int, num int) ([]*entity.Log, error) {
	return s.logRepo.GetUserLogs(entity.LogUserQueryParams{
		UserId:         userId,
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		TokenName:      tokenName,
		Offset:         startIdx,
		Limit:          num,
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

// SumUsedQuota 统计已用配额
func (s *LogService) SumUsedQuota(logType int, startTimestamp int64, endTimestamp int64, modelName string,
	username string, tokenName string, channel int) int64 {
	return s.logRepo.SumUsedQuota(entity.LogUsedQuotaQueryParams{
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		Username:       username,
		TokenName:      tokenName,
		Channel:        channel,
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
