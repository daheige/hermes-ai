package cache

import (
	"time"

	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/rds"
)

var _ repo.CacheRepository = (*RedisCacheImpl)(nil)

// RedisCacheImpl Redis缓存实现
type RedisCacheImpl struct{}

// NewRedisCache 创建Redis缓存
func NewRedisCache() repo.CacheRepository {
	return &RedisCacheImpl{}
}

// Get 获取缓存
func (r *RedisCacheImpl) Get(key string) (string, error) {
	return rds.RedisGet(key)
}

// Set 设置缓存
func (r *RedisCacheImpl) Set(key string, value string, expiration time.Duration) error {
	return rds.RedisSet(key, value, expiration)
}

// Delete 删除缓存
func (r *RedisCacheImpl) Delete(key string) error {
	return rds.RedisDel(key)
}

// Decrease 减少缓存值
func (r *RedisCacheImpl) Decrease(key string, value int64) error {
	return rds.RedisDecrease(key, value)
}

// IsEnabled 缓存是否启用
func (r *RedisCacheImpl) IsEnabled() bool {
	return rds.RedisEnabled
}
