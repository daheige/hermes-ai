package tencent

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/ginzo"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/constant"
	model2 "hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/render"
	"hermes-ai/internal/infras/utils"
)

func ConvertRequest(request model2.GeneralOpenAIRequest) *ChatRequest {
	messages := make([]*Message, 0, len(request.Messages))
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		messages = append(messages, &Message{
			Content: message.StringContent(),
			Role:    message.Role,
		})
	}
	return &ChatRequest{
		Model:       &request.Model,
		Stream:      &request.Stream,
		Messages:    messages,
		TopP:        request.TopP,
		Temperature: request.Temperature,
	}
}

func ConvertEmbeddingRequest(request model2.GeneralOpenAIRequest) *EmbeddingRequest {
	return &EmbeddingRequest{
		InputList: request.ParseInput(),
	}
}

func EmbeddingHandler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *model2.Usage) {
	var tencentResponseP EmbeddingResponseP
	err := json.NewDecoder(resp.Body).Decode(&tencentResponseP)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	tencentResponse := tencentResponseP.Response
	if tencentResponse.Error.Code != "" {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: tencentResponse.Error.Message,
				Code:    tencentResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	requestModel := c.GetString(ctxkey.RequestModel)
	fullTextResponse := embeddingResponseTencent2OpenAI(&tencentResponse)
	fullTextResponse.Model = requestModel
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func embeddingResponseTencent2OpenAI(response *EmbeddingResponse) *openai2.EmbeddingResponse {
	openAIEmbeddingResponse := openai2.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai2.EmbeddingResponseItem, 0, len(response.Data)),
		Model:  "hunyuan-embedding",
		Usage:  model2.Usage{TotalTokens: response.EmbeddingUsage.TotalTokens},
	}

	for _, item := range response.Data {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, openai2.EmbeddingResponseItem{
			Object:    item.Object,
			Index:     item.Index,
			Embedding: item.Embedding,
		})
	}
	return &openAIEmbeddingResponse
}

func responseTencent2OpenAI(response *ChatResponse) *openai2.TextResponse {
	fullTextResponse := openai2.TextResponse{
		Id:      response.ReqID,
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Usage: model2.Usage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}
	if len(response.Choices) > 0 {
		choice := openai2.TextResponseChoice{
			Index: 0,
			Message: model2.Message{
				Role:    "assistant",
				Content: response.Choices[0].Messages.Content,
			},
			FinishReason: response.Choices[0].FinishReason,
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseTencent2OpenAI(TencentResponse *ChatResponse) *openai2.ChatCompletionsStreamResponse {
	response := openai2.ChatCompletionsStreamResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", utils.UUID()),
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   "tencent-hunyuan",
	}
	if len(TencentResponse.Choices) > 0 {
		var choice openai2.ChatCompletionsStreamResponseChoice
		choice.Delta.Content = TencentResponse.Choices[0].Delta.Content
		if TencentResponse.Choices[0].FinishReason == "stop" {
			choice.FinishReason = &constant.StopFinishReason
		}
		response.Choices = append(response.Choices, choice)
	}
	return &response
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, string) {
	var responseText string
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	ginzo.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 5 || !strings.HasPrefix(data, "data:") {
			continue
		}
		data = strings.TrimPrefix(data, "data:")

		var tencentResponse ChatResponse
		err := json.Unmarshal([]byte(data), &tencentResponse)
		if err != nil {
			slog.Error("error unmarshalling stream response: " + err.Error())
			continue
		}

		response := streamResponseTencent2OpenAI(&tencentResponse)
		if len(response.Choices) != 0 {
			responseText += utils.AsString(response.Choices[0].Delta.Content)
		}

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
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}

	return nil, responseText
}

func Handler(c *gin.Context, resp *http.Response) (*model2.ErrorWithStatusCode, *model2.Usage) {
	var TencentResponse ChatResponse
	var responseP ChatResponseP
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &responseP)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	TencentResponse = responseP.Response
	if TencentResponse.Error.Code != "" {
		return &model2.ErrorWithStatusCode{
			Error: model2.Error{
				Message: TencentResponse.Error.Message,
				Code:    TencentResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseTencent2OpenAI(&TencentResponse)
	fullTextResponse.Model = "hunyuan"
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

func ParseConfig(config string) (appId int64, secretId string, secretKey string, err error) {
	parts := strings.Split(config, "|")
	if len(parts) != 3 {
		err = errors.New("invalid tencent config")
		return
	}
	appId, err = strconv.ParseInt(parts[0], 10, 64)
	secretId = parts[1]
	secretKey = parts[2]
	return
}

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func hmacSha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func GetSign(req any, adaptor *Adaptor, secId, secKey string) string {
	// build canonical request string
	host := "hunyuan.tencentcloudapi.com"
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		"application/json", host, strings.ToLower(adaptor.Action))
	signedHeaders := "content-type;host;x-tc-action"
	payload, _ := json.Marshal(req)
	hashedRequestPayload := sha256hex(string(payload))
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)
	// build string to sign
	algorithm := "TC3-HMAC-SHA256"
	requestTimestamp := strconv.FormatInt(adaptor.Timestamp, 10)
	timestamp, _ := strconv.ParseInt(requestTimestamp, 10, 64)
	t := time.Unix(timestamp, 0).UTC()
	// must be the format 2006-01-02, ref to package time for more info
	date := t.Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, "hunyuan")
	hashedCanonicalRequest := sha256hex(canonicalRequest)
	string2sign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		requestTimestamp,
		credentialScope,
		hashedCanonicalRequest)

	// sign string
	secretDate := hmacSha256(date, "TC3"+secKey)
	secretService := hmacSha256("hunyuan", secretDate)
	secretKey := hmacSha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(hmacSha256(string2sign, secretKey)))

	// build authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		secId,
		credentialScope,
		signedHeaders,
		signature)
	return authorization
}
