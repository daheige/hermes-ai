package persistence

import (
	"errors"

	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/utils"
)

var _ repo.RedemptionRepository = (*RedemptionRepoImpl)(nil)

// RedemptionRepoImpl 兑换码仓储实现
type RedemptionRepoImpl struct {
	db *gorm.DB
}

// NewRedemptionRepo 创建兑换码仓储
func NewRedemptionRepo(db *gorm.DB) repo.RedemptionRepository {
	return &RedemptionRepoImpl{db: db}
}

// GetAllRedemptions 获取所有兑换码
func (r *RedemptionRepoImpl) GetAllRedemptions(startIdx int, num int) ([]*entity.Redemption, error) {
	var redemptions []*entity.Redemption
	err := r.db.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	return redemptions, err
}

// SearchRedemptions 搜索兑换码
func (r *RedemptionRepoImpl) SearchRedemptions(keyword string) ([]*entity.Redemption, error) {
	var redemptions []*entity.Redemption
	err := r.db.Where("id = ? or name LIKE ?", keyword, keyword+"%").Find(&redemptions).Error
	return redemptions, err
}

// GetRedemptionById 根据ID获取兑换码
func (r *RedemptionRepoImpl) GetRedemptionById(id int) (*entity.Redemption, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	var redemption entity.Redemption
	err := r.db.First(&redemption, "id = ?", id).Error
	return &redemption, err
}

// Insert 插入兑换码
func (r *RedemptionRepoImpl) Insert(redemption *entity.Redemption) error {
	return r.db.Create(redemption).Error
}

// Update 更新兑换码
func (r *RedemptionRepoImpl) Update(redemption *entity.Redemption) error {
	return r.db.Model(redemption).Select("name", "status", "quota", "redeemed_time").Updates(redemption).Error
}

// Delete 删除兑换码
func (r *RedemptionRepoImpl) Delete(id int) error {
	if id == 0 {
		return errors.New("id 为空！")
	}
	var redemption entity.Redemption
	err := r.db.Where("id = ?", id).First(&redemption).Error
	if err != nil {
		return err
	}
	return r.db.Delete(&redemption).Error
}

// Redeem 兑换操作(事务)
func (r *RedemptionRepoImpl) Redeem(key string, userId int) (*entity.Redemption, error) {
	if key == "" {
		return nil, errors.New("未提供兑换码")
	}
	if userId == 0 {
		return nil, errors.New("无效的 user id")
	}

	kc := keyCol(r.db)
	var redemption entity.Redemption

	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(kc+" = ?", key).First(&redemption).Error
		if err != nil {
			return errors.New("无效的兑换码")
		}
		if redemption.Status != entity.RedemptionCodeStatusEnabled {
			return errors.New("该兑换码已被使用")
		}
		// 增加用户配额
		err = tx.Model(&entity.User{}).Where("id = ?", userId).
			Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
		if err != nil {
			return err
		}
		// 更新兑换码状态
		redemption.RedeemedTime = utils.GetTimestamp()
		redemption.Status = entity.RedemptionCodeStatusUsed
		return tx.Save(&redemption).Error
	})

	if err != nil {
		return nil, errors.New("兑换失败，" + err.Error())
	}
	return &redemption, nil
}
