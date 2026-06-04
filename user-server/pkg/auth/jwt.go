// Package auth provides JWT token generation, validation, and RSA key loading.
package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims JWT access token 声明.
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 JWT access token.
func GenerateAccessToken(userID string, key *rsa.PrivateKey, ttl time.Duration) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "knowsync-user-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("生成 JWT 失败: %w", err)
	}
	return tokenString, nil
}

// ValidateAccessToken 验证 JWT access token 并返回用户 ID.
func ValidateAccessToken(tokenString string, key *rsa.PublicKey) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return "", fmt.Errorf("JWT 校验失败: %w", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return "", errors.New("JWT 声明解析失败")
	}

	return claims.UserID, nil
}

// LoadPrivateKeyFromFile 从文件加载 RSA 私钥.
func LoadPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	//nolint:gosec // path is configured, not user-controlled
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 RSA 私钥文件失败: %w", err)
	}
	return jwt.ParseRSAPrivateKeyFromPEM(keyData)
}

// LoadPublicKeyFromFile 从文件加载 RSA 公钥.
func LoadPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	//nolint:gosec // path is configured, not user-controlled
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 RSA 公钥文件失败: %w", err)
	}
	return jwt.ParseRSAPublicKeyFromPEM(keyData)
}
