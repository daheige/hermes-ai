package repo

import (
	"hermes-ai/internal/domain/entity"
)

// RedemptionRepository 兑换码仓储接口
type RedemptionRepository interface {
	// GetAllRedemptions 获取所有兑换码
	GetAllRedemptions(startIdx int, num int) ([]*entity.Redemption, error)
	// SearchRedemptions 搜索兑换码
	SearchRedemptions(keyword string) ([]*entity.Redemption, error)
	// GetRedemptionById 根据ID获取兑换码
	GetRedemptionById(id int) (*entity.Redemption, error)
	// Insert 插入兑换码
	Insert(redemption *entity.Redemption) error
	// Update 更新兑换码
	Update(redemption *entity.Redemption) error
	// Delete 删除兑换码
	Delete(id int) error
	// Redeem 兑换操作(事务)
	Redeem(key string, userId int) (*entity.Redemption, error)
}
