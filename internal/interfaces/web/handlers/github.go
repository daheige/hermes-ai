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
	"hermes-ai/internal/infras/utils"
)

type GitHubOAuthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type GitHubUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitHubUserConfig struct {
	GitHubClientId     string
	GitHubClientSecret string
	GitHubOAuthEnabled bool
	RegisterEnabled    bool
}

type GitHubHandler struct {
	userService *application.UserService
	GitHubUserConfig
}

// NewGitHubHandler 创建github handler
func NewGitHubHandler(userService *application.UserService, conf GitHubUserConfig) *GitHubHandler {
	return &GitHubHandler{
		userService:      userService,
		GitHubUserConfig: conf,
	}
}

func (h *GitHubHandler) getGitHubUserInfoByCode(code string) (*GitHubUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}

	values := map[string]string{
		"client_id":     h.GitHubClientId,
		"client_secret": h.GitHubClientSecret,
		"code":          code,
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(jsonData))
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
		return nil, errors.New("无法连接至 GitHub 服务器，请稍后重试！")
	}
	defer func() {
		e := res.Body.Close()
		if e != nil {
			log.Println("close body err:", e)
		}
	}()

	var oAuthResponse GitHubOAuthResponse
	err = json.NewDecoder(res.Body).Decode(&oAuthResponse)
	if err != nil {
		return nil, err
	}
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oAuthResponse.AccessToken))
	res2, err := client.Do(req)
	if err != nil {
		slog.Info(err.Error())
		return nil, errors.New("无法连接至 GitHub 服务器，请稍后重试！")
	}
	defer func() {
		err2 := res2.Body.Close()
		if err2 != nil {
			log.Println("close body err:", err2)
		}
	}()

	var githubUser GitHubUser
	err = json.NewDecoder(res2.Body).Decode(&githubUser)
	if err != nil {
		return nil, err
	}
	if githubUser.Login == "" {
		return nil, errors.New("返回值非法，用户字段为空，请稍后重试！")
	}

	return &githubUser, nil
}

func (h *GitHubHandler) GitHubOAuth(c *gin.Context) {
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
		h.GitHubBind(c, currentUser)
		return
	}

	if !h.GitHubOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 GitHub 登录以及注册",
		})
		return
	}
	code := c.Query("code")
	githubUser, err := h.getGitHubUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var user *entity.User
	if h.userService.IsGitHubIdAlreadyTaken(githubUser.Login) {
		var err error
		user, err = h.userService.FillUserByGitHubId(githubUser.Login)
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
				GitHubId:    githubUser.Login,
				Username:    "github_" + strconv.Itoa(h.userService.GetMaxUserId()+1),
				DisplayName: githubUser.Name,
				Email:       githubUser.Email,
				Role:        entity.RoleCommonUser,
				Status:      entity.UserStatusEnabled,
			}
			if user.DisplayName == "" {
				user.DisplayName = "GitHub User"
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

func (h *GitHubHandler) GitHubBind(c *gin.Context, currentUser *entity.User) {
	if !h.GitHubOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 GitHub 登录以及注册",
		})
		return
	}

	code := c.Query("code")
	githubUser, err := h.getGitHubUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if h.userService.IsGitHubIdAlreadyTaken(githubUser.Login) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 GitHub 账户已被绑定",
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

	user.GitHubId = githubUser.Login
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

func (h *GitHubHandler) GenerateOAuthCode(c *gin.Context) {
	state := utils.GetRandomString(12)
	// 10 minutes for oauth_state cookie
	c.SetCookie("oauth_state", state, 600, "/", "", false, false)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    state,
	})
}
