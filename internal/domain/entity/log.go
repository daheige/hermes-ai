package entity

const (
	LogTypeUnknown = iota
	LogTypeTopup
	LogTypeConsume
	LogTypeManage
	LogTypeSystem
	LogTypeTest
)

type Log struct {
	Id                int    `json:"id"`
	UserId            int    `json:"user_id" gorm:"index"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint;index:idx_created_at_type"`
	Type              int    `json:"type" gorm:"index:idx_created_at_type"`
	Content           string `json:"content"`
	Username          string `json:"username" gorm:"index:index_username_model_name,priority:2;default:''"`
	TokenName         string `json:"token_name" gorm:"index;default:''"`
	ModelName         string `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Quota             int    `json:"quota" gorm:"default:0"`
	PromptTokens      int    `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens  int    `json:"completion_tokens" gorm:"default:0"`
	ChannelId         int    `json:"channel" gorm:"index"`
	RequestId         string `json:"request_id" gorm:"default:''"`
	ElapsedTime       int64  `json:"elapsed_time" gorm:"default:0"` // unit is ms
	IsStream          bool   `json:"is_stream" gorm:"default:false"`
	SystemPromptReset bool   `json:"system_prompt_reset" gorm:"default:false"`
}

type LogStatistic struct {
	Day              string `json:"Day" gorm:"column:day"`
	ModelName        string `json:"ModelName" gorm:"column:model_name"`
	RequestCount     int    `json:"RequestCount" gorm:"column:request_count"`
	Quota            int    `json:"Quota" gorm:"column:quota"`
	PromptTokens     int    `json:"PromptTokens" gorm:"column:prompt_tokens"`
	CompletionTokens int    `json:"CompletionTokens" gorm:"column:completion_tokens"`
}

type LogQueryParams struct {
	LogType int
	Limit   int
	Offset  int

	// 可选参数
	StartTimestamp int64
	EndTimestamp   int64
	ModelName      string
	Username       string
	TokenName      string
	Channel        int
}

type LogUserQueryParams struct {
	UserId  int
	LogType int
	Limit   int
	Offset  int

	// 可选参数
	StartTimestamp int64
	EndTimestamp   int64
	ModelName      string
	TokenName      string
}

type LogUsedQuotaQueryParams struct {
	LogType        int
	StartTimestamp int64
	EndTimestamp   int64
	ModelName      string
	Username       string
	TokenName      string
	Channel        int
}

type LogUsedTokenQueryParams struct {
	LogType        int
	StartTimestamp int64
	EndTimestamp   int64
	ModelName      string
	Username       string
	TokenName      string
}

const LogTable = "logs"

func (Log) TableName() string {
	return LogTable
}
