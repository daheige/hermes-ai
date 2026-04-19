package repo

import (
	"hermes-ai/internal/domain/entity"
)

// TokenRepository 令牌仓储接口
type TokenRepository interface {
	// GetAllUserTokens 获取用户所有令牌
	GetAllUserTokens(userId int, offset int, limit int, order string) ([]*entity.Token, error)
	// SearchUserTokens 搜索用户令牌
	SearchUserTokens(userId int, keyword string) ([]*entity.Token, error)
	// GetTokenByKey 根据Key获取令牌
	GetTokenByKey(key string) (*entity.Token, error)
	// GetTokenByIds 根据ID和用户ID获取令牌
	GetTokenByIds(id int, userId int) (*entity.Token, error)
	// GetTokenById 根据ID获取令牌
	GetTokenById(id int) (*entity.Token, error)
	// Insert 插入令牌
	Insert(token *entity.Token) error
	// Update 更新令牌
	Update(token *entity.Token) error
	// SelectUpdate 选择性更新令牌(可更新零值)
	SelectUpdate(token *entity.Token) error
	// Delete 删除令牌
	Delete(id int, userId int) error
	// IncreaseQuota 增加令牌配额
	IncreaseQuota(id int, quota int64) error
	// DecreaseQuota 减少令牌配额
	DecreaseQuota(id int, quota int64) error
}
