package openai

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math"
	"strings"

	"github.com/pkoukk/tiktoken-go"

	"hermes-ai/internal/infras/image"
	billingratio "hermes-ai/internal/infras/relay/billing/ratio"
	"hermes-ai/internal/infras/relay/model"
)

// TokenCounter 封装 token 计数逻辑，避免包级状态
type TokenCounter struct {
	tokenEncoderMap         map[string]*tiktoken.Tiktoken
	defaultTokenEncoder     *tiktoken.Tiktoken
	approximateTokenEnabled bool
}

// NewTokenCounter 创建 TokenCounter 并初始化编码器缓存
func NewTokenCounter(approximateTokenEnabled bool) *TokenCounter {
	tc := &TokenCounter{
		tokenEncoderMap:         make(map[string]*tiktoken.Tiktoken),
		approximateTokenEnabled: approximateTokenEnabled,
	}

	slog.Info("initializing token encoders")
	gpt35TokenEncoder, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		log.Fatalln(fmt.Sprintf("failed to get gpt-3.5-turbo token encoder: %s, "+
			"if you are using in offline environment, please set TIKTOKEN_CACHE_DIR to use exsited files, check this link for more information: https://stackoverflow.com/questions/76106366/how-to-use-tiktoken-in-offline-mode-computer ", err.Error()))
	}

	tc.defaultTokenEncoder = gpt35TokenEncoder
	gpt4oTokenEncoder, err := tiktoken.EncodingForModel("gpt-4o")
	if err != nil {
		log.Fatalln(fmt.Sprintf("failed to get gpt-4o token encoder: %s", err.Error()))
	}
	gpt4TokenEncoder, err := tiktoken.EncodingForModel("gpt-4")
	if err != nil {
		log.Fatalln(fmt.Sprintf("failed to get gpt-4 token encoder: %s", err.Error()))
	}
	for model := range billingratio.ModelRatio {
		if strings.HasPrefix(model, "gpt-3.5") {
			tc.tokenEncoderMap[model] = gpt35TokenEncoder
		} else if strings.HasPrefix(model, "gpt-4o") {
			tc.tokenEncoderMap[model] = gpt4oTokenEncoder
		} else if strings.HasPrefix(model, "gpt-4") {
			tc.tokenEncoderMap[model] = gpt4TokenEncoder
		} else {
			tc.tokenEncoderMap[model] = nil
		}
	}
	slog.Info("token encoders initialized")
	return tc
}

func (tc *TokenCounter) getTokenEncoder(model string) *tiktoken.Tiktoken {
	tokenEncoder, ok := tc.tokenEncoderMap[model]
	if ok && tokenEncoder != nil {
		return tokenEncoder
	}
	if ok {
		tokenEncoder, err := tiktoken.EncodingForModel(model)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to get token encoder for model %s: %s, using encoder for gpt-3.5-turbo", model, err.Error()))
			tokenEncoder = tc.defaultTokenEncoder
		}
		tc.tokenEncoderMap[model] = tokenEncoder
		return tokenEncoder
	}
	return tc.defaultTokenEncoder
}

func (tc *TokenCounter) getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
	if tc.approximateTokenEnabled {
		return int(float64(len(text)) * 0.38)
	}
	return len(tokenEncoder.Encode(text, nil, nil))
}

