# KnowSync

**KnowSync** 是一个基于微服务架构的知识库管理与协作平台，提供用户认证、知识库管理、AI 智能对话与语义搜索、社交聊天等核心功能。

## 技术栈

| 类别 | 技术选型 |
|------|---------|
| **后端语言** | Go 1.26+ |
| **RPC 框架** | gRPC（服务间通信） |
| **HTTP 框架** | CloudWeGo Hertz（网关层） |
| **AI 框架** | CloudWeGo Eino（LLM 编排） + SiliconFlow + DeepSeek |
| **数据库** | MySQL 8.0+（GORM） |
| **缓存** | Redis 7+（go-redis） |
| **向量数据库** | Chromem-go（本地嵌入式） |
| **认证** | JWT (RSA-2048) + Service Token |
| **前端** | React 19 + TypeScript + Vite + Tailwind CSS |
| **容器化** | Docker Compose |

## 项目结构

```
knowsync/
├── user-server/        # 用户认证与管理服务
├── repo-server/        # 知识库仓库管理服务
├── ai-server/          # AI 对话与语义搜索服务
├── chat-server/        # 社交聊天与即时通讯服务
├── gateway/            # HTTP API 网关（Hertz）
├── ks-proto/           # 共享 Protobuf 协议定义
├── web/                # 前端 SPA（React）
├── docs/               # 项目文档
├── scripts/            # 构建与部署脚本
├── docker-compose.yaml # 容器编排
└── Taskfile.yml        # 任务管理
```

## 系统架构

```
[浏览器] ←→ Caddy(:80/:443)
                │
                ▼
        Gateway(:8080, Hertz)
            ├── gRPC → user-server(:50051) ←→ MySQL + Redis
            ├── gRPC → repo-server(:50052) ←→ MySQL
            ├── gRPC → ai-server(:50053)   ←→ MySQL + Redis + Chromem-go
            └── gRPC → chat-server(:50054) ←→ MySQL + Redis + WebSocket(:50055)
```

各服务职责：
- **user-server**：用户注册/登录、JWT 令牌管理、邮箱验证码、密码管理
- **repo-server**：知识库 CRUD、文件/文章节点树管理、协作者权限控制
- **ai-server**：流式 AI 对话、语义搜索、文章向量化、会话管理
- **chat-server**：好友关系、私聊/群聊消息、WebSocket 实时推送、在线状态
- **gateway**：HTTP 网关，无业务逻辑，仅转换 HTTP → gRPC 并处理认证

## 快速启动

### 前置要求

- Go 1.26+
- MySQL 8.0+
- Redis 7+
- Docker（可选，用于容器化部署）

### 本地开发

1. 克隆并进入项目：

```bash
git clone <repo-url> && cd knowsync
```

2. 配置环境变量：

```bash
cp .env.example .env
# 编辑 .env 填写数据库、Redis、邮件等配置
```

3. 初始化数据库（自动迁移将在服务启动时执行）：

```bash
# 创建各服务对应的数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS knowsync_user;"
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS knowsync_repo;"
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS knowsync_ai;"
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS knowsync_chat;"
```

4. 启动服务（推荐使用 Taskfile）：

```bash
# 安装 task：go install github.com/go-task/task/v3/cmd/task@latest

# 启动所有服务
task dev:all

# 或逐个启动
task dev:user-server
task dev:repo-server
task dev:ai-server
task dev:chat-server
task dev:gateway
```

5. 构建前端（可选）：

```bash
cd web && npm install && npm run build
```

### Docker 部署

```bash
docker compose up -d
```

## 模块文档

各模块的详细文档位于对应目录的 `README.md`：

- [user-server](./user-server/README.md) — 用户认证与管理
- [repo-server](./repo-server/README.md) — 知识库管理
- [ai-server](./ai-server/README.md) — AI 对话与搜索
- [chat-server](./chat-server/README.md) — 社交聊天
- [gateway](./gateway/README.md) — HTTP 网关
- [ks-proto](./ks-proto/README.md) — 协议定义

## 设计文档

- [项目架构选型](./docs/项目架构选型.md) — 技术选型分析与决策记录
- [数据库设计](./docs/数据库设计.md) — 各服务的数据库表结构与关系
- [API 接口文档](./gateway/docs/) — 网关 HTTP API 接口文档（JSONC 格式）
  - [用户服务接口](./gateway/docs/user.jsonc)
  - [知识库服务接口](./gateway/docs/repo.jsonc)
  - [AI 服务接口](./gateway/docs/ai.jsonc)
  - [聊天服务接口](./gateway/docs/chat.jsonc)

## 贡献指南

1. 严格遵守三层架构：handler → service → repository
2. 禁止 JOIN 查询与数据库外键约束
3. 新增代码须与已有代码风格保持一致（见 AGENTS.md）
4. 所有日志使用中文，英文 snake_case 字段名
5. 提交前运行 `task lint` 和 `task test` 确保代码质量

## 协议

[GPL v3](LICENSE)
