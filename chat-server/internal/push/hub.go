// Package push 提供消息推送连接管理，当前实现为 SSE（Server-Sent Events）长连接.
// 服务层通过 Hub 将事件推送给指定用户的所有在线连接，与具体传输协议解耦.
package push

import (
	"log/slog"
	"sync"
)

// PushClient 推送客户端接口，Hub 仅依赖用户 ID 与发送通道，
// 便于在不改动服务层的前提下接入不同的推送通道（SSE、WebSocket 等）.
type PushClient interface {
	// UserID 返回客户端所属用户 ID.
	UserID() string

	// SendCh 返回客户端接收消息的通道，Hub 注销连接时会关闭该通道.
	SendCh() chan []byte
}

// Hub 连接管理器，管理所有推送连接.
type Hub struct {
	clients    map[string]map[PushClient]bool // userID → 连接集合
	register   chan PushClient                // 注册连接的通道
	unregister chan PushClient                // 注销连接的通道
	mu         sync.RWMutex
}

// NewHub 创建连接管理器并启动事件循环.
func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[string]map[PushClient]bool),
		register:   make(chan PushClient),
		unregister: make(chan PushClient),
	}
	go h.run()
	return h
}

// run 事件循环，处理注册和注销请求.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID()] == nil {
				h.clients[client.UserID()] = make(map[PushClient]bool)
			}
			h.clients[client.UserID()][client] = true
			count := len(h.clients[client.UserID()])
			h.mu.Unlock()
			slog.Info("推送客户端已注册", "user_id", client.UserID(), "connections", count)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID()]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.SendCh())
					if len(clients) == 0 {
						delete(h.clients, client.UserID())
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// SendToUser 向指定用户的所有连接发送消息.
func (h *Hub) SendToUser(userID string, msg []byte) {
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.SendCh() <- msg:
		default:
			// 发送缓冲区已满，丢弃消息
			slog.Warn("推送发送缓冲区已满，丢弃消息", "user_id", userID)
		}
	}
}

// GetOnlineUsers 获取用户在线状态，返回在线的 userID 子集.
func (h *Hub) GetOnlineUsers(userIDs []string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var online []string
	for _, userID := range userIDs {
		if clients, ok := h.clients[userID]; ok && len(clients) > 0 {
			online = append(online, userID)
		}
	}
	return online
}
