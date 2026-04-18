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

## 技术栈

- Go 1.25+
- Gin Web框架
- GORM ORM
- Redis缓存
- MySQL数据库
- JWT认证

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

3. 配置数据库
   修改 `.env` 中的数据库和redis连接信息

3. 运行
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
