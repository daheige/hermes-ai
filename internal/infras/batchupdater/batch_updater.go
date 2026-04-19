package batchupdater

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"hermes-ai/internal/domain/repo"
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
	userRepo    interface{ IncreaseUserQuota(int, int64) error }
	tokenRepo   interface{ IncreaseQuota(int, int64) error }
	channelRepo interface{ UpdateChannelUsedQuota(int, int64) error }
	userUpdater userUpdater

	stop                chan struct{}
	batchUpdateInterval time.Duration
	batchUpdateEnabled  bool
	redisClient         redis.UniversalClient
}

type userUpdater interface {
	UpdateUserUsedQuota(int, int64) error
	UpdateUserRequestCount(int, int64) error
}

var _ repo.BatchUpdater = (*BatchUpdater)(nil)

// NewBatchUpdater 创建批量更新器
func NewBatchUpdater(
	userRepo interface{ IncreaseUserQuota(int, int64) error },
	tokenRepo interface{ IncreaseQuota(int, int64) error },
	channelRepo interface{ UpdateChannelUsedQuota(int, int64) error },
	userUpdater userUpdater,
	batchUpdateInterval time.Duration,
	batchUpdateEnabled bool,
	redisClient redis.UniversalClient,
) *BatchUpdater {
	b := &BatchUpdater{
		userRepo:            userRepo,
		tokenRepo:           tokenRepo,
		channelRepo:         channelRepo,
		userUpdater:         userUpdater,
		batchUpdateEnabled:  batchUpdateEnabled,
		batchUpdateInterval: batchUpdateInterval,
		stop:                make(chan struct{}, 1),
		redisClient:         redisClient,
	}

	return b
}

// Start 启动批量更新器
func (b *BatchUpdater) Start() {
	if !b.batchUpdateEnabled {
		return
	}

	go func() {
		ticker := time.NewTicker(b.batchUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b.batchUpdate()
			case <-b.stop:
				log.Println("batch updater stopped")
				return
			}
		}
	}()
}

func (b *BatchUpdater) Stop() {
	close(b.stop)
}

func (b *BatchUpdater) hashKey(targetType int) string {
	hashKey := fmt.Sprintf("batch_update:%d", targetType)
	return hashKey
}

// AddRecord 添加新记录
func (b *BatchUpdater) AddRecord(targetType int, id int, value int64) {
	hashKey := b.hashKey(targetType)
	field := fmt.Sprintf("%d", id)
	// log.Println("update target type", targetType, "id", id, "value", value)
	_, err := b.redisClient.HIncrBy(context.Background(), hashKey, field, value).Result()
	if err != nil {
		slog.Error("hIncrBy error", "error", err.Error(), "target_type", targetType, "id", id, "value", value)
		return
	}

	slog.Info("add record success", "target_type", targetType, "id", id, "value", value)
}

func (b *BatchUpdater) batchUpdate() {
	slog.Info("batch update started")
	targetTypeHandlers := map[int]func(int, int64) error{
		BatchUpdateTypeUserQuota:        b.userRepo.IncreaseUserQuota,
		BatchUpdateTypeTokenQuota:       b.tokenRepo.IncreaseQuota,
		BatchUpdateTypeUsedQuota:        b.userUpdater.UpdateUserUsedQuota,
		BatchUpdateTypeRequestCount:     b.userUpdater.UpdateUserRequestCount,
		BatchUpdateTypeChannelUsedQuota: b.channelRepo.UpdateChannelUsedQuota,
	}

	// 开始遍历执行
	for targetType, handler := range targetTypeHandlers {
		hashKey := b.hashKey(targetType)
		err := b.scanAndHandler(hashKey, handler)
		if err != nil {
			log.Printf("failed to scan and handle target_type %d error:%v", targetType, err)
		}
	}

	slog.Info("batch update finished")
}

func (b *BatchUpdater) scanAndHandler(hashKey string, handler func(int, int64) error) error {
	ctx := context.Background()

	var cursor uint64 = 0 // 游标从 0 开始

	for {
		// 执行 HScan 命令，获取一批数据和新游标
		// 匹配所有 field，每次返回约 100 个
		keys, nextCursor, err := b.redisClient.HScan(ctx, hashKey, cursor, "*", 100).Result()
		// fmt.Println("keys:", keys)
		if err != nil {
			log.Printf("failed to scan hash key: %s error:%s", hashKey, err.Error())
			return err
		}

		// 处理返回的 field-value 对 (返回的 keys 切片是交替的)
		for i := 0; i < len(keys); i += 2 {
			field := keys[i]
			value := keys[i+1]
			// fmt.Printf("field: %s, value: %s\n", field, value)

			// 开始处理
			id, err2 := strconv.Atoi(field)
			if err2 != nil {
				log.Printf("failed to convert %s to number err: %s", field, err2.Error())
				continue
			}

			count, err2 := strconv.ParseInt(value, 10, 64)
			if err2 != nil {
				log.Printf("failed to convert field:%s value:%s to number err: %s", field, value, err2.Error())
				continue
			}
			if count == 0 {
				continue
			}

			err2 = handler(id, count)
			if err2 != nil {
				log.Printf("failed to handle id:%d value:%d err: %s", id, count, err2.Error())
			} else {
				// 执行成功后需要对应减去对应的count
				count, err2 = b.redisClient.HIncrBy(ctx, hashKey, field, -count).Result()
				if err2 != nil {
					log.Printf("failed to handler:%s hincr id:%d count:-%d err: %s", hashKey, id, count, err2.Error())
				}
			}

		}

		// 更新游标，准备下一次循环
		cursor = nextCursor
		// 当游标归零时，表示遍历完成
		if cursor == 0 {
			break
		}
	}

	return nil
}
