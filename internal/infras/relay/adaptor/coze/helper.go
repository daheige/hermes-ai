package coze

import (
	"hermes-ai/internal/infras/relay/adaptor/coze/constant/event"
)

func event2StopReason(e *string) string {
	if e == nil || *e == event.Message {
		return ""
	}
	return "stop"
}
