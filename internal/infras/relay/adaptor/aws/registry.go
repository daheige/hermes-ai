package aws

import (
	"hermes-ai/internal/infras/relay/adaptor/aws/claude"
	aws2 "hermes-ai/internal/infras/relay/adaptor/aws/llama3"
	"hermes-ai/internal/infras/relay/adaptor/aws/utils"
)

type AwsModelType int

const (
	AwsClaude AwsModelType = iota + 1
	AwsLlama3
)

var adaptors = initAdaptors()

func initAdaptors() map[string]AwsModelType {
	m := map[string]AwsModelType{}
	for model := range aws.AwsModelIDMap {
		m[model] = AwsClaude
	}
	for model := range aws2.AwsModelIDMap {
		m[model] = AwsLlama3
	}

	return m
}

func GetAdaptor(model string) utils.AwsAdapter {
	adaptorType := adaptors[model]
	switch adaptorType {
	case AwsClaude:
		return &aws.Adaptor{}
	case AwsLlama3:
		return &aws2.Adaptor{}
	default:
		return nil
	}
}
