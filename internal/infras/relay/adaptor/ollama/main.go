package ollama

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"hermes-ai/internal/infras/ginzo"
	"hermes-ai/internal/infras/image"
	"hermes-ai/internal/infras/logger"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/constant"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/render"
	"hermes-ai/internal/infras/utils"

	"github.com/gin-gonic/gin"
)

func ConvertRequest(request model2.GeneralOpenAIRequest) *ChatRequest {
	ollamaRequest := ChatRequest{
		Model: request.Model,
		Options: &Options{
			Seed:             int(request.Seed),
			Temperature:      request.Temperature,
			TopP:             request.TopP,
			FrequencyPenalty: request.FrequencyPenalty,
			PresencePenalty:  request.PresencePenalty,
			NumPredict:       request.MaxTokens,
			NumCtx:           request.NumCtx,
		},
		Stream: request.Stream,
	}
	for _, message := range request.Messages {
		openaiContent := message.ParseContent()
		var imageUrls []string
		var contentText string
		for _, part := range openaiContent {
			switch part.Type {
			case model2.ContentTypeText:
				contentText = part.Text
			case model2.ContentTypeImageURL:
				_, data, _ := image.GetImageFromUrl(part.ImageURL.Url)
				imageUrls = append(imageUrls, data)
			}
		}
		ollamaRequest.Messages = append(ollamaRequest.Messages, Message{
			Role:    message.Role,
			Content: contentText,
			Images:  imageUrls,
		})
	}
	return &ollamaRequest
}

func responseOllama2OpenAI(response *ChatResponse) *openai2.TextResponse {
	choice := openai2.TextResponseChoice{
		Index: 0,
		Message: model2.Message{
			Role:    response.Message.Role,
			Content: response.Message.Content,
		},
	}
	if response.Done {
		choice.FinishReason = "stop"
	}
	fullTextResponse := openai2.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", utils.UUID()),
		Model:   response.Model,
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Choices: []openai2.TextResponseChoice{choice},
		Usage: model2.Usage{
			PromptTokens:     response.PromptEvalCount,
			CompletionTokens: response.EvalCount,
			TotalTokens:      response.PromptEvalCount + response.EvalCount,
		},
	}
	return &fullTextResponse
}

func streamResponseOllama2OpenAI(ollamaResponse *ChatResponse) *openai2.ChatCompletionsStreamResponse {
	var choice openai2.ChatCompletionsStreamResponseChoice
	choice.Delta.Role = ollamaResponse.Message.Role
	choice.Delta.Content = ollamaResponse.Message.Content
	if ollamaResponse.Done {
		choice.FinishReason = &constant.StopFinishReason
	}
	response := openai2.ChatCompletionsStreamResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", utils.UUID()),
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   ollamaResponse.Model,
		Choices: []openai2.ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *model2.Usage) {
	var usage model2.Usage
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "}\n"); i >= 0 {
			return i + 2, data[0 : i+1], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	ginzo.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		if strings.HasPrefix(data, "}") {
			data = strings.TrimPrefix(data, "}") + "}"
		}

		var ollamaResponse ChatResponse
		err := json.Unmarshal([]byte(data), &ollamaResponse)
		if err != nil {
			slog.Error("error unmarshalling stream response: " + err.Error())
			continue
		}

		if ollamaResponse.EvalCount != 0 {
			usage.PromptTokens = ollamaResponse.PromptEvalCount
			usage.CompletionTokens = ollamaResponse.EvalCount
			usage.TotalTokens = ollamaResponse.PromptEvalCount + ollamaResponse.EvalCount
		}

		response := streamResponseOllama2OpenAI(&ollamaResponse)
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

	return nil, &usage
}

func ConvertEmbeddingRequest(request model2.GeneralOpenAIRequest) *EmbeddingRequest {
	return &EmbeddingRequest{
		Model: request.Model,
		Input: request.ParseInput(),
		Options: &Options{
			Seed:             int(request.Seed),
			Temperature:      request.Temperature,
			TopP:             request.TopP,
			FrequencyPenalty: request.FrequencyPenalty,
			PresencePenalty:  request.PresencePenalty,
		},
	}
}

func EmbeddingHandler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *model2.Usage) {
	var ollamaResponse EmbeddingResponse
	err := json.NewDecoder(resp.Body).Decode(&ollamaResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if ollamaResponse.Error != "" {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: ollamaResponse.Error,
				Type:    "ollama_error",
				Param:   "",
				Code:    "ollama_error",
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := embeddingResponseOllama2OpenAI(&ollamaResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func embeddingResponseOllama2OpenAI(response *EmbeddingResponse) *openai2.EmbeddingResponse {
	openAIEmbeddingResponse := openai2.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai2.EmbeddingResponseItem, 0, 1),
		Model:  response.Model,
		Usage:  model2.Usage{TotalTokens: 0},
	}

	for i, embedding := range response.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, openai2.EmbeddingResponseItem{
			Object:    `embedding`,
			Index:     i,
			Embedding: embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func Handler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *model2.Usage) {
	ctx := context.TODO()
	var ollamaResponse ChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	slog.With("request_id", logger.GetRequestID(ctx)).
		Debug(fmt.Sprintf("ollama response: %s", string(responseBody)))
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &ollamaResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if ollamaResponse.Error != "" {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: ollamaResponse.Error,
				Type:    "ollama_error",
				Param:   "",
				Code:    "ollama_error",
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseOllama2OpenAI(&ollamaResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
