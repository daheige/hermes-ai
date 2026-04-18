package persistence

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
)

var _ repo.UserRepository = (*UserRepoImpl)(nil)

// UserRepoImpl 用户仓储实现
type UserRepoImpl struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储
func NewUserRepo(db *gorm.DB) repo.UserRepository {
	return &UserRepoImpl{db: db}
}

// GetMaxUserId 获取最大用户ID
func (u *UserRepoImpl) GetMaxUserId() int {
	var user entity.User
	u.db.Last(&user)
	return user.Id
}

// GetAllUsers 获取所有用户
func (u *UserRepoImpl) GetAllUsers(startIdx int, num int, order string) ([]*entity.User, error) {
	var users []*entity.User
	query := u.db.Limit(num).Offset(startIdx).Omit("password").Where("status != ?", entity.UserStatusDeleted)

	switch order {
	case "quota":
		query = query.Order("quota desc")
	case "used_quota":
		query = query.Order("used_quota desc")
	case "request_count":
		query = query.Order("request_count desc")
	default:
		query = query.Order("id desc")
	}

	err := query.Find(&users).Error
	return users, err
}

// SearchUsers 搜索用户
func (u *UserRepoImpl) SearchUsers(keyword string) ([]*entity.User, error) {
	var users []*entity.User
	var err error
	if isPostgreSQL(u.db) {
		err = u.db.Omit("password").Where("username LIKE ? or email LIKE ? or display_name LIKE ?",
			keyword+"%", keyword+"%", keyword+"%").Find(&users).Error
	} else {
		err = u.db.Omit("password").Where("id = ? or username LIKE ? or email LIKE ? or display_name LIKE ?",
			keyword, keyword+"%", keyword+"%", keyword+"%").Find(&users).Error
	}
	return users, err
}

// GetUserById 根据ID获取用户
func (u *UserRepoImpl) GetUserById(id int, selectAll bool) (*entity.User, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	var user entity.User
	var err error
	if selectAll {
		err = u.db.First(&user, "id = ?", id).Error
	} else {
		err = u.db.Omit("password", "access_token").First(&user, "id = ?", id).Error
	}
	return &user, err
}

// GetUserIdByAffCode 根据邀请码获取用户ID
func (u *UserRepoImpl) GetUserIdByAffCode(affCode string) (int, error) {
	if affCode == "" {
		return 0, errors.New("affCode 为空！")
	}
	var user entity.User
	err := u.db.Select("id").First(&user, "aff_code = ?", affCode).Error
	return user.Id, err
}

// Insert 插入用户
func (u *UserRepoImpl) Insert(user *entity.User) error {
	return u.db.Create(user).Error
}

// Update 更新用户
func (u *UserRepoImpl) Update(user *entity.User) error {
	return u.db.Model(user).Updates(user).Error
}

