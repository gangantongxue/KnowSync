# chat-server — 社交聊天与即时通讯服务

社交聊天微服务，提供好友关系管理、好友申请、私聊/群聊消息、SSE 实时推送、在线状态等即时通讯功能。

## 技术栈

- **RPC 框架**：gRPC（端口 50054）
- **SSE 推送**：Server-Sent Events（端口 50055）
- **ORM**：GORM + MySQL 8.0+
- **缓存**：go-redis/v9（在线状态、消息发布订阅）
- **认证**：golang-jwt/v5（验证 access_token）
- **配置管理**：spf13/viper
- **ID 生成**：xid

## 目录结构

```
chat-server/
├── cmd/                    # 启动入口
├── internal/
│   ├── app/                # 应用初始化
│   ├── handler/            # gRPC 处理方法
│   ├── service/            # 业务逻辑层
│   ├── repository/         # 数据访问层
│   │   └── postgres/       # PostgreSQL 适配（预留）
│   └── push/               # 推送连接管理（Hub、SSE 客户端）
├── pkg/
│   ├── config/             # 配置结构体
│   ├── database/           # GORM 连接与迁移
│   │   └── schema/         # 数据模型
│   └── logger/             # 日志
├── config.yaml
└── go.mod
```

## 核心功能

### 好友关系
- 双向好友关系，`friends` 表双向存储（每对存储两条记录）
- 好友申请全生命周期：发送 → 待处理(pending) → 接受(accepted)/拒绝(rejected)
- 好友备注管理
- 用户搜索（跨系统聊天用户缓存表）

### 私聊/群聊消息
- 统一 `messages` 表存储私聊与群聊消息
- 消息类型：text（文本）、image（图片）、file（文件）、system_invitation（系统邀请）
- 消息状态：normal（正常）、recalled（撤回）
- SeqID 单调递增，支持游标分页
- 消息撤回功能
- 消息转发

### 群组管理
- 群组 CRUD
- 群主转让
- 管理员/成员管理（添加/移除/角色变更）
- 群组角色：owner（群主）、admin（管理员）、member（成员）

### SSE 实时推送
- JWT access_token 作为查询参数认证（EventSource 无法自定义请求头）
- 实时推送新消息、好友申请、在线状态变更
- 连接管理：Hub 连接池维护，30s 心跳注释行保活并续期 Redis 在线状态

### 会话管理（Conversation）
- 私聊/群聊统一会话列表
- 已读标记（基于 SeqID）
- 置顶/取消置顶
- 删除会话（仅删除本地记录，不影响消息）

### 在线状态
- 基于 Redis 维护（key: `user_online:<user_id>`, TTL: 60s）
- 心跳续期

## 配置项

| 配置项 | 环境变量 | 说明 |
|--------|---------|------|
| `server.port` | `KNOWSYNC_CHAT_SERVER_SERVER_PORT` | gRPC 端口（默认 50054） |
| `sse.port` | `KNOWSYNC_CHAT_SERVER_SSE_PORT` | SSE 推送端口（默认 50055） |
| `mysql.dsn` | `KNOWSYNC_CHAT_SERVER_MYSQL_DSN` | MySQL 连接字符串 |
| `redis.addr` | `KNOWSYNC_CHAT_SERVER_REDIS_ADDR` | Redis 地址 |
| `jwt.public_key_path` | `KNOWSYNC_CHAT_SERVER_JWT_PUBLIC_KEY_PATH` | RSA 公钥路径（验证令牌） |

## 启动

```bash
# 本地开发
task dev:chat-server

# 或直接使用 Go
go run cmd/main.go

# Docker
docker compose up chat-server
```

## API

对外接口通过网关暴露 HTTP API，详见 [gateway/docs/chat.jsonc](../gateway/docs/chat.jsonc)。

## SSE

```
http(s)://<host>:50055/sse?token=<jwt_access_token>
```

连接建立后服务端立即返回 `: connected` 确认行，此后以默认事件（message）推送
`data: {"type": "...", "data": {...}}` JSON 事件，事件类型包括：
`new_message`、`message_recalled`、`friend_request`、`friend_accepted`。
每 30s 发送 `: ping` 注释行保活并刷新 Redis 在线状态 TTL。
前端通过 Caddy `/sse*` 反向代理访问，页面打开即保持全局长连接。
