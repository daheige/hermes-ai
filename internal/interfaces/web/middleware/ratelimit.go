package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"hermes-ai/internal/infras/logger"
	"hermes-ai/internal/infras/ratelimit"
)

var (
	timeFormat = "2006-01-02T15:04:05.000Z"

	inMemoryRateLimiter ratelimit.InMemoryRateLimiter
)

type RateLimitMiddleware struct {
	rdb redis.UniversalClient

	RateLimitConfig
}

type RateLimitConfig struct {
	GlobalWebRateLimitNum      int
	GlobalWebRateLimitDuration int64

	GlobalApiRateLimitNum      int
	GlobalApiRateLimitDuration int64

	CriticalRateLimitNum      int
	CriticalRateLimitDuration int64

	DownloadRateLimitNum      int
	DownloadRateLimitDuration int64

	UploadRateLimitNum int

	UploadRateLimitDuration int64

	RateLimitKeyExpirationDuration time.Duration

	DebugEnabled bool

	RedisEnabled bool
}

// NewRateLimitMiddleware 创建ratelimit
func NewRateLimitMiddleware(rdb redis.UniversalClient, conf RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{rdb: rdb, RateLimitConfig: conf}
}

func (r *RateLimitMiddleware) redisRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	ctx := context.Background()
	key := "rateLimit:" + mark + c.ClientIP()
	listLength, err := r.rdb.LLen(ctx, key).Result()
	if err != nil {
		slog.With("request_id", logger.GetRequestID(ctx)).
			Error("failed to check rate limit len", "error", err.Error())
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}

	if listLength < int64(maxRequestNum) {
		r.rdb.LPush(ctx, key, time.Now().Format(timeFormat))
		r.rdb.Expire(ctx, key, r.RateLimitKeyExpirationDuration)
	} else {
		oldTimeStr, _ := r.rdb.LIndex(ctx, key, -1).Result()
		oldTime, err := time.Parse(timeFormat, oldTimeStr)
		if err != nil {
			slog.With("request_id", logger.GetRequestID(ctx)).
				Error("failed to get rate limit lindex", "error", err.Error())
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		nowTimeStr := time.Now().Format(timeFormat)
		nowTime, err := time.Parse(timeFormat, nowTimeStr)
		if err != nil {
			slog.With("request_id", logger.GetRequestID(ctx)).
				Error("failed to parse time", "error", err.Error())
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		// time.Since will return negative number!
		// See: https://stackoverflow.com/questions/50970900/why-is-time-since-returning-negative-durations-on-windows
		if int64(nowTime.Sub(oldTime).Seconds()) < duration {
			r.rdb.Expire(ctx, key, r.RateLimitKeyExpirationDuration)
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		r.rdb.LPush(ctx, key, time.Now().Format(timeFormat))
		r.rdb.LTrim(ctx, key, 0, int64(maxRequestNum-1))
		r.rdb.Expire(ctx, key, r.RateLimitKeyExpirationDuration)
	}
}

func (r *RateLimitMiddleware) memoryRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	key := mark + c.ClientIP()
	if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
		c.Status(http.StatusTooManyRequests)
		c.Abort()
		return
	}
}

func (r *RateLimitMiddleware) rateLimitFactory(maxRequestNum int, duration int64, mark string) gin.HandlerFunc {
	if maxRequestNum == 0 || r.DebugEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	if r.RedisEnabled {
		return func(c *gin.Context) {
			r.redisRateLimiter(c, maxRequestNum, duration, mark)
		}
	}

	// It's safe to call multi times.
	inMemoryRateLimiter.Init(r.RateLimitKeyExpirationDuration)
	return func(c *gin.Context) {
		r.memoryRateLimiter(c, maxRequestNum, duration, mark)
	}
}

func (r *RateLimitMiddleware) GlobalWebRateLimit() gin.HandlerFunc {
	return r.rateLimitFactory(r.GlobalWebRateLimitNum, r.GlobalWebRateLimitDuration, "GW")
}

func (r *RateLimitMiddleware) GlobalAPIRateLimit() gin.HandlerFunc {
	return r.rateLimitFactory(r.GlobalApiRateLimitNum, r.GlobalApiRateLimitDuration, "GA")
}

func (r *RateLimitMiddleware) CriticalRateLimit() gin.HandlerFunc {
	return r.rateLimitFactory(r.CriticalRateLimitNum, r.CriticalRateLimitDuration, "CT")
}

func (r *RateLimitMiddleware) DownloadRateLimit() gin.HandlerFunc {
	return r.rateLimitFactory(r.DownloadRateLimitNum, r.DownloadRateLimitDuration, "DW")
}

func (r *RateLimitMiddleware) UploadRateLimit() gin.HandlerFunc {
	return r.rateLimitFactory(r.UploadRateLimitNum, r.UploadRateLimitDuration, "UP")
}
