package persistence

import (
	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/utils"
)

var _ repo.TokenRepository = (*TokenRepoImpl)(nil)

// TokenRepoImpl 令牌仓储实现
type TokenRepoImpl struct {
	db *gorm.DB
}

// NewTokenRepo 创建令牌仓储
func NewTokenRepo(db *gorm.DB) repo.TokenRepository {
	return &TokenRepoImpl{db: db}
}

// GetAllUserTokens 获取用户所有令牌
func (t *TokenRepoImpl) GetAllUserTokens(userId int, startIdx int, num int, order string) ([]*entity.Token, error) {
	var tokens []*entity.Token
	query := t.db.Where("user_id = ?", userId)

	switch order {
	case "remain_quota":
		query = query.Order("unlimited_quota desc, remain_quota desc")
	case "used_quota":
		query = query.Order("used_quota desc")
	default:
		query = query.Order("id desc")
	}

	err := query.Limit(num).Offset(startIdx).Find(&tokens).Error
	return tokens, err
}

// SearchUserTokens 搜索用户令牌
func (t *TokenRepoImpl) SearchUserTokens(userId int, keyword string) ([]*entity.Token, error) {
	var tokens []*entity.Token
	err := t.db.Where("user_id = ?", userId).Where("name LIKE ?", keyword+"%").Find(&tokens).Error
	return tokens, err
}

// GetTokenByKey 根据Key获取令牌
func (t *TokenRepoImpl) GetTokenByKey(key string) (*entity.Token, error) {
	kc := keyCol(t.db)
	var token entity.Token
	err := t.db.Where(kc+" = ?", key).First(&token).Error
	return &token, err
}

// GetTokenByIds 根据ID和用户ID获取令牌
func (t *TokenRepoImpl) GetTokenByIds(id int, userId int) (*entity.Token, error) {
	var token entity.Token
	err := t.db.First(&token, "id = ? and user_id = ?", id, userId).Error
	return &token, err
}

// GetTokenById 根据ID获取令牌
func (t *TokenRepoImpl) GetTokenById(id int) (*entity.Token, error) {
	var token entity.Token
	err := t.db.First(&token, "id = ?", id).Error
	return &token, err
}

// Insert 插入令牌
func (t *TokenRepoImpl) Insert(token *entity.Token) error {
	return t.db.Create(token).Error
}

// Update 更新令牌
func (t *TokenRepoImpl) Update(token *entity.Token) error {
	return t.db.Model(token).Select("name", "status", "expired_time", "remain_quota", "unlimited_quota", "models", "subnet").
		Updates(token).Error
}

// SelectUpdate 选择性更新令牌
func (t *TokenRepoImpl) SelectUpdate(token *entity.Token) error {
	return t.db.Model(token).Select("accessed_time", "status").Updates(token).Error
}

// Delete 删除令牌
func (t *TokenRepoImpl) Delete(id int, userId int) error {
	token := entity.Token{Id: id, UserId: userId}
	err := t.db.Where(token).First(&token).Error
	if err != nil {
		return err
	}
	return t.db.Delete(&token).Error
}

// IncreaseQuota 增加令牌配额
func (t *TokenRepoImpl) IncreaseQuota(id int, quota int64) error {
	return t.db.Model(&entity.Token{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota + ?", quota),
			"used_quota":    gorm.Expr("used_quota - ?", quota),
			"accessed_time": utils.GetTimestamp(),
		},
	).Error
}

// DecreaseQuota 减少令牌配额
func (t *TokenRepoImpl) DecreaseQuota(id int, quota int64) error {
	return t.db.Model(&entity.Token{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota - ?", quota),
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"accessed_time": utils.GetTimestamp(),
		},
	).Error
}
