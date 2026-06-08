# Go 惯用法重构计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构 Go 微服务项目，消除 God Object 反模式，实现接口抽象，统一代码风格，提高可测试性和可维护性。

**Architecture:** 按领域拆分 Service 和 Handler，定义消费者接口，实现隐式依赖注入，创建共享 pkg 模块，统一 gRPC 响应风格。

**Tech Stack:** Go, gRPC, Hertz, GORM, Redis, viper, slog

---

## 阶段一：user-server 接口重构（3-4 小时）

### Task 1.1: 定义消费者接口

**Files:**
- Create: `user-server/internal/service/user_repository.go`
- Create: `user-server/internal/service/session_repository.go`
- Create: `user-server/internal/service/verify_code_repository.go`

- [ ] **Step 1: 创建 UserRepository 接口**

```go
// user-server/internal/service/user_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// UserRepository 用户数据访问接口（消费者定义）.
type UserRepository interface {
	GetUser(ctx context.Context, id string) (*schema.User, error)
	GetUserByEmail(ctx context.Context, email string) (*schema.User, error)
	CreateUser(ctx context.Context, user *schema.User) error
	UpdateUser(ctx context.Context, user *schema.User) error
	DeleteUser(ctx context.Context, id string) error
}
```

- [ ] **Step 2: 创建 SessionRepository 接口**

```go
// user-server/internal/service/session_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// SessionRepository 会话数据访问接口（消费者定义）.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *schema.UserSession) error
	GetSessionByRefreshToken(ctx context.Context, tokenHash string) (*schema.UserSession, error)
	UpdateSession(ctx context.Context, session *schema.UserSession) error
	InvalidateUserSessions(ctx context.Context, userID string) error
}
```

- [ ] **Step 3: 创建 VerifyCodeRepository 接口**

```go
// user-server/internal/service/verify_code_repository.go
package service

import "context"

// VerifyCodeRepository 验证码数据访问接口（消费者定义）.
type VerifyCodeRepository interface {
	GetVerifyCode(ctx context.Context, email string) (string, error)
	SetVerifyCode(ctx context.Context, email, code string) error
	DeleteVerifyCode(ctx context.Context, email string) error
}
```

- [ ] **Step 4: 验证接口定义**

运行: `cd user-server && go build ./...`
预期: 编译成功

### Task 1.2: 拆分 Service 结构体

**Files:**
- Create: `user-server/internal/service/user_service.go`
- Create: `user-server/internal/service/auth_service.go`
- Modify: `user-server/internal/service/service.go` (保留为兼容性包装)

- [ ] **Step 1: 创建 UserService**

```go
// user-server/internal/service/user_service.go
package service

import (
	"crypto/rsa"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// UserService 用户业务服务.
type UserService struct {
	userRepo       UserRepository
	sessionRepo    SessionRepository
	verifyCodeRepo VerifyCodeRepository
	privateKey     *rsa.PrivateKey
	cfg            *config.Config
}

// NewUserService 创建用户服务.
func NewUserService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	verifyCodeRepo VerifyCodeRepository,
	privateKey *rsa.PrivateKey,
	cfg *config.Config,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		verifyCodeRepo: verifyCodeRepo,
		privateKey:     privateKey,
		cfg:            cfg,
	}
}

// GetUser 获取用户信息.
func (s *UserService) GetUser(ctx context.Context, userID string) (*schema.User, error) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		slog.Error("获取用户信息失败", "user_id", userID, "error", err)
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user, nil
}

// UpdateUserInfo 更新用户信息.
func (s *UserService) UpdateUserInfo(ctx context.Context, userID, name, email, avatar string) (*schema.User, error) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}
	if avatar != "" {
		user.Avatar = avatar
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		slog.Error("更新用户信息失败", "user_id", userID, "error", err)
		return nil, fmt.Errorf("更新用户信息失败: %w", err)
	}

	slog.Info("更新用户信息成功", "user_id", userID)
	return user, nil
}

// SetAvatar 设置用户头像.
func (s *UserService) SetAvatar(ctx context.Context, userID, avatar string) error {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Avatar = avatar
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		slog.Error("设置头像失败", "user_id", userID, "error", err)
		return fmt.Errorf("设置头像失败: %w", err)
	}

	slog.Info("设置头像成功", "user_id", userID)
	return nil
}

// Unregister 注销用户.
func (s *UserService) Unregister(ctx context.Context, userID, email, password, verifyCode string) error {
	// 1. 验证码校验
	storedCode, err := s.verifyCodeRepo.GetVerifyCode(ctx, email)
	if err != nil {
		return errors.New("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	// 2. 查找用户并校验密码
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if err := CheckPassword(password, user.Password); err != nil {
		return errors.New("密码错误")
	}

	// 3. 删除用户
	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		slog.Error("注销用户失败", "user_id", userID, "error", err)
		return fmt.Errorf("注销用户失败: %w", err)
	}

	// 4. 清理该用户的所有会话
	if err := s.sessionRepo.InvalidateUserSessions(ctx, userID); err != nil {
		slog.Warn("注销时清理会话失败", "user_id", userID, "error", err)
	}

	// 5. 删除已使用的验证码
	if err := s.verifyCodeRepo.DeleteVerifyCode(ctx, email); err != nil {
		slog.Warn("删除验证码失败", "email", email, "error", err)
	}

	slog.Info("用户注销成功", "user_id", userID)
	return nil
}
```

- [ ] **Step 2: 创建 AuthService**

