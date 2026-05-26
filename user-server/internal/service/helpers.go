package service

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// HashPassword 使用 bcrypt 对密码进行哈希处理
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword 校验密码是否与哈希匹配
func CheckPassword(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("密码错误")
	}
	return nil
}

// validateRegisterParams 校验注册参数
func validateRegisterParams(name, email, password, verifyCode string) error {
	if name == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if len(name) > 64 {
		return fmt.Errorf("用户名不能超过 64 个字符")
	}
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("邮箱格式不正确")
	}
	if len(email) > 254 {
		return fmt.Errorf("邮箱不能超过 254 个字符")
	}
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于 6 位")
	}
	if len(password) > 72 {
		return fmt.Errorf("密码长度不能超过 72 位")
	}
	if verifyCode == "" {
		return fmt.Errorf("验证码不能为空")
	}
	return nil
}
