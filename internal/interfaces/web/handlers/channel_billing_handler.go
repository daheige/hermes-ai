package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"hermes-ai/internal/application"
	"hermes-ai/internal/domain/entity"
	"hermes-ai/internal/infras/httpclient"
	"hermes-ai/internal/infras/monitor"
	channeltype2 "hermes-ai/internal/infras/relay/channeltype"
)

// ChannelBillingHandler 渠道账单处理器
type ChannelBillingHandler struct {
	service         *application.ChannelService
	channelMonitor  *monitor.ChannelMonitor
	requestInterval time.Duration
}

// NewChannelBillingHandler 创建渠道账单处理器
func NewChannelBillingHandler(service *application.ChannelService,
	channelMonitor *monitor.ChannelMonitor, requestInterval time.Duration) *ChannelBillingHandler {
	return &ChannelBillingHandler{service: service, channelMonitor: channelMonitor, requestInterval: requestInterval}
}

// OpenAISubscriptionResponse OpenAI订阅响应
type OpenAISubscriptionResponse struct {
	Object             string  `json:"object"`
	HasPaymentMethod   bool    `json:"has_payment_method"`
	SoftLimitUSD       float64 `json:"soft_limit_usd"`
	HardLimitUSD       float64 `json:"hard_limit_usd"`
	SystemHardLimitUSD float64 `json:"system_hard_limit_usd"`
	AccessUntil        int64   `json:"access_until"`
}

// OpenAIUsageDailyCost OpenAI每日成本
type OpenAIUsageDailyCost struct {
	Timestamp float64 `json:"timestamp"`
	LineItems []struct {
		Name string  `json:"name"`
		Cost float64 `json:"cost"`
	}
}

// OpenAICreditGrants OpenAI信用额度
type OpenAICreditGrants struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalAvailable float64 `json:"total_available"`
}

// OpenAIUsageResponse OpenAI使用响应
type OpenAIUsageResponse struct {
	Object     string  `json:"object"`
	TotalUsage float64 `json:"total_usage"`
}

// OpenAISBUsageResponse OpenAI-SB使用响应
type OpenAISBUsageResponse struct {
	Msg  string `json:"msg"`
	Data *struct {
		Credit string `json:"credit"`
	} `json:"data"`
}

// AIProxyUserOverviewResponse AIProxy用户概览响应
type AIProxyUserOverviewResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode int    `json:"error_code"`
	Data      struct {
		TotalPoints float64 `json:"totalPoints"`
	} `json:"data"`
}

// API2GPTUsageResponse API2GPT使用响应
type API2GPTUsageResponse struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalRemaining float64 `json:"total_remaining"`
}

