package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
)

// VerifyCode 发送邮箱验证码
func (s *Service) VerifyCode(ctx context.Context, email string) error {
	// 检查发送频率限制，防止频繁请求
	allowed, err := s.Repository.SetVerifyCodeRateLimit(ctx, email)
	if err != nil {
		s.Logger.Logger.Error("检查发送频率限制失败", "email", email, "error", err)
		return fmt.Errorf("检查发送频率失败: %w", err)
	}
	if !allowed {
		return fmt.Errorf("发送过于频繁，请稍后再试")
	}

	// 生成 4 位随机验证码
	code, err := generateVerifyCode()
	if err != nil {
		s.Logger.Logger.Error("生成验证码失败", "email", email, "error", err)
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存入 Redis，有效期 5 分钟
	if err := s.Repository.SetVerifyCode(ctx, email, code); err != nil {
		s.Logger.Logger.Error("存储验证码到 Redis 失败", "email", email, "error", err)
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	// 发送邮件
	if err := s.Mailer.SendVerifyCode(email, code); err != nil {
		return fmt.Errorf("发送验证码失败: %w", err)
	}

	s.Logger.Logger.Info("验证码发送成功", "email", email)
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

// ForgetPassword 忘记密码
func (s *Service) ForgetPassword(ctx context.Context, email, password, verifyCode string) error {
	// TODO: implement me
	return nil
}

// ResetPassword 重置密码
func (s *Service) ResetPassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// TODO: implement me
	return nil
}
