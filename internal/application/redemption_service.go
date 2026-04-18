package application

import (
	"context"
	"fmt"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/utils"
)

// RedemptionService 兑换码服务
type RedemptionService struct {
	redemptionRepo           repo.RedemptionRepository
	logService               *LogService
	quotaPerUnit             float64
	displayInCurrencyEnabled bool
}

// NewRedemptionService 创建兑换码服务
func NewRedemptionService(
	redemptionRepo repo.RedemptionRepository,
	logService *LogService,
	quotaPerUnit float64,
	displayInCurrencyEnabled bool,
) *RedemptionService {
	return &RedemptionService{
		redemptionRepo:           redemptionRepo,
		logService:               logService,
		quotaPerUnit:             quotaPerUnit,
		displayInCurrencyEnabled: displayInCurrencyEnabled,
	}
}

// GetAllRedemptions 获取所有兑换码
func (s *RedemptionService) GetAllRedemptions(startIdx int, num int) ([]*entity.Redemption, error) {
	return s.redemptionRepo.GetAllRedemptions(startIdx, num)
}

// SearchRedemptions 搜索兑换码
func (s *RedemptionService) SearchRedemptions(keyword string) ([]*entity.Redemption, error) {
	return s.redemptionRepo.SearchRedemptions(keyword)
}

// GetRedemptionById 根据ID获取兑换码
func (s *RedemptionService) GetRedemptionById(id int) (*entity.Redemption, error) {
	return s.redemptionRepo.GetRedemptionById(id)
}

// Redeem 兑换
func (s *RedemptionService) Redeem(ctx context.Context, key string, userId int) (int64, error) {
	redemption, err := s.redemptionRepo.Redeem(key, userId)
	if err != nil {
		return 0, err
	}
	s.logService.RecordLog(ctx, userId, entity.LogTypeTopup,
		fmt.Sprintf("通过兑换码充值 %s", utils.LogQuota(redemption.Quota, s.quotaPerUnit, s.displayInCurrencyEnabled)))
	return redemption.Quota, nil
}

// Insert 插入兑换码
func (s *RedemptionService) Insert(redemption *entity.Redemption) error {
	return s.redemptionRepo.Insert(redemption)
}

// Update 更新兑换码
func (s *RedemptionService) Update(redemption *entity.Redemption) error {
	return s.redemptionRepo.Update(redemption)
}

// Delete 删除兑换码
func (s *RedemptionService) Delete(id int) error {
	return s.redemptionRepo.Delete(id)
}
