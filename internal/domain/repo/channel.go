package repo

import (
	"hermes-ai/internal/domain/entity"
)

// ChannelRepository 渠道仓储接口
type ChannelRepository interface {
	// GetAllChannels 获取所有渠道
	GetAllChannels(offset int, limit int, scope string) ([]*entity.Channel, error)
	// SearchChannels 搜索渠道
	SearchChannels(keyword string) ([]*entity.Channel, error)
	// GetChannelById 根据ID获取渠道
	GetChannelById(id int, selectAll bool) (*entity.Channel, error)
	// BatchInsert 批量插入渠道
	BatchInsert(channels []entity.Channel) error
	// Insert 插入渠道
	Insert(channel *entity.Channel) error
	// Update 更新渠道
	Update(channel *entity.Channel) error
	// Delete 删除渠道
	Delete(id int) error
	// UpdateResponseTime 更新渠道响应时间
	UpdateResponseTime(id int, responseTime int64)
	// UpdateBalance 更新渠道余额
	UpdateBalance(id int, balance float64)
	// UpdateChannelStatusById 更新渠道状态
	UpdateChannelStatusById(id int, status int)
	// UpdateChannelUsedQuota 更新渠道已用配额
	UpdateChannelUsedQuota(id int, quota int64)
	// DeleteChannelByStatus 根据状态删除渠道
	DeleteChannelByStatus(status int64) (int64, error)
	// DeleteDisabledChannel 删除禁用的渠道
	DeleteDisabledChannel() (int64, error)
	// GetEnabledChannels 获取所有已启用的渠道
	GetEnabledChannels() ([]*entity.Channel, error)
}