```go
// user-server/internal/service/auth_service.go
package service

import (
	"crypto/rsa"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gangantongxue/knowsync/user-server/pkg/auth"
	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// AuthService 认证业务服务.
type AuthService struct {
	userRepo       UserRepository
	sessionRepo    SessionRepository
	verifyCodeRepo VerifyCodeRepository
	privateKey     *rsa.PrivateKey
	cfg            *config.Config
}

// NewAuthService 创建认证服务.
func NewAuthService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	verifyCodeRepo VerifyCodeRepository,
	privateKey *rsa.PrivateKey,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		verifyCodeRepo: verifyCodeRepo,
		privateKey:     privateKey,
		cfg:            cfg,
	}
}

// Register 注册用户.
func (s *AuthService) Register(ctx context.Context, name, email, password, verifyCode string) (*schema.User, error) {
	// 1. 参数校验
	if err := validateRegisterParams(name, email, password, verifyCode); err != nil {
		slog.Warn("注册参数校验失败", "email", email, "error", err)
		return nil, err
	}

	// 2. 验证码校验
	storedCode, err := s.verifyCodeRepo.GetVerifyCode(ctx, email)
	if err != nil {
		slog.Error("获取验证码失败", "email", email, "error", err)
		return nil, errors.New("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		slog.Warn("验证码错误", "email", email)
		return nil, errors.New("验证码错误")
	}

	// 3. 检查邮箱是否已被注册
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("检查邮箱是否已注册失败", "email", email, "error", err)
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existingUser != nil {
		slog.Warn("邮箱已被注册", "email", email)
		return nil, errors.New("该邮箱已被注册")
	}

	// 4. 密码加密
	hashedPassword, err := HashPassword(password)
	if err != nil {
		slog.Error("密码加密失败", "email", email, "error", err)
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 5. 创建用户
	user := &schema.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}
	if err := s.userRepo.CreateUser(user); err != nil {
		slog.Error("创建用户失败", "email", email, "error", err)
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 6. 删除已使用的验证码
	if err := s.verifyCodeRepo.DeleteVerifyCode(ctx, email); err != nil {
		slog.Warn("删除验证码失败", "email", email, "error", err)
	}

	slog.Info("用户注册成功", "user_id", user.ID, "email", email)
	return user, nil
}

// Login 用户登录.
func (s *AuthService) Login(ctx context.Context, email, password, clientIP string) (*schema.User, string, string, error) {
	// 1. 根据邮箱查找用户
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errors.New("邮箱或密码错误")
	}

	// 2. 校验密码
	if err := CheckPassword(password, user.Password); err != nil {
		return nil, "", "", errors.New("邮箱或密码错误")
	}

	// 3. 清理该用户所有旧会话，保证一个用户至多一个活跃会话
	if err := s.sessionRepo.InvalidateUserSessions(ctx, user.ID); err != nil {
		slog.Error("清理旧会话失败", "user_id", user.ID, "error", err)
	}

	// 4. 生成 JWT access token
	accessToken, err := auth.GenerateAccessToken(user.ID, s.privateKey, s.cfg.Auth.AccessTTL)
	if err != nil {
		slog.Error("生成 access token 失败", "user_id", user.ID, "error", err)
		return nil, "", "", errors.New("生成访问凭证失败")
	}

	// 5. 生成随机 refresh token
	refreshToken, err := generateRefreshToken()
	if err != nil {
		slog.Error("生成 refresh token 失败", "user_id", user.ID, "error", err)
		return nil, "", "", errors.New("生成刷新凭证失败")
	}

	// 6. 创建新会话
	session := &schema.UserSession{
		UserID:           user.ID,
		ClientIP:         clientIP,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpireAt:         time.Now().Add(s.cfg.Auth.RefreshTTL),
		LoginAt:          time.Now(),
	}
	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		slog.Error("创建会话失败", "user_id", user.ID, "error", err)
		return nil, "", "", errors.New("创建会话失败")
	}

	slog.Info("用户登录成功", "user_id", user.ID, "client_ip", clientIP)
	return user, accessToken, refreshToken, nil
}

// Logout 用户退出登录，清空该用户全部活跃会话.
func (s *AuthService) Logout(ctx context.Context, refreshToken, _ string) error {
	// 1. 通过 refresh token 哈希值查找会话，获取 user_id
	session, err := s.sessionRepo.GetSessionByRefreshToken(ctx, hashRefreshToken(refreshToken))
	if err == nil && session != nil {
		// 2. 清空该用户所有活跃会话
		if err := s.sessionRepo.InvalidateUserSessions(ctx, session.UserID); err != nil {
			slog.Error("退出登录时清理会话失败", "user_id", session.UserID, "error", err)
			return fmt.Errorf("退出登录失败: %w", err)
		}
		slog.Info("用户退出登录成功", "user_id", session.UserID)
	}
	// 即使 refresh token 查不到也返回成功（幂等设计）
	return nil
}

// Refresh 刷新登录凭证.
func (s *AuthService) Refresh(ctx context.Context, refreshToken, clientIP string) (string, string, *schema.User, error) {
	// 1. 通过 refresh token 哈希值查找会话
	session, err := s.sessionRepo.GetSessionByRefreshToken(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		return "", "", nil, errors.New("刷新凭证无效或已过期")
	}

	// 2. 检查会话是否已退出
	if session.LogoutAt != nil {
		return "", "", nil, errors.New("刷新凭证已失效")
	}

	// 3. 检查 refresh token 是否过期
	if time.Now().After(session.ExpireAt) {
		return "", "", nil, errors.New("刷新凭证已过期")
	}

	// 4. 获取用户信息
	user, err := s.userRepo.GetUser(ctx, session.UserID)
	if err != nil {
		return "", "", nil, errors.New("获取用户信息失败")
	}

	// 5. 生成新的 access token
	accessToken, err := auth.GenerateAccessToken(session.UserID, s.privateKey, s.cfg.Auth.AccessTTL)
	if err != nil {
		slog.Error("刷新时生成 access token 失败", "user_id", session.UserID, "error", err)
		return "", "", nil, errors.New("生成访问凭证失败")
	}

	// 6. 生成新的 refresh token
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		slog.Error("刷新时生成 refresh token 失败", "user_id", session.UserID, "error", err)
		return "", "", nil, errors.New("生成刷新凭证失败")
	}

	// 7. 更新会话
	session.RefreshTokenHash = hashRefreshToken(newRefreshToken)
	session.ClientIP = clientIP
	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		slog.Error("刷新时更新会话失败", "user_id", session.UserID, "error", err)
		return "", "", nil, errors.New("更新会话失败")
	}

	slog.Info("刷新凭证成功", "user_id", session.UserID)
	return accessToken, newRefreshToken, user, nil
}
```

- [ ] **Step 3: 更新 service.go 为兼容性包装**

```go
// user-server/internal/service/service.go
package service

import (
	"crypto/rsa"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/user-server/internal/repository"
	"github.com/gangantongxue/knowsync/user-server/pkg/auth"
	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
	"github.com/gangantongxue/knowsync/user-server/pkg/mail"
)

// Service 业务服务层（兼容性包装，逐步迁移后删除）.
type Service struct {
	Cfg        *config.Config
	Repository *repository.Repository
	Mailer     *mail.Mailer
	Logger     *logger.Logger
	privateKey *rsa.PrivateKey
	
	// 新服务
	UserSvc  *UserService
	AuthSvc  *AuthService
}

// NewService 创建业务服务层.
func NewService(cfg *config.Config, repository *repository.Repository, mailer *mail.Mailer, l *logger.Logger) (*Service, error) {
	privateKey, err := auth.LoadPrivateKeyFromFile(cfg.Auth.RSAPrivateKeyPath)
	if err != nil {
		slog.Error("加载 RSA 私钥失败", "error", err)
		return nil, fmt.Errorf("加载 RSA 私钥失败: %w", err)
	}

	return &Service{
		Cfg:        cfg,
		Repository: repository,
		Mailer:     mailer,
		Logger:     l,
		privateKey: privateKey,
	}, nil
}
```

- [ ] **Step 4: 验证编译**

运行: `cd user-server && go build ./...`
预期: 编译成功

### Task 1.3: 实现 Repository 接口

**Files:**
- Create: `user-server/internal/repository/postgres/user.go`
- Create: `user-server/internal/repository/postgres/session.go`
- Create: `user-server/internal/repository/postgres/verify_code.go`

- [ ] **Step 1: 创建 UserRepo 实现**

```go
// user-server/internal/repository/postgres/user.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// UserRepo 用户仓库实现.
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓库.
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetUser 获取用户.
func (r *UserRepo) GetUser(ctx context.Context, id string) (*schema.User, error) {
	var user schema.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*schema.User, error) {
	var user schema.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户.
func (r *UserRepo) CreateUser(ctx context.Context, user *schema.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// UpdateUser 更新用户.
func (r *UserRepo) UpdateUser(ctx context.Context, user *schema.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser 删除用户.
func (r *UserRepo) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&schema.User{}, "id = ?", id).Error
}
```

- [ ] **Step 2: 创建 SessionRepo 实现**

```go
// user-server/internal/repository/postgres/session.go
package postgres

import (
	"context"
	"time"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// SessionRepo 会话仓库实现.
type SessionRepo struct {
	db *gorm.DB
}

// NewSessionRepo 创建会话仓库.
func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// CreateSession 创建会话.
func (r *SessionRepo) CreateSession(ctx context.Context, session *schema.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSessionByRefreshToken 根据 refresh token 哈希获取会话.
func (r *SessionRepo) GetSessionByRefreshToken(ctx context.Context, tokenHash string) (*schema.UserSession, error) {
	var session schema.UserSession
	if err := r.db.WithContext(ctx).Where("refresh_token_hash = ?", tokenHash).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession 更新会话.
func (r *SessionRepo) UpdateSession(ctx context.Context, session *schema.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// InvalidateUserSessions 使用户所有会话失效.
func (r *SessionRepo) InvalidateUserSessions(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&schema.UserSession{}).
		Where("user_id = ?", userID).
		Update("logout_at", time.Now()).Error
}
```

- [ ] **Step 3: 创建 VerifyCodeRepo 实现**

```go
// user-server/internal/repository/postgres/verify_code.go
package postgres

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

// VerifyCodeRepo 验证码仓库实现.
type VerifyCodeRepo struct {
	rdb *redis.Client
}

// NewVerifyCodeRepo 创建验证码仓库.
func NewVerifyCodeRepo(rdb *redis.Client) *VerifyCodeRepo {
	return &VerifyCodeRepo{rdb: rdb}
}

// GetVerifyCode 获取验证码.
func (r *VerifyCodeRepo) GetVerifyCode(ctx context.Context, email string) (string, error) {
	code, err := r.rdb.Get(ctx, "verify_code:"+email).Result()
	if err != nil {
		return "", err
	}
	return code, nil
}

// SetVerifyCode 设置验证码.
func (r *VerifyCodeRepo) SetVerifyCode(ctx context.Context, email, code string) error {
	return r.rdb.Set(ctx, "verify_code:"+email, code, 5*time.Minute).Error
}

// DeleteVerifyCode 删除验证码.
func (r *VerifyCodeRepo) DeleteVerifyCode(ctx context.Context, email string) error {
	return r.rdb.Del(ctx, "verify_code:"+email).Error
}
```

- [ ] **Step 4: 验证接口满足**

运行: `cd user-server && go build ./...`
预期: 编译成功（隐式接口满足）

### Task 1.4: 更新依赖注入

**Files:**
- Modify: `user-server/internal/app/app.go`

- [ ] **Step 1: 更新 app.go 依赖注入**

