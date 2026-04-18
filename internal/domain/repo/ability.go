package repo

import (
	"context"

	"hermes-ai/internal/domain/entity"
)

// AbilityRepository 渠道能力仓储接口
type AbilityRepository interface {
	// GetRandomSatisfiedChannel 随机获取满足条件的渠道
	GetRandomSatisfiedChannel(group, model string, ignoreFirstPriority bool) (*entity.Channel, error)
	// UpdateAbilityStatus 更新渠道能力状态
	UpdateAbilityStatus(channelId int, status bool) error
	// GetGroupModels 获取分组可用的模型列表
	GetGroupModels(ctx context.Context, group string) ([]string, error)
	// BatchCreate 批量创建能力
	BatchCreate(abilities []entity.Ability) error
	// DeleteByChannelId 根据渠道ID删除能力
	DeleteByChannelId(channelId int) error
	// GetAll 获取所有能力
	GetAll() ([]*entity.Ability, error)
}
