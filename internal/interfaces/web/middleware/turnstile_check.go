package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/config"
)

type turnstileCheckResponse struct {
	Success bool `json:"success"`
}

func TurnstileCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.TurnstileCheckEnabled {
			turnstileChecked, _ := c.Cookie("turnstile_checked")
			if turnstileChecked == "true" {
				c.Next()
				return
			}
			response := c.Query("turnstile")
			if response == "" {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Turnstile token 为空",
				})
				c.Abort()
				return
			}
			rawRes, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
				"secret":   {config.TurnstileSecretKey},
				"response": {response},
				"remoteip": {c.ClientIP()},
			})
			if err != nil {
				slog.Error(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				c.Abort()
				return
			}
			defer rawRes.Body.Close()
			var res turnstileCheckResponse
			err = json.NewDecoder(rawRes.Body).Decode(&res)
			if err != nil {
				slog.Error(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				c.Abort()
				return
			}
			if !res.Success {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Turnstile 校验失败，请刷新重试！",
				})
				c.Abort()
				return
			}
			// 1 hour for turnstile_checked cookie
			c.SetCookie("turnstile_checked", "true", 3600, "/", "", false, false)
		}
		c.Next()
	}
}
