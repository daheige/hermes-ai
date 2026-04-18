# Hermes AI API 文档

本文档包含 Hermes AI 系统的所有 API 接口信息（Relay 路由除外）。

## 目录

1. [公共接口](#公共接口)
2. [认证接口](#认证接口)
3. [用户接口](#用户接口)
4. [令牌接口](#令牌接口)
5. [渠道接口](#渠道接口)
6. [日志接口](#日志接口)
7. [兑换码接口](#兑换码接口)
8. [配置选项接口](#配置选项接口)
9. [分组接口](#分组接口)
10. [模型接口](#模型接口)
11. [账单接口](#账单接口)
12. [Relay 路由](#relay-路由)
13. [中间件说明](#中间件说明)

---

## 公共接口

### 获取系统状态

```
GET /api/status
```

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 系统状态数据 |
| data.version | string | 系统版本 |
| data.start_time | int64 | 启动时间戳 |
| data.email_verification | boolean | 是否启用邮箱验证 |
| data.github_oauth | boolean | 是否启用 GitHub OAuth |
| data.github_client_id | string | GitHub Client ID |
| data.lark_client_id | string | Lark Client ID |
| data.system_name | string | 系统名称 |
| data.logo | string | Logo URL |
| data.footer_html | string | 页脚 HTML |
| data.wechat_qrcode | string | 微信二维码 URL |
| data.wechat_login | boolean | 是否启用微信登录 |
| data.server_address | string | 服务器地址 |
| data.turnstile_check | boolean | 是否启用 Turnstile 验证 |
| data.turnstile_site_key | string | Turnstile Site Key |
| data.top_up_link | string | 充值链接 |
| data.chat_link | string | 聊天链接 |
| data.quota_per_unit | int64 | 每单位配额 |
| data.display_in_currency | boolean | 是否以货币形式显示 |
| data.oidc | boolean | 是否启用 OIDC |
| data.oidc_client_id | string | OIDC Client ID |
| data.oidc_well_known | string | OIDC Well-Known URL |
| data.oidc_authorization_endpoint | string | OIDC 授权端点 |
| data.oidc_token_endpoint | string | OIDC Token 端点 |
| data.oidc_userinfo_endpoint | string | OIDC UserInfo 端点 |

### 获取公告

```
GET /api/notice
```

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | string | 公告内容 |

### 获取关于信息

```
GET /api/about
```

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | string | 关于信息内容 |

### 获取首页内容

```
GET /api/home_page_content
```

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | string | 首页内容 |

### 发送邮箱验证邮件

```
GET /api/verification?email={email}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |

**中间件**: CriticalRateLimit, TurnstileCheck

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 发送密码重置邮件

```
GET /api/reset_password?email={email}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |

**中间件**: CriticalRateLimit, TurnstileCheck

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 重置密码

```
POST /api/user/reset
```

**请求体**:

```json
{
  "email": "user@example.com",
  "token": "reset_token"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| token | string | 是 | 重置令牌 |

**中间件**: CriticalRateLimit

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | string | 新密码（重置成功时） |

---

## 认证接口

### 用户登录

```
POST /api/user/login
```

**请求体**:

```json
{
  "username": "root",
  "password": "123456"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**中间件**: CriticalRateLimit

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 用户信息（登录成功时） |

### 用户注册

```
POST /api/user/register
```

**请求体**:

```json
{
  "username": "newuser",
  "password": "password123",
  "display_name": "New User",
  "email": "user@example.com",
  "verification_code": "123456",
  "aff_code": ""
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| display_name | string | 否 | 显示名称 |
| email | string | 条件 | 邮箱地址（启用邮箱验证时必填） |
| verification_code | string | 条件 | 验证码（启用邮箱验证时必填） |
| aff_code | string | 否 | 邀请码 |

**中间件**: CriticalRateLimit, TurnstileCheck

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 用户注销

```
GET /api/user/logout
```

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### GitHub OAuth

```
GET /api/oauth/github?code={code}&state={state}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | GitHub 授权码 |
| state | string | 是 | 状态码 |

**中间件**: CriticalRateLimit

### OIDC OAuth

```
GET /api/oauth/oidc?code={code}&state={state}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | OIDC 授权码 |
| state | string | 是 | 状态码 |

**中间件**: CriticalRateLimit

### Lark OAuth

```
GET /api/oauth/lark?code={code}&state={state}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | Lark 授权码 |
| state | string | 是 | 状态码 |

**中间件**: CriticalRateLimit

### 获取 OAuth State

```
GET /api/oauth/state
```

**中间件**: CriticalRateLimit

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | string | State 值 |

### 微信登录

```
GET /api/oauth/wechat?code={code}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 微信授权码 |

**中间件**: CriticalRateLimit

### 绑定微信

```
GET /api/oauth/wechat/bind?code={code}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 微信授权码 |

**中间件**: CriticalRateLimit, UserAuth

### 绑定邮箱

```
POST /api/oauth/email/bind
```

**中间件**: CriticalRateLimit, UserAuth

---

## 用户接口

### 管理员充值

```
POST /api/topup
```

**中间件**: AdminAuth

### 获取用户仪表盘

```
GET /api/user/dashboard
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 日志统计数据 |

### 获取当前用户信息

```
GET /api/user/self
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 用户信息 |

### 更新当前用户信息

```
PUT /api/user/self
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 删除当前用户

```
DELETE /api/user/self
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 生成访问令牌

```
GET /api/user/token
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | string | 访问令牌 |

### 获取邀请码

```
GET /api/user/aff
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | string | 邀请码 |

### 用户充值

```
POST /api/user/topup
```

**请求体**:

```json
{
  "key": "redemption_key"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | 是 | 兑换码 |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | int64 | 充值额度 |

### 获取用户可用模型

```
GET /api/user/available_models
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 可用模型列表 |

### 获取所有用户（管理员）

```
GET /api/user/?p={page}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| p | int | 否 | 页码，从 0 开始 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 用户列表 |

### 搜索用户（管理员）

```
GET /api/user/search?keyword={keyword}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 用户列表 |

### 获取单个用户（管理员）

```
GET /api/user/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 用户 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 用户信息 |

### 创建用户（管理员）

```
POST /api/user/
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 管理用户（管理员）

```
POST /api/user/manage
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 更新用户（管理员）

```
PUT /api/user/
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 删除用户（管理员）

```
DELETE /api/user/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 用户 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

---

## 令牌接口

### 获取所有令牌

```
GET /api/token/?p={page}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| p | int | 否 | 页码，从 0 开始 |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 令牌列表 |

### 搜索令牌

```
GET /api/token/search?keyword={keyword}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 令牌列表 |

### 获取单个令牌

```
GET /api/token/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 令牌 ID |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 令牌信息 |

### 创建令牌

```
POST /api/token/
```

**请求体**:

```json
{
  "name": "My Token",
  "expired_time": 0,
  "remain_quota": 500000,
  "unlimited_quota": false,
  "models": "gpt-3.5-turbo,gpt-4",
  "subnet": "192.168.0.0/24"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 令牌名称，最多 30 字符 |
| expired_time | int64 | 否 | 过期时间戳，0 或 -1 表示永不过期 |
| remain_quota | int64 | 否 | 剩余额度 |
| unlimited_quota | boolean | 否 | 是否无限额度 |
| models | string | 否 | 允许的模型，逗号分隔 |
| subnet | string | 否 | 允许的网段 |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 创建的令牌信息 |

### 更新令牌

```
PUT /api/token/?status_only={status_only}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status_only | string | 否 | 如果非空，只更新状态 |

**请求体**:

```json
{
  "id": 1,
  "name": "Updated Token",
  "expired_time": 0,
  "remain_quota": 1000000,
  "unlimited_quota": false,
  "models": "gpt-3.5-turbo",
  "subnet": "",
  "status": 1
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 令牌 ID |
| name | string | 否 | 令牌名称 |
| expired_time | int64 | 否 | 过期时间戳 |
| remain_quota | int64 | 否 | 剩余额度 |
| unlimited_quota | boolean | 否 | 是否无限额度 |
| models | string | 否 | 允许的模型 |
| subnet | string | 否 | 允许的网段 |
| status | int | 否 | 状态（1=启用） |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 更新后的令牌信息 |

### 删除令牌

```
DELETE /api/token/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 令牌 ID |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

---

## 渠道接口

### 获取所有渠道

```
GET /api/channel/?p={page}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| p | int | 否 | 页码，从 0 开始 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 渠道列表 |

### 搜索渠道

```
GET /api/channel/search?keyword={keyword}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 渠道列表 |

### 获取所有模型

```
GET /api/channel/models
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| object | string | 对象类型 |
| data | array | 模型列表 |

### 获取单个渠道

```
GET /api/channel/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 渠道 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 渠道信息 |

### 测试所有渠道

```
GET /api/channel/test
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 测试单个渠道

```
GET /api/channel/test/{id}?model={model}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 渠道 ID |

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 否 | 测试模型 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| time | float64 | 响应时间（秒） |
| modelName | string | 测试模型 |

### 更新所有渠道余额

```
GET /api/channel/update_balance
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 更新渠道余额

```
GET /api/channel/update_balance/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 渠道 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| balance | float64 | 余额 |

### 创建渠道

```
POST /api/channel/
```

**请求体**:

```json
{
  "type": 1,
  "key": "sk-xxx",
  "name": "OpenAI Channel",
  "base_url": "https://api.openai.com",
  "models": "gpt-3.5-turbo,gpt-4",
  "group": "default",
  "model_mapping": "{}",
  "status": 1
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 是 | 渠道类型 |
| key | string | 是 | 密钥 |
| name | string | 是 | 渠道名称 |
| base_url | string | 否 | 基础 URL |
| models | string | 否 | 支持的模型 |
| group | string | 否 | 分组 |
| model_mapping | string | 否 | 模型映射 |
| status | int | 否 | 状态 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

### 更新渠道

```
PUT /api/channel/
```

**请求体**:

```json
{
  "id": 1,
  "type": 1,
  "key": "sk-xxx",
  "name": "Updated Channel",
  "base_url": "https://api.openai.com",
  "models": "gpt-3.5-turbo,gpt-4",
  "group": "default",
  "model_mapping": "{}",
  "status": 1
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 渠道 ID |
| type | int | 否 | 渠道类型 |
| key | string | 否 | 密钥 |
| name | string | 否 | 渠道名称 |
| base_url | string | 否 | 基础 URL |
| models | string | 否 | 支持的模型 |
| group | string | 否 | 分组 |
| model_mapping | string | 否 | 模型映射 |
| status | int | 否 | 状态 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 更新后的渠道信息 |

### 删除禁用渠道

```
DELETE /api/channel/disabled
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | int64 | 删除数量 |

### 删除渠道

```
DELETE /api/channel/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 渠道 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

---

## 日志接口

### 获取所有日志

```
GET /api/log/?p={page}&type={type}&start_timestamp={start}&end_timestamp={end}&username={username}&token_name={token_name}&model_name={model_name}&channel={channel}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| p | int | 否 | 页码 |
| type | int | 否 | 日志类型 |
| start_timestamp | int64 | 否 | 开始时间戳 |
| end_timestamp | int64 | 否 | 结束时间戳 |
| username | string | 否 | 用户名 |
| token_name | string | 否 | 令牌名称 |
| model_name | string | 否 | 模型名称 |
| channel | int | 否 | 渠道 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 日志列表 |

### 删除历史日志

```
DELETE /api/log/?target_timestamp={timestamp}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_timestamp | int64 | 是 | 目标时间戳 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | int64 | 删除数量 |

### 获取日志统计

```
GET /api/log/stat?type={type}&start_timestamp={start}&end_timestamp={end}&token_name={token_name}&username={username}&model_name={model_name}&channel={channel}
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 统计数据 |
| data.quota | int64 | 使用额度 |

### 获取当前用户日志统计

```
GET /api/log/self/stat?type={type}&start_timestamp={start}&end_timestamp={end}&token_name={token_name}&model_name={model_name}&channel={channel}
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 统计数据 |
| data.quota | int64 | 使用额度 |

### 搜索所有日志

```
GET /api/log/search?keyword={keyword}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 日志列表 |

### 获取当前用户日志

```
GET /api/log/self?p={page}&type={type}&start_timestamp={start}&end_timestamp={end}&token_name={token_name}&model_name={model_name}
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 日志列表 |

### 搜索当前用户日志

```
GET /api/log/self/search?keyword={keyword}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 日志列表 |

---

## 兑换码接口

### 获取所有兑换码

```
GET /api/redemption/?p={page}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| p | int | 否 | 页码 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 兑换码列表 |

### 搜索兑换码

```
GET /api/redemption/search?keyword={keyword}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 兑换码列表 |

### 获取单个兑换码

```
GET /api/redemption/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 兑换码 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 兑换码信息 |

### 创建兑换码

```
POST /api/redemption/
```

**请求体**:

```json
{
  "name": "Redemption Batch",
  "count": 10,
  "quota": 500000
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 批次名称（1-20字符） |
| count | int | 是 | 生成数量（1-100） |
| quota | int64 | 是 | 每个兑换码的额度 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 兑换码列表 |

### 更新兑换码

```
PUT /api/redemption/?status_only={status_only}
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status_only | string | 否 | 如果非空，只更新状态 |

**请求体**:

```json
{
  "id": 1,
  "name": "Updated Name",
  "quota": 1000000,
  "status": 1
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 兑换码 ID |
| name | string | 否 | 名称 |
| quota | int64 | 否 | 额度 |
| status | int | 否 | 状态 |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 更新后的兑换码信息 |

### 删除兑换码

```
DELETE /api/redemption/{id}
```

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 兑换码 ID |

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

---

## 配置选项接口

### 获取所有配置选项

```
GET /api/option/
```

**中间件**: RootAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 配置选项列表 |

### 更新配置选项

```
PUT /api/option/
```

**请求体**:

```json
{
  "key": "SystemName",
  "value": "New System Name"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | 是 | 选项键 |
| value | string | 否 | 选项值 |

**中间件**: RootAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |

---

## 分组接口

### 获取所有分组

```
GET /api/group/
```

**中间件**: AdminAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | array | 分组名称列表 |

---

## 模型接口

### 仪表盘获取模型列表

```
GET /api/models
```

**中间件**: UserAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 响应消息 |
| data | object | 渠道类型到模型列表的映射 |

---

## 账单接口

### 获取订阅信息

```
GET /dashboard/billing/subscription
GET /v1/dashboard/billing/subscription
```

**中间件**: TokenAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| object | string | 对象类型 |
| has_payment_method | boolean | 是否有支付方式 |
| soft_limit_usd | float64 | 软限制（USD） |
| hard_limit_usd | float64 | 硬限制（USD） |
| system_hard_limit_usd | float64 | 系统硬限制（USD） |
| access_until | int64 | 访问截止时间 |

### 获取使用情况

```
GET /dashboard/billing/usage
GET /v1/dashboard/billing/usage
```

**中间件**: TokenAuth

**响应字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| object | string | 对象类型 |
| total_usage | float64 | 总使用量 |

---

## Relay 路由

以下接口通过 API Token（`Authorization: Bearer sk-xxx`）认证，由网关统一转发到上游渠道服务。

### 支持的 OpenAI 兼容接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/v1/models` | GET | 获取模型列表 |
| `/v1/completions` | POST | 文本补全 |
| `/v1/chat/completions` | POST | 聊天补全 |
| `/v1/embeddings` | POST | 文本嵌入 |
| `/v1/images/generations` | POST | 图像生成 |
| `/v1/audio/transcriptions` | POST | 音频转录 |
| `/v1/audio/translations` | POST | 音频翻译 |
| `/v1/audio/speech` | POST | 语音合成 |
| `/v1/moderations` | POST | 内容审核 |

### Anthropic Messages 接口

```
POST /v1/messages
```

**说明**：原生 Anthropic Messages API 转发接口，用于 Claude Code 等直接调用 Anthropic 原生协议的场景。

**请求头**：

| 字段 | 说明 |
|------|------|
| `Authorization` | `Bearer sk-xxx`（API Token） |
| `anthropic-version` | 可选，默认为 `2023-06-01` |
| `anthropic-beta` | 可选，系统会自动设置必要的 beta 头 |

**请求体示例**：

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

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 是 | 模型名称 |
| messages | array | 是 | 消息列表，content 支持字符串或数组格式 |
| max_tokens | int | 条件 | 最大生成 token 数，不传时默认 4096 |
| stream | boolean | 否 | 是否开启流式响应 |
| system | string/array | 否 | 系统提示词，支持字符串或数组格式 |
| temperature | float64 | 否 | 采样温度 |
| top_p | float64 | 否 | 核采样概率 |
| top_k | int | 否 | Top-K 采样 |
| tools | array | 否 | 工具列表 |
| tool_choice | any | 否 | 工具选择策略 |

**行为说明**：

- 该接口会基于请求体中的 `model` 字段进行渠道路由与配额校验。
- 请求体会**原样透传**到 Anthropic 上游，响应也保持原生 Anthropic JSON / SSE 格式，不做 OpenAI 格式转换。
- 支持失败重试、渠道自动禁用、配额预扣与后扣等网关通用能力。

---

## 中间件说明

| 中间件 | 说明 |
|--------|------|
| UserAuth | 需要用户登录（Session 认证） |
| AdminAuth | 需要管理员权限 |
| RootAuth | 需要超级管理员权限 |
| TokenAuth | 需要 API Token 认证 |
| CriticalRateLimit | 关键接口限流 |
| GlobalAPIRateLimit | 全局 API 限流 |
| DownloadRateLimit | 下载接口限流 |
| UploadRateLimit | 上传接口限流 |
| CORS | 跨域支持 |
| TurnstileCheck | Turnstile 验证检查 |
| Distributor | Relay 路由渠道分发 |
| RelayPanicRecover | Relay 异常恢复 |
