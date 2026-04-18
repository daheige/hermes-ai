package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	message2 "hermes-ai/internal/infras/message"
	"hermes-ai/internal/infras/utils"
)

// TokenService 令牌服务
type TokenService struct {
	tokenRepo    repo.TokenRepository
	userRepo     repo.UserRepository
	cacheRepo    repo.CacheRepository
	batchUpdater *BatchUpdater
	TokenConfig
}

type TokenConfig struct {
	SyncFrequency        int
	BatchUpdateEnabled   bool
	QuotaRemindThreshold int64
	ServerAddress        string
}

// NewTokenService 创建令牌服务
func NewTokenService(
	tokenRepo repo.TokenRepository,
	userRepo repo.UserRepository,
	cacheRepo repo.CacheRepository,
	batchUpdater *BatchUpdater,
	conf TokenConfig,
) *TokenService {
	return &TokenService{
		tokenRepo:    tokenRepo,
		userRepo:     userRepo,
		cacheRepo:    cacheRepo,
		batchUpdater: batchUpdater,
		TokenConfig:  conf,
	}
}

// GetAllUserTokens 获取用户所有令牌
func (s *TokenService) GetAllUserTokens(userId int, startIdx int, num int, order string) ([]*entity.Token, error) {
	return s.tokenRepo.GetAllUserTokens(userId, startIdx, num, order)
}

// SearchUserTokens 搜索用户令牌
func (s *TokenService) SearchUserTokens(userId int, keyword string) ([]*entity.Token, error) {
	return s.tokenRepo.SearchUserTokens(userId, keyword)
}

// ValidateUserToken 验证用户令牌
func (s *TokenService) ValidateUserToken(key string) (*entity.Token, error) {
	if key == "" {
		return nil, errors.New("未提供令牌")
	}
	token, err := s.CacheGetTokenByKey(key)
	if err != nil {
		slog.Error("CacheGetTokenByKey failed: " + err.Error())
		return nil, errors.New("无效的令牌")
	}
	if token.Status == entity.TokenStatusExhausted {
		return nil, fmt.Errorf("令牌 %s（#%d）额度已用尽", token.Name, token.Id)
	} else if token.Status == entity.TokenStatusExpired {
		return nil, errors.New("该令牌已过期")
	}
	if token.Status != entity.TokenStatusEnabled {
		return nil, errors.New("该令牌状态不可用")
	}
	if token.ExpiredTime != -1 && token.ExpiredTime < utils.GetTimestamp() {
		if !s.cacheRepo.IsEnabled() {
			token.Status = entity.TokenStatusExpired
			_ = s.tokenRepo.SelectUpdate(token)
		}
		return nil, errors.New("该令牌已过期")
	}
	if !token.UnlimitedQuota && token.RemainQuota <= 0 {
		if !s.cacheRepo.IsEnabled() {
			token.Status = entity.TokenStatusExhausted
			_ = s.tokenRepo.SelectUpdate(token)
		}
		return nil, errors.New("该令牌额度已用尽")
	}
	return token, nil
}

// CacheGetTokenByKey 带缓存的根据Key获取令牌
func (s *TokenService) CacheGetTokenByKey(key string) (*entity.Token, error) {
	if !s.cacheRepo.IsEnabled() {
		return s.tokenRepo.GetTokenByKey(key)
	}
	tokenObjectString, err := s.cacheRepo.Get(fmt.Sprintf("token:%s", key))
	if err != nil {
		token, err := s.tokenRepo.GetTokenByKey(key)
		if err != nil {
			return nil, err
		}
		jsonBytes, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		cacheErr := s.cacheRepo.Set(fmt.Sprintf("token:%s", key), string(jsonBytes),
			time.Duration(s.SyncFrequency)*time.Second)
		if cacheErr != nil {
			slog.Error("Redis set token error: " + cacheErr.Error())
		}
		return token, nil
	}
	var token entity.Token
	err = json.Unmarshal([]byte(tokenObjectString), &token)
	return &token, err
}

// GetTokenByIds 根据ID和用户ID获取令牌
func (s *TokenService) GetTokenByIds(id int, userId int) (*entity.Token, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id 或 userId 为空！")
	}
	return s.tokenRepo.GetTokenByIds(id, userId)
}

