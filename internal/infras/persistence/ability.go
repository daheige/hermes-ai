package persistence

import (
	"context"
	"sort"
	"strings"

	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
)

var _ repo.AbilityRepository = (*AbilityRepoImpl)(nil)

// AbilityRepoImpl 渠道能力仓储实现
type AbilityRepoImpl struct {
	db *gorm.DB
}

// NewAbilityRepo 创建渠道能力仓储
func NewAbilityRepo(db *gorm.DB) repo.AbilityRepository {
	return &AbilityRepoImpl{db: db}
}

// GetRandomSatisfiedChannel 随机获取满足条件的渠道
func (a *AbilityRepoImpl) GetRandomSatisfiedChannel(group, model string, ignoreFirstPriority bool) (*entity.Channel, error) {
	gc := groupCol(a.db)
	tv := trueVal(a.db)
	ro := randomOrder(a.db)

	var ability entity.Ability
	var channelQuery *gorm.DB
	if ignoreFirstPriority {
		channelQuery = a.db.Where(gc+" = ? and model = ? and enabled = "+tv, group, model)
	} else {
		maxPrioritySubQuery := a.db.Model(&entity.Ability{}).Select("MAX(priority)").
			Where(gc+" = ? and model = ? and enabled = "+tv, group, model)
		channelQuery = a.db.Where(gc+" = ? and model = ? and enabled = "+tv+" and priority = (?)", group, model, maxPrioritySubQuery)
	}

	err := channelQuery.Order(ro).First(&ability).Error
	if err != nil {
		return nil, err
	}

	var channel entity.Channel
	err = a.db.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}

// UpdateAbilityStatus 更新渠道能力状态
func (a *AbilityRepoImpl) UpdateAbilityStatus(channelId int, status bool) error {
	return a.db.Model(&entity.Ability{}).Where("channel_id = ?", channelId).
		Select("enabled").Update("enabled", status).Error
}

// GetGroupModels 获取分组可用的模型列表
func (a *AbilityRepoImpl) GetGroupModels(ctx context.Context, group string) ([]string, error) {
	gc := groupCol(a.db)
	tv := trueVal(a.db)

	var models []string
	err := a.db.WithContext(ctx).Model(&entity.Ability{}).Distinct("model").
		Where(gc+" = ? and enabled = "+tv, group).Pluck("model", &models).Error
	if err != nil {
		return nil, err
	}
	sort.Strings(models)
	return models, nil
}

// BatchCreate 批量创建能力
func (a *AbilityRepoImpl) BatchCreate(abilities []entity.Ability) error {
	return a.db.Create(&abilities).Error
}

// DeleteByChannelId 根据渠道ID删除能力
func (a *AbilityRepoImpl) DeleteByChannelId(channelId int) error {
	return a.db.Where("channel_id = ?", channelId).Delete(&entity.Ability{}).Error
}

// GetAll 获取所有能力
func (a *AbilityRepoImpl) GetAll() ([]*entity.Ability, error) {
	var abilities []*entity.Ability
	err := a.db.Find(&abilities).Error
	return abilities, err
}

// BuildAbilities 根据渠道构建能力列表
func BuildAbilities(channel *entity.Channel) []entity.Ability {
	models := strings.Split(channel.Models, ",")
	// 去重
	seen := make(map[string]bool)
	var uniqueModels []string
	for _, m := range models {
		if !seen[m] {
			seen[m] = true
			uniqueModels = append(uniqueModels, m)
		}
	}

	groups := strings.Split(channel.Group, ",")
	abilities := make([]entity.Ability, 0, len(uniqueModels)*len(groups))
	for _, model := range uniqueModels {
		for _, group := range groups {
			ability := entity.Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == entity.ChannelStatusEnabled,
				Priority:  channel.Priority,
			}
			abilities = append(abilities, ability)
		}
	}
	return abilities
}
