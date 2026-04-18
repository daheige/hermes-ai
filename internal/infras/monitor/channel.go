package monitor

import (
	"fmt"
	"log/slog"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/config"
	message2 "hermes-ai/internal/infras/message"
)

type ChannelMonitor struct {
	userService    *application.UserService
	channelService *application.ChannelService
}

func NewChannelMonitor(userService *application.UserService, channelService *application.ChannelService) *ChannelMonitor {
	return &ChannelMonitor{userService: userService, channelService: channelService}
}

func (monitor *ChannelMonitor) notifyRootUser(subject string, content string) {
	if config.MessagePusherAddress != "" {
		err := message2.SendMessage(subject, content, content)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to send message: %s", err.Error()))
		} else {
			return
		}
	}

	err := message2.SendEmail(subject, monitor.userService.GetRootUserEmail(), content)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

// DisableChannel disable & notify
func (monitor *ChannelMonitor) DisableChannel(channelId int, channelName string, reason string) {
	monitor.channelService.UpdateChannelStatusById(channelId, entity.ChannelStatusAutoDisabled)
	slog.Info(fmt.Sprintf("channel #%d has been disabled: %s", channelId, reason))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message2.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道「<strong>%s</strong>」（#%d）已被禁用。</p>
			<p>禁用原因：</p>
			<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px;">%s</p>
		`, channelName, channelId, reason),
	)

	monitor.notifyRootUser(subject, content)
}

func (monitor *ChannelMonitor) MetricDisableChannel(channelId int, successRate float64) {
	monitor.channelService.UpdateChannelStatusById(channelId, entity.ChannelStatusAutoDisabled)
	slog.Info(fmt.Sprintf("channel #%d has been disabled due to low success rate: %.2f", channelId, successRate*100))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message2.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道 #%d 已被系统自动禁用。</p>
			<p>禁用原因：</p>
			<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px;">该渠道在最近 %d 次调用中成功率为 <strong>%.2f%%</strong>，低于系统阈值 <strong>%.2f%%</strong>。</p>
		`, channelId, config.MetricQueueSize, successRate*100, config.MetricSuccessRateThreshold*100),
	)
	monitor.notifyRootUser(subject, content)
}

// EnableChannel enable & notify
func (monitor *ChannelMonitor) EnableChannel(channelId int, channelName string) {
	monitor.channelService.UpdateChannelStatusById(channelId, entity.ChannelStatusEnabled)
	slog.Info(fmt.Sprintf("channel #%d has been enabled", channelId))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message2.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道「<strong>%s</strong>」（#%d）已被重新启用。</p>
			<p>您现在可以继续使用该渠道了。</p>
		`, channelName, channelId),
	)
	monitor.notifyRootUser(subject, content)
}
