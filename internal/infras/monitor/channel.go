package monitor

import (
	"fmt"
	"log/slog"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/message"
)

type ChannelMonitor struct {
	userService    *application.UserService
	channelService *application.ChannelService
	smtpCfg        message.SMTPConfig
	pusherCfg      message.MessagePusherConfig
	systemName     string
	rootEmail      string
}

type ChannelMonitorDeps struct {
	UserService    *application.UserService
	ChannelService *application.ChannelService
}

type ChannelMonitorConfig struct {
	SmtpCfg    message.SMTPConfig
	PusherCfg  message.MessagePusherConfig
	SystemName string
	RootEmail  string
}

func NewChannelMonitor(deps ChannelMonitorDeps, cfg ChannelMonitorConfig) *ChannelMonitor {
	return &ChannelMonitor{
		userService:    deps.UserService,
		channelService: deps.ChannelService,
		smtpCfg:        cfg.SmtpCfg,
		pusherCfg:      cfg.PusherCfg,
		systemName:     cfg.SystemName,
		rootEmail:      cfg.RootEmail,
	}
}

func (monitor *ChannelMonitor) notifyRootUser(subject string, content string) {
	if monitor.pusherCfg.Address != "" {
		err := message.SendMessagePusher(monitor.pusherCfg, message.MessageContent{
			Title:       subject,
			Description: content,
			Content:     content,
		})
		if err != nil {
			slog.Error(fmt.Sprintf("failed to send message: %s", err.Error()))
		} else {
			return
		}
	}

	err := message.SendEmail(monitor.smtpCfg, message.EmailMessage{
		SystemName: monitor.systemName,
		Subject:    subject,
		Receiver:   monitor.userService.GetRootUserEmail(),
		Content:    content,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

// NotifyRootUser 通知root用户
func (monitor *ChannelMonitor) NotifyRootUser(subject string, content string) {
	monitor.notifyRootUser(subject, content)
}

// DisableChannel disable & notify
func (monitor *ChannelMonitor) DisableChannel(channelId int, channelName string, reason string) {
	monitor.channelService.UpdateChannelStatusById(channelId, entity.ChannelStatusAutoDisabled)
	slog.Info(fmt.Sprintf("channel #%d has been disabled: %s", channelId, reason))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message.EmailTemplate(monitor.systemName,
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

type MetricDisableParams struct {
	ChannelId            int
	SuccessRate          float64
	QueueSize            int
	SuccessRateThreshold float64
}

func (monitor *ChannelMonitor) MetricDisableChannel(p MetricDisableParams) {
	monitor.channelService.UpdateChannelStatusById(p.ChannelId, entity.ChannelStatusAutoDisabled)
	slog.Info(fmt.Sprintf("channel #%d has been disabled due to low success rate: %.2f", p.ChannelId, p.SuccessRate*100))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message.EmailTemplate(monitor.systemName,
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道 #%d 已被系统自动禁用。</p>
			<p>禁用原因：</p>
			<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px;">该渠道在最近 %d 次调用中成功率为 <strong>%.2f%%</strong>，低于系统阈值 <strong>%.2f%%</strong>。</p>
		`, p.ChannelId, p.QueueSize, p.SuccessRate*100, p.SuccessRateThreshold*100),
	)
	monitor.notifyRootUser(subject, content)
}

// EnableChannel enable & notify
func (monitor *ChannelMonitor) EnableChannel(channelId int, channelName string) {
	monitor.channelService.UpdateChannelStatusById(channelId, entity.ChannelStatusEnabled)
	slog.Info(fmt.Sprintf("channel #%d has been enabled", channelId))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message.EmailTemplate(monitor.systemName,
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道「<strong>%s</strong>」（#%d）已被重新启用。</p>
			<p>您现在可以继续使用该渠道了。</p>
		`, channelName, channelId),
	)
	monitor.notifyRootUser(subject, content)
}
