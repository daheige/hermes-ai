package application

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/relay/billing/ratio"
)

// OptionService 配置选项服务
type OptionService struct {
	optionRepo repo.OptionRepository
	cfg        *config.AppConfig
}

// NewOptionService 创建配置选项服务
func NewOptionService(optionRepo repo.OptionRepository, cfg *config.AppConfig) *OptionService {
	return &OptionService{optionRepo: optionRepo, cfg: cfg}
}

// AllOption 获取所有配置选项
func (s *OptionService) AllOption() ([]*entity.Option, error) {
	return s.optionRepo.GetAll()
}

// InitOptionMap 初始化配置选项Map
func (s *OptionService) InitOptionMap() {
	s.cfg.OptionMapRWMutex.Lock()
	s.cfg.OptionMap = make(map[string]string)
	s.cfg.OptionMap["PasswordLoginEnabled"] = strconv.FormatBool(s.cfg.PasswordLoginEnabled)
	s.cfg.OptionMap["PasswordRegisterEnabled"] = strconv.FormatBool(s.cfg.PasswordRegisterEnabled)
	s.cfg.OptionMap["EmailVerificationEnabled"] = strconv.FormatBool(s.cfg.EmailVerificationEnabled)
	s.cfg.OptionMap["GitHubOAuthEnabled"] = strconv.FormatBool(s.cfg.GitHubOAuthEnabled)
	s.cfg.OptionMap["OidcEnabled"] = strconv.FormatBool(s.cfg.OidcEnabled)
	s.cfg.OptionMap["WeChatAuthEnabled"] = strconv.FormatBool(s.cfg.WeChatAuthEnabled)
	s.cfg.OptionMap["TurnstileCheckEnabled"] = strconv.FormatBool(s.cfg.TurnstileCheckEnabled)
	s.cfg.OptionMap["RegisterEnabled"] = strconv.FormatBool(s.cfg.RegisterEnabled)
	s.cfg.OptionMap["AutomaticDisableChannelEnabled"] = strconv.FormatBool(s.cfg.AutomaticDisableChannelEnabled)
	s.cfg.OptionMap["AutomaticEnableChannelEnabled"] = strconv.FormatBool(s.cfg.AutomaticEnableChannelEnabled)
	s.cfg.OptionMap["ApproximateTokenEnabled"] = strconv.FormatBool(s.cfg.ApproximateTokenEnabled)
	s.cfg.OptionMap["LogConsumeEnabled"] = strconv.FormatBool(s.cfg.LogConsumeEnabled)
	s.cfg.OptionMap["DisplayInCurrencyEnabled"] = strconv.FormatBool(s.cfg.DisplayInCurrencyEnabled)
	s.cfg.OptionMap["DisplayTokenStatEnabled"] = strconv.FormatBool(s.cfg.DisplayTokenStatEnabled)
	s.cfg.OptionMap["ChannelDisableThreshold"] = strconv.FormatFloat(s.cfg.ChannelDisableThreshold, 'f', -1, 64)
	s.cfg.OptionMap["EmailDomainRestrictionEnabled"] = strconv.FormatBool(s.cfg.EmailDomainRestrictionEnabled)
	s.cfg.OptionMap["EmailDomainWhitelist"] = strings.Join(s.cfg.EmailDomainWhitelist, ",")
	s.cfg.OptionMap["SMTPServer"] = ""
	s.cfg.OptionMap["SMTPFrom"] = ""
	s.cfg.OptionMap["SMTPPort"] = strconv.Itoa(s.cfg.SMTPPort)
	s.cfg.OptionMap["SMTPAccount"] = ""
	s.cfg.OptionMap["SMTPToken"] = ""
	s.cfg.OptionMap["Notice"] = ""
	s.cfg.OptionMap["About"] = ""
	s.cfg.OptionMap["HomePageContent"] = ""
	s.cfg.OptionMap["Footer"] = s.cfg.Footer
	s.cfg.OptionMap["SystemName"] = s.cfg.SystemName
	s.cfg.OptionMap["Logo"] = s.cfg.Logo
	s.cfg.OptionMap["ServerAddress"] = ""
	s.cfg.OptionMap["GitHubClientId"] = ""
	s.cfg.OptionMap["GitHubClientSecret"] = ""
	s.cfg.OptionMap["WeChatServerAddress"] = ""
	s.cfg.OptionMap["WeChatServerToken"] = ""
	s.cfg.OptionMap["WeChatAccountQRCodeImageURL"] = ""
	s.cfg.OptionMap["MessagePusherAddress"] = ""
	s.cfg.OptionMap["MessagePusherToken"] = ""
	s.cfg.OptionMap["TurnstileSiteKey"] = ""
	s.cfg.OptionMap["TurnstileSecretKey"] = ""
	s.cfg.OptionMap["QuotaForNewUser"] = strconv.FormatInt(s.cfg.QuotaForNewUser, 10)
	s.cfg.OptionMap["QuotaForInviter"] = strconv.FormatInt(s.cfg.QuotaForInviter, 10)
	s.cfg.OptionMap["QuotaForInvitee"] = strconv.FormatInt(s.cfg.QuotaForInvitee, 10)
	s.cfg.OptionMap["QuotaRemindThreshold"] = strconv.FormatInt(s.cfg.QuotaRemindThreshold, 10)
	s.cfg.OptionMap["PreConsumedQuota"] = strconv.FormatInt(s.cfg.PreConsumedQuota, 10)
	s.cfg.OptionMap["ModelRatio"] = ratio.ModelRatio2JSONString()
	s.cfg.OptionMap["GroupRatio"] = ratio.GroupRatio2JSONString()
	s.cfg.OptionMap["CompletionRatio"] = ratio.CompletionRatio2JSONString()
	s.cfg.OptionMap["TopUpLink"] = s.cfg.TopUpLink
	s.cfg.OptionMap["ChatLink"] = s.cfg.ChatLink
	s.cfg.OptionMap["QuotaPerUnit"] = strconv.FormatFloat(s.cfg.QuotaPerUnit, 'f', -1, 64)
	s.cfg.OptionMap["RetryTimes"] = strconv.Itoa(s.cfg.RetryTimes)
	s.cfg.OptionMap["Theme"] = s.cfg.Theme
	s.cfg.OptionMapRWMutex.Unlock()
	s.loadOptionsFromDatabase()
}

