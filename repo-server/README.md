# repo-server — 知识库管理服务

知识库仓库管理微服务，负责知识库的 CRUD、文件/文章节点树管理、协作者权限控制、关注管理等功能。

## 技术栈

- **RPC 框架**：gRPC（端口 50052）
- **ORM**：GORM + MySQL 8.0+
- **配置管理**：spf13/viper
- **ID 生成**：xid

## 目录结构

```
repo-server/
├── cmd/                # 启动入口
├── internal/
│   ├── app/            # 应用初始化
│   ├── handler/        # gRPC 处理方法
│   ├── service/        # 业务逻辑层（含权限校验）
│   └── repository/     # 数据访问层
│       └── postgres/   # PostgreSQL 适配（预留）
├── pkg/
│   ├── config/         # 配置结构体
│   ├── database/       # GORM 连接与迁移
│   │   └── schema/     # 数据模型
│   └── logger/         # 日志
├── config.yaml
└── go.mod
```

## 核心功能

### 知识库管理
- 创建/获取/更新/删除知识库（软删除）
- 知识库可见性：PUBLIC（公开）/ PRIVATE（私有）
- 知识库创建者自动成为 ADMIN 角色

### 文件/文章节点树
- 树形目录结构，支持无限层级
- 节点类型：FOLDER（文件夹）/ ARTICLE（文章）
- 文章内容以 Markdown 文件形式上传到本地存储
- 获取签名临时 URL 访问文章内容（防越权）
- 递归删除节点及子节点

### 协作者权限控制
- 角色模型：ADMIN（管理员）/ DEVELOPER（开发者）/ VIEWER（只读者）
- 仅 ADMIN 可管理协作者与修改知识库配置
- 协作者邀请通过聊天系统发送系统消息

### 关注管理
- 用户可关注感兴趣的知识库
- 关注列表独立查询

## 配置项

| 配置项 | 环境变量 | 说明 |
|--------|---------|------|
| `server.port` | `KNOWSYNC_REPO_SERVER_SERVER_PORT` | gRPC 端口（默认 50052） |
| `mysql.dsn` | `KNOWSYNC_REPO_SERVER_MYSQL_DSN` | MySQL 连接字符串 |
| `storage.path` | `KNOWSYNC_REPO_SERVER_STORAGE_PATH` | 文件存储路径 |

## 启动

```bash
# 本地开发
task dev:repo-server

# 或直接使用 Go
go run cmd/main.go

# Docker
docker compose up repo-server
```

## API

对外接口通过网关暴露 HTTP API，详见 [gateway/docs/repo.jsonc](../gateway/docs/repo.jsonc)。
