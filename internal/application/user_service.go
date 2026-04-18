package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/blacklist"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/crypto"
	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/utils"
)

// UserService 用户服务
type UserService struct {
	userRepo     repo.UserRepository
	tokenRepo    repo.TokenRepository
	cacheRepo    repo.CacheRepository
	logService   *LogService
	batchUpdater *BatchUpdater
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo repo.UserRepository,
	tokenRepo repo.TokenRepository,
	cacheRepo repo.CacheRepository,
	logService *LogService,
	batchUpdater *BatchUpdater,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		cacheRepo:    cacheRepo,
		logService:   logService,
		batchUpdater: batchUpdater,
	}
}

// GetMaxUserId 获取最大用户ID
func (s *UserService) GetMaxUserId() int {
	return s.userRepo.GetMaxUserId()
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers(startIdx int, num int, order string) ([]*entity.User, error) {
	return s.userRepo.GetAllUsers(startIdx, num, order)
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(keyword string) ([]*entity.User, error) {
	return s.userRepo.SearchUsers(keyword)
}

// GetUserById 根据ID获取用户
func (s *UserService) GetUserById(id int, selectAll bool) (*entity.User, error) {
	return s.userRepo.GetUserById(id, selectAll)
}

// GetUserById 根据ID获取用户
func (s *UserService) GetUserByName(username string) (*entity.User, error) {
	return s.userRepo.GetByUsername(username)
}

// GetUserIdByAffCode 根据邀请码获取用户ID
func (s *UserService) GetUserIdByAffCode(affCode string) (int, error) {
	return s.userRepo.GetUserIdByAffCode(affCode)
}

// DeleteUserById 根据ID删除用户(软删除)
func (s *UserService) DeleteUserById(id int) error {
	if id == 0 {
		return errors.New("id 为空！")
	}
	user, err := s.userRepo.GetUserById(id, true)
	if err != nil {
		return err
	}
	blacklist.BanUser(user.Id)
	user.Username = fmt.Sprintf("deleted_%s", utils.UUID())
	user.Status = entity.UserStatusDeleted
	return s.userRepo.Update(user)
}

// Insert 插入用户
func (s *UserService) Insert(ctx context.Context, user *entity.User, inviterId int) error {
	var err error
	if user.Password != "" {
		user.Password, err = crypto.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	user.Quota = config.QuotaForNewUser
	user.AccessToken = utils.UUID()
	user.AffCode = utils.GetRandomString(4)
	err = s.userRepo.Insert(user)
	if err != nil {
		return err
	}
	if config.QuotaForNewUser > 0 {
		s.logService.RecordLog(ctx, user.Id, entity.LogTypeSystem,
			fmt.Sprintf("新用户注册赠送 %s", utils.LogQuota(config.QuotaForNewUser, config.QuotaPerUnit, config.DisplayInCurrencyEnabled)))
	}
	if inviterId != 0 {
		if config.QuotaForInvitee > 0 {
			_ = s.IncreaseUserQuota(user.Id, config.QuotaForInvitee)
			s.logService.RecordLog(ctx, user.Id, entity.LogTypeSystem,
				fmt.Sprintf("使用邀请码赠送 %s", utils.LogQuota(config.QuotaForInvitee, config.QuotaPerUnit, config.DisplayInCurrencyEnabled)))
		}
		if config.QuotaForInviter > 0 {
			_ = s.IncreaseUserQuota(inviterId, config.QuotaForInviter)
			s.logService.RecordLog(ctx, inviterId, entity.LogTypeSystem,
				fmt.Sprintf("邀请用户赠送 %s", utils.LogQuota(config.QuotaForInviter, config.QuotaPerUnit, config.DisplayInCurrencyEnabled)))
		}
	}
	// 创建默认令牌
	cleanToken := &entity.Token{
		UserId:         user.Id,
		Name:           "default",
		Key:            utils.GenerateKey(),
		CreatedTime:    utils.GetTimestamp(),
		AccessedTime:   utils.GetTimestamp(),
		ExpiredTime:    -1,
		RemainQuota:    -1,
		UnlimitedQuota: true,
	}
	err = s.tokenRepo.Insert(cleanToken)
	if err != nil {
		slog.Error(fmt.Sprintf("create default token for user %d failed: %s", user.Id, err.Error()))
	}
	return nil
}

// Update 更新用户
func (s *UserService) Update(user *entity.User, updatePassword bool) error {
	var err error
	if updatePassword {
		user.Password, err = crypto.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	if user.Status == entity.UserStatusDisabled {
		blacklist.BanUser(user.Id)
	} else if user.Status == entity.UserStatusEnabled {
		blacklist.UnbanUser(user.Id)
	}
	return s.userRepo.Update(user)
}

// ValidateAndFill 验证用户名密码并填充用户信息
func (s *UserService) ValidateAndFill(username string, password string) (*entity.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("用户名或密码为空")
	}
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		user, err = s.userRepo.GetByEmail(username)
		if err != nil {
			return nil, errors.New("用户名或密码错误，或用户已被封禁")
		}
	}
	okay := crypto.ValidatePasswordAndHash(password, user.Password)
	if !okay || user.Status != entity.UserStatusEnabled {
		return nil, errors.New("用户名或密码错误，或用户已被封禁")
	}
	return user, nil
}

// FillUserById 根据ID填充用户信息
func (s *UserService) FillUserById(id int) (*entity.User, error) {
	return s.userRepo.FillUserById(id)
}

// FillUserByEmail 根据邮箱填充用户信息
func (s *UserService) FillUserByEmail(email string) (*entity.User, error) {
	return s.userRepo.FillUserByEmail(email)
}

// FillUserByGitHubId 根据GitHub ID填充用户信息
func (s *UserService) FillUserByGitHubId(githubId string) (*entity.User, error) {
	return s.userRepo.FillUserByGitHubId(githubId)
}

// FillUserByLarkId 根据Lark ID填充用户信息
func (s *UserService) FillUserByLarkId(larkId string) (*entity.User, error) {
	return s.userRepo.FillUserByLarkId(larkId)
}

// FillUserByOidcId 根据OIDC ID填充用户信息
func (s *UserService) FillUserByOidcId(oidcId string) (*entity.User, error) {
	return s.userRepo.FillUserByOidcId(oidcId)
}

// FillUserByWeChatId 根据WeChat ID填充用户信息
func (s *UserService) FillUserByWeChatId(wechatId string) (*entity.User, error) {
	return s.userRepo.FillUserByWeChatId(wechatId)
}

// IsEmailAlreadyTaken 邮箱是否已被使用
func (s *UserService) IsEmailAlreadyTaken(email string) bool {
	return s.userRepo.IsEmailAlreadyTaken(email)
}

// IsWeChatIdAlreadyTaken WeChat ID是否已被使用
func (s *UserService) IsWeChatIdAlreadyTaken(wechatId string) bool {
	return s.userRepo.IsWeChatIdAlreadyTaken(wechatId)
}

// IsGitHubIdAlreadyTaken GitHub ID是否已被使用
func (s *UserService) IsGitHubIdAlreadyTaken(githubId string) bool {
	return s.userRepo.IsGitHubIdAlreadyTaken(githubId)
}

// IsLarkIdAlreadyTaken Lark ID是否已被使用
func (s *UserService) IsLarkIdAlreadyTaken(larkId string) bool {
	return s.userRepo.IsLarkIdAlreadyTaken(larkId)
}

// IsOidcIdAlreadyTaken OIDC ID是否已被使用
func (s *UserService) IsOidcIdAlreadyTaken(oidcId string) bool {
	return s.userRepo.IsOidcIdAlreadyTaken(oidcId)
}

// IsUsernameAlreadyTaken 用户名是否已被使用
func (s *UserService) IsUsernameAlreadyTaken(username string) bool {
	return s.userRepo.IsUsernameAlreadyTaken(username)
}

// ResetUserPasswordByEmail 根据邮箱重置密码
func (s *UserService) ResetUserPasswordByEmail(email string, password string) error {
	if email == "" || password == "" {
		return errors.New("邮箱地址或密码为空！")
	}
	hashedPassword, err := crypto.Password2Hash(password)
	if err != nil {
		return err
	}
	return s.userRepo.ResetUserPasswordByEmail(email, hashedPassword)
}

// IsAdmin 是否为管理员
func (s *UserService) IsAdmin(userId int) bool {
	return s.userRepo.IsAdmin(userId)
}

// IsUserEnabled 用户是否启用
func (s *UserService) IsUserEnabled(userId int) (bool, error) {
	return s.userRepo.IsUserEnabled(userId)
}

// CacheIsUserEnabled 带缓存的检查用户是否启用
func (s *UserService) CacheIsUserEnabled(userId int) (bool, error) {
	if !s.cacheRepo.IsEnabled() {
		return s.IsUserEnabled(userId)
	}
	enabled, err := s.cacheRepo.Get(fmt.Sprintf("user_enabled:%d", userId))
	if err == nil {
		return enabled == "1", nil
	}
	userEnabled, err := s.IsUserEnabled(userId)
	if err != nil {
		return false, err
	}
	enabledStr := "0"
	if userEnabled {
		enabledStr = "1"
	}

	cacheErr := s.cacheRepo.Set(fmt.Sprintf("user_enabled:%d", userId), enabledStr,
		time.Duration(config.SyncFrequency)*time.Second)
	if cacheErr != nil {
		slog.Error("Redis set user enabled error: " + cacheErr.Error())
	}

	return userEnabled, nil
}

// ValidateAccessToken 验证访问令牌
func (s *UserService) ValidateAccessToken(token string) (*entity.User, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}

	token = strings.Replace(token, "Bearer ", "", 1)
	return s.userRepo.ValidateAccessToken(token)
}

// GetUserQuota 获取用户配额
func (s *UserService) GetUserQuota(id int) (int64, error) {
	return s.userRepo.GetUserQuota(id)
}

// GetUserUsedQuota 获取用户已用配额
func (s *UserService) GetUserUsedQuota(id int) (int64, error) {
	return s.userRepo.GetUserUsedQuota(id)
}

// GetUserEmail 获取用户邮箱
func (s *UserService) GetUserEmail(id int) (string, error) {
	return s.userRepo.GetUserEmail(id)
}

// GetUserGroup 获取用户分组
func (s *UserService) GetUserGroup(id int) (string, error) {
	return s.userRepo.GetUserGroup(id)
}

// CacheGetUserGroup 带缓存的获取用户分组
func (s *UserService) CacheGetUserGroup(id int) (string, error) {
	if !s.cacheRepo.IsEnabled() {
		return s.GetUserGroup(id)
	}
	group, err := s.cacheRepo.Get(fmt.Sprintf("user_group:%d", id))
	if err != nil {
		group, err = s.GetUserGroup(id)
		if err != nil {
			return "", err
		}
		cacheErr := s.cacheRepo.Set(fmt.Sprintf("user_group:%d", id), group,
			time.Duration(config.SyncFrequency)*time.Second)
		if cacheErr != nil {
			slog.Error("Redis set user group error: " + cacheErr.Error())
		}
	}
	return group, nil
}

// CacheGetUserQuota 带缓存的获取用户配额
func (s *UserService) CacheGetUserQuota(ctx context.Context, id int) (int64, error) {
	if !s.cacheRepo.IsEnabled() {
		return s.GetUserQuota(id)
	}

	quotaString, err := s.cacheRepo.Get(fmt.Sprintf("user_quota:%d", id))
	if err != nil {
		return s.fetchAndUpdateUserQuota(ctx, id)
	}
	quota, err := strconv.ParseInt(quotaString, 10, 64)
	if err != nil {
		return 0, nil
	}
	if quota <= config.PreConsumedQuota {
		slog.With("request_id", logger.GetRequestID(ctx)).Info("user %d's cached quota is too low: %d, refreshing from db", quota, id)
		return s.fetchAndUpdateUserQuota(ctx, id)
	}
	return quota, nil
}

func (s *UserService) fetchAndUpdateUserQuota(ctx context.Context, id int) (int64, error) {
	quota, err := s.GetUserQuota(id)
	if err != nil {
		return 0, err
	}
	cacheErr := s.cacheRepo.Set(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota),
		time.Duration(config.SyncFrequency)*time.Second)
	if cacheErr != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).Error("Redis set user quota error: " + cacheErr.Error())
	}
	return quota, nil
}

