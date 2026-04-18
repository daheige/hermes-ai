package vertexai

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/relay/adaptor/vertexai/claude"
	gemini "hermes-ai/internal/infras/relay/adaptor/vertexai/gemini"
	"hermes-ai/internal/infras/relay/meta"
	model2 "hermes-ai/internal/infras/relay/model"
)

type VertexAIModelType int

const (
	VerterAIClaude VertexAIModelType = iota + 1
	VerterAIGemini
)

var modelMapping = map[string]VertexAIModelType{}
var modelList = []string{}

func init() {
	modelList = append(modelList, vertexai.ModelList...)
	for _, model := range vertexai.ModelList {
		modelMapping[model] = VerterAIClaude
	}

	modelList = append(modelList, gemini.ModelList...)
	for _, model := range gemini.ModelList {
		modelMapping[model] = VerterAIGemini
	}
}

type innerAIAdapter interface {
	ConvertRequest(c *gin.Context, relayMode int, request *model2.GeneralOpenAIRequest) (any, error)
	DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model2.Usage, err *model2.ErrorWithStatusCode)
}

func GetAdaptor(model string) innerAIAdapter {
	adaptorType := modelMapping[model]
	switch adaptorType {
	case VerterAIClaude:
		return &vertexai.Adaptor{}
	case VerterAIGemini:
		return &gemini.Adaptor{}
	default:
		return nil
	}
}
