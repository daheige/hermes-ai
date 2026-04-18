package application

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/domain/repo"
	"hermes-ai/internal/infras/persistence"
	"hermes-ai/internal/infras/utils"
)

// ChannelService 渠道服务
type ChannelService struct {
	channelRepo repo.ChannelRepository
	abilityRepo repo.AbilityRepository
	cacheRepo   repo.CacheRepository

	// 内存缓存
	group2model2channels map[string]map[string][]*entity.Channel
	channelSyncLock      sync.RWMutex

	batchUpdater       *BatchUpdater
	batchUpdateEnabled bool
	syncFrequency      int
	cacheEnabled       bool
}

// NewChannelService 创建渠道服务
func NewChannelService(
	channelRepo repo.ChannelRepository,
	abilityRepo repo.AbilityRepository,
	cacheRepo repo.CacheRepository,
	batchUpdater *BatchUpdater,
	batchUpdateEnabled bool,
	syncFrequency int,
	cacheEnabled bool,
) *ChannelService {
	return &ChannelService{
		channelRepo:        channelRepo,
		abilityRepo:        abilityRepo,
		cacheRepo:          cacheRepo,
		batchUpdater:       batchUpdater,
		batchUpdateEnabled: batchUpdateEnabled,
		syncFrequency:      syncFrequency,
		cacheEnabled:       cacheEnabled,
	}
}

// GetAllChannels 获取所有渠道
func (s *ChannelService) GetAllChannels(startIdx int, num int, scope string) ([]*entity.Channel, error) {
	return s.channelRepo.GetAllChannels(startIdx, num, scope)
}

// SearchChannels 搜索渠道
func (s *ChannelService) SearchChannels(keyword string) ([]*entity.Channel, error) {
	return s.channelRepo.SearchChannels(keyword)
}

// GetChannelById 根据ID获取渠道
func (s *ChannelService) GetChannelById(id int, selectAll bool) (*entity.Channel, error) {
	return s.channelRepo.GetChannelById(id, selectAll)
}

// BatchInsertChannels 批量插入渠道
func (s *ChannelService) BatchInsertChannels(channels []entity.Channel) error {
	err := s.channelRepo.BatchInsert(channels)
	if err != nil {
		return err
	}
	for i := range channels {
		abilities := persistence.BuildAbilities(&channels[i])
		err = s.abilityRepo.BatchCreate(abilities)
		if err != nil {
			return err
		}
	}
	return nil
}

// Insert 插入渠道
func (s *ChannelService) Insert(channel *entity.Channel) error {
	err := s.channelRepo.Insert(channel)
	if err != nil {
		return err
	}
	abilities := persistence.BuildAbilities(channel)
	return s.abilityRepo.BatchCreate(abilities)
}

// Update 更新渠道
func (s *ChannelService) Update(channel *entity.Channel) error {
	err := s.channelRepo.Update(channel)
	if err != nil {
		return err
	}
	// 更新能力: 先删除旧的，再添加新的
	err = s.abilityRepo.DeleteByChannelId(channel.Id)
	if err != nil {
		return err
	}
	abilities := persistence.BuildAbilities(channel)
	return s.abilityRepo.BatchCreate(abilities)
}

// Delete 删除渠道
func (s *ChannelService) Delete(id int) error {
	err := s.channelRepo.Delete(id)
	if err != nil {
		return err
	}
	return s.abilityRepo.DeleteByChannelId(id)
}

// UpdateResponseTime 更新渠道响应时间
func (s *ChannelService) UpdateResponseTime(id int, responseTime int64) {
	s.channelRepo.UpdateResponseTime(id, responseTime)
}

// UpdateBalance 更新渠道余额
func (s *ChannelService) UpdateBalance(id int, balance float64) {
	s.channelRepo.UpdateBalance(id, balance)
}

// UpdateChannelStatusById 更新渠道状态
func (s *ChannelService) UpdateChannelStatusById(id int, status int) {
	err := s.abilityRepo.UpdateAbilityStatus(id, status == entity.ChannelStatusEnabled)
	if err != nil {
		slog.Error("failed to update ability status: " + err.Error())
	}
	s.channelRepo.UpdateChannelStatusById(id, status)
}

// UpdateChannelUsedQuota 更新渠道已用配额
func (s *ChannelService) UpdateChannelUsedQuota(id int, quota int64) {
	if s.batchUpdateEnabled && s.batchUpdater != nil {
		s.batchUpdater.AddRecord(BatchUpdateTypeChannelUsedQuota, id, quota)
		return
	}

	s.channelRepo.UpdateChannelUsedQuota(id, quota)
}

