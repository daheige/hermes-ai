package repo

import (
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
	// Decrease 减少缓存值
	Decrease(key string, value int64) error
	// IsEnabled 缓存是否启用
	IsEnabled() bool
}
