// Package serviceauth provides service-to-service token management.
package serviceauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func newJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

const (
	redisKeyPrefix = "service_token:"
	defaultTTL     = 30 * time.Minute
)

// TokenInfo 存储在 Redis 中的 token 信息.
type TokenInfo struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// serviceTokenClaims JWT 声明.
type serviceTokenClaims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// Manager service token 管理器，仅用于 gateway 侧.
type Manager struct {
	rdb    *redis.Client
	secret []byte
	ttl    time.Duration
}

// NewManager 创建 service token 管理器.
func NewManager(rdb *redis.Client, secret string) *Manager {
	ttl := defaultTTL
	return &Manager{
		rdb:    rdb,
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// Generate 生成 service token，存入 Redis.
func (m *Manager) Generate(ctx context.Context, userID, sessionID string) (tokenString, jti string, err error) {
	jti = newJTI()
	now := time.Now()

	claims := serviceTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(m.secret)
	if err != nil {
		return "", "", fmt.Errorf("签发 service token 失败: %w", err)
	}

	// 存入 Redis
	info := TokenInfo{UserID: userID, SessionID: sessionID}
	data, _ := json.Marshal(info)
	if err := m.rdb.Set(ctx, redisKeyPrefix+jti, data, m.ttl).Err(); err != nil {
		return "", "", fmt.Errorf("存储 service token 失败: %w", err)
	}

	return tokenString, jti, nil
}

// Validate 验证 service token，返回 token 信息.
func (m *Manager) Validate(ctx context.Context, tokenString string) (*TokenInfo, error) {
	claims := &serviceTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方式: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("service token 无效: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("service token 不可用")
	}

	// 检查 Redis 中是否存在
	data, err := m.rdb.Get(ctx, redisKeyPrefix+claims.ID).Bytes()
	if err != nil {
		return nil, errors.New("service token 不存在或已失效")
	}

	var info TokenInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("解析 token 信息失败: %w", err)
	}

	return &info, nil
}

// Revoke 立即吊销 token.
func (m *Manager) Revoke(ctx context.Context, tokenString string) error {
	claims := &serviceTokenClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err != nil {
		return fmt.Errorf("解析 token 失败: %w", err)
	}
	return m.rdb.Del(ctx, redisKeyPrefix+claims.ID).Err()
}

// RevokeByJTI 按 jti 吊销 token.
func (m *Manager) RevokeByJTI(ctx context.Context, jti string) error {
	return m.rdb.Del(ctx, redisKeyPrefix+jti).Err()
}

// ValidateInternalSecret 验证共享密钥（用于异步 worker 等无 JWT 场景）.
func (m *Manager) ValidateInternalSecret(rawSecret string) bool {
	return rawSecret == string(m.secret)
}