// DeleteChannelByStatus 根据状态删除渠道
func (s *ChannelService) DeleteChannelByStatus(status int64) (int64, error) {
	return s.channelRepo.DeleteChannelByStatus(status)
}

// DeleteDisabledChannel 删除禁用的渠道
func (s *ChannelService) DeleteDisabledChannel() (int64, error) {
	return s.channelRepo.DeleteDisabledChannel()
}

// GetRandomSatisfiedChannel 随机获取满足条件的渠道(直接数据库查询)
func (s *ChannelService) GetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*entity.Channel, error) {
	return s.abilityRepo.GetRandomSatisfiedChannel(group, model, ignoreFirstPriority)
}

// UpdateAbilityStatus 更新渠道能力状态
func (s *ChannelService) UpdateAbilityStatus(channelId int, status bool) error {
	return s.abilityRepo.UpdateAbilityStatus(channelId, status)
}

// GetGroupModels 获取分组可用模型列表
func (s *ChannelService) GetGroupModels(ctx context.Context, group string) ([]string, error) {
	return s.abilityRepo.GetGroupModels(ctx, group)
}

// CacheGetGroupModels 带缓存的获取分组可用模型列表
func (s *ChannelService) CacheGetGroupModels(ctx context.Context, group string) ([]string, error) {
	if !s.cacheRepo.IsEnabled() {
		return s.GetGroupModels(ctx, group)
	}
	modelsStr, err := s.cacheRepo.Get("group_models:" + group)
	if err == nil {
		return strings.Split(modelsStr, ","), nil
	}
	models, err := s.GetGroupModels(ctx, group)
	if err != nil {
		return nil, err
	}
	cacheErr := s.cacheRepo.Set("group_models:"+group, strings.Join(models, ","),
		time.Duration(s.syncFrequency)*time.Second)
	if cacheErr != nil {
		slog.Error("Redis set group models error: " + cacheErr.Error())
	}
	return models, nil
}

// InitChannelCache 初始化渠道内存缓存
func (s *ChannelService) InitChannelCache() {
	channels, err := s.channelRepo.GetEnabledChannels()
	if err != nil {
		slog.Error("failed to get enabled channels: " + err.Error())
		return
	}

	abilities, err := s.abilityRepo.GetAll()
	if err != nil {
		slog.Error("failed to get abilities: " + err.Error())
		return
	}

	groups := make(map[string]bool)
	for _, ability := range abilities {
		groups[ability.Group] = true
	}

	newGroup2model2channels := make(map[string]map[string][]*entity.Channel)
	for group := range groups {
		newGroup2model2channels[group] = make(map[string][]*entity.Channel)
	}

	for _, channel := range channels {
		channelGroups := strings.Split(channel.Group, ",")
		for _, group := range channelGroups {
			models := strings.Split(channel.Models, ",")
			for _, model := range models {
				if _, ok := newGroup2model2channels[group]; !ok {
					newGroup2model2channels[group] = make(map[string][]*entity.Channel)
				}
				newGroup2model2channels[group][model] = append(newGroup2model2channels[group][model], channel)
			}
		}
	}

	// 按优先级排序
	for group, model2channels := range newGroup2model2channels {
		for model, chs := range model2channels {
			sort.Slice(chs, func(i, j int) bool {
				return chs[i].GetPriority() > chs[j].GetPriority()
			})
			newGroup2model2channels[group][model] = chs
		}
	}

	s.channelSyncLock.Lock()
	s.group2model2channels = newGroup2model2channels
	s.channelSyncLock.Unlock()
	slog.Info("channels synced from database")
}

// SyncChannelCache 同步渠道缓存
func (s *ChannelService) SyncChannelCache(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Second)
		slog.Info("syncing channels from database")
		s.InitChannelCache()
	}
}

// CacheGetRandomSatisfiedChannel 带内存缓存的随机获取满足条件的渠道
func (s *ChannelService) CacheGetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*entity.Channel, error) {
	if !s.cacheEnabled {
		return s.GetRandomSatisfiedChannel(group, model, ignoreFirstPriority)
	}

	s.channelSyncLock.RLock()
	defer s.channelSyncLock.RUnlock()

	channels := s.group2model2channels[group][model]
	if len(channels) == 0 {
		return nil, errors.New("channel not found")
	}

	endIdx := len(channels)
	firstChannel := channels[0]
	if firstChannel.GetPriority() > 0 {
		for i := range channels {
			if channels[i].GetPriority() != firstChannel.GetPriority() {
				endIdx = i
				break
			}
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(endIdx)
	if ignoreFirstPriority {
		if endIdx < len(channels) {
			idx = utils.RandRange(endIdx, len(channels))
		}
	}

	return channels[idx], nil
}
