package repo

import (
	"hermes-ai/internal/domain/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// GetMaxUserId 获取最大用户ID
	GetMaxUserId() int
	// GetAllUsers 获取所有用户
	GetAllUsers(startIdx int, num int, order string) ([]*entity.User, error)
	// SearchUsers 搜索用户
	SearchUsers(keyword string) ([]*entity.User, error)
	// GetUserById 根据ID获取用户
	GetUserById(id int, selectAll bool) (*entity.User, error)
	// GetUserIdByAffCode 根据邀请码获取用户ID
	GetUserIdByAffCode(affCode string) (int, error)
	// Insert 插入用户
	Insert(user *entity.User) error
	// Update 更新用户
	Update(user *entity.User) error
	// GetByUsername 根据用户名获取用户
	GetByUsername(username string) (*entity.User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(email string) (*entity.User, error)

	// FillUserById 根据ID填充用户信息
	FillUserById(id int) (*entity.User, error)
	// FillUserByEmail 根据邮箱填充用户信息
	FillUserByEmail(email string) (*entity.User, error)
	// FillUserByGitHubId 根据GitHub ID填充用户信息
	FillUserByGitHubId(githubId string) (*entity.User, error)
	// FillUserByLarkId 根据Lark ID填充用户信息
	FillUserByLarkId(larkId string) (*entity.User, error)
	// FillUserByOidcId 根据OIDC ID填充用户信息
	FillUserByOidcId(oidcId string) (*entity.User, error)
	// FillUserByWeChatId 根据WeChat ID填充用户信息
	FillUserByWeChatId(wechatId string) (*entity.User, error)

	// IsEmailAlreadyTaken 邮箱是否已被使用
	IsEmailAlreadyTaken(email string) bool
	// IsWeChatIdAlreadyTaken WeChat ID是否已被使用
	IsWeChatIdAlreadyTaken(wechatId string) bool
	// IsGitHubIdAlreadyTaken GitHub ID是否已被使用
	IsGitHubIdAlreadyTaken(githubId string) bool
	// IsLarkIdAlreadyTaken Lark ID是否已被使用
	IsLarkIdAlreadyTaken(larkId string) bool
	// IsOidcIdAlreadyTaken OIDC ID是否已被使用
	IsOidcIdAlreadyTaken(oidcId string) bool
	// IsUsernameAlreadyTaken 用户名是否已被使用
	IsUsernameAlreadyTaken(username string) bool

	// ResetUserPasswordByEmail 根据邮箱重置密码
	ResetUserPasswordByEmail(email string, hashedPassword string) error
	// IsAdmin 是否为管理员
	IsAdmin(userId int) bool
	// IsUserEnabled 用户是否启用
	IsUserEnabled(userId int) (bool, error)
	// ValidateAccessToken 验证访问令牌
	ValidateAccessToken(token string) (*entity.User, error)

	// GetUserQuota 获取用户配额
	GetUserQuota(id int) (int64, error)
	// GetUserUsedQuota 获取用户已用配额
	GetUserUsedQuota(id int) (int64, error)
	// GetUserEmail 获取用户邮箱
	GetUserEmail(id int) (string, error)
	// GetUserGroup 获取用户分组
	GetUserGroup(id int) (string, error)
	// GetUsernameById 根据ID获取用户名
	GetUsernameById(id int) string
	// GetRootUserEmail 获取root用户邮箱
	GetRootUserEmail() string
	// IncreaseUserQuota 增加用户配额
	IncreaseUserQuota(id int, quota int64) error
	// DecreaseUserQuota 减少用户配额
	DecreaseUserQuota(id int, quota int64) error
	// UpdateUserUsedQuotaAndRequestCount 更新用户已用配额和请求计数
	UpdateUserUsedQuotaAndRequestCount(id int, quota int64, count int) error
	// UpdateUserUsedQuota 更新用户已用配额
	UpdateUserUsedQuota(id int, quota int64) error
	// UpdateUserRequestCount 更新用户请求计数
	UpdateUserRequestCount(id int, count int) error
}
