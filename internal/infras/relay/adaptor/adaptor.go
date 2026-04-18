package adaptor

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/relay/meta"
	model2 "hermes-ai/internal/infras/relay/model"
)

type Adaptor interface {
	Init(meta *meta.Meta)
	GetRequestURL(meta *meta.Meta) (string, error)
	SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error
	ConvertRequest(c *gin.Context, relayMode int, request *model2.GeneralOpenAIRequest) (any, error)
	ConvertImageRequest(request *model2.ImageRequest) (any, error)
	DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error)
	DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model2.Usage, err *model2.ErrorWithStatusCode)
	GetModelList() []string
	GetChannelName() string
}
