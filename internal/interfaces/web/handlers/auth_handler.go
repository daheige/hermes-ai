package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/i18n"
	"hermes-ai/internal/interfaces/web/handlers/validate"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	service *application.UserService
	AuthConfig
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(service *application.UserService, conf AuthConfig) *AuthHandler {
	return &AuthHandler{service: service, AuthConfig: conf}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthConfig struct {
	PasswordLoginEnabled     bool
	PasswordRegisterEnabled  bool
	RegisterEnabled          bool
	EmailVerificationEnabled bool
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	if !h.PasswordLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了密码登录",
			"success": false,
		})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": i18n.Translate(c, "invalid_parameter"),
			"success": false,
		})
		return
	}

	user, err := h.service.ValidateAndFill(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}

	SetupLogin(user, c)
}

// Logout 用户注销
func (h *AuthHandler) Logout(c *gin.Context) {
	ClearAuthCookie(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
	})
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username         string `json:"username" binding:"required"`
	Password         string `json:"password" binding:"required"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	VerificationCode string `json:"verification_code"`
	AffCode          string `json:"aff_code"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	if !h.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了新用户注册",
			"success": false,
		})
		return
	}
	if !h.PasswordRegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册",
			"success": false,
		})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if err := validate.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}

	if h.EmailVerificationEnabled {
		if req.Email == "" || req.VerificationCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员开启了邮箱验证，请输入邮箱地址和验证码",
			})
			return
		}
		if !validate.VerifyCodeWithKey(req.Email, req.VerificationCode, validate.EmailVerificationPurpose) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "验证码错误或已过期",
			})
			return
		}
	}

	affCode := req.AffCode // this code is the inviter's code, not the user's own code
	inviterId, _ := h.service.GetUserIdByAffCode(affCode)
	cleanUser := entity.User{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.Username,
		InviterId:   inviterId,
	}
	if h.EmailVerificationEnabled {
		cleanUser.Email = req.Email
	}
	if err := h.service.Insert(ctx, &cleanUser, inviterId); err != nil {
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
