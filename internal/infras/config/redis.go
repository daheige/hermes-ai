package config

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig redis配置
type RedisConfig struct {
	ConnString    string
	EnableCluster bool
	Password      string
	Username      string
}

// InitRedisClient 初始化redis client
func InitRedisClient(cfg RedisConfig) (redis.UniversalClient, error) {
	if cfg.ConnString == "" {
		log.Println("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil, errors.New("REDIS_CONN_STRING not set")
	}

	var client redis.UniversalClient
	if !cfg.EnableCluster {
		opt, err := redis.ParseURL(cfg.ConnString)
		if err != nil {
			return nil, err
		}

		client = redis.NewClient(opt)
	} else {
		// cluster mode
		log.Println("Redis cluster mode enabled")
		client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:    strings.Split(cfg.ConnString, ","),
			Password: cfg.Password,
			Username: cfg.Username,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
