package message

import "fmt"

const (
	ByAll           = "all"
	ByEmail         = "email"
	ByMessagePusher = "message_pusher"
)

type NotifyParams struct {
	RootEmail   string
	SmtpCfg     SMTPConfig
	PusherCfg   MessagePusherConfig
	SystemName  string
	By          string
	Title       string
	Description string
	Content     string
}

func Notify(p NotifyParams) error {
	if p.By == ByEmail {
		return SendEmail(p.SmtpCfg, EmailMessage{
			SystemName: p.SystemName,
			Subject:    p.Title,
			Receiver:   p.RootEmail,
			Content:    p.Content,
		})
	}
	if p.By == ByMessagePusher {
		return SendMessagePusher(p.PusherCfg, MessageContent{
			Title:       p.Title,
			Description: p.Description,
			Content:     p.Content,
		})
	}
	return fmt.Errorf("unknown notify method: %s", p.By)
}
