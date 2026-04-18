package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/ctxkey"
)

type wechatLoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type WeChatUserConfig struct {
	WeChatServerAddress string
	WeChatServerToken   string
	WeChatAuthEnabled   bool
	RegisterEnabled     bool
}

type WeChatUserHandler struct {
	userService *application.UserService
	WeChatUserConfig
}

func NewWechatLoginHandler(userService *application.UserService, conf WeChatUserConfig) *WeChatUserHandler {
	return &WeChatUserHandler{userService: userService, WeChatUserConfig: conf}
}

func (h *WeChatUserHandler) getWeChatIdByCode(code string) (string, error) {
	if code == "" {
		return "", errors.New("无效的参数")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/wechat/user?code=%s", h.WeChatServerAddress, code), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", h.WeChatServerToken)
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	httpResponse, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()

	var res wechatLoginResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	if !res.Success {
		return "", errors.New(res.Message)
	}
	if res.Data == "" {
		return "", errors.New("验证码错误或已过期")
	}
	return res.Data, nil
}

func (h *WeChatUserHandler) WeChatAuth(c *gin.Context) {
	ctx := c.Request.Context()
	if !h.WeChatAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员未开启通过微信登录以及注册",
			"success": false,
		})
		return
	}
	code := c.Query("code")
	wechatId, err := h.getWeChatIdByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}

	var user *entity.User
	if h.userService.IsWeChatIdAlreadyTaken(wechatId) {
		var err error
		user, err = h.userService.FillUserByWeChatId(wechatId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		if h.RegisterEnabled {
			user = &entity.User{
				WeChatId:    wechatId,
				Username:    "wechat_" + strconv.Itoa(h.userService.GetMaxUserId()+1),
				DisplayName: "WeChat User",
				Role:        entity.RoleCommonUser,
				Status:      entity.UserStatusEnabled,
			}
			if err := h.userService.Insert(ctx, user, 0); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员关闭了新用户注册",
			})
			return
		}
	}

	if user.Status != entity.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}

	SetupLogin(user, c)
}

func (h *WeChatUserHandler) WeChatBind(c *gin.Context) {
	if !h.WeChatAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员未开启通过微信登录以及注册",
			"success": false,
		})
		return
	}
	code := c.Query("code")
	wechatId, err := h.getWeChatIdByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	if h.userService.IsWeChatIdAlreadyTaken(wechatId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该微信账号已被绑定",
		})
		return
	}
	id := c.GetInt(ctxkey.Id)
	user, err := h.userService.FillUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	user.WeChatId = wechatId
	err = h.userService.Update(user, false)
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
	return
}
