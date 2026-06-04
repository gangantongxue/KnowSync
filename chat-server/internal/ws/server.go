package ws

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gangantongxue/knowsync/chat-server/pkg/config"
	"github.com/gangantongxue/knowsync/chat-server/pkg/redis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// AccessTokenClaims JWT 访问令牌声明.
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// Server WebSocket 服务器.
type Server struct {
	hub       *Hub
	cfg       *config.Config
	redis     *redis.Redis
	publicKey *rsa.PublicKey
	httpSrv   *http.Server
}

// NewServer 创建 WebSocket 服务器.
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
	mux.HandleFunc("/ws", s.handleWS)

	s.httpSrv = &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.WS.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s, nil
}

// Start 启动 WebSocket HTTP 服务.
func (s *Server) Start() error {
	slog.Info("WebSocket 服务启动成功", "port", s.cfg.WS.Port)
	return s.httpSrv.ListenAndServe()
}

// Shutdown 优雅关闭 WebSocket 服务.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

// handleWS 处理 WebSocket 升级请求.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "缺少 token 参数", http.StatusUnauthorized)
		return
	}

	userID, err := s.validateToken(token)
	if err != nil {
		slog.Warn("WebSocket token 验证失败", "error", err)
		http.Error(w, "token 无效或已过期", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket 升级失败", "error", err)
		return
	}

	client := NewClient(s.hub, conn, userID)

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

	s.hub.register <- client

	// 设置 Redis 在线状态
	ctx := context.Background()
	if err := s.redis.RDB.Set(ctx, userOnlineKey, 1, 60*time.Second).Err(); err != nil {
		slog.Warn("设置在线状态失败", "user_id", userID, "error", err)
	}

	// 启动读写 goroutine
	go client.writePump()
	go client.readPump()
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
