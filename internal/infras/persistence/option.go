package persistence

import (
	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
)

var _ repo.OptionRepository = (*OptionRepoImpl)(nil)

// OptionRepoImpl 配置选项仓储实现
type OptionRepoImpl struct {
	db *gorm.DB
}

// NewOptionRepo 创建配置选项仓储
func NewOptionRepo(db *gorm.DB) repo.OptionRepository {
	return &OptionRepoImpl{db: db}
}

// GetAll 获取所有配置选项
func (o *OptionRepoImpl) GetAll() ([]*entity.Option, error) {
	var options []*entity.Option
	err := o.db.Find(&options).Error
	return options, err
}

// Update 更新配置选项
func (o *OptionRepoImpl) Update(key string, value string) error {
	option := entity.Option{Key: key}
	o.db.FirstOrCreate(&option, entity.Option{Key: key})
	option.Value = value
	return o.db.Save(&option).Error
}
