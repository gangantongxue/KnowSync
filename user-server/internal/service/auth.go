package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gangantongxue/knowsync/user-server/pkg/auth"
	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// Register 注册用户
func (s *Service) Register(ctx context.Context, name, email, password, verifyCode string) (*schema.User, error) {
	// 1. 参数校验
	if err := validateRegisterParams(name, email, password, verifyCode); err != nil {
		slog.Warn("注册参数校验失败", "email", email, "error", err)
		return nil, err
	}

	// 2. 验证码校验
	storedCode, err := s.Repository.GetVerifyCode(ctx, email)
	if err != nil {
		slog.Error("获取验证码失败", "email", email, "error", err)
		return nil, fmt.Errorf("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		slog.Warn("验证码错误", "email", email)
		return nil, fmt.Errorf("验证码错误")
	}

	// 3. 检查邮箱是否已被注册
	existingUser, err := s.Repository.GetUserByEmail(ctx, email)
	if err != nil && err != gorm.ErrRecordNotFound {
		slog.Error("检查邮箱是否已注册失败", "email", email, "error", err)
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existingUser != nil {
		slog.Warn("邮箱已被注册", "email", email)
		return nil, fmt.Errorf("该邮箱已被注册")
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
	if err := s.Repository.CreateUser(user); err != nil {
		slog.Error("创建用户失败", "email", email, "error", err)
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 6. 删除已使用的验证码
	if err := s.Repository.DeleteVerifyCode(ctx, email); err != nil {
		slog.Warn("删除验证码失败", "email", email, "error", err)
	}

	slog.Info("用户注册成功", "user_id", user.ID, "email", email)
	return user, nil
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, email, password, clientIP string) (*schema.User, string, string, error) {
	// 1. 根据邮箱查找用户
	user, err := s.Repository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("邮箱或密码错误")
	}

	// 2. 校验密码
	if err := CheckPassword(password, user.Password); err != nil {
		return nil, "", "", fmt.Errorf("邮箱或密码错误")
	}

	// 3. 清理该用户所有旧会话，保证一个用户至多一个活跃会话
	if err := s.Repository.InvalidateUserSessions(ctx, user.ID); err != nil {
		slog.Error("清理旧会话失败", "user_id", user.ID, "error", err)
	}

	// 4. 生成 JWT access token
	accessToken, err := auth.GenerateAccessToken(user.ID, s.privateKey, s.Cfg.Auth.AccessTTL)
	if err != nil {
		slog.Error("生成 access token 失败", "user_id", user.ID, "error", err)
		return nil, "", "", fmt.Errorf("生成访问凭证失败")
	}

	// 5. 生成随机 refresh token
	refreshToken, err := generateRefreshToken()
	if err != nil {
		slog.Error("生成 refresh token 失败", "user_id", user.ID, "error", err)
		return nil, "", "", fmt.Errorf("生成刷新凭证失败")
	}

	// 6. 创建新会话
	session := &schema.UserSession{
		UserID:           user.ID,
		ClientIP:         clientIP,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpireAt:         time.Now().Add(s.Cfg.Auth.RefreshTTL),
		LoginAt:          time.Now(),
	}
	if err := s.Repository.CreateSession(ctx, session); err != nil {
		slog.Error("创建会话失败", "user_id", user.ID, "error", err)
		return nil, "", "", fmt.Errorf("创建会话失败")
	}

	slog.Info("用户登录成功", "user_id", user.ID, "client_ip", clientIP)
	return user, accessToken, refreshToken, nil
}

// Logout 用户退出登录，清空该用户全部活跃会话
func (s *Service) Logout(ctx context.Context, refreshToken, clientIP string) error {
	// 1. 通过 refresh token 哈希值查找会话，获取 user_id
	session, err := s.Repository.GetSessionByRefreshToken(ctx, hashRefreshToken(refreshToken))
	if err == nil && session != nil {
		// 2. 清空该用户所有活跃会话
		if err := s.Repository.InvalidateUserSessions(ctx, session.UserID); err != nil {
			slog.Error("退出登录时清理会话失败", "user_id", session.UserID, "error", err)
			return fmt.Errorf("退出登录失败: %w", err)
		}
		slog.Info("用户退出登录成功", "user_id", session.UserID)
	}
	// 即使 refresh token 查不到也返回成功（幂等设计）
	return nil
}

// Refresh 刷新登录凭证
func (s *Service) Refresh(ctx context.Context, refreshToken, clientIP string) (string, string, error) {
	// 1. 通过 refresh token 哈希值查找会话
	session, err := s.Repository.GetSessionByRefreshToken(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		return "", "", fmt.Errorf("刷新凭证无效或已过期")
	}

	// 2. 检查会话是否已退出
	if session.LogoutAt != nil {
		return "", "", fmt.Errorf("刷新凭证已失效")
	}

	// 3. 检查 refresh token 是否过期
	if time.Now().After(session.ExpireAt) {
		return "", "", fmt.Errorf("刷新凭证已过期")
	}

	// 4. 生成新的 access token
	accessToken, err := auth.GenerateAccessToken(session.UserID, s.privateKey, s.Cfg.Auth.AccessTTL)
	if err != nil {
		slog.Error("刷新时生成 access token 失败", "user_id", session.UserID, "error", err)
		return "", "", fmt.Errorf("生成访问凭证失败")
	}

	// 5. 生成新的 refresh token
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		slog.Error("刷新时生成 refresh token 失败", "user_id", session.UserID, "error", err)
		return "", "", fmt.Errorf("生成刷新凭证失败")
	}

	// 6. 更新会话
	session.RefreshTokenHash = hashRefreshToken(newRefreshToken)
	session.ClientIP = clientIP
	if err := s.Repository.UpdateSession(ctx, session); err != nil {
		slog.Error("刷新时更新会话失败", "user_id", session.UserID, "error", err)
		return "", "", fmt.Errorf("更新会话失败")
	}

	slog.Info("刷新凭证成功", "user_id", session.UserID)
	return accessToken, newRefreshToken, nil
}
