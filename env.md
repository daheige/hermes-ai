# 环境变量配置文档

本文档整理了项目中所有环境变量的含义、用途和配置方法。

## 目录

- [后端环境变量](#后端环境变量)
  - [基础服务配置](#基础服务配置)
  - [数据库配置](#数据库配置)
  - [Redis 配置](#redis-配置)
  - [调试与日志](#调试与日志)
  - [性能与限流](#性能与限流)
  - [安全与认证](#安全与认证)
  - [监控与指标](#监控与指标)
  - [Gemini 配置](#gemini-配置)
  - [主题配置](#主题配置)
  - [其他配置](#其他配置)
- [前端环境变量](#前端环境变量)
- [配置示例](#配置示例)
- [注意事项](#注意事项)

---

## 后端环境变量

### 基础服务配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `PORT` | int | `1337` | 服务器监听端口 |
| `GRACEFUL_WAIT` | int | `5` | 平滑退出等待时间（秒） |
| `GIN_MODE` | string | - | Gin 框架运行模式，设置为 `debug` 开启调试模式 |
| `FRONTEND_BASE_URL` | string | - | 前端基础 URL，主节点时会被忽略 |
| `NODE_TYPE` | string | - | 节点类型，`slave` 表示从节点，其他值表示主节点 |
| `POLLING_INTERVAL` | int | - | 轮询间隔（秒） |

### 数据库配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `SQL_DSN` | string | - | **必填**，主数据库连接字符串，支持 MySQL 和 PostgreSQL |
| `LOG_SQL_DSN` | string | - | 日志数据库连接字符串，为空则使用主数据库 |
| `SQL_MAX_IDLE_CONNS` | int | `100` | 数据库最大空闲连接数 |
| `SQL_MAX_OPEN_CONNS` | int | `1000` | 数据库最大打开连接数 |
| `SQL_MAX_LIFETIME` | int | `60` | 连接最大生命周期（秒） |

**连接字符串格式：**
- MySQL: `user:password@tcp(host:port)/dbname`
- PostgreSQL: `postgres://user:password@host:port/dbname`

### Redis 配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `REDIS_CONN_STRING` | string | - | Redis 连接字符串，为空则禁用 Redis |
| `REDIS_ENABLE_CLUSTER` | bool | `false` | 是否启用 Redis 集群模式 |
| `REDIS_PASSWORD` | string | - | Redis 密码（集群模式使用） |
| `REDIS_USERNAME` | string | - | Redis 用户名（集群模式使用） |
| `SYNC_FREQUENCY` | int | `600` | 数据同步频率（秒） |

**连接字符串格式：**
- 单机模式：`redis://[:password@]host:port/db[?options]`
- 集群模式：传入多个地址，用逗号分隔，如 `redis://host1:6379,redis://host2:6379`

示例：
```
redis://:@localhost:6379/0?dial_timeout=3&db=1&read_timeout=6s&max_retries=2
```

### 调试与日志

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `DEBUG` | bool | `false` | 启用调试模式 |
| `DEBUG_SQL` | bool | `false` | 启用 SQL 调试日志 |
| `LOG_LEVEL` | string | `info` | 日志级别，可选：`debug`、`info`、`warn`、`error` |
| `LOG_DIR` | string | - | 日志输出目录，为空则输出到标准输出 |

### 安全与认证

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `INITIAL_ROOT_TOKEN` | string | - | 初始化 Root 用户 Token |
| `INITIAL_ROOT_ACCESS_TOKEN` | string | - | 初始化 Root 用户 Access Token |

### 性能优化

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `MEMORY_CACHE_ENABLED` | bool | `false` | 启用内存缓存 |
| `BATCH_UPDATE_ENABLED` | bool | `false` | 启用批量更新 |
| `BATCH_UPDATE_INTERVAL` | int | `5` | 批量更新间隔（秒） |
| `CHANNEL_TEST_FREQUENCY` | int | - | 渠道测试频率（秒），设置后自动测试渠道 |

### 限流配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `GLOBAL_API_RATE_LIMIT` | int | `480` | 全局 API 限流数量 |
| `GLOBAL_WEB_RATE_LIMIT` | int | `240` | 全局 Web 限流数量 |

### 代理配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `RELAY_PROXY` | string | - | 中继代理地址 |
| `USER_CONTENT_REQUEST_PROXY` | string | - | 用户内容请求代理 |
| `USER_CONTENT_REQUEST_TIMEOUT` | int | `30` | 用户内容请求超时（秒） |
| `RELAY_TIMEOUT` | int | `0` | 中继超时时间（秒），0 表示不限制 |

### 监控指标

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ENABLE_METRIC` | bool | `false` | 启用指标监控 |
| `METRIC_QUEUE_SIZE` | int | `10` | 指标队列大小 |
| `METRIC_SUCCESS_RATE_THRESHOLD` | float | `0.8` | 成功率阈值 |
| `METRIC_SUCCESS_CHAN_SIZE` | int | `1024` | 成功指标通道大小 |
| `METRIC_FAIL_CHAN_SIZE` | int | `128` | 失败指标通道大小 |

### Gemini 配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `GEMINI_SAFETY_SETTING` | string | `BLOCK_NONE` | Gemini 安全设置 |
| `GEMINI_VERSION` | string | `v1` | Gemini API 版本 |

### 主题配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `THEME` | string | `default` | 默认主题，可选：`default`、`berry`、`air` |

### 其他配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ENFORCE_INCLUDE_USAGE` | bool | `false` | 强制包含使用量信息 |
| `TEST_PROMPT` | string | `Output only your specific model name...` | 测试提示词 |

---

## 前端环境变量

### 通用环境变量（所有主题）

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `REACT_APP_SERVER` | string | `/` 或空 | API 服务器地址 |
| `REACT_APP_VERSION` | string | - | 应用版本号，显示在页面上 |
| `NODE_ENV` | string | - | 构建环境，`development` 或 `production` |
| `PUBLIC_URL` | string | - | 应用公共 URL（用于 Service Worker） |

### 使用位置

1. **API 基础 URL 配置**
   - `web/air/src/helpers/api.js:5`
   - `web/berry/src/utils/api.js:8`
   - `web/default/src/helpers/api.js:5`

2. **版本号显示**
   - `web/berry/src/contexts/StatusContext.js:26`
   - `web/berry/src/layout/MainLayout/Sidebar/index.js:43,58`
   - `web/berry/src/views/Setting/component/OtherSetting.js:115`
   - `web/berry/src/ui-component/Footer.js:19`
   - `web/air/src/components/Footer.js:44`
   - `web/default/src/App.js:63`
   - `web/default/src/components/Footer.js:42`
   - `web/default/src/components/OtherSetting.js:106`

3. **Service Worker**（仅在 production 模式）
   - `web/berry/src/serviceWorker.js:88,90,99`

4. **i18n 调试**
   - `web/default/src/i18n.js:12`

---

## 配置示例

### 最小配置（开发环境）

```bash
# 基础服务
PORT=1337

# 数据库（MySQL）
SQL_DSN=root:password@tcp(localhost:3306)/ai_gateway

# Redis（可选，但建议启用）
REDIS_CONN_STRING=redis://:@localhost:6379/0
SYNC_FREQUENCY=600
```

### 生产环境配置

```bash
# 基础服务
PORT=1337
GRACEFUL_WAIT=30
GIN_MODE=release

# 数据库
SQL_DSN=user:password@tcp(db:3306)/ai_gateway
LOG_SQL_DSN=user:password@tcp(logdb:3306)/ai_gateway_logs
SQL_MAX_IDLE_CONNS=100
SQL_MAX_OPEN_CONNS=1000

# Redis 集群
REDIS_CONN_STRING=redis://:password@redis1:6379,redis://:password@redis2:6379
REDIS_ENABLE_CLUSTER=true
REDIS_PASSWORD=password
REDIS_USERNAME=default
SYNC_FREQUENCY=60

# 日志
LOG_LEVEL=info
LOG_DIR=./logs

# 性能优化
MEMORY_CACHE_ENABLED=true
BATCH_UPDATE_ENABLED=true
BATCH_UPDATE_INTERVAL=10

# 监控
ENABLE_METRIC=true

# 代理（如需要）
RELAY_PROXY=http://proxy:8080
```

### 多节点部署配置

**主节点：**
```bash
NODE_TYPE=master
# 或省略 NODE_TYPE
```

**从节点：**
```bash
NODE_TYPE=slave
FRONTEND_BASE_URL=https://frontend.example.com
```

---

## 注意事项

1. **敏感信息**：所有包含 `Secret`、`Token`、`Password` 的环境变量不会通过 API 返回
2. **优先级**：环境变量优先级高于配置文件
3. **热更新**：部分配置修改后需要重启服务才能生效
4. **从节点**：从节点的 `FRONTEND_BASE_URL` 配置有效，主节点该配置会被忽略
