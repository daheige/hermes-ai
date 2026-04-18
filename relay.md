# Hermes AI Relay API 文档

本文档包含 Hermes AI 系统的 Relay 路由接口，兼容 OpenAI API 格式。

## 目录

1. [认证方式](#认证方式)
2. [模型接口](#模型接口)
3. [文本生成](#文本生成)
4. [图像生成](#图像生成)
5. [音频处理](#音频处理)
6. [嵌入向量](#嵌入向量)
7. [内容审核](#内容审核)
8. [代理接口](#代理接口)
9. [Anthropic Messages](#anthropic-messages)
10. [文件操作（未实现）](#文件操作未实现)
11. [微调操作（未实现）](#微调操作未实现)
12. [助手操作（未实现）](#助手操作未实现)
13. [线程操作（未实现）](#线程操作未实现)

---

## 认证方式

所有 Relay 接口需要在请求头中携带认证信息：

```
Authorization: Bearer {your-api-key}
```

API Key 可以通过管理后台的令牌管理功能创建。

---

## 模型接口

### 列出所有模型

```
GET /v1/models
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |

**响应体**:

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-3.5-turbo",
      "object": "model",
      "created": 1626777600,
      "owned_by": "openai",
      "permission": [
        {
          "id": "modelperm-LwHkVFn8AcMItP432fKKDIKJ",
          "object": "model_permission",
          "created": 1626777600,
          "allow_create_engine": true,
          "allow_sampling": true,
          "allow_logprobs": true,
          "allow_search_indices": false,
          "allow_view": true,
          "allow_fine_tuning": false,
          "organization": "*",
          "group": null,
          "is_blocking": false
        }
      ],
      "root": "gpt-3.5-turbo",
      "parent": null
    }
  ]
}
```

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| object | string | 对象类型，固定为 "list" |
| data | array | 模型列表 |
| data[].id | string | 模型 ID |
| data[].object | string | 对象类型，固定为 "model" |
| data[].created | int | 创建时间戳 |
| data[].owned_by | string | 拥有者 |
| data[].permission | array | 权限列表 |
| data[].root | string | 根模型 |
| data[].parent | string | 父模型 |

### 获取单个模型

```
GET /v1/models/{model}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| model | string | 模型 ID |

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |

**响应体**: 单个模型对象，字段同上

---

## 文本生成

### 创建对话补全

```
POST /v1/chat/completions
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**:

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 150,
  "top_p": 1,
  "frequency_penalty": 0,
  "presence_penalty": 0,
  "stream": false,
  "user": "user_id"
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 是 | 模型 ID |
| messages | array | 是 | 消息列表 |
| messages[].role | string | 是 | 角色（system/user/assistant） |
| messages[].content | string | 是 | 消息内容 |
| temperature | float | 否 | 采样温度（0-2），默认 1 |
| max_tokens | int | 否 | 最大生成 token 数 |
| top_p | float | 否 | 核采样概率，默认 1 |
| frequency_penalty | float | 否 | 频率惩罚（-2 到 2），默认 0 |
| presence_penalty | float | 否 | 存在惩罚（-2 到 2），默认 0 |
| stream | boolean | 否 | 是否流式返回，默认 false |
| user | string | 否 | 用户标识 |

**响应体（非流式）**:

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-3.5-turbo",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I assist you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 19,
    "completion_tokens": 9,
    "total_tokens": 28
  }
}
```

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 补全 ID |
| object | string | 对象类型 |
| created | int | 创建时间戳 |
| model | string | 使用的模型 |
| choices | array | 生成结果列表 |
| choices[].index | int | 结果索引 |
| choices[].message | object | 消息对象 |
| choices[].message.role | string | 角色 |
| choices[].message.content | string | 内容 |
| choices[].finish_reason | string | 结束原因 |
| usage | object | Token 使用统计 |
| usage.prompt_tokens | int | 提示 token 数 |
| usage.completion_tokens | int | 补全 token 数 |
| usage.total_tokens | int | 总 token 数 |

### 创建文本补全（Legacy）

```
POST /v1/completions
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**:

```json
{
  "model": "text-davinci-003",
  "prompt": "Once upon a time",
  "max_tokens": 100,
  "temperature": 0.7,
  "top_p": 1,
  "n": 1,
  "stream": false,
  "logprobs": null,
  "stop": null
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 是 | 模型 ID |
| prompt | string/array | 是 | 提示文本 |
| suffix | string | 否 | 后缀文本 |
| max_tokens | int | 否 | 最大 token 数 |
| temperature | float | 否 | 采样温度 |
| top_p | float | 否 | 核采样 |
| n | int | 否 | 生成数量 |
| stream | boolean | 否 | 是否流式 |
| logprobs | int | 否 | 返回的 logprobs 数量 |
| echo | boolean | 否 | 是否回显提示 |
| stop | string/array | 否 | 停止序列 |
| presence_penalty | float | 否 | 存在惩罚 |
| frequency_penalty | float | 否 | 频率惩罚 |
| best_of | int | 否 | 最佳结果数 |
| user | string | 否 | 用户标识 |

### 创建文本编辑

```
POST /v1/edits
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**:

```json
{
  "model": "text-davinci-edit-001",
  "input": "What day of the wek is it?",
  "instruction": "Fix the spelling mistakes"
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 是 | 模型 ID |
| input | string | 否 | 输入文本 |
| instruction | string | 是 | 编辑指令 |
| n | int | 否 | 生成数量 |
| temperature | float | 否 | 采样温度 |
| top_p | float | 否 | 核采样 |

---

## 图像生成

### 创建图像

```
POST /v1/images/generations
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**:

```json
{
  "prompt": "A cute baby sea otter",
  "n": 1,
  "size": "1024x1024",
  "response_format": "url",
  "user": "user_id"
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| prompt | string | 是 | 图像描述 |
| n | int | 否 | 生成数量（1-10） |
| size | string | 否 | 尺寸（256x256, 512x512, 1024x1024） |
| response_format | string | 否 | 响应格式（url 或 b64_json） |
| user | string | 否 | 用户标识 |

**响应体**:

```json
{
  "created": 1589478378,
  "data": [
    {
      "url": "https://...",
      "b64_json": null
    }
  ]
}
```

### 创建图像编辑（未实现）

```
POST /v1/images/edits
```

**状态**: 未实现

**响应**:

```json
{
  "error": {
    "message": "API not implemented",
    "type": "one_api_error",
    "code": "api_not_implemented"
  }
}
```

### 创建图像变体（未实现）

```
POST /v1/images/variations
```

**状态**: 未实现

**响应**:

```json
{
  "error": {
    "message": "API not implemented",
    "type": "one_api_error",
    "code": "api_not_implemented"
  }
}
```

---

## 音频处理

### 创建转录

```
POST /v1/audio/transcriptions
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | multipart/form-data |

**请求参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 音频文件（mp3, mp4, mpeg, mpga, m4a, wav, webm） |
| model | string | 是 | 模型 ID（如 whisper-1） |
| prompt | string | 否 | 提示文本 |
| response_format | string | 否 | 响应格式（json, text, srt, verbose_json, vtt） |
| temperature | float | 否 | 采样温度（0-1） |
| language | string | 否 | 语言代码 |

**响应体**:

```json
{
  "text": "Hello, this is a transcription."
}
```

### 创建翻译

```
POST /v1/audio/translations
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | multipart/form-data |

**请求参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 音频文件 |
| model | string | 是 | 模型 ID |
| prompt | string | 否 | 提示文本 |
| response_format | string | 否 | 响应格式 |
| temperature | float | 否 | 采样温度 |

**响应体**:

```json
{
  "text": "Hello, this is a translation."
}
```

### 创建语音

```
POST /v1/audio/speech
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**:

```json
{
  "model": "tts-1",
  "input": "Hello, world!",
  "voice": "alloy",
  "response_format": "mp3",
  "speed": 1.0
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 是 | 模型 ID（tts-1 或 tts-1-hd） |
| input | string | 是 | 文本内容（最大 4096 字符） |
| voice | string | 是 | 声音（alloy, echo, fable, onyx, nova, shimmer） |
| response_format | string | 否 | 格式（mp3, opus, aac, flac） |
| speed | float | 否 | 语速（0.25-4.0） |

**响应**: 音频文件流

---

## 嵌入向量

### 创建嵌入

```
POST /v1/embeddings
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**:

```json
{
  "input": "The food was delicious and the waiter...",
  "model": "text-embedding-ada-002",
  "user": "user_id"
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| input | string/array | 是 | 输入文本（最大 8192 tokens） |
| model | string | 是 | 模型 ID |
| user | string | 否 | 用户标识 |

**响应体**:

```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.0023064255, -0.009327292, ...],
      "index": 0
    }
  ],
  "model": "text-embedding-ada-002",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

### 引擎嵌入（Legacy）

```
POST /v1/engines/{model}/embeddings
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| model | string | 模型 ID |

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**: 同 `/v1/embeddings`

---

## 内容审核

### 创建审核

```
POST /v1/moderations
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |

**请求体**:

```json
{
  "input": "I want to kill them.",
  "model": "text-moderation-latest"
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| input | string/array | 是 | 需要审核的文本 |
| model | string | 否 | 审核模型 |

**响应体**:

```json
{
  "id": "modr-xxx",
  "model": "text-moderation-004",
  "results": [
    {
      "flagged": true,
      "categories": {
        "sexual": false,
        "hate": false,
        "harassment": false,
        "self-harm": false,
        "sexual/minors": false,
        "hate/threatening": false,
        "violence/graphic": false,
        "violence": true
      },
      "category_scores": {
        "sexual": 0.0001,
        "hate": 0.0001,
        ...
      }
    }
  ]
}
```

---

## 代理接口

### OneAPI 代理

```
ANY /v1/oneapi/proxy/{channelid}/*target
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| channelid | string | 渠道 ID |
| target | string | 目标路径 |

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |

**说明**: 此接口用于直接代理请求到指定渠道。

---

## Anthropic Messages

### 创建消息

```
POST /v1/messages
```

**请求头**:

| 字段 | 必填 | 说明 |
|------|------|------|
| Authorization | 是 | Bearer Token |
| Content-Type | 是 | application/json |
| anthropic-version | 否 | Anthropic API 版本，默认 `2023-06-01` |
| anthropic-beta | 否 | Beta 功能标识，系统会自动设置必要的 beta 头 |

**请求体**:

```json
{
  "model": "claude-3-5-sonnet-20241022",
  "max_tokens": 4096,
  "messages": [
    {
      "role": "user",
      "content": "Hello, world!"
    }
  ],
  "stream": false
}
```

**请求字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 是 | 模型名称 |
| messages | array | 是 | 消息列表，content 支持字符串或数组格式 |
| messages[].role | string | 是 | 角色（user/assistant） |
| messages[].content | string/array | 是 | 消息内容 |
| max_tokens | int | 条件 | 最大生成 token 数，不传时默认 4096 |
| stream | boolean | 否 | 是否开启流式响应，默认 false |
| system | string/array | 否 | 系统提示词，支持字符串或数组格式 |
| temperature | float | 否 | 采样温度 |
| top_p | float | 否 | 核采样概率 |
| top_k | int | 否 | Top-K 采样 |
| tools | array | 否 | 工具列表 |
| tool_choice | any | 否 | 工具选择策略 |

**行为说明**:

- 该接口为原生 Anthropic Messages API 转发接口，请求体原样透传到 Anthropic 上游。
- 响应保持原生 Anthropic JSON / SSE 格式，不做 OpenAI 格式转换。
- 支持失败重试、渠道自动禁用、配额预扣与后扣等网关通用能力。

---

## 文件操作（未实现）

以下文件相关接口暂未实现：

### 列出文件
```
GET /v1/files
```
**状态**: 未实现

### 上传文件
```
POST /v1/files
```
**状态**: 未实现

### 删除文件
```
DELETE /v1/files/{id}
```
**状态**: 未实现

### 获取文件信息
```
GET /v1/files/{id}
```
**状态**: 未实现

### 获取文件内容
```
GET /v1/files/{id}/content
```
**状态**: 未实现

---

## 微调操作（未实现）

以下微调相关接口暂未实现：

### 创建微调任务
```
POST /v1/fine_tuning/jobs
```
**状态**: 未实现

### 列出微调任务
```
GET /v1/fine_tuning/jobs
```
**状态**: 未实现

### 获取微调任务
```
GET /v1/fine_tuning/jobs/{id}
```
**状态**: 未实现

### 取消微调任务
```
POST /v1/fine_tuning/jobs/{id}/cancel
```
**状态**: 未实现

### 列出微调事件
```
GET /v1/fine_tuning/jobs/{id}/events
```
**状态**: 未实现

### 删除模型
```
DELETE /v1/models/{model}
```
**状态**: 未实现

---

## 助手操作（未实现）

以下助手相关接口暂未实现：

### 创建助手
```
POST /v1/assistants
```
**状态**: 未实现

### 获取助手
```
GET /v1/assistants/{id}
```
**状态**: 未实现

### 更新助手
```
POST /v1/assistants/{id}
```
**状态**: 未实现

### 删除助手
```
DELETE /v1/assistants/{id}
```
**状态**: 未实现

### 列出助手
```
GET /v1/assistants
```
**状态**: 未实现

### 创建助手文件
```
POST /v1/assistants/{id}/files
```
**状态**: 未实现

### 获取助手文件
```
GET /v1/assistants/{id}/files/{fileId}
```
**状态**: 未实现

### 删除助手文件
```
DELETE /v1/assistants/{id}/files/{fileId}
```
**状态**: 未实现

### 列出助手文件
```
GET /v1/assistants/{id}/files
```
**状态**: 未实现

---

## 线程操作（未实现）

以下线程相关接口暂未实现：

### 创建线程
```
POST /v1/threads
```
**状态**: 未实现

### 获取线程
```
GET /v1/threads/{id}
```
**状态**: 未实现

### 更新线程
```
POST /v1/threads/{id}
```
**状态**: 未实现

### 删除线程
```
DELETE /v1/threads/{id}
```
**状态**: 未实现

### 创建消息
```
POST /v1/threads/{id}/messages
```
**状态**: 未实现

### 获取消息
```
GET /v1/threads/{id}/messages/{messageId}
```
**状态**: 未实现

### 更新消息
```
POST /v1/threads/{id}/messages/{messageId}
```
**状态**: 未实现

### 获取消息文件
```
GET /v1/threads/{id}/messages/{messageId}/files/{filesId}
```
**状态**: 未实现

### 列出消息文件
```
GET /v1/threads/{id}/messages/{messageId}/files
```
**状态**: 未实现

### 创建运行
```
POST /v1/threads/{id}/runs
```
**状态**: 未实现

### 获取运行
```
GET /v1/threads/{id}/runs/{runsId}
```
**状态**: 未实现

### 更新运行
```
POST /v1/threads/{id}/runs/{runsId}
```
**状态**: 未实现

### 列出运行
```
GET /v1/threads/{id}/runs
```
**状态**: 未实现

### 提交工具输出
```
POST /v1/threads/{id}/runs/{runsId}/submit_tool_outputs
```
**状态**: 未实现

### 取消运行
```
POST /v1/threads/{id}/runs/{runsId}/cancel
```
**状态**: 未实现

### 获取运行步骤
```
GET /v1/threads/{id}/runs/{runsId}/steps/{stepId}
```
**状态**: 未实现

### 列出运行步骤
```
GET /v1/threads/{id}/runs/{runsId}/steps
```
**状态**: 未实现

---

## 错误响应格式

当接口返回错误时，统一使用以下格式：

```json
{
  "error": {
    "message": "Error message here",
    "type": "error_type",
    "param": "parameter_name",
    "code": "error_code"
  }
}
```

**错误字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| message | string | 错误消息 |
| type | string | 错误类型 |
| param | string | 相关参数（可选） |
| code | string | 错误代码（可选） |

### 常见错误类型

| 类型 | 说明 |
|------|------|
| invalid_request_error | 请求参数无效 |
| one_api_error | 系统内部错误 |
| upstream_error | 上游服务错误 |
| authentication_error | 认证失败 |
| rate_limit_error | 请求频率限制 |
| api_not_implemented | 接口未实现 |

---

## 流式响应（SSE）

对于支持流式输出的接口（如 `stream: true`），服务器会返回 SSE（Server-Sent Events）格式的数据：

```
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

**说明**: 每个 `data:` 行包含一个 JSON 对象，最后以 `data: [DONE]` 表示流结束。
