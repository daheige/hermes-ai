# hermes-ai
AI LLM大模型网关系统，支持多租户管理、API Key管理、限流、日志审计等核心功能，其设计原型来自 OneAPI，该框架对分层设计和代码可维护性、可读性、可拓展性等方面做了大量优化，让ai 大模型 provider能快速接入和使用。

# 项目命名说明
赫尔墨斯 (Hermes) —— 希腊神话信使神‌。在希腊神话中，‌赫尔墨斯‌（Hermes）是宙斯之子，掌管商业、旅行、沟通、发明等众多领域。

- 象征物‌：双蛇杖（商神杖）、插翼凉鞋、翼帽。
- 罗马对应神‌：墨丘利（Mercury），象征速度与商业。
- 神话事迹‌：出生当晚即偷走阿波罗的牛，后发明里拉琴作为补偿，展现其机智与创造力 。
- 现代影响‌：其形象常被用于象征“快速传递信息”，如物流、通信行业品牌命名灵感来源 。

# 核心功能
- 多租户管理
- 虚拟API Key配置和删除
- 请求日志和审计
- Token消费统计
- 流量控制和限流
- 用户注册登录
- JWT身份认证
- RBAC权限管理
- 数据加密和脱敏
- 原生 Anthropic Messages API 转发
- 渠道自动测试与监控
- 批量更新与内存缓存
- 多主题前端支持（default / berry / air）

## 技术栈

- Go 1.25+
- Gin Web框架
- GORM ORM
- Redis缓存
- MySQL/PostgreSQL数据库
- JWT认证

## 架构特点

- **显式配置管理**：所有环境变量统一收敛到 `SystemConfig` 结构体，通过 `InitSystemConfig()` 显式初始化，彻底消除全局配置变量
- **依赖注入**：服务、仓库、中间件、Handler 均通过构造函数注入依赖，便于测试和替换实现
- **分层设计**：严格遵循 DDD 分层（Domain / Application / Infrastructure / Interfaces），职责边界清晰
- **AES-GCM 加密**：支持对 tokens、redemptions、channels 的 key 字段进行加密存储，保护敏感信息
- **原生 Anthropic Messages API**：支持 `/v1/messages` 原生转发，不做 OpenAI 格式转换，兼容 Claude Code 等客户端
- **渠道自动测试**：可配置自动轮询测试渠道可用性，自动禁用/启用异常渠道
- **批量更新**：支持配额批量更新，降低数据库写入压力
- **多主题前端**：内置 default、berry、air 三种前端主题

## 快速开始

1. 安装go依赖
```bash
go mod tidy
```

2. 初始化db
```shell
# 进入mysql终端后，执行该命令
source db.sql
```

3. 配置环境变量
   复制 `.env.example` 为 `.env`，修改数据库、Redis 连接信息，并设置 `AES_SECRET_KEY`（建议至少 16 字符的随机字符串，用于加密敏感字段）

4. 运行
```bash
go run main.go
```

## 网关代理
- POST /v1/chat/completions - AI模型请求代理，基于basic认证
- 请求demo见client中代码

### curl请求方式
```shell
curl http://127.0.0.1:1337/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxx" \
  -d '{
        "model": "deepseek-chat",
        "messages": [
          {"role": "system", "content": "You are a helpful assistant."},
          {"role": "user", "content": "go语言是什么"}
        ],
        "stream": false
      }'
```

# AI Gateway Relay API 接口文档
参考： [relay.md](relay.md)

# 关于优化
后续将对internal目录中所有不规范的db,redis操作以及分层设计逐步进行整改，满足以最小的人力成本，持续维护和构建项目。
