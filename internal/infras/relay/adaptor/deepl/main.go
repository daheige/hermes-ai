package deepl

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/customevent"
	"hermes-ai/internal/infras/ginzo"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/constant"
	"hermes-ai/internal/infras/relay/constant/finishreason"
	"hermes-ai/internal/infras/relay/constant/role"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/utils"
)

// https://developers.deepl.com/docs/getting-started/your-first-api-request

func ConvertRequest(textRequest model2.GeneralOpenAIRequest) (*Request, string) {
	var text string
	if len(textRequest.Messages) != 0 {
		text = textRequest.Messages[len(textRequest.Messages)-1].StringContent()
	}
	deeplRequest := Request{
		TargetLang: parseLangFromModelName(textRequest.Model),
		Text:       []string{text},
	}
	return &deeplRequest, text
}

func StreamResponseDeepL2OpenAI(deeplResponse *Response) *openai2.ChatCompletionsStreamResponse {
	var choice openai2.ChatCompletionsStreamResponseChoice
	if len(deeplResponse.Translations) != 0 {
		choice.Delta.Content = deeplResponse.Translations[0].Text
	}
	choice.Delta.Role = role.Assistant
	choice.FinishReason = &constant.StopFinishReason
	openaiResponse := openai2.ChatCompletionsStreamResponse{
		Object:  constant.StreamObject,
		Created: utils.GetTimestamp(),
		Choices: []openai2.ChatCompletionsStreamResponseChoice{choice},
	}
	return &openaiResponse
}

func ResponseDeepL2OpenAI(deeplResponse *Response) *openai2.TextResponse {
	var responseText string
	if len(deeplResponse.Translations) != 0 {
		responseText = deeplResponse.Translations[0].Text
	}
	choice := openai2.TextResponseChoice{
		Index: 0,
		Message: model2.Message{
			Role:    role.Assistant,
			Content: responseText,
			Name:    nil,
		},
		FinishReason: finishreason.Stop,
	}
	fullTextResponse := openai2.TextResponse{
		Object:  constant.NonStreamObject,
		Created: utils.GetTimestamp(),
		Choices: []openai2.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamHandler(c *gin.Context, resp *http.Response, modelName string) *model2.ErrorWithStatusCode {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	var deeplResponse Response
	err = json.Unmarshal(responseBody, &deeplResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}
	fullTextResponse := StreamResponseDeepL2OpenAI(&deeplResponse)
	fullTextResponse.Model = modelName
	fullTextResponse.Id = ginzo.GetResponseID(c)
	jsonData, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError)
	}
	ginzo.SetEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		if jsonData != nil {
			c.Render(-1, customevent.CustomEvent{Data: "data: " + string(jsonData)})
			jsonData = nil
			return true
		}
		c.Render(-1, customevent.CustomEvent{Data: "data: [DONE]"})
		return false
	})
	_ = resp.Body.Close()
	return nil
}

func Handler(c *gin.Context, resp *http.Response, modelName string) *model2.ErrorWithStatusCode {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	var deeplResponse Response
	err = json.Unmarshal(responseBody, &deeplResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}
	if deeplResponse.Message != "" {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: deeplResponse.Message,
				Code:    "deepl_error",
			},
			StatusCode: resp.StatusCode,
		}
	}
	fullTextResponse := ResponseDeepL2OpenAI(&deeplResponse)
	fullTextResponse.Model = modelName
	fullTextResponse.Id = ginzo.GetResponseID(c)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError)
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil
}