// APGC2DGPTUsageResponse AIGC2D使用响应
type APGC2DGPTUsageResponse struct {
	Object         string  `json:"object"`
	TotalAvailable float64 `json:"total_available"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
}

// SiliconFlowUsageResponse SiliconFlow使用响应
type SiliconFlowUsageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
	Data    struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Image         string `json:"image"`
		Email         string `json:"email"`
		IsAdmin       bool   `json:"isAdmin"`
		Balance       string `json:"balance"`
		Status        string `json:"status"`
		Introduction  string `json:"introduction"`
		Role          string `json:"role"`
		ChargeBalance string `json:"chargeBalance"`
		TotalBalance  string `json:"totalBalance"`
		Category      string `json:"category"`
	} `json:"data"`
}

// DeepSeekUsageResponse DeepSeek使用响应
type DeepSeekUsageResponse struct {
	IsAvailable  bool `json:"is_available"`
	BalanceInfos []struct {
		Currency        string `json:"currency"`
		TotalBalance    string `json:"total_balance"`
		GrantedBalance  string `json:"granted_balance"`
		ToppedUpBalance string `json:"topped_up_balance"`
	} `json:"balance_infos"`
}

// OpenRouterResponse OpenRouter响应
type OpenRouterResponse struct {
	Data struct {
		TotalCredits float64 `json:"total_credits"`
		TotalUsage   float64 `json:"total_usage"`
	} `json:"data"`
}

// GetAuthHeader 获取认证头
func (h *ChannelBillingHandler) GetAuthHeader(token string) http.Header {
	headers := http.Header{}
	headers.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return headers
}

// GetResponseBody 获取响应体
func (h *ChannelBillingHandler) GetResponseBody(method, url string, channel *entity.Channel, headers http.Header) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	for k := range headers {
		req.Header.Add(k, headers.Get(k))
	}
	res, err := httpclient.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (h *ChannelBillingHandler) updateChannelCloseAIBalance(channel *entity.Channel) (float64, error) {
	url := fmt.Sprintf("%s/dashboard/billing/credit_grants", channel.GetBaseURL())
	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := OpenAICreditGrants{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	h.service.UpdateBalance(channel.Id, response.TotalAvailable)
	return response.TotalAvailable, nil
}

func (h *ChannelBillingHandler) updateChannelOpenAISBBalance(channel *entity.Channel) (float64, error) {
	url := fmt.Sprintf("https://api.openai-sb.com/sb-api/user/status?api_key=%s", channel.Key)
	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := OpenAISBUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if response.Data == nil {
		return 0, errors.New(response.Msg)
	}
	balance, err := strconv.ParseFloat(response.Data.Credit, 64)
	if err != nil {
		return 0, err
	}
	h.service.UpdateBalance(channel.Id, balance)
	return balance, nil
}

func (h *ChannelBillingHandler) updateChannelAIProxyBalance(channel *entity.Channel) (float64, error) {
	url := "https://aiproxy.io/api/report/getUserOverview"
	headers := http.Header{}
	headers.Add("Api-Key", channel.Key)
	body, err := h.GetResponseBody("GET", url, channel, headers)
	if err != nil {
		return 0, err
	}
	response := AIProxyUserOverviewResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if !response.Success {
		return 0, fmt.Errorf("code: %d, message: %s", response.ErrorCode, response.Message)
	}
	h.service.UpdateBalance(channel.Id, response.Data.TotalPoints)
	return response.Data.TotalPoints, nil
}

func (h *ChannelBillingHandler) updateChannelAPI2GPTBalance(channel *entity.Channel) (float64, error) {
	url := "https://api.api2gpt.com/dashboard/billing/credit_grants"
	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := API2GPTUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	h.service.UpdateBalance(channel.Id, response.TotalRemaining)
	return response.TotalRemaining, nil
}

func (h *ChannelBillingHandler) updateChannelAIGC2DBalance(channel *entity.Channel) (float64, error) {
	url := "https://api.aigc2d.com/dashboard/billing/credit_grants"
	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := APGC2DGPTUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	h.service.UpdateBalance(channel.Id, response.TotalAvailable)
	return response.TotalAvailable, nil
}

func (h *ChannelBillingHandler) updateChannelSiliconFlowBalance(channel *entity.Channel) (float64, error) {
	url := "https://api.siliconflow.cn/v1/user/info"
	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := SiliconFlowUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if response.Code != 20000 {
		return 0, fmt.Errorf("code: %d, message: %s", response.Code, response.Message)
	}
	balance, err := strconv.ParseFloat(response.Data.TotalBalance, 64)
	if err != nil {
		return 0, err
	}
	h.service.UpdateBalance(channel.Id, balance)
	return balance, nil
}

func (h *ChannelBillingHandler) updateChannelDeepSeekBalance(channel *entity.Channel) (float64, error) {
	url := "https://api.deepseek.com/user/balance"
	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := DeepSeekUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	index := -1
	for i, balanceInfo := range response.BalanceInfos {
		if balanceInfo.Currency == "CNY" {
			index = i
			break
		}
	}
	if index == -1 {
		return 0, errors.New("currency CNY not found")
	}
	balance, err := strconv.ParseFloat(response.BalanceInfos[index].TotalBalance, 64)
	if err != nil {
		return 0, err
	}
	h.service.UpdateBalance(channel.Id, balance)
	return balance, nil
}

func (h *ChannelBillingHandler) updateChannelOpenRouterBalance(channel *entity.Channel) (float64, error) {
	url := "https://openrouter.ai/api/v1/credits"
	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := OpenRouterResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	balance := response.Data.TotalCredits - response.Data.TotalUsage
	h.service.UpdateBalance(channel.Id, balance)
	return balance, nil
}

func (h *ChannelBillingHandler) updateChannelBalance(channel *entity.Channel) (float64, error) {
	baseURL := channeltype2.ChannelBaseURLs[channel.Type]
	if channel.GetBaseURL() == "" {
		channel.BaseURL = &baseURL
	}
	switch channel.Type {
	case channeltype2.OpenAI:
		if channel.GetBaseURL() != "" {
			baseURL = channel.GetBaseURL()
		}
	case channeltype2.Azure:
		return 0, errors.New("尚未实现")
	case channeltype2.Custom:
		baseURL = channel.GetBaseURL()
	case channeltype2.CloseAI:
		return h.updateChannelCloseAIBalance(channel)
	case channeltype2.OpenAISB:
		return h.updateChannelOpenAISBBalance(channel)
	case channeltype2.AIProxy:
		return h.updateChannelAIProxyBalance(channel)
	case channeltype2.API2GPT:
		return h.updateChannelAPI2GPTBalance(channel)
	case channeltype2.AIGC2D:
		return h.updateChannelAIGC2DBalance(channel)
	case channeltype2.SiliconFlow:
		return h.updateChannelSiliconFlowBalance(channel)
	case channeltype2.DeepSeek:
		return h.updateChannelDeepSeekBalance(channel)
	case channeltype2.OpenRouter:
		return h.updateChannelOpenRouterBalance(channel)
	default:
		return 0, errors.New("尚未实现")
	}
	url := fmt.Sprintf("%s/v1/dashboard/billing/subscription", baseURL)

	body, err := h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	subscription := OpenAISubscriptionResponse{}
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	startDate := fmt.Sprintf("%s-01", now.Format("2006-01"))
	endDate := now.Format("2006-01-02")
	if !subscription.HasPaymentMethod {
		startDate = now.AddDate(0, 0, -100).Format("2006-01-02")
	}
	url = fmt.Sprintf("%s/v1/dashboard/billing/usage?start_date=%s&end_date=%s", baseURL, startDate, endDate)
	body, err = h.GetResponseBody("GET", url, channel, h.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	usage := OpenAIUsageResponse{}
	err = json.Unmarshal(body, &usage)
	if err != nil {
		return 0, err
	}
	balance := subscription.HardLimitUSD - usage.TotalUsage/100
	h.service.UpdateBalance(channel.Id, balance)
	return balance, nil
}

// UpdateChannelBalance 更新渠道余额
func (h *ChannelBillingHandler) UpdateChannelBalance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	channel, err := h.service.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	balance, err := h.updateChannelBalance(channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"balance": balance,
	})
}

func (h *ChannelBillingHandler) updateAllChannelsBalance() error {
	channels, err := h.service.GetAllChannels(0, 0, "all")
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if channel.Status != entity.ChannelStatusEnabled {
			continue
		}
		if channel.Type != channeltype2.OpenAI && channel.Type != channeltype2.Custom {
			continue
		}

		balance, err := h.updateChannelBalance(channel)
		if err != nil {
			continue
		} else {
			if balance <= 0 {
				h.channelMonitor.DisableChannel(channel.Id, channel.Name, "余额不足")
			}
		}

		time.Sleep(h.requestInterval)
	}

	return nil
}

// UpdateAllChannelsBalance 更新所有渠道余额
func (h *ChannelBillingHandler) UpdateAllChannelsBalance(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// AutomaticallyUpdateChannels 自动更新渠道
func AutomaticallyUpdateChannels(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		slog.Info("updating all channels")
		// 这里需要在 main 中初始化 handler 后调用
		slog.Info("channels update done")
	}
}
