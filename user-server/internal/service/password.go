package service

import (
	"context"
)

// VerifyCode 发送邮箱验证码
func (s *Service) VerifyCode(ctx context.Context, email string) error {
	// TODO: implement me
	return nil
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
