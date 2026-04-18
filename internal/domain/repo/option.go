package repo

import (
	"hermes-ai/internal/domain/entity"
)

// OptionRepository 配置选项仓储接口
type OptionRepository interface {
	// GetAll 获取所有配置选项
	GetAll() ([]*entity.Option, error)
	// Update 更新配置选项
	Update(key string, value string) error
}
