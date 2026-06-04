package ws

import (
	"log/slog"
	"sync"
)

// Hub 连接管理器，管理所有 WebSocket 连接.
type Hub struct {
	clients    map[string]map[*Client]bool // userID → 连接集合
	register   chan *Client                // 注册连接的通道
	unregister chan *Client                // 注销连接的通道
	mu         sync.RWMutex
}

// NewHub 创建连接管理器并启动事件循环.
func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
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
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			count := len(h.clients[client.userID])
			h.mu.Unlock()
			slog.Info("WebSocket 客户端已注册", "user_id", client.userID, "connections", count)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.userID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.userID)
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
		case client.send <- msg:
		default:
			// 发送缓冲区已满，丢弃消息
			slog.Warn("WebSocket 发送缓冲区已满，丢弃消息", "user_id", userID)
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
