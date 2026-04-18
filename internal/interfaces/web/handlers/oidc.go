package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
)

type OidcResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type OidcUser struct {
	OpenID            string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
}

type OidcUserConfig struct {
	OidcClientId          string
	OidcClientSecret      string
	ServerAddress         string
	OidcTokenEndpoint     string
	OidcUserinfoEndpoint  string
	OidcEnabled           bool
	RegisterEnabled       bool
}

type OidcUserHandler struct {
	userService *application.UserService
	OidcUserConfig
}

func NewOidcUserHandler(userService *application.UserService, conf OidcUserConfig) *OidcUserHandler {
	return &OidcUserHandler{userService: userService, OidcUserConfig: conf}
}

func (h *OidcUserHandler) getOidcUserInfoByCode(code string) (*OidcUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}
	values := map[string]string{
		"client_id":     h.OidcClientId,
		"client_secret": h.OidcClientSecret,
		"code":          code,
		"grant_type":    "authorization_code",
		"redirect_uri":  fmt.Sprintf("%s/oauth/oidc", h.ServerAddress),
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", h.OidcTokenEndpoint, bytes.NewBuffer(jsonData))
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
		return nil, errors.New("无法连接至 OIDC 服务器，请稍后重试！")
	}
	defer res.Body.Close()

	var oidcResponse OidcResponse
	err = json.NewDecoder(res.Body).Decode(&oidcResponse)
	if err != nil {
		return nil, err
	}
	req, err = http.NewRequest("GET", h.OidcUserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+oidcResponse.AccessToken)
	res2, err := client.Do(req)
	if err != nil {
		slog.Info(err.Error())
		return nil, errors.New("无法连接至 OIDC 服务器，请稍后重试！")
	}
	var oidcUser OidcUser
	err = json.NewDecoder(res2.Body).Decode(&oidcUser)
	if err != nil {
		return nil, err
	}
	return &oidcUser, nil
}

func (h *OidcUserHandler) OidcAuth(c *gin.Context) {
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
		h.OidcBind(c, currentUser)
		return
	}

	if !h.OidcEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 OIDC 登录以及注册",
		})
		return
	}
	code := c.Query("code")
	oidcUser, err := h.getOidcUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	var user *entity.User
	if h.userService.IsOidcIdAlreadyTaken(oidcUser.OpenID) {
		var err error
		user, err = h.userService.FillUserByOidcId(oidcUser.OpenID)
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
				OidcId:      oidcUser.OpenID,
				Email:       oidcUser.Email,
				DisplayName: oidcUser.Name,
				Role:        entity.RoleCommonUser,
				Status:      entity.UserStatusEnabled,
			}
			if oidcUser.PreferredUsername != "" {
				user.Username = oidcUser.PreferredUsername
			} else {
				user.Username = "oidc_" + strconv.Itoa(h.userService.GetMaxUserId()+1)
			}
			if user.DisplayName == "" {
				user.DisplayName = "OIDC User"
			}
			err := h.userService.Insert(ctx, user, 0)
			if err != nil {
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

func (h *OidcUserHandler) OidcBind(c *gin.Context, currentUser *entity.User) {
	if !h.OidcEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 OIDC 登录以及注册",
		})
		return
	}

	code := c.Query("code")
	oidcUser, err := h.getOidcUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if h.userService.IsOidcIdAlreadyTaken(oidcUser.OpenID) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 OIDC 账户已被绑定",
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

	user.OidcId = oidcUser.OpenID
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
	return
}