// GetTokenById 根据ID获取令牌
func (s *TokenService) GetTokenById(id int) (*entity.Token, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	return s.tokenRepo.GetTokenById(id)
}

// Insert 插入令牌
func (s *TokenService) Insert(token *entity.Token) error {
	return s.tokenRepo.Insert(token)
}

// Update 更新令牌
func (s *TokenService) Update(token *entity.Token) error {
	return s.tokenRepo.Update(token)
}

// Delete 删除令牌
func (s *TokenService) Delete(id int, userId int) error {
	if id == 0 || userId == 0 {
		return errors.New("id 或 userId 为空！")
	}
	return s.tokenRepo.Delete(id, userId)
}

// IncreaseTokenQuota 增加令牌配额
func (s *TokenService) IncreaseTokenQuota(id int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if s.BatchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeTokenQuota, id, quota)
		return nil
	}

	return s.tokenRepo.IncreaseQuota(id, quota)
}

// DecreaseTokenQuota 减少令牌配额
func (s *TokenService) DecreaseTokenQuota(id int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if s.BatchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeTokenQuota, id, -quota)
		return nil
	}
	return s.tokenRepo.DecreaseQuota(id, quota)
}

// PreConsumeTokenQuota 预消费令牌配额
func (s *TokenService) PreConsumeTokenQuota(tokenId int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	token, err := s.GetTokenById(tokenId)
	if err != nil {
		return err
	}
	if !token.UnlimitedQuota && token.RemainQuota < quota {
		return errors.New("令牌额度不足")
	}
	userQuota, err := s.userRepo.GetUserQuota(token.UserId)
	if err != nil {
		return err
	}
	if userQuota < quota {
		return errors.New("用户额度不足")
	}

	quotaTooLow := userQuota >= s.QuotaRemindThreshold && userQuota-quota < s.QuotaRemindThreshold
	noMoreQuota := userQuota-quota <= 0
	if quotaTooLow || noMoreQuota {
		go func() {
			email, err := s.userRepo.GetUserEmail(token.UserId)
			if err != nil {
				slog.Error("failed to fetch user email: " + err.Error())
			}
			prompt := "额度提醒"
			var contentText string
			if noMoreQuota {
				contentText = "您的额度已用尽"
			} else {
				contentText = "您的额度即将用尽"
			}
			if email != "" {
				topUpLink := fmt.Sprintf("%s/topup", s.ServerAddress)
				content := message2.EmailTemplate(
					prompt,
					fmt.Sprintf(`
						<p>您好！</p>
						<p>%s，当前剩余额度为 <strong>%d</strong>。</p>
						<p>为了不影响您的使用，请及时充值。</p>
						<p style="text-align: center; margin: 30px 0;">
							<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">立即充值</a>
						</p>
						<p style="color: #666;">如果按钮无法点击，请复制以下链接到浏览器中打开：</p>
						<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px; word-break: break-all;">%s</p>
					`, contentText, userQuota, topUpLink, topUpLink),
				)
				err = message2.SendEmail(prompt, email, content)
				if err != nil {
					slog.Error("failed to send email: " + err.Error())
				}
			}
		}()
	}

	if !token.UnlimitedQuota {
		err = s.DecreaseTokenQuota(tokenId, quota)
		if err != nil {
			return err
		}
	}
	return s.decreaseUserQuota(token.UserId, quota)
}

// PostConsumeTokenQuota 后消费令牌配额
func (s *TokenService) PostConsumeTokenQuota(tokenId int, quota int64) error {
	token, err := s.GetTokenById(tokenId)
	if err != nil {
		return err
	}
	if quota > 0 {
		err = s.decreaseUserQuota(token.UserId, quota)
	} else {
		err = s.increaseUserQuota(token.UserId, -quota)
	}
	if !token.UnlimitedQuota {
		if quota > 0 {
			err = s.DecreaseTokenQuota(tokenId, quota)
		} else {
			err = s.IncreaseTokenQuota(tokenId, -quota)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TokenService) increaseUserQuota(id int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if s.BatchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeUserQuota, id, quota)
		return nil
	}
	return s.userRepo.IncreaseUserQuota(id, quota)
}

func (s *TokenService) decreaseUserQuota(id int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if s.BatchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeUserQuota, id, -quota)
		return nil
	}

	return s.userRepo.DecreaseUserQuota(id, quota)
}
