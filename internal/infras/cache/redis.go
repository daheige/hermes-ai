package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"hermes-ai/internal/domain/repo"
)

var _ repo.CacheRepository = (*RedisCacheImpl)(nil)

// RedisCacheImpl Redis缓存实现
type RedisCacheImpl struct {
	redisClient redis.UniversalClient
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(redisClient redis.UniversalClient) repo.CacheRepository {
	return &RedisCacheImpl{redisClient: redisClient}
}

// Get 获取缓存
func (r *RedisCacheImpl) Get(key string) (string, error) {
	ctx := context.Background()
	return r.redisClient.Get(ctx, key).Result()
}

// Set 设置缓存
func (r *RedisCacheImpl) Set(key string, value string, expiration time.Duration) error {
	ctx := context.Background()
	return r.redisClient.Set(ctx, key, value, expiration).Err()
}

// Delete 删除缓存
func (r *RedisCacheImpl) Delete(key string) error {
	return r.redisClient.Del(context.Background(), key).Err()
}

// Decrease 减少缓存值
func (r *RedisCacheImpl) Decrease(key string, value int64) error {
	return r.redisClient.DecrBy(context.Background(), key, value).Err()
}

// IsEnabled 缓存是否启用
func (r *RedisCacheImpl) IsEnabled() bool {
	return true
}
