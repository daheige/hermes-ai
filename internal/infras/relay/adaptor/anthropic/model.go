package anthropic

import (
	"encoding/json"
	"errors"
	"strings"
)

// https://docs.anthropic.com/claude/reference/messages_post

type Metadata struct {
	UserId string `json:"user_id"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Content struct {
	Type   string       `json:"type"`
	Text   string       `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
	// tool_calls
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Input     any    `json:"input,omitempty"`
	Content   string `json:"content,omitempty"`
	ToolUseId string `json:"tool_use_id,omitempty"`
}

type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// GetContentItems 解析 content 字段，支持字符串或数组格式
func (m *Message) GetContentItems() ([]Content, error) {
	// 首先尝试解析为字符串
	var textContent string
	if err := json.Unmarshal(m.Content, &textContent); err == nil {
		return []Content{{Type: "text", Text: textContent}}, nil
	}

	// 尝试解析为数组
	var items []Content
	if err := json.Unmarshal(m.Content, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// StringContent 获取纯文本格式的 content
func (m *Message) StringContent() string {
	items, err := m.GetContentItems()
	if err != nil {
		return ""
	}
	var textParts []string
	for _, item := range items {
		if item.Type == "text" {
			textParts = append(textParts, item.Text)
		}
	}
	return strings.Join(textParts, "")
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string `json:"type"`
	Properties any    `json:"properties,omitempty"`
	Required   any    `json:"required,omitempty"`
}

type SystemContent json.RawMessage

func (s SystemContent) MarshalJSON() ([]byte, error) {
	if len(s) == 0 {
		return []byte("null"), nil
	}
	return s, nil
}

func (s *SystemContent) UnmarshalJSON(data []byte) error {
	if s == nil {
		return errors.New("SystemContent: UnmarshalJSON on nil pointer")
	}
	*s = append((*s)[0:0], data...)
	return nil
}

func (s SystemContent) IsEmpty() bool {
	return len(s) == 0 || string(s) == `""` || string(s) == "null"
}

func (s SystemContent) String() string {
	var str string
	if err := json.Unmarshal(s, &str); err == nil {
		return str
	}
	var items []Content
	if err := json.Unmarshal(s, &items); err == nil {
		var parts []string
		for _, item := range items {
			if item.Type == "text" {
				parts = append(parts, item.Text)
			}
		}
		return strings.Join(parts, "")
	}
	return ""
}

type Request struct {
	Model         string       `json:"model"`
	Messages      []Message    `json:"messages"`
	System        SystemContent `json:"system,omitempty"`
	MaxTokens     int          `json:"max_tokens,omitempty"`
	StopSequences []string     `json:"stop_sequences,omitempty"`
	Stream        bool         `json:"stream,omitempty"`
	Temperature   *float64     `json:"temperature,omitempty"`
	TopP          *float64     `json:"top_p,omitempty"`
	TopK          int          `json:"top_k,omitempty"`
	Tools         []Tool       `json:"tools,omitempty"`
	ToolChoice    any          `json:"tool_choice,omitempty"`
	//Metadata    `json:"metadata,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Response struct {
	Id           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Model        string    `json:"model"`
	StopReason   *string   `json:"stop_reason"`
	StopSequence *string   `json:"stop_sequence"`
	Usage        Usage     `json:"usage"`
	Error        Error     `json:"error"`
}

type Delta struct {
	Type         string  `json:"type"`
	Text         string  `json:"text"`
	PartialJson  string  `json:"partial_json,omitempty"`
	StopReason   *string `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
}

type StreamResponse struct {
	Type         string    `json:"type"`
	Message      *Response `json:"message"`
	Index        int       `json:"index"`
	ContentBlock *Content  `json:"content_block"`
	Delta        *Delta    `json:"delta"`
	Usage        *Usage    `json:"usage"`
}
