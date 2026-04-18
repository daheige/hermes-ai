package ali

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/infras/ctxkey"
	openai2 "hermes-ai/internal/infras/relay/adaptor/openai"
	"hermes-ai/internal/infras/relay/model"
	"hermes-ai/internal/infras/utils"
)

func ImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	responseFormat := c.GetString(ctxkey.ResponseFormat)

	var aliTaskResponse TaskResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai2.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai2.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &aliTaskResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if aliTaskResponse.Message != "" {
		slog.Error("aliAsyncTask err: " + string(responseBody))
		return openai2.ErrorWrapper(errors.New(aliTaskResponse.Message), "ali_async_task_failed", http.StatusInternalServerError), nil
	}

	aliResponse, _, err := asyncTaskWait(aliTaskResponse.Output.TaskId, apiKey)
	if err != nil {
		return openai2.ErrorWrapper(err, "ali_async_task_wait_failed", http.StatusInternalServerError), nil
	}

	if aliResponse.Output.TaskStatus != "SUCCEEDED" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: aliResponse.Output.Message,
				Type:    "ali_error",
				Param:   "",
				Code:    aliResponse.Output.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := responseAli2OpenAIImage(aliResponse, responseFormat)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai2.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, nil
}

func asyncTask(taskID string, key string) (*TaskResponse, error, []byte) {
	url := fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/tasks/%s", taskID)

	var aliResponse TaskResponse

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &aliResponse, err, nil
	}

	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("aliAsyncTask client.Do err: " + err.Error())
		return &aliResponse, err, nil
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	var response TaskResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		slog.Error("aliAsyncTask NewDecoder err: " + err.Error())
		return &aliResponse, err, nil
	}

	return &response, nil, responseBody
}

func asyncTaskWait(taskID string, key string) (*TaskResponse, []byte, error) {
	waitSeconds := 2
	step := 0
	maxStep := 20

	var taskResponse TaskResponse
	var responseBody []byte

	for {
		step++
		rsp, err, body := asyncTask(taskID, key)
		responseBody = body
		if err != nil {
			return &taskResponse, responseBody, err
		}

		if rsp.Output.TaskStatus == "" {
			return &taskResponse, responseBody, nil
		}

		switch rsp.Output.TaskStatus {
		case "FAILED":
			fallthrough
		case "CANCELED":
			fallthrough
		case "SUCCEEDED":
			fallthrough
		case "UNKNOWN":
			return rsp, responseBody, nil
		}
		if step >= maxStep {
			break
		}
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	return nil, nil, fmt.Errorf("aliAsyncTaskWait timeout")
}

func responseAli2OpenAIImage(response *TaskResponse, responseFormat string) *openai2.ImageResponse {
	imageResponse := openai2.ImageResponse{
		Created: utils.GetTimestamp(),
	}

	for _, data := range response.Output.Results {
		var b64Json string
		if responseFormat == "b64_json" {
			// 读取 data.Url 的图片数据并转存到 b64Json
			imageData, err := getImageData(data.Url)
			if err != nil {
				// 处理获取图片数据失败的情况
				slog.Error("getImageData Error getting image data: " + err.Error())
				continue
			}

			// 将图片数据转为 Base64 编码的字符串
			b64Json = Base64Encode(imageData)
		} else {
			// 如果 responseFormat 不是 "b64_json"，则直接使用 data.B64Image
			b64Json = data.B64Image
		}

		imageResponse.Data = append(imageResponse.Data, openai2.ImageData{
			Url:           data.Url,
			B64Json:       b64Json,
			RevisedPrompt: "",
		})
	}
	return &imageResponse
}

func getImageData(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}

func Base64Encode(data []byte) string {
	b64Json := base64.StdEncoding.EncodeToString(data)
	return b64Json
}
