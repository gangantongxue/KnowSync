package push

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// sseSendBufferSize 发送缓冲区大小.
	sseSendBufferSize = 256
)

// sseHeartbeatInterval 心跳注释行发送间隔，同时用于刷新 Redis 在线状态 TTL.
// 使用变量而非常量，便于单元测试缩短间隔验证心跳逻辑.
var sseHeartbeatInterval = 30 * time.Second

// SSEClient SSE 长连接客户端，注册进 Hub 后通过发送通道接收服务层推送的事件.
type SSEClient struct {
	hub          *Hub
	w            http.ResponseWriter
	flusher      http.Flusher
	send         chan []byte
	userID       string
	onPing       func() // 心跳回调，用于刷新 Redis 在线状态 TTL
	onDisconnect func() // 断开连接回调，用于清理 Redis 在线状态
}

// NewSSEClient 创建 SSE 长连接客户端.
func NewSSEClient(hub *Hub, w http.ResponseWriter, userID string) *SSEClient {
	return &SSEClient{
		hub:    hub,
		w:      w,
		send:   make(chan []byte, sseSendBufferSize),
		userID: userID,
	}
}

// UserID 实现 PushClient 接口，返回客户端所属用户 ID.
func (c *SSEClient) UserID() string {
	return c.userID
}

// SendCh 实现 PushClient 接口，返回接收消息的通道.
func (c *SSEClient) SendCh() chan []byte {
	return c.send
}

// Serve 阻塞运行写入循环：从发送通道取出事件写入响应（data: 行），
// 定时发送心跳注释行保持连接存活，请求上下文取消（客户端断开）时退出.
// 事件 payload 复用服务层已有的 {"type": "...", "data": {...}} JSON 结构，
// 以单条默认事件（message）下发，前端解析逻辑与原先 WebSocket 完全一致.
func (c *SSEClient) Serve(ctx context.Context) {
	ticker := time.NewTicker(sseHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// hub 已注销连接并关闭发送通道
				return
			}
			if _, err := fmt.Fprintf(c.w, "data: %s\n\n", message); err != nil {
				return
			}
			c.flusher.Flush()

		case <-ticker.C:
			// 心跳注释行（SSE 规范中以冒号开头的行会被客户端忽略）
			if _, err := io.WriteString(c.w, ": ping\n\n"); err != nil {
				return
			}
			c.flusher.Flush()
			if c.onPing != nil {
				c.onPing()
			}

		case <-ctx.Done():
			return
		}
	}
}
