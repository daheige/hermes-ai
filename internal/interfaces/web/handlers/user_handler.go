package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/config"
	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/i18n"
	"hermes-ai/internal/infras/utils"
	"hermes-ai/internal/interfaces/web/handlers/validate"
)

// UserHandler 用户处理器
type UserHandler struct {
	service           *application.UserService
	logService        *application.LogService
	redemptionService *application.RedemptionService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(service *application.UserService, logService *application.LogService,
	redemptionService *application.RedemptionService) *UserHandler {
	return &UserHandler{service: service, logService: logService, redemptionService: redemptionService}
}

// GetAllUsers 获取所有用户
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.DefaultQuery("order", "")
	users, err := h.service.GetAllUsers(p*config.ItemsPerPage, config.ItemsPerPage, order)

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
		"data":    users,
	})
}

// SearchUsers 搜索用户
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	users, err := h.service.SearchUsers(keyword)
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
		"data":    users,
	})
}

// GetUser 根据ID获取用户
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user, err := h.service.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= user.Role && myRole != entity.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权获取同级或更高等级用户的信息",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
}

// GetUserDashboard 获取用户仪表盘数据
func (h *UserHandler) GetUserDashboard(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	now := time.Now()
	startOfDay := now.Truncate(24*time.Hour).AddDate(0, 0, -6).Unix()
	endOfDay := now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second).Unix()

	dashboards, err := h.logService.SearchLogsByDayAndModel(id, int(startOfDay), int(endOfDay))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取统计信息",
			"data":    []*entity.LogStatistic{},
		})
		return
	}
	if len(dashboards) == 0 {
		dashboards = make([]*entity.LogStatistic, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    dashboards,
	})
}

// GenerateAccessToken 生成访问令牌
func (h *UserHandler) GenerateAccessToken(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := h.service.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.AccessToken = utils.UUID()

	if config.DB.Where("access_token = ?", user.AccessToken).First(user).RowsAffected != 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请重试，系统生成的 UUID 竟然重复了！",
		})
		return
	}

	if err := h.service.Update(user, false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	SetAuthCookie(c, user.AccessToken)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AccessToken,
	})
}

// GetAffCode 获取邀请码
func (h *UserHandler) GetAffCode(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := h.service.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.AffCode == "" {
		user.AffCode = utils.GetRandomString(4)
		if err := h.service.Update(user, false); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AffCode,
	})
}

// GetSelf 获取当前用户信息
func (h *UserHandler) GetSelf(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := h.service.GetUserById(id, false)
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
		"data":    user,
	})
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	ID          int    `json:"id" binding:"required"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Quota       int64  `json:"quota"`
	Role        int    `json:"role"`
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	if req.Password == "" {
		req.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := validate.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_input"),
		})
		return
	}

	originUser, err := h.service.GetUserById(req.ID, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	myRole := c.GetInt(ctxkey.Role)
	if myRole <= originUser.Role && myRole != entity.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}
	if myRole <= req.Role && myRole != entity.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权将其他用户权限等级提升到大于等于自己的权限等级",
		})
		return
	}

	updatedUser := &entity.User{
		Id:          req.ID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Quota:       req.Quota,
		Role:        req.Role,
	}

	updatePassword := req.Password != "$I_LOVE_U"
	if updatePassword {
		updatedUser.Password = req.Password
	}

	if err := h.service.Update(updatedUser, updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if originUser.Quota != req.Quota {
		h.logService.RecordLog(ctx, originUser.Id, entity.LogTypeManage, fmt.Sprintf("管理员将用户额度从 %s修改为 %s",
			utils.LogQuota(originUser.Quota, config.QuotaPerUnit, config.DisplayInCurrencyEnabled),
			utils.LogQuota(req.Quota, config.QuotaPerUnit, config.DisplayInCurrencyEnabled)),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// UserSelfUpdateRequest 当前用户更新请求
type UserSelfUpdateRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// UpdateSelf 更新当前用户信息
func (h *UserHandler) UpdateSelf(c *gin.Context) {
	var req UserSelfUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if req.Password == "" {
		req.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := validate.Validate.Struct(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}

	cleanUser := &entity.User{
		Id:          c.GetInt(ctxkey.Id),
		Username:    req.Username,
		DisplayName: req.DisplayName,
	}
	updatePassword := req.Password != "$I_LOVE_U"
	if updatePassword {
		cleanUser.Password = req.Password
	}

	if err := h.service.Update(cleanUser, updatePassword); err != nil {
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

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	originUser, err := h.service.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= originUser.Role {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权删除同权限等级或更高权限等级的用户",
		})
		return
	}
	err = h.service.DeleteUserById(id)
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

// DeleteSelf 删除当前用户
func (h *UserHandler) DeleteSelf(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, _ := h.service.GetUserById(id, false)

	if user.Role == entity.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "不能删除超级管理员账户",
		})
		return
	}

	err := h.service.DeleteUserById(id)
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

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
	Role        int    `json:"role"`
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req UserCreateRequest
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
	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}
	myRole := c.GetInt(ctxkey.Role)
	if req.Role >= myRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法创建权限大于等于自己的用户",
		})
		return
	}

	cleanUser := &entity.User{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}
	if err := h.service.Insert(ctx, cleanUser, 0); err != nil {
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

// UserManageRequest 用户管理请求
type UserManageRequest struct {
	Username string `json:"username" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// ManageUser 管理用户（仅管理员可操作）
