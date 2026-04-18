# 快速开始

本文档帮助你在本地或生产环境快速启动 hermes-ai（AI Gateway）。

---

## 目录

- [前置条件](#前置条件)
- [本地开发](#本地开发)
- [前端构建](#前端构建)
- [Docker 部署](#docker-部署)
- [生产部署](#生产部署)
- [首次使用](#首次使用)
- [项目结构](#项目结构)
- [相关文档](#相关文档)

---

## 前置条件

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| Go | 1.25+ | 后端运行时 |
| Node.js | 16+ | 前端构建（如需自定义主题） |
| MySQL | 5.7+ | 主数据库，也可用 PostgreSQL |
| Redis | 6+ | 可选，用于缓存和限流 |

---

## 本地开发

### 1. 克隆项目

```bash
git clone <repository-url>
cd hermes-ai
```

### 2. 安装 Go 依赖

```bash
go mod tidy
```

### 3. 初始化数据库

进入 MySQL 终端，执行：

```bash
mysql -u root -p
```

```sql
source db.sql;
```

或使用命令行直接导入：

```bash
mysql -u root -p < db.sql
```

### 4. 配置环境变量

复制示例配置文件：

```bash
cp .env.example .env
```

编辑 `.env`，修改数据库连接信息：

```dotenv
PORT=1337
SQL_DSN=root:root123456@tcp(localhost:3306)/ai_gateway
REDIS_CONN_STRING=redis://:@localhost:6379/0?dial_timeout=3&db=1&read_timeout=6s&max_retries=2
DEBUG=false
DEBUG_SQL=true
```

> 完整环境变量说明请参考 [env.md](env.md)。

### 5. 启动服务

```bash
go run main.go
```

服务启动后，控制台会输出：

```
server started on http://localhost:1337
```

打开浏览器访问 `http://localhost:1337` 即可进入管理后台。

> 系统启动时会自动初始化 root 管理员账号（如未配置，请检查数据库连接和日志）。

---

## 前端构建

项目内置三套前端主题：`default`、`berry`、`air`。默认已提供预构建的静态文件，一般无需重新构建。

如需修改前端或提交新主题：

```bash
cd web
bash build.sh
```

构建完成后，静态文件会输出到 `web/build/<theme>/` 目录，后端启动时会自动嵌入服务。

> 新主题提交指南请参考 [web/README.md](web/README.md)。

---

## Docker 部署

项目提供多阶段 Dockerfile，同时构建前端和后端：

```bash
# 构建镜像
docker build -t hermes-ai:latest .

# 运行容器
docker run -d \
  -p 1337:3000 \
  -v /path/to/data:/data \
  -e SQL_DSN="root:root123456@tcp(host.docker.internal:3306)/ai_gateway" \
  -e REDIS_CONN_STRING="redis://host.docker.internal:6379/0" \
  --name hermes-ai \
  hermes-ai:latest
```

> 注意：容器内部监听端口为 `3000`，映射到宿主机 `1337`。

---

## 生产部署

### 使用 Systemd（推荐）

项目提供了 systemd 服务模板 [ai-gateway.service](ai-gateway.service)：

```bash
# 1. 编译二进制
go build -ldflags "-s -w" -o ai-gateway main.go

# 2. 复制服务文件
sudo cp ai-gateway.service /etc/systemd/system/

# 3. 编辑服务文件，修改路径和用户信息
sudo vim /etc/systemd/system/ai-gateway.service

# 4. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable ai-gateway
sudo systemctl start ai-gateway
sudo systemctl status ai-gateway
```

### 平滑重启

服务支持平滑退出，通过 `GRACEFUL_WAIT` 环境变量控制等待时间（默认 5 秒）。

```bash
# 查看日志
sudo journalctl -u ai-gateway -f
```

---

## 首次使用

### 1. 登录管理后台

访问 `http://localhost:1337`，使用初始管理员账号登录。

### 2. 配置渠道（Channel）

进入 **渠道管理**，添加 AI 提供商的接入信息：

- 选择渠道类型（OpenAI、DeepSeek、Anthropic 等）
- 填写 API Key 和 Base URL
- 配置支持的模型列表
- 设置分组和权重

### 3. 创建 API Key（令牌）

进入 **令牌管理**，创建用于 API 调用的 Key：

- 设置额度限制
- 选择可访问的模型
- 复制生成的 `sk-xxx` 密钥

### 4. 测试 API 调用

使用 curl 测试网关代理：

```bash
curl http://127.0.0.1:1337/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxx" \
  -d '{
    "model": "deepseek-chat",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "你好"}
    ],
    "stream": false
  }'
```

或使用 Go SDK：[voocel/litellm](https://github.com/voocel/litellm)。

更多 API 说明请参考 [relay.md](relay.md)。

---

## 项目结构

```
.
├── main.go                 # 程序入口
├── go.mod / go.sum         # Go 依赖
├── db.sql                  # 数据库初始化脚本
├── .env.example            # 环境变量示例
├── Dockerfile              # 容器构建
├── ai-gateway.service      # systemd 服务模板
│
├── internal/               # 核心业务代码
│   ├── domain/             # 领域层（实体、仓库接口）
│   ├── application/        # 应用层
│   ├── infras/             # 基础设施层（配置、数据库、缓存、日志、限流等）
│   ├── interfaces/         # 接口层（HTTP handlers、路由、中间件）
│   └── providers/          # 服务提供者注册
│
├── web/                    # 前端 React 项目
│   ├── default/            # 默认主题
│   ├── berry/              # Berry 主题
│   ├── air/                # Air 主题
│   └── build.sh            # 前端构建脚本
│
├── client/                 # 客户端调用示例
├── docs/                   # 文档与截图
└── bin/                    # 辅助脚本
```

---

## 相关文档

| 文档 | 说明 |
|------|------|
| [README.md](README.md) | 项目简介与核心功能 |
| [env.md](env.md) | 完整环境变量配置说明 |
| [relay.md](relay.md) | Relay API 接口文档（OpenAI 兼容格式） |
| [docs/API.md](docs/API.md) | 管理后台扩展 API 文档 |
| [web/README.md](web/README.md) | 前端主题开发指南 |
| [client/readme.md](client/readme.md) | 客户端调用示例 |

---

## 常见问题

**Q: 启动时提示数据库连接失败？**

确保 MySQL 服务已启动，且 `.env` 中的 `SQL_DSN` 配置正确。如果 MySQL 通过 Docker 运行，注意使用 `host.docker.internal` 或容器网络地址。

**Q: Redis 是必需的吗？**

不是。如果 `REDIS_CONN_STRING` 为空，系统会禁用 Redis 相关功能（如分布式限流和缓存同步），部分功能会退化为本地内存实现。

**Q: 如何切换前端主题？**

目前主题在构建时确定，多主题切换需要重新构建前端。默认主题为 `default`。

**Q: 如何查看请求日志？**

开启 `DEBUG_SQL=true` 可在控制台查看 SQL 日志；也可配置 `LOG_DIR` 将日志写入文件。
