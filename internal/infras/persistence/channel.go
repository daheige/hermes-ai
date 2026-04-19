package persistence

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
)

var _ repo.ChannelRepository = (*ChannelRepoImpl)(nil)

// ChannelRepoImpl 渠道仓储实现
type ChannelRepoImpl struct {
	db *gorm.DB
}

// NewChannelRepo 创建渠道仓储
func NewChannelRepo(db *gorm.DB) repo.ChannelRepository {
	return &ChannelRepoImpl{db: db}
}

// GetAllChannels 获取所有渠道
func (c *ChannelRepoImpl) GetAllChannels(offset int, limit int, scope string) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	var err error
	switch scope {
	case "all":
		err = c.db.Order("id desc").Find(&channels).Error
	case "disabled":
		err = c.db.Order("id desc").Where("status = ? or status = ?",
			entity.ChannelStatusAutoDisabled, entity.ChannelStatusManuallyDisabled).Find(&channels).Error
	default:
		err = c.db.Order("id desc").Limit(limit).Offset(offset).Omit("key").Find(&channels).Error
	}

	return channels, err
}

// SearchChannels 搜索渠道
func (c *ChannelRepoImpl) SearchChannels(keyword string) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	id, _ := strconv.Atoi(keyword)
	err := c.db.Omit("key").Where("id = ? or name LIKE ?", id, keyword+"%").
		Find(&channels).Error
	return channels, err
}

// GetChannelById 根据ID获取渠道
func (c *ChannelRepoImpl) GetChannelById(id int, selectAll bool) (*entity.Channel, error) {
	var channel entity.Channel
	var err error
	if selectAll {
		err = c.db.First(&channel, "id = ?", id).Error
	} else {
		err = c.db.Omit("key").First(&channel, "id = ?", id).Error
	}

	return &channel, err
}

// BatchInsert 批量插入渠道
func (c *ChannelRepoImpl) BatchInsert(channels []entity.Channel) error {
	return c.db.Create(&channels).Error
}

// Insert 插入渠道
func (c *ChannelRepoImpl) Insert(channel *entity.Channel) error {
	return c.db.Create(channel).Error
}

// Update 更新渠道
func (c *ChannelRepoImpl) Update(channel *entity.Channel) error {
	err := c.db.Model(channel).Updates(channel).Error
	if err != nil {
		return err
	}

	// 重新加载更新后的数据
	c.db.Model(channel).First(channel, "id = ?", channel.Id)
	return nil
}

// Delete 删除渠道
func (c *ChannelRepoImpl) Delete(id int) error {
	return c.db.Delete(&entity.Channel{}, "id = ?", id).Error
}

// UpdateResponseTime 更新渠道响应时间
func (c *ChannelRepoImpl) UpdateResponseTime(id int, responseTime int64) {
	c.db.Model(&entity.Channel{Id: id}).Select("response_time", "test_time").Updates(entity.Channel{
		TestTime:     time.Now().Unix(),
		ResponseTime: int(responseTime),
	})
}

// UpdateBalance 更新渠道余额
func (c *ChannelRepoImpl) UpdateBalance(id int, balance float64) {
	c.db.Model(&entity.Channel{Id: id}).Select("balance_updated_time", "balance").Updates(entity.Channel{
		BalanceUpdatedTime: time.Now().Unix(),
		Balance:            balance,
	})
}

// UpdateChannelStatusById 更新渠道状态
func (c *ChannelRepoImpl) UpdateChannelStatusById(id int, status int) {
	c.db.Model(&entity.Channel{}).Where("id = ?", id).Update("status", status)
}

// UpdateChannelUsedQuota 更新渠道已用配额
func (c *ChannelRepoImpl) UpdateChannelUsedQuota(id int, quota int64) {
	c.db.Model(&entity.Channel{}).Where("id = ?", id).Update("used_quota", gorm.Expr("used_quota + ?", quota))
}

// DeleteChannelByStatus 根据状态删除渠道
func (c *ChannelRepoImpl) DeleteChannelByStatus(status int64) (int64, error) {
	result := c.db.Where("status = ?", status).Delete(&entity.Channel{})
	return result.RowsAffected, result.Error
}

// DeleteDisabledChannel 删除禁用的渠道
func (c *ChannelRepoImpl) DeleteDisabledChannel() (int64, error) {
	result := c.db.Where("status = ? or status = ?",
		entity.ChannelStatusAutoDisabled, entity.ChannelStatusManuallyDisabled).Delete(&entity.Channel{})
	return result.RowsAffected, result.Error
}

// GetEnabledChannels 获取所有已启用的渠道
func (c *ChannelRepoImpl) GetEnabledChannels() ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := c.db.Where("status = ?", entity.ChannelStatusEnabled).Find(&channels).Error
	return channels, err
}