```go
// user-server/internal/app/app.go
package app

import (
	// ... 其他导入
	"github.com/gangantongxue/knowsync/user-server/internal/repository/postgres"
	"github.com/gangantongxue/knowsync/user-server/internal/service"
)

func NewApp() error {
	// ... 初始化配置、数据库、Redis 等
	
	// 创建具体实现
	userRepo := postgres.NewUserRepo(db.DB)
	sessionRepo := postgres.NewSessionRepo(db.DB)
	verifyCodeRepo := postgres.NewVerifyCodeRepo(redis.RDB)
	
	// 加载私钥
	privateKey, err := auth.LoadPrivateKeyFromFile(cfg.Auth.RSAPrivateKeyPath)
	if err != nil {
		return err
	}
	
	// 创建服务（注入接口，Go 隐式满足）
	userSvc := service.NewUserService(userRepo, sessionRepo, verifyCodeRepo, privateKey, cfg)
	authSvc := service.NewAuthService(userRepo, sessionRepo, verifyCodeRepo, privateKey, cfg)
	
	// 创建 handler
	h := handler.NewHandler(userSvc, authSvc)
	
	// ... 启动服务
}
```

- [ ] **Step 2: 验证编译**

运行: `cd user-server && go build ./...`
预期: 编译成功

### Task 1.5: 更新 Handler 层

**Files:**
- Modify: `user-server/internal/handler/user.go`
- Modify: `user-server/internal/handler/auth.go`

- [ ] **Step 1: 更新 user.go 使用 UserService**

```go
// user-server/internal/handler/user.go
package handler

import (
	"context"
	"github.com/gangantongxue/knowsync/user-server/internal/service"
	"github.com/gangantongxue/knowsync/user-server/pkg/errcode"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserHandler 用户处理器.
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建用户处理器.
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetUser 获取用户信息.
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.userSvc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户信息失败")
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	
	return &pb.GetUserResponse{
		User: &pb.User{
			Id:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Avatar: user.Avatar,
		},
	}, nil
}
```

- [ ] **Step 2: 更新 auth.go 使用 AuthService**

```go
// user-server/internal/handler/auth.go
package handler

import (
	"context"
	"github.com/gangantongxue/knowsync/user-server/internal/service"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler 认证处理器.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler 创建认证处理器.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register 用户注册.
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.authSvc.Register(ctx, req.Name, req.Email, req.Password, req.VerifyCode)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	
	return &pb.RegisterResponse{
		User: &pb.User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}
```

- [ ] **Step 3: 验证编译**

运行: `cd user-server && go build ./...`
预期: 编译成功

---

## 阶段二：repo-server 接口重构（3-4 小时）

### Task 2.1: 定义消费者接口

**Files:**
- Create: `repo-server/internal/service/repo_repository.go`
- Create: `repo-server/internal/service/collaborator_repository.go`
- Create: `repo-server/internal/service/follow_repository.go`

- [ ] **Step 1: 创建 RepoRepository 接口**

```go
// repo-server/internal/service/repo_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// RepoRepository 知识库数据访问接口（消费者定义）.
type RepoRepository interface {
	CreateRepo(ctx context.Context, repo *schema.Repo) error
	GetRepo(ctx context.Context, id string) (*schema.Repo, error)
	UpdateRepo(ctx context.Context, repo *schema.Repo) error
	DeleteRepo(ctx context.Context, id string) error
	ListUserRepos(ctx context.Context, userID string, page, pageSize int) ([]*schema.Repo, int64, error)
	ListPublicRepos(ctx context.Context, page, pageSize int) ([]*schema.Repo, int64, error)
	SearchRepos(ctx context.Context, keyword string, page, pageSize int) ([]*schema.Repo, int64, error)
}
```

- [ ] **Step 2: 创建 CollaboratorRepository 接口**

```go
// repo-server/internal/service/collaborator_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CollaboratorRepository 协作者数据访问接口（消费者定义）.
type CollaboratorRepository interface {
	AddCollaborator(ctx context.Context, collab *schema.Collaborator) error
	RemoveCollaborator(ctx context.Context, repoID, userID string) error
	UpdateCollaboratorRole(ctx context.Context, repoID, userID, role string) error
	GetCollaborator(ctx context.Context, repoID, userID string) (*schema.Collaborator, error)
	ListCollaborators(ctx context.Context, repoID string) ([]*schema.Collaborator, error)
}
```

- [ ] **Step 3: 创建 FollowRepository 接口**

```go
// repo-server/internal/service/follow_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// FollowRepository 关注数据访问接口（消费者定义）.
type FollowRepository interface {
	FollowRepo(ctx context.Context, follow *schema.Follow) error
	UnfollowRepo(ctx context.Context, repoID, userID string) error
	IsFollowing(ctx context.Context, repoID, userID string) (bool, error)
	ListFollowedRepos(ctx context.Context, userID string, page, pageSize int) ([]*schema.Repo, int64, error)
}
```

- [ ] **Step 4: 验证接口定义**

运行: `cd repo-server && go build ./...`
预期: 编译成功

### Task 2.2: 拆分 Service 结构体

**Files:**
- Create: `repo-server/internal/service/repo_service.go`
- Create: `repo-server/internal/service/collaborator_service.go`
- Create: `repo-server/internal/service/follow_service.go`

- [ ] **Step 1: 创建 RepoService**

```go
// repo-server/internal/service/repo_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"gorm.io/gorm"
)

// RepoService 知识库业务服务.
type RepoService struct {
	repoRepo       RepoRepository
	collabRepo     CollaboratorRepository
	followRepo     FollowRepository
}

// NewRepoService 创建知识库服务.
func NewRepoService(
	repoRepo RepoRepository,
	collabRepo CollaboratorRepository,
	followRepo FollowRepository,
) *RepoService {
	return &RepoService{
		repoRepo:   repoRepo,
		collabRepo: collabRepo,
		followRepo: followRepo,
	}
}

// CreateRepo 创建知识库.
func (s *RepoService) CreateRepo(ctx context.Context, userID, name, description string) (*schema.Repo, error) {
	repo := &schema.Repo{
		UserID:      userID,
		Name:        name,
		Description: description,
		IsPublic:    true,
	}
	
	if err := s.repoRepo.CreateRepo(ctx, repo); err != nil {
		slog.Error("创建知识库失败", "user_id", userID, "error", err)
		return nil, fmt.Errorf("创建知识库失败: %w", err)
	}
	
	slog.Info("创建知识库成功", "repo_id", repo.ID, "user_id", userID)
	return repo, nil
}

// GetRepo 获取知识库信息.
func (s *RepoService) GetRepo(ctx context.Context, id string) (*schema.Repo, error) {
	repo, err := s.repoRepo.GetRepo(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		slog.Error("获取知识库信息失败", "repo_id", id, "error", err)
		return nil, fmt.Errorf("获取知识库信息失败: %w", err)
	}
	return repo, nil
}

// UpdateRepo 更新知识库信息.
func (s *RepoService) UpdateRepo(ctx context.Context, id, name, description string) (*schema.Repo, error) {
	repo, err := s.repoRepo.GetRepo(ctx, id)
	if err != nil {
		return nil, errors.New("知识库不存在")
	}
	
	if name != "" {
		repo.Name = name
	}
	if description != "" {
		repo.Description = description
	}
	
	if err := s.repoRepo.UpdateRepo(ctx, repo); err != nil {
		slog.Error("更新知识库信息失败", "repo_id", id, "error", err)
		return nil, fmt.Errorf("更新知识库信息失败: %w", err)
	}
	
	slog.Info("更新知识库信息成功", "repo_id", id)
	return repo, nil
}

// DeleteRepo 删除知识库.
func (s *RepoService) DeleteRepo(ctx context.Context, id string) error {
	if err := s.repoRepo.DeleteRepo(ctx, id); err != nil {
		slog.Error("删除知识库失败", "repo_id", id, "error", err)
		return fmt.Errorf("删除知识库失败: %w", err)
	}
	
	slog.Info("删除知识库成功", "repo_id", id)
	return nil
}

// ListUserRepos 列出用户的知识库.
func (s *RepoService) ListUserRepos(ctx context.Context, userID string, page, pageSize int) ([]*schema.Repo, int64, error) {
	return s.repoRepo.ListUserRepos(ctx, userID, page, pageSize)
}

// ListPublicRepos 列出公开知识库.
func (s *RepoService) ListPublicRepos(ctx context.Context, page, pageSize int) ([]*schema.Repo, int64, error) {
	return s.repoRepo.ListPublicRepos(ctx, page, pageSize)
}

// SearchRepos 搜索知识库.
func (s *RepoService) SearchRepos(ctx context.Context, keyword string, page, pageSize int) ([]*schema.Repo, int64, error) {
	return s.repoRepo.SearchRepos(ctx, keyword, page, pageSize)
}
```

- [ ] **Step 2: 创建 CollaboratorService**

