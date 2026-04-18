package aiproxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"hermes-ai/internal/infras/ginzo"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/constant"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/render"
	"hermes-ai/internal/infras/utils"

	"github.com/gin-gonic/gin"
)

// https://docs.aiproxy.io/dev/library#使用已经定制好的知识库进行对话问答

func ConvertRequest(request model2.GeneralOpenAIRequest) *LibraryRequest {
	query := ""
	if len(request.Messages) != 0 {
		query = request.Messages[len(request.Messages)-1].StringContent()
	}
	return &LibraryRequest{
		Model:  request.Model,
		Stream: request.Stream,
		Query:  query,
	}
}

func aiProxyDocuments2Markdown(documents []LibraryDocument) string {
	if len(documents) == 0 {
		return ""
	}
	content := "\n\n参考文档：\n"
	for i, document := range documents {
		content += fmt.Sprintf("%d. [%s](%s)\n", i+1, document.Title, document.URL)
	}
	return content
}

func responseAIProxyLibrary2OpenAI(response *LibraryResponse) *openai2.TextResponse {
	content := response.Answer + aiProxyDocuments2Markdown(response.Documents)
	choice := openai2.TextResponseChoice{
		Index: 0,
		Message: model2.Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: "stop",
	}
	fullTextResponse := openai2.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", utils.UUID()),
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Choices: []openai2.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

func documentsAIProxyLibrary(documents []LibraryDocument) *openai2.ChatCompletionsStreamResponse {
	var choice openai2.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = aiProxyDocuments2Markdown(documents)
	choice.FinishReason = &constant.StopFinishReason
	return &openai2.ChatCompletionsStreamResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", utils.UUID()),
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   "",
		Choices: []openai2.ChatCompletionsStreamResponseChoice{choice},
	}
}

func streamResponseAIProxyLibrary2OpenAI(response *LibraryStreamResponse) *openai2.ChatCompletionsStreamResponse {
	var choice openai2.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = response.Content
	return &openai2.ChatCompletionsStreamResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", utils.UUID()),
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   response.Model,
		Choices: []openai2.ChatCompletionsStreamResponseChoice{choice},
	}
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *model2.Usage) {
	var usage model2.Usage
	var documents []LibraryDocument
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	ginzo.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 5 || data[:5] != "data:" {
			continue
		}
		data = data[5:]

		var AIProxyLibraryResponse LibraryStreamResponse
		err := json.Unmarshal([]byte(data), &AIProxyLibraryResponse)
		if err != nil {
			slog.Error("error unmarshalling stream response: " + err.Error())
			continue
		}
		if len(AIProxyLibraryResponse.Documents) != 0 {
			documents = AIProxyLibraryResponse.Documents
		}
		response := streamResponseAIProxyLibrary2OpenAI(&AIProxyLibraryResponse)
		err = render.ObjectData(c, response)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("error reading stream: " + err.Error())
	}

	response := documentsAIProxyLibrary(documents)
	err := render.ObjectData(c, response)
	if err != nil {
		slog.Error(err.Error())
	}
	render.Done(c)

	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	return nil, &usage
}

func Handler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *model2.Usage) {
	var AIProxyLibraryResponse LibraryResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &AIProxyLibraryResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if AIProxyLibraryResponse.ErrCode != 0 {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: AIProxyLibraryResponse.Message,
				Type:    strconv.Itoa(AIProxyLibraryResponse.ErrCode),
				Code:    AIProxyLibraryResponse.ErrCode,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseAIProxyLibrary2OpenAI(&AIProxyLibraryResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "write_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &fullTextResponse.Usage
}
