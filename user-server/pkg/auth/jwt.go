package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims JWT access token 声明
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 JWT access token
func GenerateAccessToken(userID, secret string, ttl time.Duration) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "knowsync-user-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("生成 JWT 失败: %w", err)
	}
	return tokenString, nil
}

// ValidateAccessToken 校验并解析 JWT access token，返回 user_id
func ValidateAccessToken(tokenString, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("JWT 校验失败: %w", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("JWT 声明解析失败")
	}

	return claims.UserID, nil
}
