package monitor

type MetricCollector struct {
	store                  map[int][]bool
	metricSuccessChan      chan int
	metricFailChan         chan int
	queueSize              int
	successRateThreshold   float64
	channelMonitor         *ChannelMonitor
}

func NewMetricCollector(channelMonitor *ChannelMonitor, queueSize int, successRateThreshold float64, successChanSize, failChanSize int) *MetricCollector {
	mc := &MetricCollector{
		store:                make(map[int][]bool),
		metricSuccessChan:    make(chan int, successChanSize),
		metricFailChan:       make(chan int, failChanSize),
		queueSize:            queueSize,
		successRateThreshold: successRateThreshold,
		channelMonitor:       channelMonitor,
	}
	go mc.consumeSuccess()
	go mc.consumeFail()
	return mc
}

func (mc *MetricCollector) consumeSuccess() {
	for {
		select {
		case channelId := <-mc.metricSuccessChan:
			if len(mc.store[channelId]) > mc.queueSize {
				mc.store[channelId] = mc.store[channelId][1:]
			}
			mc.store[channelId] = append(mc.store[channelId], true)
		}
	}
}

func (mc *MetricCollector) consumeFail() {
	for {
		select {
		case channelId := <-mc.metricFailChan:
			if len(mc.store[channelId]) > mc.queueSize {
				mc.store[channelId] = mc.store[channelId][1:]
			}
			mc.store[channelId] = append(mc.store[channelId], false)
			successCount := 0
			for _, success := range mc.store[channelId] {
				if success {
					successCount++
				}
			}
			successRate := float64(successCount) / float64(len(mc.store[channelId]))
			if len(mc.store[channelId]) < mc.queueSize {
				return
			}
			if successRate < mc.successRateThreshold {
				mc.store[channelId] = make([]bool, 0)
				go mc.channelMonitor.MetricDisableChannel(MetricDisableParams{
					ChannelId:            channelId,
					SuccessRate:          successRate,
					QueueSize:            mc.queueSize,
					SuccessRateThreshold: mc.successRateThreshold,
				})
			}
		}
	}
}

func (mc *MetricCollector) Emit(channelId int, success bool) {
	go func() {
		if success {
			mc.metricSuccessChan <- channelId
		} else {
			mc.metricFailChan <- channelId
		}
	}()
}
