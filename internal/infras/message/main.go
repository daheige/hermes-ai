package message

import "fmt"

const (
	ByAll           = "all"
	ByEmail         = "email"
	ByMessagePusher = "message_pusher"
)

func Notify(rootEmail string, smtpCfg SMTPConfig, pusherCfg MessagePusherConfig, systemName, by, title, description, content string) error {
	if by == ByEmail {
		return SendEmail(smtpCfg, systemName, title, rootEmail, content)
	}
	if by == ByMessagePusher {
		return SendMessagePusher(pusherCfg, title, description, content)
	}
	return fmt.Errorf("unknown notify method: %s", by)
}