func (h *UserHandler) ManageUser(c *gin.Context) {
	var req UserManageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}

	user := &entity.User{
		Username: req.Username,
	}
	// Fill attributes
	config.DB.Where(user).First(user)
	if user.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= user.Role && myRole != entity.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}

	switch req.Action {
	case "disable":
		user.Status = entity.UserStatusDisabled
		if user.Role == entity.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法禁用超级管理员用户",
			})
			return
		}
	case "enable":
		user.Status = entity.UserStatusEnabled
	case "delete":
		if user.Role == entity.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法删除超级管理员用户",
			})
			return
		}
		if err := h.service.DeleteUserById(user.Id); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "promote":
		if myRole != entity.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "普通管理员用户无法提升其他用户为管理员",
			})
			return
		}
		if user.Role >= entity.RoleAdminUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是管理员",
			})
			return
		}
		user.Role = entity.RoleAdminUser
	case "demote":
		if user.Role == entity.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法降级超级管理员用户",
			})
			return
		}
		if user.Role == entity.RoleCommonUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是普通用户",
			})
			return
		}
		user.Role = entity.RoleCommonUser
	}

	if err := h.service.Update(user, false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	clearUser := &entity.User{
		Role:   user.Role,
		Status: user.Status,
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    clearUser,
	})
}

// EmailBind 绑定邮箱
func (h *UserHandler) EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !validate.VerifyCodeWithKey(email, code, validate.EmailVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码错误或已过期",
		})
		return
	}
	id := c.GetInt(ctxkey.Id)
	user, err := h.service.FillUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.Email = email
	// no need to check if this email already taken, because we have used verification code to check it
	err = h.service.Update(user, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.Role == entity.RoleRootUser {
		config.RootUserEmail = email
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// TopUpRequest 充值请求
type TopUpRequest struct {
	Key string `json:"key" binding:"required"`
}

// TopUp 用户充值
func (h *UserHandler) TopUp(c *gin.Context) {
	ctx := c.Request.Context()
	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	id := c.GetInt(ctxkey.Id)
	quota, err := h.redemptionService.Redeem(ctx, req.Key, id)
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
		"data":    quota,
	})
}

// AdminTopUpRequest 管理员充值请求
type AdminTopUpRequest struct {
	UserId int    `json:"user_id" binding:"required"`
	Quota  int    `json:"quota" binding:"required"`
	Remark string `json:"remark"`
}

// AdminTopUp 管理员给用户充值
func (h *UserHandler) AdminTopUp(c *gin.Context) {
	ctx := c.Request.Context()
	var req AdminTopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err := h.service.IncreaseUserQuota(req.UserId, int64(req.Quota))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if req.Remark == "" {
		req.Remark = fmt.Sprintf("通过 API 充值 %s", utils.LogQuota(int64(req.Quota), config.QuotaPerUnit, config.DisplayInCurrencyEnabled))
	}

	h.logService.RecordTopupLog(ctx, req.UserId, req.Remark, req.Quota)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// SetupLogin 设置登录会话并返回用户信息（供AuthHandler使用）
func SetupLogin(user *entity.User, c *gin.Context) {
	SetAuthCookie(c, user.AccessToken)

	// 清理敏感字段，避免返回给前端存入 localStorage
	user.Password = ""
	user.AccessToken = ""
	user.VerificationCode = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data":    user,
	})
}
