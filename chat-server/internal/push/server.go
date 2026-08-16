package push

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gangantongxue/knowsync/chat-server/pkg/config"
	"github.com/gangantongxue/knowsync/chat-server/pkg/redis"
	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims JWT 访问令牌声明.
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Server SSE 推送服务器，负责鉴权并维护用户的 SSE 长连接.
type Server struct {
	hub       *Hub
	cfg       *config.Config
	redis     *redis.Redis
	publicKey *rsa.PublicKey
	httpSrv   *http.Server
}

// NewServer 创建 SSE 推送服务器.
func NewServer(hub *Hub, cfg *config.Config, r *redis.Redis) (*Server, error) {
	publicKey, err := loadPublicKey(cfg.JWT.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载 RSA 公钥失败: %w", err)
	}

	s := &Server{
		hub:       hub,
		cfg:       cfg,
		redis:     r,
		publicKey: publicKey,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", s.handleSSE)

	s.httpSrv = &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.SSE.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s, nil
}

// Start 启动 SSE HTTP 服务.
func (s *Server) Start() error {
	slog.Info("SSE 推送服务启动成功", "port", s.cfg.SSE.Port)
	return s.httpSrv.ListenAndServe()
}

// Shutdown 优雅关闭 SSE 服务.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

// handleSSE 处理 SSE 长连接请求：校验 token、注册连接、维持连接直至客户端断开.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "缺少 token 参数", http.StatusUnauthorized)
		return
	}

	userID, err := s.validateToken(token)
	if err != nil {
		slog.Warn("SSE token 验证失败", "error", err)
		http.Error(w, "token 无效或已过期", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持流式响应", http.StatusInternalServerError)
		return
	}

	// 设置 SSE 响应头，禁用代理缓冲以便事件即时下发
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// 立即写出连接确认注释行：客户端（EventSource）收到首个字节才会触发 onopen，
	// 同时让前置代理（Caddy）完成握手后不再缓冲后续响应
	if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	client := NewSSEClient(s.hub, w, userID)
	client.flusher = flusher

	userOnlineKey := "user_online:" + userID

	// 设置心跳回调，刷新 Redis 在线状态 TTL
	client.onPing = func() {
		ctx := context.Background()
		if err := s.redis.RDB.Expire(ctx, userOnlineKey, 60*time.Second).Err(); err != nil {
			slog.Warn("刷新在线状态 TTL 失败", "user_id", userID, "error", err)
		}
	}

	// 设置断开连接回调，清理 Redis 在线状态
	client.onDisconnect = func() {
		ctx := context.Background()
		if err := s.redis.RDB.Del(ctx, userOnlineKey).Err(); err != nil {
			slog.Warn("清理在线状态失败", "user_id", userID, "error", err)
		}
	}

	// 注册连接并设置 Redis 在线状态
	s.hub.register <- client

	ctx := context.Background()
	if err := s.redis.RDB.Set(ctx, userOnlineKey, 1, 60*time.Second).Err(); err != nil {
		slog.Warn("设置在线状态失败", "user_id", userID, "error", err)
	}

	// 连接结束后注销并清理在线状态
	defer func() {
		s.hub.unregister <- client
		if client.onDisconnect != nil {
			client.onDisconnect()
		}
	}()

	client.Serve(r.Context())
}

// validateToken 验证 JWT token 并提取 user_id.
func (s *Server) validateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.publicKey, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	return claims.UserID, nil
}

// loadPublicKey 从文件加载 RSA 公钥.
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path) //nolint:gosec // 路径由配置文件指定，非用户输入
	if err != nil {
		return nil, fmt.Errorf("读取 RSA 公钥文件失败: %w", err)
	}
	return jwt.ParseRSAPublicKeyFromPEM(keyData)
}
