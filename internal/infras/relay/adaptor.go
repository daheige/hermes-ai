package relay

import (
	"hermes-ai/internal/infras/relay/adaptor"
	"hermes-ai/internal/infras/relay/adaptor/aiproxy"
	"hermes-ai/internal/infras/relay/adaptor/ali"
	"hermes-ai/internal/infras/relay/adaptor/anthropic"
	"hermes-ai/internal/infras/relay/adaptor/aws"
	"hermes-ai/internal/infras/relay/adaptor/baidu"
	"hermes-ai/internal/infras/relay/adaptor/cloudflare"
	"hermes-ai/internal/infras/relay/adaptor/cohere"
	"hermes-ai/internal/infras/relay/adaptor/coze"
	"hermes-ai/internal/infras/relay/adaptor/deepl"
	"hermes-ai/internal/infras/relay/adaptor/gemini"
	"hermes-ai/internal/infras/relay/adaptor/ollama"
	"hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/adaptor/palm"
	"hermes-ai/internal/infras/relay/adaptor/proxy"
	"hermes-ai/internal/infras/relay/adaptor/replicate"
	"hermes-ai/internal/infras/relay/adaptor/tencent"
	"hermes-ai/internal/infras/relay/adaptor/vertexai"
	"hermes-ai/internal/infras/relay/adaptor/xunfei"
	"hermes-ai/internal/infras/relay/adaptor/zhipu"
	"hermes-ai/internal/infras/relay/apitype"
)

func GetAdaptor(apiType int) adaptor.Adaptor {
	switch apiType {
	case apitype.AIProxyLibrary:
		return &aiproxy.Adaptor{}
	case apitype.Ali:
		return &ali.Adaptor{}
	case apitype.Anthropic:
		return &anthropic.Adaptor{}
	case apitype.AwsClaude:
		return &aws.Adaptor{}
	case apitype.Baidu:
		return &baidu.Adaptor{}
	case apitype.Gemini:
		return &gemini.Adaptor{}
	case apitype.OpenAI:
		return &openai.Adaptor{}
	case apitype.PaLM:
		return &palm.Adaptor{}
	case apitype.Tencent:
		return &tencent.Adaptor{}
	case apitype.Xunfei:
		return &xunfei.Adaptor{}
	case apitype.Zhipu:
		return &zhipu.Adaptor{}
	case apitype.Ollama:
		return &ollama.Adaptor{}
	case apitype.Coze:
		return &coze.Adaptor{}
	case apitype.Cohere:
		return &cohere.Adaptor{}
	case apitype.Cloudflare:
		return &cloudflare.Adaptor{}
	case apitype.DeepL:
		return &deepl.Adaptor{}
	case apitype.VertexAI:
		return &vertexai.Adaptor{}
	case apitype.Proxy:
		return &proxy.Adaptor{}
	case apitype.Replicate:
		return &replicate.Adaptor{}
	}
	return nil
}
