package ws

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// 读取等待时间
	readWait = 60 * time.Second

	// 心跳发送间隔
	pingPeriod = 30 * time.Second

	// 发送缓冲区大小
	sendBufferSize = 256
)

// Client WebSocket 客户端连接
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan []byte
	userID       string
	onPing       func() // 心跳回调，用于刷新 Redis 在线状态 TTL
	onDisconnect func() // 断开连接回调，用于清理 Redis 在线状态
}

// NewClient 创建 WebSocket 客户端连接
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
		userID: userID,
	}
}

// readPump 读取消息循环，处理心跳和消息接收
func (c *Client) readPump() {
	defer func() {
		if c.onDisconnect != nil {
			c.onDisconnect()
		}
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(readWait))

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(readWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				slog.Warn("WebSocket 读取异常", "user_id", c.userID, "error", err)
			}
			break
		}
	}
}

// writePump 写入消息循环，发送心跳和消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub 已关闭发送通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			if c.onPing != nil {
				c.onPing()
			}
		}
	}
}
