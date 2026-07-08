# gateway — HTTP API 网关

KnowSync 的 HTTP API 网关，基于 CloudWeGo Hertz 框架。网关不包含业务逻辑，仅负责 HTTP 到 gRPC 的协议转换、认证校验、文件存储与静态文件服务。

## 技术栈

- **HTTP 框架**：CloudWeGo Hertz
- **中间件**：JWT 认证中间件、CORS 中间件、ServiceToken 认证中间件
- **JWT 验证**：golang-jwt/v5（持有 RSA 公钥）
- **Redis**：go-redis/v9（令牌黑名单）
- **gRPC 客户端**：连接 4 个后端 gRPC 服务
- **配置管理**：spf13/viper

## 目录结构

```
gateway/
├── cmd/                    # 启动入口
├── internal/
│   ├── app/                # 应用初始化
│   ├── handler/            # HTTP 处理方法（22 个文件，直接调用 gRPC）
│   └── router/             # 路由注册
├── pkg/
│   ├── config/             # 配置结构体
│   ├── errcode/            # 业务错误码定义
│   ├── grpcclient/         # gRPC 客户端连接管理
│   ├── middleware/          # HTTP 中间件
│   ├── response/           # 统一响应格式
│   ├── serviceauth/        # 服务间认证
│   └── storage/            # 文件存储（头像、文章）
├── docs/                   # API 接口文档（JSONC 格式）
│   ├── user.jsonc          # 用户服务 API 文档
│   ├── repo.jsonc          # 知识库服务 API 文档
│   ├── ai.jsonc            # AI 服务 API 文档
│   └── chat.jsonc          # 聊天服务 API 文档
├── config.yaml
└── go.mod
```

## 路由概览

所有 API 以 `/api/v1` 为前缀。

### 公共路由（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/refresh` | 刷新令牌 |
| POST | `/api/v1/verify-codes` | 发送验证码 |
| POST | `/api/v1/password/forget` | 忘记密码 |

### 已认证路由（需 Bearer Token）

涵盖用户管理、知识库管理、AI 对话、聊天社交四大模块，详见各 JSONC 接口文档。

### 内部路由（需 Service Token）

- `/internal/file` - 文件读取
- `/internal/repos/*` - 知识库内部操作
- `/internal/users/search` - 用户搜索

由 ai-server 通过内部 HTTP 调用，无需用户令牌。

### 静态文件

- `/files/*filepath` - 头像、文章 Markdown 文件等静态资源

## gRPC 客户端目标

| 服务 | 地址 | 端口 |
|------|------|------|
| user-server | `user-server` | 50051 |
| repo-server | `repo-server` | 50052 |
| ai-server | `ai-server` | 50053 |
| chat-server | `chat-server` | 50054 |

## 启动

```bash
# 本地开发
task dev:gateway

# 或直接使用 Go
go run cmd/main.go

# Docker
docker compose up gateway
```

## API 文档

所有 HTTP API 接口文档位于 `docs/` 目录，使用 JSONC 格式：

- [用户服务 API](docs/user.jsonc)
- [知识库服务 API](docs/repo.jsonc)
- [AI 服务 API](docs/ai.jsonc)
- [聊天服务 API](docs/chat.jsonc)