func (s *OptionService) loadOptionsFromDatabase() {
	options, _ := s.optionRepo.GetAll()
	for _, option := range options {
		if option.Key == "ModelRatio" {
			option.Value = ratio.AddNewMissingRatio(option.Value)
		}

		err := s.UpdateOptionMap(option.Key, option.Value)
		if err != nil {
			slog.Error("failed to update option map: " + err.Error())
		}
	}
}

// SyncOptions 同步配置选项
func (s *OptionService) SyncOptions(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		slog.Info("syncing options from database")
		s.loadOptionsFromDatabase()
	}
}

// UpdateOption 更新配置选项
func (s *OptionService) UpdateOption(key string, value string) error {
	err := s.optionRepo.Update(key, value)
	if err != nil {
		return err
	}

	return s.UpdateOptionMap(key, value)
}

// UpdateOptionMap 更新配置选项Map
func (s *OptionService) UpdateOptionMap(key string, value string) (err error) {
	s.cfg.OptionMapRWMutex.Lock()
	defer s.cfg.OptionMapRWMutex.Unlock()

	s.cfg.OptionMap[key] = value
	if strings.HasSuffix(key, "Enabled") {
		boolValue := value == "true"
		switch key {
		case "PasswordRegisterEnabled":
			s.cfg.PasswordRegisterEnabled = boolValue
		case "PasswordLoginEnabled":
			s.cfg.PasswordLoginEnabled = boolValue
		case "EmailVerificationEnabled":
			s.cfg.EmailVerificationEnabled = boolValue
		case "GitHubOAuthEnabled":
			s.cfg.GitHubOAuthEnabled = boolValue
		case "OidcEnabled":
			s.cfg.OidcEnabled = boolValue
		case "WeChatAuthEnabled":
			s.cfg.WeChatAuthEnabled = boolValue
		case "TurnstileCheckEnabled":
			s.cfg.TurnstileCheckEnabled = boolValue
		case "RegisterEnabled":
			s.cfg.RegisterEnabled = boolValue
		case "EmailDomainRestrictionEnabled":
			s.cfg.EmailDomainRestrictionEnabled = boolValue
		case "AutomaticDisableChannelEnabled":
			s.cfg.AutomaticDisableChannelEnabled = boolValue
		case "AutomaticEnableChannelEnabled":
			s.cfg.AutomaticEnableChannelEnabled = boolValue
		case "ApproximateTokenEnabled":
			s.cfg.ApproximateTokenEnabled = boolValue
		case "LogConsumeEnabled":
			s.cfg.LogConsumeEnabled = boolValue
		case "DisplayInCurrencyEnabled":
			s.cfg.DisplayInCurrencyEnabled = boolValue
		case "DisplayTokenStatEnabled":
			s.cfg.DisplayTokenStatEnabled = boolValue
		}
	}
	switch key {
	case "EmailDomainWhitelist":
		s.cfg.EmailDomainWhitelist = strings.Split(value, ",")
	case "SMTPServer":
		s.cfg.SMTPServer = value
	case "SMTPPort":
		intValue, _ := strconv.Atoi(value)
		s.cfg.SMTPPort = intValue
	case "SMTPAccount":
		s.cfg.SMTPAccount = value
	case "SMTPFrom":
		s.cfg.SMTPFrom = value
	case "SMTPToken":
		s.cfg.SMTPToken = value
	case "ServerAddress":
		s.cfg.ServerAddress = value
	case "GitHubClientId":
		s.cfg.GitHubClientId = value
	case "GitHubClientSecret":
		s.cfg.GitHubClientSecret = value
	case "LarkClientId":
		s.cfg.LarkClientId = value
	case "LarkClientSecret":
		s.cfg.LarkClientSecret = value
	case "OidcClientId":
		s.cfg.OidcClientId = value
	case "OidcClientSecret":
		s.cfg.OidcClientSecret = value
	case "OidcWellKnown":
		s.cfg.OidcWellKnown = value
	case "OidcAuthorizationEndpoint":
		s.cfg.OidcAuthorizationEndpoint = value
	case "OidcTokenEndpoint":
		s.cfg.OidcTokenEndpoint = value
	case "OidcUserinfoEndpoint":
		s.cfg.OidcUserinfoEndpoint = value
	case "Footer":
		s.cfg.Footer = value
	case "SystemName":
		s.cfg.SystemName = value
	case "Logo":
		s.cfg.Logo = value
	case "WeChatServerAddress":
		s.cfg.WeChatServerAddress = value
	case "WeChatServerToken":
		s.cfg.WeChatServerToken = value
	case "WeChatAccountQRCodeImageURL":
		s.cfg.WeChatAccountQRCodeImageURL = value
	case "MessagePusherAddress":
		s.cfg.MessagePusherAddress = value
	case "MessagePusherToken":
		s.cfg.MessagePusherToken = value
	case "TurnstileSiteKey":
		s.cfg.TurnstileSiteKey = value
	case "TurnstileSecretKey":
		s.cfg.TurnstileSecretKey = value
	case "QuotaForNewUser":
		s.cfg.QuotaForNewUser, _ = strconv.ParseInt(value, 10, 64)
	case "QuotaForInviter":
		s.cfg.QuotaForInviter, _ = strconv.ParseInt(value, 10, 64)
	case "QuotaForInvitee":
		s.cfg.QuotaForInvitee, _ = strconv.ParseInt(value, 10, 64)
	case "QuotaRemindThreshold":
		s.cfg.QuotaRemindThreshold, _ = strconv.ParseInt(value, 10, 64)
	case "PreConsumedQuota":
		s.cfg.PreConsumedQuota, _ = strconv.ParseInt(value, 10, 64)
	case "RetryTimes":
		s.cfg.RetryTimes, _ = strconv.Atoi(value)
	case "ModelRatio":
		err = ratio.UpdateModelRatioByJSONString(value)
	case "GroupRatio":
		err = ratio.UpdateGroupRatioByJSONString(value)
	case "CompletionRatio":
		err = ratio.UpdateCompletionRatioByJSONString(value)
	case "TopUpLink":
		s.cfg.TopUpLink = value
	case "ChatLink":
		s.cfg.ChatLink = value
	case "ChannelDisableThreshold":
		s.cfg.ChannelDisableThreshold, _ = strconv.ParseFloat(value, 64)
	case "QuotaPerUnit":
		s.cfg.QuotaPerUnit, _ = strconv.ParseFloat(value, 64)
	case "Theme":
		s.cfg.Theme = value
	}
	return err
}

// GetOptionValue 安全地获取指定 key 的配置值
func (s *OptionService) GetOptionValue(key string) string {
	s.cfg.OptionMapRWMutex.RLock()
	defer s.cfg.OptionMapRWMutex.RUnlock()
	return s.cfg.OptionMap[key]
}

// GetAllOptions 安全地获取所有配置项的副本（排除敏感字段）
func (s *OptionService) GetAllOptions() map[string]string {
	s.cfg.OptionMapRWMutex.RLock()
	defer s.cfg.OptionMapRWMutex.RUnlock()
	result := make(map[string]string, len(s.cfg.OptionMap))
	for k, v := range s.cfg.OptionMap {
		result[k] = v
	}
	return result
}
