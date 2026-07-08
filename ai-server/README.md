# ai-server — AI 对话与语义搜索服务

AI 智能服务，提供流式 AI 对话、知识库语义搜索、文章向量化等核心 AI 功能，是 KnowSync 的智能核心。

## 技术栈

- **RPC 框架**：gRPC（端口 50053）
- **LLM 编排**：CloudWeGo Eino（DeepSeek API）
- **向量嵌入**：Eino-ext（SiliconFlow API，模型 Qwen/Qwen3-Embedding-8B，1536 维）
- **向量存储**：Chromem-go（本地嵌入式向量数据库）
- **数据库**：GORM + MySQL 8.0+
- **缓存**：go-redis/v9（用于搜索缓存）
- **配置管理**：spf13/viper
- **ID 生成**：xid

## 目录结构

```
ai-server/
├── cmd/                    # 启动入口
├── internal/
│   ├── app/                # 应用初始化
│   ├── chunker/            # 文本分块策略
│   ├── embedder/           # 向量嵌入生成（SiliconFlow API）
│   ├── handler/            # gRPC 处理方法
│   ├── llm/                # LLM 交互（Eino + DeepSeek）
│   ├── repository/         # 数据访问层（会话、消息）
│   ├── service/            # 业务逻辑层
│   │   └── compactor/      # 对话上下文管理（溢出检测、摘要、裁剪）
│   ├── vectorstore/        # 向量存储（Chromem-go）
│   └── worker/             # 异步向量化工作队列
├── config.yaml
└── go.mod
```

## 核心功能

### AI 流式对话
- 基于 DeepSeek 模型的流式对话（通过 Eino 框架）
- 支持思考过程（Thinking）与回答（Content）分 SSE Event 返回
- 支持反问用户（Ask User）事件（单选/多选）
- 首次对话自动生成会话标题

### 语义搜索
- 基于向量嵌入的语义搜索
- 搜索 Redis 缓存（MD5 哈希关键词，不同用户共享搜索结果）
- 仅搜索公开知识库

### 文章向量化
- 自动将知识库文章向量化
- 异步工作队列处理，不阻塞主流程
- 支持全量向量化和增量更新

### 会话管理
- 会话创建、消息分页查询、删除
- 上下文管理：LLM 上下文窗口溢出时自动摘要压缩（compactor）

### 可见性同步
- 知识库可见性变更时同步更新向量数据访问权限

## 配置项

| 配置项 | 环境变量 | 说明 |
|--------|---------|------|
| `server.port` | `KNOWSYNC_AI_SERVER_SERVER_PORT` | gRPC 端口（默认 50053） |
| `mysql.dsn` | `KNOWSYNC_AI_SERVER_MYSQL_DSN` | MySQL 连接字符串 |
| `redis.addr` | `KNOWSYNC_AI_SERVER_REDIS_ADDR` | Redis 地址 |
| `llm.api_key` | `KNOWSYNC_AI_SERVER_LLM_API_KEY` | DeepSeek API Key |
| `embedder.api_key` | `KNOWSYNC_AI_SERVER_EMBEDDER_API_KEY` | SiliconFlow API Key |
| `vectorstore.path` | `KNOWSYNC_AI_SERVER_VECTORSTORE_PATH` | 向量数据库持久化路径 |
| `worker.concurrency` | `KNOWSYNC_AI_SERVER_WORKER_CONCURRENCY` | 向量化工作并发数 |

## 启动

```bash
# 本地开发
task dev:ai-server

# 或直接使用 Go
go run cmd/main.go

# Docker
docker compose up ai-server
```

## API

对外接口通过网关暴露 HTTP API，详见 [gateway/docs/ai.jsonc](../gateway/docs/ai.jsonc)。
