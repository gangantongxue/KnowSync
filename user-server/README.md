# user-server — 用户认证与管理服务

用户认证与管理微服务，负责用户注册、登录、JWT 令牌管理、邮箱验证、密码管理等核心认证功能。

## 技术栈

- **RPC 框架**：gRPC（端口 50051）
- **ORM**：GORM + MySQL 8.0+
- **缓存**：go-redis/v9
- **认证**：golang-jwt/v5（RSA-2048 签名）
- **密码加密**：bcrypt（golang.org/x/crypto）
- **邮件**：gomail.v2（SMTP）
- **配置管理**：spf13/viper
- **ID 生成**：用户 ID 数字自增，会话 ID 使用 xid

## 目录结构

```
user-server/
├── cmd/                # 启动入口
├── internal/
│   ├── app/            # 应用初始化与生命周期管理
│   ├── handler/        # gRPC 处理方法
│   ├── service/        # 业务逻辑层（含权限校验）
│   └── repository/     # 数据访问层
├── pkg/
│   ├── auth/           # JWT 签发与验证
│   ├── config/         # 配置结构体与解析
│   ├── database/       # GORM 自动迁移与连接
│   │   └── schema/     # 数据模型定义
│   ├── logger/         # 日志初始化
│   ├── mail/           # 邮件发送
│   └── redis/          # Redis 连接与操作
├── config.yaml         # 默认配置文件
└── go.mod
```

## 核心功能

### 用户注册
- 邮箱 + 密码 + 验证码 + 头像注册
- 自动触发邮箱验证码发送，验证码存 Redis（5 分钟有效）
- 头像上传到本地文件存储，返回 URL

### 用户登录
- 邮箱 + 密码登录，验证通过后签发 access_token + refresh_token
- 刷新令牌的哈希存入 `user_session` 表，支持多设备登录

### JWT 令牌管理
- Access Token：15 分钟有效，RSA-2048 签名
- Refresh Token：30 天有效，支持轮换（刷新时旧令牌失效）
- 登出时服务端清除会话记录

### 密码管理
- 忘记密码：邮箱验证码验证后重置
- 修改密码：需验证旧密码

### 用户资料
- 获取/更新用户信息
- 设置头像（支持 jpg/png/webp，最大 2MB）

## 配置项

| 配置项 | 环境变量 | 说明 |
|--------|---------|------|
| `server.port` | `KNOWSYNC_USER_SERVER_SERVER_PORT` | gRPC 端口（默认 50051） |
| `mysql.dsn` | `KNOWSYNC_USER_SERVER_MYSQL_DSN` | MySQL 连接字符串 |
| `redis.addr` | `KNOWSYNC_USER_SERVER_REDIS_ADDR` | Redis 地址 |
| `jwt.private_key_path` | `KNOWSYNC_USER_SERVER_JWT_PRIVATE_KEY_PATH` | RSA 私钥路径 |
| `mail.smtp_host` | `KNOWSYNC_USER_SERVER_MAIL_SMTP_HOST` | SMTP 服务器地址 |
| `mail.from` | `KNOWSYNC_USER_SERVER_MAIL_FROM` | 发件人邮箱 |

完整配置项见 `config.yaml`。

## 启动

```bash
# 本地开发
task dev:user-server

# 或直接使用 Go
go run cmd/main.go

# Docker
docker compose up user-server
```

## API

对外接口通过网关暴露 HTTP API，详见 [gateway/docs/user.jsonc](../gateway/docs/user.jsonc)。
