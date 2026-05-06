package config

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitRedisClient 初始化redis client
func InitRedisClient(connString string, enableCluster bool, password, username string) (redis.UniversalClient, error) {
	if connString == "" {
		log.Println("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil, errors.New("REDIS_CONN_STRING not set")
	}

	var client redis.UniversalClient
	if !enableCluster {
		opt, err := redis.ParseURL(connString)
		if err != nil {
			return nil, err
		}

		client = redis.NewClient(opt)
	} else {
		// cluster mode
		log.Println("Redis cluster mode enabled")
		client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:    strings.Split(connString, ","),
			Password: password,
			Username: username,
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