// GetByUsername 根据用户名获取用户
func (u *UserRepoImpl) GetByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := u.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetByEmail 根据邮箱获取用户
func (u *UserRepoImpl) GetByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := u.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// FillUserById 根据ID填充用户信息
func (u *UserRepoImpl) FillUserById(id int) (*entity.User, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	var user entity.User
	err := u.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

// FillUserByEmail 根据邮箱填充用户信息
func (u *UserRepoImpl) FillUserByEmail(email string) (*entity.User, error) {
	if email == "" {
		return nil, errors.New("email 为空！")
	}
	var user entity.User
	err := u.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// FillUserByGitHubId 根据GitHub ID填充用户信息
func (u *UserRepoImpl) FillUserByGitHubId(githubId string) (*entity.User, error) {
	if githubId == "" {
		return nil, errors.New("GitHub id 为空！")
	}
	var user entity.User
	err := u.db.Where("github_id = ?", githubId).First(&user).Error
	return &user, err
}

// FillUserByLarkId 根据Lark ID填充用户信息
func (u *UserRepoImpl) FillUserByLarkId(larkId string) (*entity.User, error) {
	if larkId == "" {
		return nil, errors.New("lark id 为空！")
	}
	var user entity.User
	err := u.db.Where("lark_id = ?", larkId).First(&user).Error
	return &user, err
}

// FillUserByOidcId 根据OIDC ID填充用户信息
func (u *UserRepoImpl) FillUserByOidcId(oidcId string) (*entity.User, error) {
	if oidcId == "" {
		return nil, errors.New("oidc id 为空！")
	}
	var user entity.User
	err := u.db.Where("oidc_id = ?", oidcId).First(&user).Error
	return &user, err
}

// FillUserByWeChatId 根据WeChat ID填充用户信息
func (u *UserRepoImpl) FillUserByWeChatId(wechatId string) (*entity.User, error) {
	if wechatId == "" {
		return nil, errors.New("WeChat id 为空！")
	}
	var user entity.User
	err := u.db.Where("wechat_id = ?", wechatId).First(&user).Error
	return &user, err
}

// IsEmailAlreadyTaken 邮箱是否已被使用
func (u *UserRepoImpl) IsEmailAlreadyTaken(email string) bool {
	return u.db.Where("email = ?", email).Find(&entity.User{}).RowsAffected == 1
}

// IsWeChatIdAlreadyTaken WeChat ID是否已被使用
func (u *UserRepoImpl) IsWeChatIdAlreadyTaken(wechatId string) bool {
	return u.db.Where("wechat_id = ?", wechatId).Find(&entity.User{}).RowsAffected == 1
}

// IsGitHubIdAlreadyTaken GitHub ID是否已被使用
func (u *UserRepoImpl) IsGitHubIdAlreadyTaken(githubId string) bool {
	return u.db.Where("github_id = ?", githubId).Find(&entity.User{}).RowsAffected == 1
}

// IsLarkIdAlreadyTaken Lark ID是否已被使用
func (u *UserRepoImpl) IsLarkIdAlreadyTaken(larkId string) bool {
	return u.db.Where("lark_id = ?", larkId).Find(&entity.User{}).RowsAffected == 1
}

// IsOidcIdAlreadyTaken OIDC ID是否已被使用
func (u *UserRepoImpl) IsOidcIdAlreadyTaken(oidcId string) bool {
	return u.db.Where("oidc_id = ?", oidcId).Find(&entity.User{}).RowsAffected == 1
}

// IsUsernameAlreadyTaken 用户名是否已被使用
func (u *UserRepoImpl) IsUsernameAlreadyTaken(username string) bool {
	return u.db.Where("username = ?", username).Find(&entity.User{}).RowsAffected == 1
}

// ResetUserPasswordByEmail 根据邮箱重置密码
func (u *UserRepoImpl) ResetUserPasswordByEmail(email string, hashedPassword string) error {
	return u.db.Model(&entity.User{}).Where("email = ?", email).Update("password", hashedPassword).Error
}

// IsAdmin 是否为管理员
func (u *UserRepoImpl) IsAdmin(userId int) bool {
	if userId == 0 {
		return false
	}
	var user entity.User
	err := u.db.Where("id = ?", userId).Select("role").Find(&user).Error
	if err != nil {
		return false
	}
	return user.Role >= entity.RoleAdminUser
}

// IsUserEnabled 用户是否启用
func (u *UserRepoImpl) IsUserEnabled(userId int) (bool, error) {
	if userId == 0 {
		return false, errors.New("user id is empty")
	}
	var user entity.User
	err := u.db.Where("id = ?", userId).Select("status").Find(&user).Error
	if err != nil {
		return false, err
	}
	return user.Status == entity.UserStatusEnabled, nil
}

// ValidateAccessToken 验证访问令牌
func (u *UserRepoImpl) ValidateAccessToken(token string) *entity.User {
	if token == "" {
		return nil
	}
	token = strings.Replace(token, "Bearer ", "", 1)
	user := &entity.User{}
	if u.db.Where("access_token = ?", token).First(user).RowsAffected == 1 {
		return user
	}
	return nil
}

// GetUserQuota 获取用户配额
func (u *UserRepoImpl) GetUserQuota(id int) (int64, error) {
	var quota int64
	err := u.db.Model(&entity.User{}).Where("id = ?", id).Select("quota").Find(&quota).Error
	return quota, err
}

// GetUserUsedQuota 获取用户已用配额
func (u *UserRepoImpl) GetUserUsedQuota(id int) (int64, error) {
	var quota int64
	err := u.db.Model(&entity.User{}).Where("id = ?", id).Select("used_quota").Find(&quota).Error
	return quota, err
}

// GetUserEmail 获取用户邮箱
func (u *UserRepoImpl) GetUserEmail(id int) (string, error) {
	var email string
	err := u.db.Model(&entity.User{}).Where("id = ?", id).Select("email").Find(&email).Error
	return email, err
}

// GetUserGroup 获取用户分组
func (u *UserRepoImpl) GetUserGroup(id int) (string, error) {
	gc := groupCol(u.db)
	var group string
	err := u.db.Model(&entity.User{}).Where("id = ?", id).Select(gc).Find(&group).Error
	return group, err
}

// GetUsernameById 根据ID获取用户名
func (u *UserRepoImpl) GetUsernameById(id int) string {
	var username string
	u.db.Model(&entity.User{}).Where("id = ?", id).Select("username").Find(&username)
	return username
}

// GetRootUserEmail 获取root用户邮箱
func (u *UserRepoImpl) GetRootUserEmail() string {
	var email string
	u.db.Model(&entity.User{}).Where("role = ?", entity.RoleRootUser).Select("email").Find(&email)
	return email
}

// IncreaseUserQuota 增加用户配额
func (u *UserRepoImpl) IncreaseUserQuota(id int, quota int64) error {
	return u.db.Model(&entity.User{}).Where("id = ?", id).
		Update("quota", gorm.Expr("quota + ?", quota)).Error
}

// DecreaseUserQuota 减少用户配额
func (u *UserRepoImpl) DecreaseUserQuota(id int, quota int64) error {
	return u.db.Model(&entity.User{}).Where("id = ?", id).
		Update("quota", gorm.Expr("quota - ?", quota)).Error
}

// UpdateUserUsedQuotaAndRequestCount 更新用户已用配额和请求计数
func (u *UserRepoImpl) UpdateUserUsedQuotaAndRequestCount(id int, quota int64, count int) error {
	return u.db.Model(&entity.User{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"request_count": gorm.Expr("request_count + ?", count),
		},
	).Error
}

// UpdateUserUsedQuota 更新用户已用配额
func (u *UserRepoImpl) UpdateUserUsedQuota(id int, quota int64) error {
	return u.db.Model(&entity.User{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"used_quota": gorm.Expr("used_quota + ?", quota),
		},
	).Error
}

// UpdateUserRequestCount 更新用户请求计数
func (u *UserRepoImpl) UpdateUserRequestCount(id int, count int) error {
	return u.db.Model(&entity.User{}).Where("id = ?", id).
		Update("request_count", gorm.Expr("request_count + ?", count)).Error
}

// DeleteUserById 软删除用户(重命名用户名并更新状态)
func DeleteUserById(db *gorm.DB, id int) error {
	if id == 0 {
		return errors.New("id 为空！")
	}
	var user entity.User
	err := db.First(&user, "id = ?", id).Error
	if err != nil {
		return err
	}
	user.Username = fmt.Sprintf("deleted_%d", id)
	user.Status = entity.UserStatusDeleted
	return db.Model(&user).Updates(&user).Error
}