```go
// repo-server/internal/service/collaborator_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CollaboratorService 协作者业务服务.
type CollaboratorService struct {
	repoRepo   RepoRepository
	collabRepo CollaboratorRepository
}

// NewCollaboratorService 创建协作者服务.
func NewCollaboratorService(
	repoRepo RepoRepository,
	collabRepo CollaboratorRepository,
) *CollaboratorService {
	return &CollaboratorService{
		repoRepo:   repoRepo,
		collabRepo: collabRepo,
	}
}

// AddCollaborator 添加协作者.
func (s *CollaboratorService) AddCollaborator(ctx context.Context, repoID, userID, role string) error {
	// 检查知识库是否存在
	repo, err := s.repoRepo.GetRepo(ctx, repoID)
	if err != nil || repo == nil {
		return errors.New("知识库不存在")
	}
	
	// 检查是否已是协作者
	existing, err := s.collabRepo.GetCollaborator(ctx, repoID, userID)
	if err == nil && existing != nil {
		return errors.New("用户已是协作者")
	}
	
	// 添加协作者
	collab := &schema.Collaborator{
		RepoID: repoID,
		UserID: userID,
		Role:   role,
	}
	
	if err := s.collabRepo.AddCollaborator(ctx, collab); err != nil {
		slog.Error("添加协作者失败", "repo_id", repoID, "user_id", userID, "error", err)
		return fmt.Errorf("添加协作者失败: %w", err)
	}
	
	slog.Info("添加协作者成功", "repo_id", repoID, "user_id", userID)
	return nil
}

// RemoveCollaborator 移除协作者.
func (s *CollaboratorService) RemoveCollaborator(ctx context.Context, repoID, userID string) error {
	if err := s.collabRepo.RemoveCollaborator(ctx, repoID, userID); err != nil {
		slog.Error("移除协作者失败", "repo_id", repoID, "user_id", userID, "error", err)
		return fmt.Errorf("移除协作者失败: %w", err)
	}
	
	slog.Info("移除协作者成功", "repo_id", repoID, "user_id", userID)
	return nil
}

// UpdateCollaboratorRole 更新协作者角色.
func (s *CollaboratorService) UpdateCollaboratorRole(ctx context.Context, repoID, userID, role string) error {
	if err := s.collabRepo.UpdateCollaboratorRole(ctx, repoID, userID, role); err != nil {
		slog.Error("更新协作者角色失败", "repo_id", repoID, "user_id", userID, "error", err)
		return fmt.Errorf("更新协作者角色失败: %w", err)
	}
	
	slog.Info("更新协作者角色成功", "repo_id", repoID, "user_id", userID)
	return nil
}

// ListCollaborators 列出协作者.
func (s *CollaboratorService) ListCollaborators(ctx context.Context, repoID string) ([]*schema.Collaborator, error) {
	return s.collabRepo.ListCollaborators(ctx, repoID)
}
```

- [ ] **Step 3: 创建 FollowService**

```go
// repo-server/internal/service/follow_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"
	
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// FollowService 关注业务服务.
type FollowService struct {
	repoRepo   RepoRepository
	followRepo FollowRepository
}

// NewFollowService 创建关注服务.
func NewFollowService(
	repoRepo RepoRepository,
	followRepo FollowRepository,
) *FollowService {
	return &FollowService{
		repoRepo:   repoRepo,
		followRepo: followRepo,
	}
}

// FollowRepo 关注知识库.
func (s *FollowService) FollowRepo(ctx context.Context, repoID, userID string) error {
	// 检查知识库是否存在
	repo, err := s.repoRepo.GetRepo(ctx, repoID)
	if err != nil || repo == nil {
		return fmt.Errorf("知识库不存在")
	}
	
	// 检查是否已关注
	following, err := s.followRepo.IsFollowing(ctx, repoID, userID)
	if err != nil {
		return fmt.Errorf("检查关注状态失败: %w", err)
	}
	if following {
		return fmt.Errorf("已关注该知识库")
	}
	
	// 关注
	follow := &schema.Follow{
		RepoID: repoID,
		UserID: userID,
	}
	
	if err := s.followRepo.FollowRepo(ctx, follow); err != nil {
		slog.Error("关注知识库失败", "repo_id", repoID, "user_id", userID, "error", err)
		return fmt.Errorf("关注知识库失败: %w", err)
	}
	
	slog.Info("关注知识库成功", "repo_id", repoID, "user_id", userID)
	return nil
}

// UnfollowRepo 取消关注知识库.
func (s *FollowService) UnfollowRepo(ctx context.Context, repoID, userID string) error {
	if err := s.followRepo.UnfollowRepo(ctx, repoID, userID); err != nil {
		slog.Error("取消关注知识库失败", "repo_id", repoID, "user_id", userID, "error", err)
		return fmt.Errorf("取消关注知识库失败: %w", err)
	}
	
	slog.Info("取消关注知识库成功", "repo_id", repoID, "user_id", userID)
	return nil
}

// ListFollowedRepos 列出关注的知识库.
func (s *FollowService) ListFollowedRepos(ctx context.Context, userID string, page, pageSize int) ([]*schema.Repo, int64, error) {
	return s.followRepo.ListFollowedRepos(ctx, userID, page, pageSize)
}
```

- [ ] **Step 4: 验证编译**

运行: `cd repo-server && go build ./...`
预期: 编译成功

### Task 2.3: 实现 Repository 接口

**Files:**
- Create: `repo-server/internal/repository/postgres/repo.go`
- Create: `repo-server/internal/repository/postgres/collaborator.go`
- Create: `repo-server/internal/repository/postgres/follow.go`

- [ ] **Step 1: 创建 RepoRepo 实现**

```go
// repo-server/internal/repository/postgres/repo.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"gorm.io/gorm"
)

// RepoRepo 知识库仓库实现.
type RepoRepo struct {
	db *gorm.DB
}

// NewRepoRepo 创建知识库仓库.
func NewRepoRepo(db *gorm.DB) *RepoRepo {
	return &RepoRepo{db: db}
}

// CreateRepo 创建知识库.
func (r *RepoRepo) CreateRepo(ctx context.Context, repo *schema.Repo) error {
	return r.db.WithContext(ctx).Create(repo).Error
}

// GetRepo 获取知识库.
func (r *RepoRepo) GetRepo(ctx context.Context, id string) (*schema.Repo, error) {
	var repo schema.Repo
	if err := r.db.WithContext(ctx).First(&repo, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

// UpdateRepo 更新知识库.
func (r *RepoRepo) UpdateRepo(ctx context.Context, repo *schema.Repo) error {
	return r.db.WithContext(ctx).Save(repo).Error
}

// DeleteRepo 删除知识库.
func (r *RepoRepo) DeleteRepo(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&schema.Repo{}, "id = ?", id).Error
}

// ListUserRepos 列出用户的知识库.
func (r *RepoRepo) ListUserRepos(ctx context.Context, userID string, page, pageSize int) ([]*schema.Repo, int64, error) {
	var repos []*schema.Repo
	var total int64
	
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&repos).Error; err != nil {
		return nil, 0, err
	}
	
	return repos, total, nil
}

// ListPublicRepos 列出公开知识库.
func (r *RepoRepo) ListPublicRepos(ctx context.Context, page, pageSize int) ([]*schema.Repo, int64, error) {
	var repos []*schema.Repo
	var total int64
	
	query := r.db.WithContext(ctx).Where("is_public = ?", true)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&repos).Error; err != nil {
		return nil, 0, err
	}
	
	return repos, total, nil
}

// SearchRepos 搜索知识库.
func (r *RepoRepo) SearchRepos(ctx context.Context, keyword string, page, pageSize int) ([]*schema.Repo, int64, error) {
	var repos []*schema.Repo
	var total int64
	
	query := r.db.WithContext(ctx).Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&repos).Error; err != nil {
		return nil, 0, err
	}
	
	return repos, total, nil
}
```

- [ ] **Step 2: 创建 CollaboratorRepo 实现**

```go
// repo-server/internal/repository/postgres/collaborator.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"gorm.io/gorm"
)

// CollaboratorRepo 协作者仓库实现.
type CollaboratorRepo struct {
	db *gorm.DB
}

// NewCollaboratorRepo 创建协作者仓库.
func NewCollaboratorRepo(db *gorm.DB) *CollaboratorRepo {
	return &CollaboratorRepo{db: db}
}

// AddCollaborator 添加协作者.
func (r *CollaboratorRepo) AddCollaborator(ctx context.Context, collab *schema.Collaborator) error {
	return r.db.WithContext(ctx).Create(collab).Error
}

// RemoveCollaborator 移除协作者.
func (r *CollaboratorRepo) RemoveCollaborator(ctx context.Context, repoID, userID string) error {
	return r.db.WithContext(ctx).Where("repo_id = ? AND user_id = ?", repoID, userID).Delete(&schema.Collaborator{}).Error
}

// UpdateCollaboratorRole 更新协作者角色.
func (r *CollaboratorRepo) UpdateCollaboratorRole(ctx context.Context, repoID, userID, role string) error {
	return r.db.WithContext(ctx).Model(&schema.Collaborator{}).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		Update("role", role).Error
}

// GetCollaborator 获取协作者.
func (r *CollaboratorRepo) GetCollaborator(ctx context.Context, repoID, userID string) (*schema.Collaborator, error) {
	var collab schema.Collaborator
	if err := r.db.WithContext(ctx).Where("repo_id = ? AND user_id = ?", repoID, userID).First(&collab).Error; err != nil {
		return nil, err
	}
	return &collab, nil
}

// ListCollaborators 列出协作者.
func (r *CollaboratorRepo) ListCollaborators(ctx context.Context, repoID string) ([]*schema.Collaborator, error) {
	var collabs []*schema.Collaborator
	if err := r.db.WithContext(ctx).Where("repo_id = ?", repoID).Find(&collabs).Error; err != nil {
		return nil, err
	}
	return collabs, nil
}
```

