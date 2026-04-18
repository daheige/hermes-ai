package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/infras/i18n"
	message2 "hermes-ai/internal/infras/message"
	"hermes-ai/internal/interfaces/web/handlers/validate"
)

type MiscConfig struct {
	EmailVerificationEnabled      bool
	GitHubOAuthEnabled            bool
	GitHubClientId                string
	LarkClientId                  string
	SystemName                    string
	Logo                          string
	Footer                        string
	WeChatAccountQRCodeImageURL   string
	WeChatAuthEnabled             bool
	ServerAddress                 string
	TurnstileCheckEnabled         bool
	TurnstileSiteKey              string
	TopUpLink                     string
	ChatLink                      string
	QuotaPerUnit                  float64
	DisplayInCurrencyEnabled      bool
	OidcEnabled                   bool
	OidcClientId                  string
	OidcWellKnown                 string
	OidcAuthorizationEndpoint     string
	OidcTokenEndpoint             string
	OidcUserinfoEndpoint          string
	EmailDomainRestrictionEnabled bool
	EmailDomainWhitelist          []string
	OptionMap                     map[string]string
}

// MiscHandler 杂项处理器
type MiscHandler struct {
	userService *application.UserService
	MiscConfig
}

// NewMiscHandler 创建杂项处理器
func NewMiscHandler(userService *application.UserService, conf MiscConfig) *MiscHandler {
	return &MiscHandler{userService: userService, MiscConfig: conf}
}

// GetStatus 获取系统状态
func (h *MiscHandler) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"version":                     "1.0.1",
			"start_time":                  time.Now().Unix(),
			"email_verification":          h.EmailVerificationEnabled,
			"github_oauth":                h.GitHubOAuthEnabled,
			"github_client_id":            h.GitHubClientId,
			"lark_client_id":              h.LarkClientId,
			"system_name":                 h.SystemName,
			"logo":                        h.Logo,
			"footer_html":                 h.Footer,
			"wechat_qrcode":               h.WeChatAccountQRCodeImageURL,
			"wechat_login":                h.WeChatAuthEnabled,
			"server_address":              h.ServerAddress,
			"turnstile_check":             h.TurnstileCheckEnabled,
			"turnstile_site_key":          h.TurnstileSiteKey,
			"top_up_link":                 h.TopUpLink,
			"chat_link":                   h.ChatLink,
			"quota_per_unit":              h.QuotaPerUnit,
			"display_in_currency":         h.DisplayInCurrencyEnabled,
			"oidc":                        h.OidcEnabled,
			"oidc_client_id":              h.OidcClientId,
			"oidc_well_known":             h.OidcWellKnown,
			"oidc_authorization_endpoint": h.OidcAuthorizationEndpoint,
			"oidc_token_endpoint":         h.OidcTokenEndpoint,
			"oidc_userinfo_endpoint":      h.OidcUserinfoEndpoint,
		},
	})
}

// GetNotice 获取公告
func (h *MiscHandler) GetNotice(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    h.OptionMap["Notice"],
	})
}

// GetAbout 获取关于信息
func (h *MiscHandler) GetAbout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    h.OptionMap["About"],
	})
}

// GetHomePageContent 获取首页内容
func (h *MiscHandler) GetHomePageContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    h.OptionMap["HomePageContent"],
	})
}

// SendEmailVerification 发送邮箱验证邮件
func (h *MiscHandler) SendEmailVerification(c *gin.Context) {
	email := c.Query("email")
	if err := validate.Validate.Var(email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if h.EmailDomainRestrictionEnabled {
		allowed := false
		for _, domain := range h.EmailDomainWhitelist {
			if strings.HasSuffix(email, "@"+domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员启用了邮箱域名白名单，您的邮箱地址的域名不在白名单中",
			})
			return
		}
	}
	if h.userService.IsEmailAlreadyTaken(email) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "邮箱地址已被占用",
		})
		return
	}

	code := validate.GenerateVerificationCode(6)
	validate.RegisterVerificationCodeWithKey(email, code, validate.EmailVerificationPurpose)
	subject := fmt.Sprintf("%s 邮箱验证邮件", h.SystemName)
	content := message2.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>您正在进行 %s 邮箱验证。</p>
			<p>您的验证码为：</p>
			<p style="font-size: 24px; font-weight: bold; color: #333; background-color: #f8f8f8; padding: 10px; text-align: center; border-radius: 4px;">%s</p>
			<p style="color: #666;">验证码 %d 分钟内有效，如果不是本人操作，请忽略。</p>
		`, h.SystemName, code, validate.VerificationValidMinutes),
	)
	err := message2.SendEmail(subject, email, content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// SendPasswordResetEmail 发送密码重置邮件
func (h *MiscHandler) SendPasswordResetEmail(c *gin.Context) {
	email := c.Query("email")
	if err := validate.Validate.Var(email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if !h.userService.IsEmailAlreadyTaken(email) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该邮箱地址未注册",
		})
		return
	}

	code := validate.GenerateVerificationCode(0)
	validate.RegisterVerificationCodeWithKey(email, code, validate.PasswordResetPurpose)
	link := fmt.Sprintf("%s/user/reset?email=%s&token=%s", h.ServerAddress, email, code)
	subject := fmt.Sprintf("%s 密码重置", h.SystemName)
	content := message2.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>您正在进行 %s 密码重置。</p>
			<p>请点击下面的按钮进行密码重置：</p>
			<p style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">重置密码</a>
			</p>
			<p style="color: #666;">如果按钮无法点击，请复制以下链接到浏览器中打开：</p>
			<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px; word-break: break-all;">%s</p>
			<p style="color: #666;">重置链接 %d 分钟内有效，如果不是本人操作，请忽略。</p>
		`, h.SystemName, link, link, validate.VerificationValidMinutes),
	)
	err := message2.SendEmail(subject, email, content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("%s%s", i18n.Translate(c, "send_email_failed"), err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// PasswordResetRequest 密码重置请求
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
	Token string `json:"token" binding:"required"`
}

// ResetPassword 重置密码
func (h *MiscHandler) ResetPassword(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	if !validate.VerifyCodeWithKey(req.Email, req.Token, validate.PasswordResetPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "重置链接非法或已过期",
		})
		return
	}
	password := validate.GenerateVerificationCode(12)
	err := h.userService.ResetUserPasswordByEmail(req.Email, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	validate.DeleteKey(req.Email, validate.PasswordResetPurpose)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    password,
	})
}
