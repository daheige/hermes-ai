package monitor

import (
	"hermes-ai/internal/infras/config"
)

var (
	store             = make(map[int][]bool)
	metricSuccessChan = make(chan int, config.MetricSuccessChanSize)
	metricFailChan    = make(chan int, config.MetricFailChanSize)
)

func consumeSuccess(channelId int) {
	if len(store[channelId]) > config.MetricQueueSize {
		store[channelId] = store[channelId][1:]
	}
	store[channelId] = append(store[channelId], true)
}

func consumeFail(channelId int) (bool, float64) {
	if len(store[channelId]) > config.MetricQueueSize {
		store[channelId] = store[channelId][1:]
	}
	store[channelId] = append(store[channelId], false)
	successCount := 0
	for _, success := range store[channelId] {
		if success {
			successCount++
		}
	}
	successRate := float64(successCount) / float64(len(store[channelId]))
	if len(store[channelId]) < config.MetricQueueSize {
		return false, successRate
	}
	if successRate < config.MetricSuccessRateThreshold {
		store[channelId] = make([]bool, 0)
		return true, successRate
	}
	return false, successRate
}

func metricSuccessConsumer() {
	for {
		select {
		case channelId := <-metricSuccessChan:
			consumeSuccess(channelId)
		}
	}
}

func metricFailConsumer(channelMonitor *ChannelMonitor) {
	for {
		select {
		case channelId := <-metricFailChan:
			disable, successRate := consumeFail(channelId)
			if disable {
				go channelMonitor.MetricDisableChannel(channelId, successRate)
			}
		}
	}
}

// InitMetric 初始化metrics
func InitMetric(channelMonitor *ChannelMonitor) {
	if config.EnableMetric {
		go metricSuccessConsumer()
		go metricFailConsumer(channelMonitor)
	}
}

func Emit(channelId int, success bool) {
	if !config.EnableMetric {
		return
	}

	go func() {
		if success {
			metricSuccessChan <- channelId
		} else {
			metricFailChan <- channelId
		}
	}()
}