- [ ] **Step 3: 创建 FollowRepo 实现**

```go
// repo-server/internal/repository/postgres/follow.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"gorm.io/gorm"
)

// FollowRepo 关注仓库实现.
type FollowRepo struct {
	db *gorm.DB
}

// NewFollowRepo 创建关注仓库.
func NewFollowRepo(db *gorm.DB) *FollowRepo {
	return &FollowRepo{db: db}
}

// FollowRepo 关注知识库.
func (r *FollowRepo) FollowRepo(ctx context.Context, follow *schema.Follow) error {
	return r.db.WithContext(ctx).Create(follow).Error
}

// UnfollowRepo 取消关注知识库.
func (r *FollowRepo) UnfollowRepo(ctx context.Context, repoID, userID string) error {
	return r.db.WithContext(ctx).Where("repo_id = ? AND user_id = ?", repoID, userID).Delete(&schema.Follow{}).Error
}

// IsFollowing 检查是否已关注.
func (r *FollowRepo) IsFollowing(ctx context.Context, repoID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&schema.Follow{}).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		Count(&count).Error
	return count > 0, err
}

// ListFollowedRepos 列出关注的知识库.
func (r *FollowRepo) ListFollowedRepos(ctx context.Context, userID string, page, pageSize int) ([]*schema.Repo, int64, error) {
	var repos []*schema.Repo
	var total int64
	
	// 先获取关注的知识库 ID
	var followIDs []string
	if err := r.db.WithContext(ctx).Model(&schema.Follow{}).
		Where("user_id = ?", userID).
		Pluck("repo_id", &followIDs).Error; err != nil {
		return nil, 0, err
	}
	
	if len(followIDs) == 0 {
		return nil, 0, nil
	}
	
	// 再获取知识库详情
	query := r.db.WithContext(ctx).Where("id IN ?", followIDs)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&repos).Error; err != nil {
		return nil, 0, err
	}
	
	return repos, total, nil
}
```

- [ ] **Step 4: 验证编译**

运行: `cd repo-server && go build ./...`
预期: 编译成功

### Task 2.4: 更新依赖注入和 Handler

**Files:**
- Modify: `repo-server/internal/app/app.go`
- Modify: `repo-server/internal/handler/repo.go`
- Modify: `repo-server/internal/handler/collaborator.go`
- Modify: `repo-server/internal/handler/follow.go`

- [ ] **Step 1: 更新 app.go 依赖注入**

```go
// repo-server/internal/app/app.go
package app

import (
	// ... 其他导入
	"github.com/gangantongxue/knowsync/repo-server/internal/repository/postgres"
	"github.com/gangantongxue/knowsync/repo-server/internal/service"
)

func NewApp() error {
	// ... 初始化配置、数据库等
	
	// 创建具体实现
	repoRepo := postgres.NewRepoRepo(db.DB)
	collabRepo := postgres.NewCollaboratorRepo(db.DB)
	followRepo := postgres.NewFollowRepo(db.DB)
	
	// 创建服务（注入接口，Go 隐式满足）
	repoSvc := service.NewRepoService(repoRepo, collabRepo, followRepo)
	collabSvc := service.NewCollaboratorService(repoRepo, collabRepo)
	followSvc := service.NewFollowService(repoRepo, followRepo)
	
	// 创建 handler
	h := handler.NewHandler(repoSvc, collabSvc, followSvc)
	
	// ... 启动服务
}
```

- [ ] **Step 2: 更新 handler 层**

```go
// repo-server/internal/handler/repo.go
package handler

import (
	"context"
	"github.com/gangantongxue/knowsync/repo-server/internal/service"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RepoHandler 知识库处理器.
type RepoHandler struct {
	repoSvc *service.RepoService
}

// NewRepoHandler 创建知识库处理器.
func NewRepoHandler(repoSvc *service.RepoService) *RepoHandler {
	return &RepoHandler{repoSvc: repoSvc}
}

// CreateRepo 创建知识库.
func (h *RepoHandler) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*pb.CreateRepoResponse, error) {
	repo, err := h.repoSvc.CreateRepo(ctx, req.UserId, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, "创建知识库失败")
	}
	
	return &pb.CreateRepoResponse{
		Repo: &pb.Repo{
			Id:          repo.ID,
			Name:        repo.Name,
			Description: repo.Description,
		},
	}, nil
}
```

- [ ] **Step 3: 验证编译**

运行: `cd repo-server && go build ./...`
预期: 编译成功

---

## 阶段三：chat-server 接口重构（4-5 小时）

### Task 3.1: 定义消费者接口

**Files:**
- Create: `chat-server/internal/service/message_repository.go`
- Create: `chat-server/internal/service/conversation_repository.go`
- Create: `chat-server/internal/service/friend_repository.go`
- Create: `chat-server/internal/service/group_repository.go`

- [ ] **Step 1: 创建 MessageRepository 接口**

```go
// chat-server/internal/service/message_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// MessageRepository 消息数据访问接口（消费者定义）.
type MessageRepository interface {
	CreateMessage(ctx context.Context, msg *schema.Message) error
	GetMessage(ctx context.Context, id string) (*schema.Message, error)
	ListMessages(ctx context.Context, conversationID string, page, pageSize int) ([]*schema.Message, int64, error)
	UpdateMessageStatus(ctx context.Context, msgID, status string) error
	DeleteMessage(ctx context.Context, id string) error
}
```

- [ ] **Step 2: 创建 ConversationRepository 接口**

```go
// chat-server/internal/service/conversation_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// ConversationRepository 会话数据访问接口（消费者定义）.
type ConversationRepository interface {
	CreateConversation(ctx context.Context, conv *schema.Conversation) error
	GetConversation(ctx context.Context, id string) (*schema.Conversation, error)
	UpdateConversation(ctx context.Context, conv *schema.Conversation) error
	ListUserConversations(ctx context.Context, userID string) ([]*schema.Conversation, error)
	DeleteConversation(ctx context.Context, id string) error
}
```

- [ ] **Step 3: 创建 FriendRepository 接口**

```go
// chat-server/internal/service/friend_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// FriendRepository 好友数据访问接口（消费者定义）.
type FriendRepository interface {
	AddFriend(ctx context.Context, friend *schema.Friend) error
	RemoveFriend(ctx context.Context, userID, friendID string) error
	IsFriend(ctx context.Context, userID, friendID string) (bool, error)
	ListFriends(ctx context.Context, userID string) ([]*schema.Friend, error)
}
```

- [ ] **Step 4: 创建 GroupRepository 接口**

```go
// chat-server/internal/service/group_repository.go
package service

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// GroupRepository 群组数据访问接口（消费者定义）.
type GroupRepository interface {
	CreateGroup(ctx context.Context, group *schema.Group) error
	GetGroup(ctx context.Context, id string) (*schema.Group, error)
	UpdateGroup(ctx context.Context, group *schema.Group) error
	DeleteGroup(ctx context.Context, id string) error
	AddGroupMember(ctx context.Context, member *schema.GroupMember) error
	RemoveGroupMember(ctx context.Context, groupID, userID string) error
	ListGroupMembers(ctx context.Context, groupID string) ([]*schema.GroupMember, error)
}
```

- [ ] **Step 5: 验证接口定义**

运行: `cd chat-server && go build ./...`
预期: 编译成功

### Task 3.2: 拆分 Service 结构体

**Files:**
- Create: `chat-server/internal/service/message_service.go`
- Create: `chat-server/internal/service/conversation_service.go`
- Create: `chat-server/internal/service/friend_service.go`
- Create: `chat-server/internal/service/group_service.go`

- [ ] **Step 1: 创建 MessageService**

