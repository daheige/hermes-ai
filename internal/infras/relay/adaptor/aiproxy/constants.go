package aiproxy

import (
	"hermes-ai/internal/infras/relay/adaptor/openai"
)

var ModelList = []string{""}

func init() {
	ModelList = openai.ModelList
}
