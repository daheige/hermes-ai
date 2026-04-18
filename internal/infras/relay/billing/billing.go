package billing

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/relay/services"
)

func ReturnPreConsumedQuota(ctx context.Context, preConsumedQuota int64, tokenId int) {
	if preConsumedQuota != 0 {
		go func(ctx context.Context) {
			// return pre-consumed quota
			err := services.TokenService.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
			if err != nil {
				log.Println("failed to consume pre consumed quota: " + err.Error())
			}
		}(ctx)
	}
}

type PostConsumeParams struct {
	TokenId    int
	QuotaDelta int64
	TotalQuota int64
	UserId     int
	ChannelId  int
	ModelRatio float64
	GroupRatio float64
	ModelName  string
	TokenName  string
}

func PostConsumeQuota(ctx context.Context, p PostConsumeParams) {
	// quotaDelta is remaining quota to be consumed
	_ = services.TokenService.PostConsumeTokenQuota(p.TokenId, p.QuotaDelta)
	_ = services.UserService.CacheUpdateUserQuota(ctx, p.UserId)

	// totalQuota is total quota consumed
	if p.TotalQuota != 0 {
		logContent := fmt.Sprintf("倍率：%.2f × %.2f", p.ModelRatio, p.GroupRatio)
		logEntry := &entity.Log{
			UserId:           p.UserId,
			ChannelId:        p.ChannelId,
			PromptTokens:     int(p.TotalQuota),
			CompletionTokens: 0,
			ModelName:        p.ModelName,
			TokenName:        p.TokenName,
			Quota:            int(p.TotalQuota),
			Content:          logContent,
		}

		services.LogService.RecordConsumeLog(ctx, logEntry)
		services.UserService.UpdateUserUsedQuotaAndRequestCount(p.UserId, p.TotalQuota)
		services.ChannelService.UpdateChannelUsedQuota(p.ChannelId, p.TotalQuota)
	}

	if p.TotalQuota <= 0 {
		slog.With("request_id", logger.GetRequestID(ctx)).Error(fmt.Sprintf("totalQuota consumed is %d, something is wrong", p.TotalQuota))
	}
}