```go
// chat-server/internal/service/message_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"
	
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// MessageService 消息业务服务.
type MessageService struct {
	msgRepo  MessageRepository
	convRepo ConversationRepository
}

// NewMessageService 创建消息服务.
func NewMessageService(
	msgRepo MessageRepository,
	convRepo ConversationRepository,
) *MessageService {
	return &MessageService{
		msgRepo:  msgRepo,
		convRepo: convRepo,
	}
}

// SendMessage 发送消息.
func (s *MessageService) SendMessage(ctx context.Context, senderID, conversationID, content, msgType string) (*schema.Message, error) {
	// 检查会话是否存在
	conv, err := s.convRepo.GetConversation(ctx, conversationID)
	if err != nil || conv == nil {
		return nil, fmt.Errorf("会话不存在")
	}
	
	msg := &schema.Message{
		SenderID:       senderID,
		ConversationID: conversationID,
		Content:        content,
		Type:           msgType,
		Status:         "sent",
	}
	
	if err := s.msgRepo.CreateMessage(ctx, msg); err != nil {
		slog.Error("发送消息失败", "sender_id", senderID, "conversation_id", conversationID, "error", err)
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}
	
	slog.Info("发送消息成功", "msg_id", msg.ID, "sender_id", senderID)
	return msg, nil
}

// GetMessage 获取消息.
func (s *MessageService) GetMessage(ctx context.Context, id string) (*schema.Message, error) {
	return s.msgRepo.GetMessage(ctx, id)
}

// ListMessages 列出消息.
func (s *MessageService) ListMessages(ctx context.Context, conversationID string, page, pageSize int) ([]*schema.Message, int64, error) {
	return s.msgRepo.ListMessages(ctx, conversationID, page, pageSize)
}
```

- [ ] **Step 2: 创建 ConversationService**

```go
// chat-server/internal/service/conversation_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"
	
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// ConversationService 会话业务服务.
type ConversationService struct {
	convRepo ConversationRepository
}

// NewConversationService 创建会话服务.
func NewConversationService(convRepo ConversationRepository) *ConversationService {
	return &ConversationService{convRepo: convRepo}
}

// CreateConversation 创建会话.
func (s *ConversationService) CreateConversation(ctx context.Context, name, convType string, creatorID string) (*schema.Conversation, error) {
	conv := &schema.Conversation{
		Name:      name,
		Type:      convType,
		CreatorID: creatorID,
	}
	
	if err := s.convRepo.CreateConversation(ctx, conv); err != nil {
		slog.Error("创建会话失败", "creator_id", creatorID, "error", err)
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}
	
	slog.Info("创建会话成功", "conv_id", conv.ID, "creator_id", creatorID)
	return conv, nil
}

// GetConversation 获取会话.
func (s *ConversationService) GetConversation(ctx context.Context, id string) (*schema.Conversation, error) {
	return s.convRepo.GetConversation(ctx, id)
}

// ListUserConversations 列出用户会话.
func (s *ConversationService) ListUserConversations(ctx context.Context, userID string) ([]*schema.Conversation, error) {
	return s.convRepo.ListUserConversations(ctx, userID)
}
```

- [ ] **Step 3: 创建 FriendService**

```go
// chat-server/internal/service/friend_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"
)

// FriendService 好友业务服务.
type FriendService struct {
	friendRepo FriendRepository
}

// NewFriendService 创建好友服务.
func NewFriendService(friendRepo FriendRepository) *FriendService {
	return &FriendService{friendRepo: friendRepo}
}

// AddFriend 添加好友.
func (s *FriendService) AddFriend(ctx context.Context, userID, friendID string) error {
	// 检查是否已是好友
	isFriend, err := s.friendRepo.IsFriend(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("检查好友关系失败: %w", err)
	}
	if isFriend {
		return fmt.Errorf("已是好友关系")
	}
	
	// 添加好友关系（双向）
	friend1 := &schema.Friend{
		UserID:   userID,
		FriendID: friendID,
	}
	friend2 := &schema.Friend{
		UserID:   friendID,
		FriendID: userID,
	}
	
	if err := s.friendRepo.AddFriend(ctx, friend1); err != nil {
		slog.Error("添加好友失败", "user_id", userID, "friend_id", friendID, "error", err)
		return fmt.Errorf("添加好友失败: %w", err)
	}
	
	if err := s.friendRepo.AddFriend(ctx, friend2); err != nil {
		slog.Error("添加好友失败", "user_id", friendID, "friend_id", userID, "error", err)
		return fmt.Errorf("添加好友失败: %w", err)
	}
	
	slog.Info("添加好友成功", "user_id", userID, "friend_id", friendID)
	return nil
}

// RemoveFriend 移除好友.
func (s *FriendService) RemoveFriend(ctx context.Context, userID, friendID string) error {
	// 移除双向好友关系
	if err := s.friendRepo.RemoveFriend(ctx, userID, friendID); err != nil {
		slog.Error("移除好友失败", "user_id", userID, "friend_id", friendID, "error", err)
		return fmt.Errorf("移除好友失败: %w", err)
	}
	
	if err := s.friendRepo.RemoveFriend(ctx, friendID, userID); err != nil {
		slog.Error("移除好友失败", "user_id", friendID, "friend_id", userID, "error", err)
		return fmt.Errorf("移除好友失败: %w", err)
	}
	
	slog.Info("移除好友成功", "user_id", userID, "friend_id", friendID)
	return nil
}

// ListFriends 列出好友.
func (s *FriendService) ListFriends(ctx context.Context, userID string) ([]*schema.Friend, error) {
	return s.friendRepo.ListFriends(ctx, userID)
}
```

- [ ] **Step 4: 创建 GroupService**

```go
// chat-server/internal/service/group_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"
	
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// GroupService 群组业务服务.
type GroupService struct {
	groupRepo GroupRepository
}

// NewGroupService 创建群组服务.
func NewGroupService(groupRepo GroupRepository) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

// CreateGroup 创建群组.
func (s *GroupService) CreateGroup(ctx context.Context, name, creatorID string) (*schema.Group, error) {
	group := &schema.Group{
		Name:      name,
		CreatorID: creatorID,
	}
	
	if err := s.groupRepo.CreateGroup(ctx, group); err != nil {
		slog.Error("创建群组失败", "creator_id", creatorID, "error", err)
		return nil, fmt.Errorf("创建群组失败: %w", err)
	}
	
	// 创建者自动成为群主
	member := &schema.GroupMember{
		GroupID: group.ID,
		UserID:  creatorID,
		Role:    "owner",
	}
	
	if err := s.groupRepo.AddGroupMember(ctx, member); err != nil {
		slog.Error("添加群主失败", "group_id", group.ID, "user_id", creatorID, "error", err)
		return nil, fmt.Errorf("添加群主失败: %w", err)
	}
	
	slog.Info("创建群组成功", "group_id", group.ID, "creator_id", creatorID)
	return group, nil
}

// GetGroup 获取群组.
func (s *GroupService) GetGroup(ctx context.Context, id string) (*schema.Group, error) {
	return s.groupRepo.GetGroup(ctx, id)
}

// AddGroupMember 添加群成员.
func (s *GroupService) AddGroupMember(ctx context.Context, groupID, userID, role string) error {
	member := &schema.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    role,
	}
	
	if err := s.groupRepo.AddGroupMember(ctx, member); err != nil {
		slog.Error("添加群成员失败", "group_id", groupID, "user_id", userID, "error", err)
		return fmt.Errorf("添加群成员失败: %w", err)
	}
	
	slog.Info("添加群成员成功", "group_id", groupID, "user_id", userID)
	return nil
}

// ListGroupMembers 列出群成员.
func (s *GroupService) ListGroupMembers(ctx context.Context, groupID string) ([]*schema.GroupMember, error) {
	return s.groupRepo.ListGroupMembers(ctx, groupID)
}
```

- [ ] **Step 5: 验证编译**

运行: `cd chat-server && go build ./...`
预期: 编译成功

### Task 3.3: 实现 Repository 接口

**Files:**
- Create: `chat-server/internal/repository/postgres/message.go`
- Create: `chat-server/internal/repository/postgres/conversation.go`
- Create: `chat-server/internal/repository/postgres/friend.go`
- Create: `chat-server/internal/repository/postgres/group.go`

- [ ] **Step 1: 创建 MessageRepo 实现**