func (tc *TokenCounter) CountTokenMessages(messages []model.Message, model string) int {
	tokenEncoder := tc.getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	var tokensPerName int
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		switch v := message.Content.(type) {
		case string:
			tokenNum += tc.getTokenNum(tokenEncoder, v)
		case []any:
			for _, it := range v {
				m := it.(map[string]any)
				switch m["type"] {
				case "text":
					if textValue, ok := m["text"]; ok {
						if textString, ok := textValue.(string); ok {
							tokenNum += tc.getTokenNum(tokenEncoder, textString)
						}
					}
				case "image_url":
					imageUrl, ok := m["image_url"].(map[string]any)
					if ok {
						url := imageUrl["url"].(string)
						detail := ""
						if imageUrl["detail"] != nil {
							detail = imageUrl["detail"].(string)
						}
						imageTokens, err := countImageTokens(url, detail, model)
						if err != nil {
							slog.Error("error counting image tokens: " + err.Error())
						} else {
							tokenNum += imageTokens
						}
					}
				}
			}
		}
		tokenNum += tc.getTokenNum(tokenEncoder, message.Role)
		if message.Name != nil {
			tokenNum += tokensPerName
			tokenNum += tc.getTokenNum(tokenEncoder, *message.Name)
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum
}

const (
	lowDetailCost         = 85
	highDetailCostPerTile = 170
	additionalCost        = 85
	// gpt-4o-mini cost higher than other model
	gpt4oMiniLowDetailCost  = 2833
	gpt4oMiniHighDetailCost = 5667
	gpt4oMiniAdditionalCost = 2833
)

// https://platform.openai.com/docs/guides/vision/calculating-costs
// https://github.com/openai/openai-cookbook/blob/05e3f9be4c7a2ae7ecf029a7c32065b024730ebe/examples/How_to_count_tokens_with_tiktoken.ipynb
func countImageTokens(url string, detail string, model string) (_ int, err error) {
	var fetchSize = true
	var width, height int
	// Reference: https://platform.openai.com/docs/guides/vision/low-or-high-fidelity-image-understanding
	// detail == "auto" is undocumented on how it works, it just said the model will use the auto setting which will look at the image input size and decide if it should use the low or high setting.
	// According to the official guide, "low" disable the high-res model,
	// and only receive low-res 512px x 512px version of the image, indicating
	// that image is treated as low-res when size is smaller than 512px x 512px,
	// then we can assume that image size larger than 512px x 512px is treated
	// as high-res. Then we have the following logic:
	// if detail == "" || detail == "auto" {
	// 	width, height, err = image.GetImageSize(url)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// 	fetchSize = false
	// 	// not sure if this is correct
	// 	if width > 512 || height > 512 {
	// 		detail = "high"
	// 	} else {
	// 		detail = "low"
	// 	}
	// }

	// However, in my test, it seems to be always the same as "high".
	// The following image, which is 125x50, is still treated as high-res, taken
	// 255 tokens in the response of non-stream chat completion api.
	// https://upload.wikimedia.org/wikipedia/commons/1/10/18_Infantry_Division_Messina.jpg
	if detail == "" || detail == "auto" {
		// assume by test, not sure if this is correct
		detail = "high"
	}
	switch detail {
	case "low":
		if strings.HasPrefix(model, "gpt-4o-mini") {
			return gpt4oMiniLowDetailCost, nil
		}
		return lowDetailCost, nil
	case "high":
		if fetchSize {
			width, height, err = image.GetImageSize(url)
			if err != nil {
				return 0, err
			}
		}
		if width > 2048 || height > 2048 { // max(width, height) > 2048
			ratio := float64(2048) / math.Max(float64(width), float64(height))
			width = int(float64(width) * ratio)
			height = int(float64(height) * ratio)
		}
		if width > 768 && height > 768 { // min(width, height) > 768
			ratio := float64(768) / math.Min(float64(width), float64(height))
			width = int(float64(width) * ratio)
			height = int(float64(height) * ratio)
		}
		numSquares := int(math.Ceil(float64(width)/512) * math.Ceil(float64(height)/512))
		if strings.HasPrefix(model, "gpt-4o-mini") {
			return numSquares*gpt4oMiniHighDetailCost + gpt4oMiniAdditionalCost, nil
		}
		result := numSquares*highDetailCostPerTile + additionalCost
		return result, nil
	default:
		return 0, errors.New("invalid detail option")
	}
}

func (tc *TokenCounter) CountTokenInput(input any, model string) int {
	switch v := input.(type) {
	case string:
		return tc.CountTokenText(v, model)
	case []string:
		text := ""
		for _, s := range v {
			text += s
		}
		return tc.CountTokenText(text, model)
	}
	return 0
}

func (tc *TokenCounter) CountTokenText(text string, model string) int {
	tokenEncoder := tc.getTokenEncoder(model)
	return tc.getTokenNum(tokenEncoder, text)
}

func (tc *TokenCounter) CountToken(text string) int {
	return tc.CountTokenInput(text, "gpt-3.5-turbo")
}
