package config

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"hermes-ai/internal/infras/env"
)

// InitRedisClient 初始化redis client
func InitRedisClient() (redis.UniversalClient, error) {
	if os.Getenv("REDIS_CONN_STRING") == "" {
		log.Println("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil, errors.New("REDIS_CONN_STRING not set")
	}

	var client redis.UniversalClient
	redisConnString := os.Getenv("REDIS_CONN_STRING")
	if !env.Bool("REDIS_ENABLE_CLUSTER", false) {
		opt, err := redis.ParseURL(redisConnString)
		if err != nil {
			return nil, err
		}

		client = redis.NewClient(opt)
	} else {
		// cluster mode
		log.Println("Redis cluster mode enabled")
		client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:    strings.Split(redisConnString, ","),
			Password: os.Getenv("REDIS_PASSWORD"),
			Username: os.Getenv("REDIS_USERNAME"),
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