```go
// chat-server/internal/repository/postgres/message.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// MessageRepo 消息仓库实现.
type MessageRepo struct {
	db *gorm.DB
}

// NewMessageRepo 创建消息仓库.
func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// CreateMessage 创建消息.
func (r *MessageRepo) CreateMessage(ctx context.Context, msg *schema.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// GetMessage 获取消息.
func (r *MessageRepo) GetMessage(ctx context.Context, id string) (*schema.Message, error) {
	var msg schema.Message
	if err := r.db.WithContext(ctx).First(&msg, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// ListMessages 列出消息.
func (r *MessageRepo) ListMessages(ctx context.Context, conversationID string, page, pageSize int) ([]*schema.Message, int64, error) {
	var msgs []*schema.Message
	var total int64
	
	query := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}
	
	return msgs, total, nil
}

// UpdateMessageStatus 更新消息状态.
func (r *MessageRepo) UpdateMessageStatus(ctx context.Context, msgID, status string) error {
	return r.db.WithContext(ctx).Model(&schema.Message{}).
		Where("id = ?", msgID).
		Update("status", status).Error
}

// DeleteMessage 删除消息.
func (r *MessageRepo) DeleteMessage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&schema.Message{}, "id = ?", id).Error
}
```

- [ ] **Step 2: 创建 ConversationRepo 实现**

```go
// chat-server/internal/repository/postgres/conversation.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// ConversationRepo 会话仓库实现.
type ConversationRepo struct {
	db *gorm.DB
}

// NewConversationRepo 创建会话仓库.
func NewConversationRepo(db *gorm.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

// CreateConversation 创建会话.
func (r *ConversationRepo) CreateConversation(ctx context.Context, conv *schema.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

// GetConversation 获取会话.
func (r *ConversationRepo) GetConversation(ctx context.Context, id string) (*schema.Conversation, error) {
	var conv schema.Conversation
	if err := r.db.WithContext(ctx).First(&conv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// UpdateConversation 更新会话.
func (r *ConversationRepo) UpdateConversation(ctx context.Context, conv *schema.Conversation) error {
	return r.db.WithContext(ctx).Save(conv).Error
}

// ListUserConversations 列出用户会话.
func (r *ConversationRepo) ListUserConversations(ctx context.Context, userID string) ([]*schema.Conversation, error) {
	var convs []*schema.Conversation
	if err := r.db.WithContext(ctx).
		Joins("JOIN conversation_members ON conversations.id = conversation_members.conversation_id").
		Where("conversation_members.user_id = ?", userID).
		Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

// DeleteConversation 删除会话.
func (r *ConversationRepo) DeleteConversation(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&schema.Conversation{}, "id = ?", id).Error
}
```

- [ ] **Step 3: 创建 FriendRepo 实现**

```go
// chat-server/internal/repository/postgres/friend.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// FriendRepo 好友仓库实现.
type FriendRepo struct {
	db *gorm.DB
}

// NewFriendRepo 创建好友仓库.
func NewFriendRepo(db *gorm.DB) *FriendRepo {
	return &FriendRepo{db: db}
}

// AddFriend 添加好友.
func (r *FriendRepo) AddFriend(ctx context.Context, friend *schema.Friend) error {
	return r.db.WithContext(ctx).Create(friend).Error
}

// RemoveFriend 移除好友.
func (r *FriendRepo) RemoveFriend(ctx context.Context, userID, friendID string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&schema.Friend{}).Error
}

// IsFriend 检查是否是好友.
func (r *FriendRepo) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&schema.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error
	return count > 0, err
}

// ListFriends 列出好友.
func (r *FriendRepo) ListFriends(ctx context.Context, userID string) ([]*schema.Friend, error) {
	var friends []*schema.Friend
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}
```

- [ ] **Step 4: 创建 GroupRepo 实现**

```go
// chat-server/internal/repository/postgres/group.go
package postgres

import (
	"context"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// GroupRepo 群组仓库实现.
type GroupRepo struct {
	db *gorm.DB
}

// NewGroupRepo 创建群组仓库.
func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

// CreateGroup 创建群组.
func (r *GroupRepo) CreateGroup(ctx context.Context, group *schema.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroup 获取群组.
func (r *GroupRepo) GetGroup(ctx context.Context, id string) (*schema.Group, error) {
	var group schema.Group
	if err := r.db.WithContext(ctx).First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup 更新群组.
func (r *GroupRepo) UpdateGroup(ctx context.Context, group *schema.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteGroup 删除群组.
func (r *GroupRepo) DeleteGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&schema.Group{}, "id = ?", id).Error
}

// AddGroupMember 添加群成员.
func (r *GroupRepo) AddGroupMember(ctx context.Context, member *schema.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveGroupMember 移除群成员.
func (r *GroupRepo) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	return r.db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&schema.GroupMember{}).Error
}

// ListGroupMembers 列出群成员.
func (r *GroupRepo) ListGroupMembers(ctx context.Context, groupID string) ([]*schema.GroupMember, error) {
	var members []*schema.GroupMember
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}
```

- [ ] **Step 5: 验证编译**

运行: `cd chat-server && go build ./...`
预期: 编译成功

### Task 3.4: 更新依赖注入和 Handler

**Files:**
- Modify: `chat-server/internal/app/app.go`
- Modify: `chat-server/internal/handler/message.go`
- Modify: `chat-server/internal/handler/conversation.go`
- Modify: `chat-server/internal/handler/friend.go`
- Modify: `chat-server/internal/handler/group.go`

- [ ] **Step 1: 更新 app.go 依赖注入**

```go
// chat-server/internal/app/app.go
package app

import (
	// ... 其他导入
	"github.com/gangantongxue/knowsync/chat-server/internal/repository/postgres"
	"github.com/gangantongxue/knowsync/chat-server/internal/service"
)

func NewApp() error {
	// ... 初始化配置、数据库等
	
	// 创建具体实现
	msgRepo := postgres.NewMessageRepo(db.DB)
	convRepo := postgres.NewConversationRepo(db.DB)
	friendRepo := postgres.NewFriendRepo(db.DB)
	groupRepo := postgres.NewGroupRepo(db.DB)
	
	// 创建服务（注入接口，Go 隐式满足）
	msgSvc := service.NewMessageService(msgRepo, convRepo)
	convSvc := service.NewConversationService(convRepo)
	friendSvc := service.NewFriendService(friendRepo)
	groupSvc := service.NewGroupService(groupRepo)
	
	// 创建 handler
	h := handler.NewHandler(msgSvc, convSvc, friendSvc, groupSvc)
	
	// ... 启动服务
}
```

- [ ] **Step 2: 验证编译**

运行: `cd chat-server && go build ./...`
预期: 编译成功

---

## 阶段四：gateway handler 拆分（2-3 小时）

### Task 4.1: 拆分 Handler 结构体

**Files:**
- Create: `gateway/internal/handler/repo_handler.go`
- Create: `gateway/internal/handler/user_handler.go`
- Create: `gateway/internal/handler/ai_handler.go`
- Create: `gateway/internal/handler/chat_handler.go`

- [ ] **Step 1: 创建 RepoHandler**

```go
// gateway/internal/handler/repo_handler.go
package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// RepoHandler 知识库处理器.
type RepoHandler struct {
	grpcClient  *grpcclient.Client
	store       *storage.Store
	authManager *serviceauth.Manager
}

// NewRepoHandler 创建知识库处理器.
func NewRepoHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *RepoHandler {
	return &RepoHandler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}

// CreateRepo 创建知识库.
func (h *RepoHandler) CreateRepo() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 实现创建知识库逻辑
	}
}

// GetRepo 获取知识库信息.
func (h *RepoHandler) GetRepo() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 实现获取知识库信息逻辑
	}
}
```

- [ ] **Step 2: 创建 UserHandler**

```go
// gateway/internal/handler/user_handler.go
package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// UserHandler 用户处理器.
type UserHandler struct {
	grpcClient  *grpcclient.Client
	store       *storage.Store
	authManager *serviceauth.Manager
}

// NewUserHandler 创建用户处理器.
func NewUserHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *UserHandler {
	return &UserHandler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}

// GetUser 获取用户信息.
func (h *UserHandler) GetUser() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 实现获取用户信息逻辑
	}
}
```

- [ ] **Step 3: 更新主 Handler 结构体**

```go
// gateway/internal/handler/handler.go
package handler

import (
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// Handler HTTP 处理器容器.
type Handler struct {
	RepoHandler    *RepoHandler
	UserHandler    *UserHandler
	AIHandler      *AIHandler
	ChatHandler    *ChatHandler
}

// NewHandler 创建 Handler 实例.
func NewHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *Handler {
	return &Handler{
		RepoHandler: NewRepoHandler(grpcClient, store, authManager),
		UserHandler: NewUserHandler(grpcClient, store, authManager),
		AIHandler:   NewAIHandler(grpcClient, store, authManager),
		ChatHandler: NewChatHandler(grpcClient, store, authManager),
	}
}
```

- [ ] **Step 4: 验证编译**

运行: `cd gateway && go build ./...`
预期: 编译成功

### Task 4.2: 更新路由注册

**Files:**
- Modify: `gateway/internal/router/router.go`

- [ ] **Step 1: 更新路由注册**

