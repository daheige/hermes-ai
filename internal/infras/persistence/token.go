package persistence

import (
	"time"

	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/crypto"
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

// encryptToken 加密 token 的 key 并计算 hash
func encryptToken(token *entity.Token) error {
	if token.Key == "" {
		return nil
	}
	// 如果已经是加密格式，跳过
	if crypto.IsEncrypted(token.Key) {
		token.KeyHash = crypto.KeyHash(token.Key)
		return nil
	}
	plainKey := token.Key
	encrypted, err := crypto.Encrypt(plainKey)
	if err != nil {
		return err
	}
	token.Key = encrypted
	token.KeyHash = crypto.KeyHash(plainKey)
	return nil
}

// decryptToken 解密 token 的 key
func decryptToken(token *entity.Token) {
	if token == nil || token.Key == "" {
		return
	}
	plain, err := crypto.Decrypt(token.Key)
	if err == nil {
		token.Key = plain
	}
}

// decryptTokens 批量解密 token 的 key
func decryptTokens(tokens []*entity.Token) {
	for _, t := range tokens {
		decryptToken(t)
	}
}

// GetAllUserTokens 获取用户所有令牌
func (t *TokenRepoImpl) GetAllUserTokens(userId int, offset int, limit int, order string) ([]*entity.Token, error) {
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

	err := query.Limit(limit).Offset(offset).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	decryptTokens(tokens)
	return tokens, nil
}

// SearchUserTokens 搜索用户令牌
func (t *TokenRepoImpl) SearchUserTokens(userId int, keyword string) ([]*entity.Token, error) {
	var tokens []*entity.Token
	err := t.db.Where("user_id = ?", userId).Where("name LIKE ?", keyword+"%").Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	decryptTokens(tokens)
	return tokens, nil
}

// GetTokenByKey 根据Key获取令牌
func (t *TokenRepoImpl) GetTokenByKey(key string) (*entity.Token, error) {
	hash := crypto.KeyHash(key)
	var token entity.Token
	err := t.db.Where("key_hash = ?", hash).First(&token).Error
	if err != nil {
		return nil, err
	}
	decryptToken(&token)
	return &token, nil
}

// GetTokenByIds 根据ID和用户ID获取令牌
func (t *TokenRepoImpl) GetTokenByIds(id int, userId int) (*entity.Token, error) {
	var token entity.Token
	err := t.db.Where("id = ? and user_id = ?", id, userId).First(&token).Error
	if err != nil {
		return nil, err
	}
	decryptToken(&token)
	return &token, nil
}

// GetTokenById 根据ID获取令牌
func (t *TokenRepoImpl) GetTokenById(id int) (*entity.Token, error) {
	var token entity.Token
	err := t.db.Where("id = ?", id).First(&token).Error
	if err != nil {
		return nil, err
	}
	decryptToken(&token)
	return &token, nil
}

// Insert 插入令牌
func (t *TokenRepoImpl) Insert(token *entity.Token) error {
	if err := encryptToken(token); err != nil {
		return err
	}
	return t.db.Create(token).Error
}

// Update 更新令牌
func (t *TokenRepoImpl) Update(token *entity.Token) error {
	if token.Key != "" {
		if err := encryptToken(token); err != nil {
			return err
		}
	}
	return t.db.Model(token).
		Select("name", "status", "expired_time", "remain_quota", "unlimited_quota", "models", "subnet", "key", "key_hash").
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
			"accessed_time": time.Now().Unix(),
		},
	).Error
}

// DecreaseQuota 减少令牌配额
func (t *TokenRepoImpl) DecreaseQuota(id int, quota int64) error {
	return t.db.Model(&entity.Token{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota - ?", quota),
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"accessed_time": time.Now().Unix(),
		},
	).Error
}
