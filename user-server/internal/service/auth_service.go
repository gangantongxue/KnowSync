// Package service provides business logic for the user-server.
package service

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gangantongxue/knowsync/user-server/pkg/auth"
	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"github.com/gangantongxue/knowsync/user-server/pkg/mail"
	"gorm.io/gorm"
)

// AuthService 认证服务.
type AuthService struct {
	userRepo       UserRepository
	sessionRepo    SessionRepository
	verifyCodeRepo VerifyCodeRepository
	Cfg            *config.Config
	privateKey     *rsa.PrivateKey
	Mailer         *mail.Mailer
}

// NewAuthService 创建认证服务.
func NewAuthService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	verifyCodeRepo VerifyCodeRepository,
	cfg *config.Config,
	privateKey *rsa.PrivateKey,
	mailer *mail.Mailer,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		verifyCodeRepo: verifyCodeRepo,
		Cfg:            cfg,
		privateKey:     privateKey,
		Mailer:         mailer,
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
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
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
	accessToken, err := auth.GenerateAccessToken(user.ID, s.privateKey, s.Cfg.Auth.AccessTTL)
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
		ExpireAt:         time.Now().Add(s.Cfg.Auth.RefreshTTL),
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
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
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
	accessToken, err := auth.GenerateAccessToken(session.UserID, s.privateKey, s.Cfg.Auth.AccessTTL)
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

// VerifyCode 发送邮箱验证码.
func (s *AuthService) VerifyCode(ctx context.Context, email string) error {
	// 检查发送频率限制，防止频繁请求
	allowed, err := s.verifyCodeRepo.SetVerifyCodeRateLimit(ctx, email)
	if err != nil {
		slog.Error("检查发送频率限制失败", "email", email, "error", err)
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	if !allowed {
		return errors.New("发送过于频繁，请稍后再试")
	}

	// 生成 4 位随机验证码
	code, err := generateVerifyCode()
	if err != nil {
		slog.Error("生成验证码失败", "email", email, "error", err)
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存入 Redis，有效期 5 分钟
	if err := s.verifyCodeRepo.SetVerifyCode(ctx, email, code); err != nil {
		slog.Error("存储验证码到 Redis 失败", "email", email, "error", err)
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	// 发送邮件
	if err := s.Mailer.SendVerifyCode(email, code); err != nil {
		return fmt.Errorf("发送验证码失败: %w", err)
	}

	slog.Info("验证码发送成功", "email", email)
	return nil
}

// ForgetPassword 忘记密码（通过邮箱验证码重置密码）.
func (s *AuthService) ForgetPassword(ctx context.Context, email, password, verifyCode string) error {
	// 1. 验证码校验
	storedCode, err := s.verifyCodeRepo.GetVerifyCode(ctx, email)
	if err != nil {
		return errors.New("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	// 2. 根据邮箱查找用户
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return errors.New("该邮箱未注册")
	}

	// 3. 对新密码进行哈希处理
	hashedPassword, err := HashPassword(password)
	if err != nil {
		slog.Error("密码加密失败", "email", email, "error", err)
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 更新密码
	user.Password = hashedPassword
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		slog.Error("重置密码失败", "email", email, "error", err)
		return fmt.Errorf("重置密码失败: %w", err)
	}

	// 5. 清理该用户的所有会话，强制重新登录
	if err := s.sessionRepo.InvalidateUserSessions(ctx, user.ID); err != nil {
		slog.Warn("重置密码时清理会话失败", "user_id", user.ID, "error", err)
	}

	// 6. 删除已使用的验证码
	if err := s.verifyCodeRepo.DeleteVerifyCode(ctx, email); err != nil {
		slog.Warn("删除验证码失败", "email", email, "error", err)
	}

	slog.Info("密码重置成功", "email", email)
	return nil
}

// ResetPassword 重置密码（通过旧密码验证）.
func (s *AuthService) ResetPassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// 1. 查找用户
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 2. 校验旧密码
	if err := CheckPassword(oldPassword, user.Password); err != nil {
		return errors.New("原密码错误")
	}

	// 3. 对新密码进行哈希处理
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		slog.Error("密码加密失败", "user_id", userID, "error", err)
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 更新密码
	user.Password = hashedPassword
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		slog.Error("修改密码失败", "user_id", userID, "error", err)
		return fmt.Errorf("修改密码失败: %w", err)
	}

	// 5. 清理该用户的所有会话，强制重新登录
	if err := s.sessionRepo.InvalidateUserSessions(ctx, userID); err != nil {
		slog.Warn("修改密码时清理会话失败", "user_id", userID, "error", err)
	}

	slog.Info("密码修改成功", "user_id", userID)
	return nil
}