// CacheUpdateUserQuota 更新用户配额缓存
func (s *UserService) CacheUpdateUserQuota(ctx context.Context, id int) error {
	if !s.cacheRepo.IsEnabled() {
		return nil
	}
	quota, err := s.CacheGetUserQuota(ctx, id)
	if err != nil {
		return err
	}
	return s.cacheRepo.Set(fmt.Sprintf("user_quota:%d", id), fmt.Sprintf("%d", quota),
		time.Duration(config.SyncFrequency)*time.Second)
}

// CacheDecreaseUserQuota 减少用户配额缓存
func (s *UserService) CacheDecreaseUserQuota(id int, quota int64) error {
	if !s.cacheRepo.IsEnabled() {
		return nil
	}
	return s.cacheRepo.Decrease(fmt.Sprintf("user_quota:%d", id), quota)
}

// GetUsernameById 根据ID获取用户名
func (s *UserService) GetUsernameById(id int) string {
	return s.userRepo.GetUsernameById(id)
}

// GetRootUserEmail 获取root用户邮箱
func (s *UserService) GetRootUserEmail() string {
	return s.userRepo.GetRootUserEmail()
}

// IncreaseUserQuota 增加用户配额
func (s *UserService) IncreaseUserQuota(id int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if config.BatchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeUserQuota, id, quota)
		return nil
	}

	return s.userRepo.IncreaseUserQuota(id, quota)
}

// DecreaseUserQuota 减少用户配额
func (s *UserService) DecreaseUserQuota(id int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if config.BatchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeUserQuota, id, -quota)
		return nil
	}

	return s.userRepo.DecreaseUserQuota(id, quota)
}

// UpdateUserUsedQuotaAndRequestCount 更新用户已用配额和请求计数
func (s *UserService) UpdateUserUsedQuotaAndRequestCount(id int, quota int64) {
	if config.BatchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeUsedQuota, id, quota)
		s.batchUpdater.AddRecord(BatchUpdateTypeRequestCount, id, 1)
		return
	}

	err := s.userRepo.UpdateUserUsedQuotaAndRequestCount(id, quota, 1)
	if err != nil {
		slog.Error("failed to update user used quota and request count: " + err.Error())
	}
}
