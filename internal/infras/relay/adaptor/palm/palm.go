package palm

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"hermes-ai/internal/infras/ginzo"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/constant"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/render"
	"hermes-ai/internal/infras/utils"

	"github.com/gin-gonic/gin"
)

// https://developers.generativeai.google/api/rest/generativelanguage/models/generateMessage#request-body
// https://developers.generativeai.google/api/rest/generativelanguage/models/generateMessage#response-body

func ConvertRequest(textRequest model2.GeneralOpenAIRequest) *ChatRequest {
	palmRequest := ChatRequest{
		Prompt: Prompt{
			Messages: make([]ChatMessage, 0, len(textRequest.Messages)),
		},
		Temperature:    textRequest.Temperature,
		CandidateCount: textRequest.N,
		TopP:           textRequest.TopP,
		TopK:           textRequest.MaxTokens,
	}
	for _, message := range textRequest.Messages {
		palmMessage := ChatMessage{
			Content: message.StringContent(),
		}
		if message.Role == "user" {
			palmMessage.Author = "0"
		} else {
			palmMessage.Author = "1"
		}
		palmRequest.Prompt.Messages = append(palmRequest.Prompt.Messages, palmMessage)
	}
	return &palmRequest
}

func responsePaLM2OpenAI(response *ChatResponse) *openai2.TextResponse {
	fullTextResponse := openai2.TextResponse{
		Choices: make([]openai2.TextResponseChoice, 0, len(response.Candidates)),
	}
	for i, candidate := range response.Candidates {
		choice := openai2.TextResponseChoice{
			Index: i,
			Message: model2.Message{
				Role:    "assistant",
				Content: candidate.Content,
			},
			FinishReason: "stop",
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponsePaLM2OpenAI(palmResponse *ChatResponse) *openai2.ChatCompletionsStreamResponse {
	var choice openai2.ChatCompletionsStreamResponseChoice
	if len(palmResponse.Candidates) > 0 {
		choice.Delta.Content = palmResponse.Candidates[0].Content
	}
	choice.FinishReason = &constant.StopFinishReason
	var response openai2.ChatCompletionsStreamResponse
	response.Object = "chat.completion.chunk"
	response.Model = "palm2"
	response.Choices = []openai2.ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, string) {
	responseText := ""
	responseId := fmt.Sprintf("chatcmpl-%s", utils.UUID())
	createdTime := utils.GetTimestamp()

	ginzo.SetEventStreamHeaders(c)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading stream response: " + err.Error())
		err := resp.Body.Close()
		if err != nil {
			return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
		}
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), ""
	}

	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}

	var palmResponse ChatResponse
	err = json.Unmarshal(responseBody, &palmResponse)
	if err != nil {
		slog.Error("error unmarshalling stream response: " + err.Error())
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), ""
	}

	fullTextResponse := streamResponsePaLM2OpenAI(&palmResponse)
	fullTextResponse.Id = responseId
	fullTextResponse.Created = createdTime
	if len(palmResponse.Candidates) > 0 {
		responseText = palmResponse.Candidates[0].Content
	}

	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		slog.Error("error marshalling stream response: " + err.Error())
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), ""
	}

	err = render.ObjectData(c, string(jsonResponse))
	if err != nil {
		slog.Error(err.Error())
	}

	render.Done(c)

	return nil, responseText
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model2.ErrorWithStatusCode, *model2.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var palmResponse ChatResponse
	err = json.Unmarshal(responseBody, &palmResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if palmResponse.Error.Code != 0 || len(palmResponse.Candidates) == 0 {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: palmResponse.Error.Message,
				Type:    palmResponse.Error.Status,
				Param:   "",
				Code:    palmResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responsePaLM2OpenAI(&palmResponse)
	fullTextResponse.Model = modelName
	completionTokens := openai2.CountTokenText(palmResponse.Candidates[0].Content, modelName)
	usage := model2.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}
