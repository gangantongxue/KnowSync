package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
)

// VerifyCode 发送邮箱验证码
func (s *Service) VerifyCode(ctx context.Context, email string) error {
	// 检查发送频率限制，防止频繁请求
	allowed, err := s.Repository.SetVerifyCodeRateLimit(ctx, email)
	if err != nil {
		slog.Error("检查发送频率限制失败", "email", email, "error", err)
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	if !allowed {
		return fmt.Errorf("发送过于频繁，请稍后再试")
	}

	// 生成 4 位随机验证码
	code, err := generateVerifyCode()
	if err != nil {
		slog.Error("生成验证码失败", "email", email, "error", err)
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存入 Redis，有效期 5 分钟
	if err := s.Repository.SetVerifyCode(ctx, email, code); err != nil {
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

// generateVerifyCode 使用密码学安全随机数生成 4 位数字验证码
func generateVerifyCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

// ForgetPassword 忘记密码（通过邮箱验证码重置密码）
func (s *Service) ForgetPassword(ctx context.Context, email, password, verifyCode string) error {
	// 1. 验证码校验
	storedCode, err := s.Repository.GetVerifyCode(ctx, email)
	if err != nil {
		return fmt.Errorf("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		return fmt.Errorf("验证码错误")
	}

	// 2. 根据邮箱查找用户
	user, err := s.Repository.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("该邮箱未注册")
	}

	// 3. 对新密码进行哈希处理
	hashedPassword, err := HashPassword(password)
	if err != nil {
		slog.Error("密码加密失败", "email", email, "error", err)
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 更新密码
	user.Password = hashedPassword
	if err := s.Repository.UpdateUser(ctx, user); err != nil {
		slog.Error("重置密码失败", "email", email, "error", err)
		return fmt.Errorf("重置密码失败: %w", err)
	}

	// 5. 清理该用户的所有会话，强制重新登录
	if err := s.Repository.InvalidateUserSessions(ctx, user.ID); err != nil {
		slog.Warn("重置密码时清理会话失败", "user_id", user.ID, "error", err)
	}

	// 6. 删除已使用的验证码
	if err := s.Repository.DeleteVerifyCode(ctx, email); err != nil {
		slog.Warn("删除验证码失败", "email", email, "error", err)
	}

	slog.Info("密码重置成功", "email", email)
	return nil
}

// ResetPassword 重置密码（通过旧密码验证）
func (s *Service) ResetPassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	// 1. 查找用户
	user, err := s.Repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 2. 校验旧密码
	if err := CheckPassword(oldPassword, user.Password); err != nil {
		return fmt.Errorf("原密码错误")
	}

	// 3. 对新密码进行哈希处理
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		slog.Error("密码加密失败", "user_id", userID, "error", err)
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 更新密码
	user.Password = hashedPassword
	if err := s.Repository.UpdateUser(ctx, user); err != nil {
		slog.Error("修改密码失败", "user_id", userID, "error", err)
		return fmt.Errorf("修改密码失败: %w", err)
	}

	// 5. 清理该用户的所有会话，强制重新登录
	if err := s.Repository.InvalidateUserSessions(ctx, userID); err != nil {
		slog.Warn("修改密码时清理会话失败", "user_id", userID, "error", err)
	}

	slog.Info("密码修改成功", "user_id", userID)
	return nil
}
