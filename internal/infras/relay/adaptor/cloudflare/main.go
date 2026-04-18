package cloudflare

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/ginzo"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/render"

	"github.com/gin-gonic/gin"
)

func ConvertCompletionsRequest(textRequest model2.GeneralOpenAIRequest) *Request {
	p, _ := textRequest.Prompt.(string)
	return &Request{
		Prompt:      p,
		MaxTokens:   textRequest.MaxTokens,
		Stream:      textRequest.Stream,
		Temperature: textRequest.Temperature,
	}
}

func StreamHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model2.ErrorWithStatusCode, *model2.Usage) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	ginzo.SetEventStreamHeaders(c)
	id := ginzo.GetResponseID(c)
	responseModel := c.GetString(ctxkey.OriginalModel)
	var responseText string

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < len("data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\r")

		if data == "[DONE]" {
			break
		}

		var response openai2.ChatCompletionsStreamResponse
		err := json.Unmarshal([]byte(data), &response)
		if err != nil {
			slog.Error("error unmarshalling stream response: " + err.Error())
			continue
		}
		for _, v := range response.Choices {
			v.Delta.Role = "assistant"
			responseText += v.Delta.StringContent()
		}
		response.Id = id
		response.Model = modelName
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

	usage := openai2.ResponseText2Usage(responseText, responseModel, promptTokens)
	return nil, usage
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
	var response openai2.TextResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	response.Model = modelName
	var responseText string
	for _, v := range response.Choices {
		responseText += v.Message.Content.(string)
	}
	usage := openai2.ResponseText2Usage(responseText, modelName, promptTokens)
	response.Usage = *usage
	response.Id = ginzo.GetResponseID(c)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)
	return nil, usage
}
