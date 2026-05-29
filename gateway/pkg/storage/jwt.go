package storage

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// FileAccessClaims JWT 文件访问声明
type FileAccessClaims struct {
	Path   string `json:"path"`   // 文件相对路径
	Bucket string `json:"bucket"` // 存储桶
	jwt.RegisteredClaims
}

// SignTempURL 签发认证读桶的临时访问 URL
// 返回格式: /files/auth/{jwt_token}
func (s *Store) SignTempURL(bucket Bucket, key string) (string, error) {
	if bucket != BucketAuth {
		return "", ErrPublicBucketNoSign
	}

	now := time.Now()
	claims := FileAccessClaims{
		Path:   key,
		Bucket: string(bucket),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtTTL())),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("签发 JWT 失败: %w", err)
	}

	return fmt.Sprintf("/files/auth/%s", tokenString), nil
}

// ValidateToken 校验 JWT 令牌并返回声明
func (s *Store) ValidateToken(tokenString string) (*FileAccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &FileAccessClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
			}
			return s.jwtSecret, nil
		})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*FileAccessClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
