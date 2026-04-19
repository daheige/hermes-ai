package repo

import (
	"context"
	"time"
)

// CacheRepository 缓存仓储接口
type CacheRepository interface {
	// Get 获取缓存
	Get(key string) (string, error)

	// Set 设置缓存
	Set(key string, value string, expiration time.Duration) error

	// Delete 删除缓存
	Delete(key string) error

	// IncrBy
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	HIncrBy(ctx context.Context, key string, field string, value int64) (int64, error)

	// Decrease 减少缓存值
	Decrease(key string, value int64) error

	// IsEnabled 缓存是否启用
	IsEnabled() bool

	// Exists 判断key是否存在
	Exists(key string) (bool, error)
}
