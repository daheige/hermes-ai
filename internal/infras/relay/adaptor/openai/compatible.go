package openai

import (
	"hermes-ai/internal/infras/relay/adaptor/ai360"
	"hermes-ai/internal/infras/relay/adaptor/alibailian"
	"hermes-ai/internal/infras/relay/adaptor/baichuan"
	"hermes-ai/internal/infras/relay/adaptor/baiduv2"
	"hermes-ai/internal/infras/relay/adaptor/deepseek"
	"hermes-ai/internal/infras/relay/adaptor/doubao"
	"hermes-ai/internal/infras/relay/adaptor/geminiv2"
	"hermes-ai/internal/infras/relay/adaptor/groq"
	"hermes-ai/internal/infras/relay/adaptor/lingyiwanwu"
	"hermes-ai/internal/infras/relay/adaptor/minimax"
	"hermes-ai/internal/infras/relay/adaptor/mistral"
	"hermes-ai/internal/infras/relay/adaptor/moonshot"
	"hermes-ai/internal/infras/relay/adaptor/novita"
	"hermes-ai/internal/infras/relay/adaptor/openrouter"
	"hermes-ai/internal/infras/relay/adaptor/siliconflow"
	"hermes-ai/internal/infras/relay/adaptor/stepfun"
	"hermes-ai/internal/infras/relay/adaptor/togetherai"
	"hermes-ai/internal/infras/relay/adaptor/xai"
	"hermes-ai/internal/infras/relay/adaptor/xunfeiv2"
	"hermes-ai/internal/infras/relay/channeltype"
)

var CompatibleChannels = []int{
	channeltype.Azure,
	channeltype.AI360,
	channeltype.Moonshot,
	channeltype.Baichuan,
	channeltype.Minimax,
	channeltype.Doubao,
	channeltype.Mistral,
	channeltype.Groq,
	channeltype.LingYiWanWu,
	channeltype.StepFun,
	channeltype.DeepSeek,
	channeltype.TogetherAI,
	channeltype.Novita,
	channeltype.SiliconFlow,
	channeltype.XAI,
	channeltype.BaiduV2,
	channeltype.XunfeiV2,
}

func GetCompatibleChannelMeta(channelType int) (string, []string) {
	switch channelType {
	case channeltype.Azure:
		return "azure", ModelList
	case channeltype.AI360:
		return "360", ai360.ModelList
	case channeltype.Moonshot:
		return "moonshot", moonshot.ModelList
	case channeltype.Baichuan:
		return "baichuan", baichuan.ModelList
	case channeltype.Minimax:
		return "minimax", minimax.ModelList
	case channeltype.Mistral:
		return "mistralai", mistral.ModelList
	case channeltype.Groq:
		return "groq", groq.ModelList
	case channeltype.LingYiWanWu:
		return "lingyiwanwu", lingyiwanwu.ModelList
	case channeltype.StepFun:
		return "stepfun", stepfun.ModelList
	case channeltype.DeepSeek:
		return "deepseek", deepseek.ModelList
	case channeltype.TogetherAI:
		return "together.ai", togetherai.ModelList
	case channeltype.Doubao:
		return "doubao", doubao.ModelList
	case channeltype.Novita:
		return "novita", novita.ModelList
	case channeltype.SiliconFlow:
		return "siliconflow", siliconflow.ModelList
	case channeltype.XAI:
		return "xai", xai.ModelList
	case channeltype.BaiduV2:
		return "baiduv2", baiduv2.ModelList
	case channeltype.XunfeiV2:
		return "xunfeiv2", xunfeiv2.ModelList
	case channeltype.OpenRouter:
		return "openrouter", openrouter.ModelList
	case channeltype.AliBailian:
		return "alibailian", alibailian.ModelList
	case channeltype.GeminiOpenAICompatible:
		return "geminiv2", geminiv2.ModelList
	default:
		return "openai", ModelList
	}
}