```go
// gateway/internal/router/router.go
package router

import (
	"github.com/gangantongxue/knowsync/gateway/internal/handler"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func SetupRoutes(h *handler.Handler, r *server.Hertz) {
	// 用户相关路由
	userGroup := r.Group("/api/user")
	{
		userGroup.GET("/:user_id", h.UserHandler.GetUser())
		userGroup.PUT("/:user_id", h.UserHandler.UpdateUser())
	}
	
	// 知识库相关路由
	repoGroup := r.Group("/api/repo")
	{
		repoGroup.POST("", h.RepoHandler.CreateRepo())
		repoGroup.GET("/:repo_id", h.RepoHandler.GetRepo())
		repoGroup.PUT("/:repo_id", h.RepoHandler.UpdateRepo())
		repoGroup.DELETE("/:repo_id", h.RepoHandler.DeleteRepo())
	}
	
	// ... 其他路由
}
```

- [ ] **Step 2: 验证编译**

运行: `cd gateway && go build ./...`
预期: 编译成功

---

## 阶段五：创建共享 pkg 模块（2-3 小时）

### Task 5.1: 创建共享包结构

**Files:**
- Create: `knowsync/pkg/config/config.go`
- Create: `knowsync/pkg/database/database.go`
- Create: `knowsync/pkg/logger/logger.go`
- Create: `knowsync/pkg/redis/redis.go`
- Create: `knowsync/go.mod`

- [ ] **Step 1: 创建 go.mod**

```bash
mkdir -p knowsync/pkg
cd knowsync && go mod init github.com/gangantongxue/knowsync/pkg
```

- [ ] **Step 2: 创建共享配置包**

```go
// knowsync/pkg/config/config.go
package config

import (
	"github.com/spf13/viper"
)

// Config 通用配置结构.
type Config struct {
	Database DatabaseConfig `yaml:"database" mapstructure:"database"`
	Redis    RedisConfig    `yaml:"redis" mapstructure:"redis"`
}

// DatabaseConfig 数据库配置.
type DatabaseConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	DBName   string `yaml:"dbname" mapstructure:"dbname"`
}

// RedisConfig Redis 配置.
type RedisConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

// LoadConfig 加载配置.
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}
```

- [ ] **Step 3: 创建共享数据库包**

```go
// knowsync/pkg/database/database.go
package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database 数据库连接.
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接.
func NewDatabase(cfg *Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	return &Database{DB: db}, nil
}
```

- [ ] **Step 4: 创建共享日志包**

```go
// knowsync/pkg/logger/logger.go
package logger

import (
	"log/slog"
	"os"
)

// Init 初始化全局日志.
func Init(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
```

- [ ] **Step 5: 创建共享 Redis 包**

```go
// knowsync/pkg/redis/redis.go
package redis

import (
	"fmt"
	"github.com/redis/go-redis/v9"
)

// Redis Redis 连接.
type Redis struct {
	RDB *redis.Client
}

// NewRedis 创建 Redis 连接.
func NewRedis(cfg *Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	
	return &Redis{RDB: rdb}, nil
}
```

- [ ] **Step 6: 验证编译**

运行: `cd knowsync/pkg && go build ./...`
预期: 编译成功

### Task 5.2: 更新各服务引用

**Files:**
- Modify: `user-server/go.mod`
- Modify: `repo-server/go.mod`
- Modify: `chat-server/go.mod`
- Modify: `gateway/go.mod`
- Modify: 各服务的导入路径

- [ ] **Step 1: 更新 user-server go.mod**

```bash
cd user-server
go mod edit -require github.com/gangantongxue/knowsync/pkg@v0.0.0
go mod edit -replace github.com/gangantongxue/knowsync/pkg=../knowsync/pkg
go mod tidy
```

- [ ] **Step 2: 更新导入路径**

将各服务中的导入路径从：
```go
import "github.com/gangantongxue/knowsync/user-server/pkg/config"
```

改为：
```go
import "github.com/gangantongxue/knowsync/pkg/config"
```

- [ ] **Step 3: 验证编译**

运行: `cd user-server && go build ./...`
预期: 编译成功

- [ ] **Step 4: 重复步骤 1-3 为其他服务**

为 repo-server、chat-server、gateway 重复上述步骤。

---

## 阶段六：移除 Logger 注入（1-2 小时）

### Task 6.1: 移除 Logger 字段

**Files:**
- Modify: `user-server/internal/service/service.go`
- Modify: `user-server/pkg/database/database.go`
- Modify: `repo-server/internal/service/service.go`
- Modify: `repo-server/pkg/database/database.go`
- Modify: `chat-server/internal/service/service.go`
- Modify: `chat-server/pkg/database/database.go`

- [ ] **Step 1: 移除 user-server Logger 字段**

```go
// user-server/internal/service/service.go
type Service struct {
	Cfg        *config.Config
	Repository *repository.Repository
	Mailer     *mail.Mailer
	privateKey *rsa.PrivateKey
	
	// 移除 Logger *logger.Logger 字段
}
```

- [ ] **Step 2: 更新所有 s.Logger.Info() 调用为 slog.Info()**

```go
// 搜索并替换
s.Logger.Info("操作成功", "key", value)
// 替换为
slog.Info("操作成功", "key", value)
```

- [ ] **Step 3: 验证编译**

运行: `cd user-server && go build ./...`
预期: 编译成功

- [ ] **Step 4: 重复步骤 1-3 为其他服务**

为 repo-server、chat-server 重复上述步骤。

---

## 阶段七：统一 gRPC 响应风格（2-3 小时）

### Task 7.1: 修改 Proto 定义

**Files:**
- Modify: `ks-proto/proto/user.proto`
- Modify: `ks-proto/proto/repo.proto`
- Modify: `ks-proto/proto/chat.proto`

- [ ] **Step 1: 移除 success 字段**

```protobuf
// 修改前
message GetUserResponse {
  bool success = 1;
  string msg = 2;
  User user = 3;
}

// 修改后
message GetUserResponse {
  User user = 1;
}
```

- [ ] **Step 2: 重新生成 proto 文件**

```bash
cd ks-proto && make generate
```

- [ ] **Step 3: 验证生成**

运行: `cd ks-proto && go build ./...`
预期: 编译成功

### Task 7.2: 更新 Handler 返回值

**Files:**
- Modify: `user-server/internal/handler/user.go`
- Modify: `user-server/internal/handler/auth.go`
- Modify: `repo-server/internal/handler/repo.go`
- Modify: `chat-server/internal/handler/message.go`

- [ ] **Step 1: 更新 user-server handler**

```go
// 修改前
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.userSvc.GetUser(ctx, req.UserId)
	if err != nil {
		return &pb.GetUserResponse{
			Success: false,
			Msg:     "获取用户信息失败",
		}, nil
	}
	
	return &pb.GetUserResponse{
		Success: true,
		Msg:     "success",
		User: &pb.User{
			Id:   user.ID,
			Name: user.Name,
		},
	}, nil
}

// 修改后
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.userSvc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户信息失败")
	}
	
	if user == nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	
	return &pb.GetUserResponse{
		User: &pb.User{
			Id:   user.ID,
			Name: user.Name,
		},
	}, nil
}
```

- [ ] **Step 2: 验证编译**

运行: `cd user-server && go build ./...`
预期: 编译成功

- [ ] **Step 3: 重复步骤 1-2 为其他服务**

为 repo-server、chat-server 重复上述步骤。

---

## 阶段八：测试和验证（2-3 小时）

### Task 8.1: 运行现有测试

**Files:**
- All test files

- [ ] **Step 1: 运行 user-server 测试**

```bash
cd user-server && go test ./...
```

- [ ] **Step 2: 运行 repo-server 测试**

```bash
cd repo-server && go test ./...
```

- [ ] **Step 3: 运行 chat-server 测试**

```bash
cd chat-server && go test ./...
```

- [ ] **Step 4: 运行 gateway 测试**

```bash
cd gateway && go test ./...
```

### Task 8.2: 验证重构结果

- [ ] **Step 1: 检查接口满足**

```bash
# 检查隐式接口满足
go build ./...
```

- [ ] **Step 2: 检查代码重复率**

使用工具检查代码重复率，确保降低 50% 以上。

- [ ] **Step 3: 检查结构体方法数**

确保每个 Service 结构体方法数从 20+ 降到 5-10 个。

- [ ] **Step 4: 运行 linter**

```bash
golangci-lint run ./...
```

---

## 完成

**计划完成并保存到 `docs/superpowers/plans/2026-06-08-go-refactoring-plan.md`。**

**两种执行选项：**

**1. Subagent-Driven (推荐)** - 我为每个任务分发一个新的子代理，任务间进行审查，快速迭代

**2. Inline Execution** - 在当前会话中使用 executing-plans 执行任务，批量执行并设置检查点

**选择哪种方法？**