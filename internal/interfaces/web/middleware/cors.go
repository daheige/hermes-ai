package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	conf := cors.DefaultConfig()
	conf.AllowAllOrigins = true
	conf.AllowCredentials = true
	conf.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	conf.AllowHeaders = []string{"*"}
	return cors.New(conf)
}
