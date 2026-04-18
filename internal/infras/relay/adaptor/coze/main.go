package coze

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/relay/adaptor/coze/constant/messagetype"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/render"
	"hermes-ai/internal/infras/utils"

	"github.com/gin-gonic/gin"
)

// https://www.coze.com/open

func stopReasonCoze2OpenAI(reason *string) string {
	if reason == nil {
		return ""
	}
	switch *reason {
	case "end_turn":
		return "stop"
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return *reason
	}
}

func ConvertRequest(textRequest model2.GeneralOpenAIRequest) *Request {
	cozeRequest := Request{
		Stream: textRequest.Stream,
		User:   textRequest.User,
		BotId:  strings.TrimPrefix(textRequest.Model, "bot-"),
	}
	for i, message := range textRequest.Messages {
		if i == len(textRequest.Messages)-1 {
			cozeRequest.Query = message.StringContent()
			continue
		}
		cozeMessage := Message{
			Role:    message.Role,
			Content: message.StringContent(),
		}
		cozeRequest.ChatHistory = append(cozeRequest.ChatHistory, cozeMessage)
	}
	return &cozeRequest
}

func StreamResponseCoze2OpenAI(cozeResponse *StreamResponse) (*openai2.ChatCompletionsStreamResponse, *Response) {
	var response *Response
	var stopReason string
	var choice openai2.ChatCompletionsStreamResponseChoice

	if cozeResponse.Message != nil {
		if cozeResponse.Message.Type != messagetype.Answer {
			return nil, nil
		}
		choice.Delta.Content = cozeResponse.Message.Content
	}
	choice.Delta.Role = "assistant"
	finishReason := stopReasonCoze2OpenAI(&stopReason)
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	var openaiResponse openai2.ChatCompletionsStreamResponse
	openaiResponse.Object = "chat.completion.chunk"
	openaiResponse.Choices = []openai2.ChatCompletionsStreamResponseChoice{choice}
	openaiResponse.Id = cozeResponse.ConversationId
	return &openaiResponse, response
}

func ResponseCoze2OpenAI(cozeResponse *Response) *openai2.TextResponse {
	var responseText string
	for _, message := range cozeResponse.Messages {
		if message.Type == messagetype.Answer {
			responseText = message.Content
			break
		}
	}
	choice := openai2.TextResponseChoice{
		Index: 0,
		Message: model2.Message{
			Role:    "assistant",
			Content: responseText,
			Name:    nil,
		},
		FinishReason: "stop",
	}
	fullTextResponse := openai2.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", cozeResponse.ConversationId),
		Model:   "coze-bot",
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Choices: []openai2.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *string) {
	var responseText string
	createdTime := utils.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	ginzo.SetEventStreamHeaders(c)
	var modelName string

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 5 || !strings.HasPrefix(data, "data:") {
			continue
		}
		data = strings.TrimPrefix(data, "data:")
		data = strings.TrimSuffix(data, "\r")

		var cozeResponse StreamResponse
		err := json.Unmarshal([]byte(data), &cozeResponse)
		if err != nil {
			slog.Error("error unmarshalling stream response: " + err.Error())
			continue
		}

		response, _ := StreamResponseCoze2OpenAI(&cozeResponse)
		if response == nil {
			continue
		}

		for _, choice := range response.Choices {
			responseText += utils.AsString(choice.Delta.Content)
		}
		response.Model = modelName
		response.Created = createdTime

		err = render.ObjectData(c, response)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	return nil, &responseText
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model2.ErrorWithStatusCode, *string) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var cozeResponse Response
	err = json.Unmarshal(responseBody, &cozeResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if cozeResponse.Code != 0 {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: cozeResponse.Msg,
				Code:    cozeResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := ResponseCoze2OpenAI(&cozeResponse)
	fullTextResponse.Model = modelName
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	var responseText string
	if len(fullTextResponse.Choices) > 0 {
		responseText = fullTextResponse.Choices[0].Message.StringContent()
	}
	return nil, &responseText
}
