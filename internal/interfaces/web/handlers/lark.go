package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
)

type LarkOAuthResponse struct {
	AccessToken string `json:"access_token"`
}

type LarkUser struct {
	Name   string `json:"name"`
	OpenID string `json:"open_id"`
}

type LarkUserHandler struct {
	userService *application.UserService
	LarkUserConfig
}

type LarkUserConfig struct {
	LarkClientId     string
	LarkClientSecret string
	ServerAddress    string
	RegisterEnabled  bool
}

func NewLarkUserHandler(userService *application.UserService, conf LarkUserConfig) *LarkUserHandler {
	return &LarkUserHandler{userService: userService, LarkUserConfig: conf}
}

func (h *LarkUserHandler) getLarkUserInfoByCode(code string) (*LarkUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}
	values := map[string]string{
		"client_id":     h.LarkClientId,
		"client_secret": h.LarkClientSecret,
		"code":          code,
		"grant_type":    "authorization_code",
		"redirect_uri":  fmt.Sprintf("%s/oauth/lark", h.ServerAddress),
	}

	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://open.feishu.cn/open-apis/authen/v2/oauth/token", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		slog.Info(err.Error())
		return nil, errors.New("无法连接至飞书服务器，请稍后重试！")
	}
	defer func() {
		err2 := res.Body.Close()
		if err2 != nil {
			log.Println("failed to close response body err:", err2)
		}
	}()

	var oAuthResponse LarkOAuthResponse
	err = json.NewDecoder(res.Body).Decode(&oAuthResponse)
	if err != nil {
		return nil, err
	}
	req, err = http.NewRequest("GET", "https://passport.feishu.cn/suite/passport/oauth/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oAuthResponse.AccessToken))
	res2, err := client.Do(req)
	if err != nil {
		slog.Info(err.Error())
		return nil, errors.New("无法连接至飞书服务器，请稍后重试！")
	}
	var larkUser LarkUser
	err = json.NewDecoder(res2.Body).Decode(&larkUser)
	if err != nil {
		return nil, err
	}
	return &larkUser, nil
}

func (h *LarkUserHandler) LarkOAuth(c *gin.Context) {
	ctx := c.Request.Context()
	state := c.Query("state")
	cookieState, _ := c.Cookie("oauth_state")
	if state == "" || cookieState == "" || state != cookieState {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "state is empty or not same",
		})
		return
	}
	currentUser := GetUserFromAuthCookie(c, h.userService)
	if currentUser != nil {
		h.LarkBind(c, currentUser)
		return
	}

	code := c.Query("code")
	larkUser, err := h.getLarkUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var user *entity.User
	if h.userService.IsLarkIdAlreadyTaken(larkUser.OpenID) {
		var err error
		user, err = h.userService.FillUserByLarkId(larkUser.OpenID)
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
				LarkId:      larkUser.OpenID,
				Username:    "lark_" + strconv.Itoa(h.userService.GetMaxUserId()+1),
				DisplayName: larkUser.Name,
				Role:        entity.RoleCommonUser,
				Status:      entity.UserStatusEnabled,
			}
			if user.DisplayName == "" {
				user.DisplayName = "Lark User"
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

func (h *LarkUserHandler) LarkBind(c *gin.Context, currentUser *entity.User) {
	code := c.Query("code")
	larkUser, err := h.getLarkUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if h.userService.IsLarkIdAlreadyTaken(larkUser.OpenID) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该飞书账户已被绑定",
		})
		return
	}

	user, err := h.userService.FillUserById(currentUser.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.LarkId = larkUser.OpenID
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
		"message": "bind",
	})
}
