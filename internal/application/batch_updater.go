package application

import (
	"log/slog"
	"sync"
	"time"

	"hermes-ai/internal/infras/config"
)

const (
	BatchUpdateTypeUserQuota = iota
	BatchUpdateTypeTokenQuota
	BatchUpdateTypeUsedQuota
	BatchUpdateTypeChannelUsedQuota
	BatchUpdateTypeRequestCount
	BatchUpdateTypeCount
)

// BatchUpdater 批量更新器
type BatchUpdater struct {
	stores []map[int]int64
	locks  []sync.Mutex

	userRepo    interface{ IncreaseUserQuota(int, int64) error }
	tokenRepo   interface{ IncreaseQuota(int, int64) error }
	channelRepo interface{ UpdateChannelUsedQuota(int, int64) }
	userUpdater userUpdater
}

type userUpdater interface {
	UpdateUserUsedQuota(int, int64) error
	UpdateUserRequestCount(int, int) error
}

// NewBatchUpdater 创建批量更新器
func NewBatchUpdater(
	userRepo interface{ IncreaseUserQuota(int, int64) error },
	tokenRepo interface{ IncreaseQuota(int, int64) error },
	channelRepo interface{ UpdateChannelUsedQuota(int, int64) },
	userUpdater userUpdater,
) *BatchUpdater {
	b := &BatchUpdater{
		stores:      make([]map[int]int64, BatchUpdateTypeCount),
		locks:       make([]sync.Mutex, BatchUpdateTypeCount),
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		channelRepo: channelRepo,
		userUpdater: userUpdater,
	}
	for i := 0; i < BatchUpdateTypeCount; i++ {
		b.stores[i] = make(map[int]int64)
	}
	return b
}

// Start 启动批量更新器
func (b *BatchUpdater) Start() {
	if !config.BatchUpdateEnabled {
		return
	}

	go func() {
		for {
			time.Sleep(time.Duration(config.BatchUpdateInterval) * time.Second)
			b.batchUpdate()
		}
	}()
}

// AddRecord 添加新记录
func (b *BatchUpdater) AddRecord(targetType int, id int, value int64) {
	b.locks[targetType].Lock()
	defer b.locks[targetType].Unlock()
	if _, ok := b.stores[targetType][id]; !ok {
		b.stores[targetType][id] = value
	} else {
		b.stores[targetType][id] += value
	}
}

func (b *BatchUpdater) batchUpdate() {
	slog.Info("batch update started")
	for i := 0; i < BatchUpdateTypeCount; i++ {
		b.locks[i].Lock()
		store := b.stores[i]
		b.stores[i] = make(map[int]int64)
		b.locks[i].Unlock()

		for key, value := range store {
			switch i {
			case BatchUpdateTypeUserQuota:
				if err := b.userRepo.IncreaseUserQuota(key, value); err != nil {
					slog.Error("failed to batch update user quota: " + err.Error())
				}
			case BatchUpdateTypeTokenQuota:
				if err := b.tokenRepo.IncreaseQuota(key, value); err != nil {
					slog.Error("failed to batch update token quota: " + err.Error())
				}
			case BatchUpdateTypeUsedQuota:
				if err := b.userUpdater.UpdateUserUsedQuota(key, value); err != nil {
					slog.Error("failed to batch update user used quota: " + err.Error())
				}
			case BatchUpdateTypeRequestCount:
				if err := b.userUpdater.UpdateUserRequestCount(key, int(value)); err != nil {
					slog.Error("failed to batch update user request count: " + err.Error())
				}
			case BatchUpdateTypeChannelUsedQuota:
				b.channelRepo.UpdateChannelUsedQuota(key, value)
			default:
			}
		}
	}

	slog.Info("batch update finished")
}
