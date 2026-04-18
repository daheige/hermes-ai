package cache

import (
	"time"

	"hermes-ai/internal/domain/repo"
)

func NewNoCache() repo.CacheRepository {
	return &noCacheImpl{}
}

// noCacheImpl 无缓存实现
type noCacheImpl struct{}

func (n *noCacheImpl) Get(key string) (string, error) {
	return "", nil
}

func (n *noCacheImpl) Set(key string, value string, expiration time.Duration) error {
	return nil
}

func (n *noCacheImpl) Delete(key string) error {
	return nil
}

func (n *noCacheImpl) Decrease(key string, value int64) error {
	return nil
}

func (n *noCacheImpl) IsEnabled() bool {
	return false
}
