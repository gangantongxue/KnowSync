package push

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// newTestSSEServer 创建测试用 SSE 服务器：连接建立后注册进 Hub 并进入写入循环.
// 与生产 handler 一致，先写出连接确认行，使客户端请求立即返回.
func newTestSSEServer(t *testing.T, hub *Hub, userID string, onPing func()) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "不支持流式响应", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
			return
		}
		flusher.Flush()
		client := NewSSEClient(hub, w, userID)
		client.flusher = flusher
		client.onPing = onPing
		hub.register <- client
		defer func() { hub.unregister <- client }()
		client.Serve(r.Context())
	}))
}

// waitOnline 轮询等待用户在 Hub 中上线，避免注册与请求之间的竞态.
func waitOnline(t *testing.T, hub *Hub, userID string, expectOnline bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		online := len(hub.GetOnlineUsers([]string{userID})) > 0
		if online == expectOnline {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("等待用户在线状态超时: user=%s expectOnline=%v", userID, expectOnline)
}

// readSSELines 后台读取响应流并按行收集到通道.
func readSSELines(resp *http.Response) <-chan string {
	lines := make(chan string, 64)
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			lines <- line
		}
	}()
	return lines
}

// TestHubSSEDelivery 验证 SSE 连接注册后能收到 SendToUser 推送的事件，断开后正确注销.
func TestHubSSEDelivery(t *testing.T) {
	hub := NewHub()
	srv := newTestSSEServer(t, hub, "user-1", nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("发起 SSE 请求失败: %v", err)
	}
	defer resp.Body.Close()

	waitOnline(t, hub, "user-1", true)

	lines := readSSELines(resp)

	// 推送事件并验证收到 data 行
	msg := []byte(`{"type":"new_message","data":{"id":"m1"}}`)
	hub.SendToUser("user-1", msg)

	timeout := time.After(3 * time.Second)
	for {
		select {
		case line := <-lines:
			if line == "data: "+string(msg)+"\n" {
				t.Logf("收到事件行: %q", line)
				goto delivered
			}
		case <-timeout:
			t.Fatalf("超时未收到推送事件")
		}
	}
delivered:

	// 客户端断开后应自动注销
	resp.Body.Close()
	waitOnline(t, hub, "user-1", false)
}

// TestSSEClientHeartbeat 验证心跳注释行会定时写出且回调 onPing.
func TestSSEClientHeartbeat(t *testing.T) {
	old := sseHeartbeatInterval
	sseHeartbeatInterval = 50 * time.Millisecond
	defer func() { sseHeartbeatInterval = old }()

	hub := NewHub()
	var mu sync.Mutex
	pingCalls := 0

	srv := newTestSSEServer(t, hub, "user-2", func() {
		mu.Lock()
		pingCalls++
		mu.Unlock()
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("发起 SSE 请求失败: %v", err)
	}
	defer resp.Body.Close()

	lines := readSSELines(resp)

	// 等待心跳注释行
	timeout := time.After(3 * time.Second)
	for {
		select {
		case line := <-lines:
			if line == ": ping\n" {
				t.Logf("收到心跳行: %q", line)
				goto pinged
			}
		case <-timeout:
			t.Fatal("未在超时时间内收到心跳注释行")
		}
	}
pinged:

	mu.Lock()
	calls := pingCalls
	mu.Unlock()
	if calls == 0 {
		t.Fatal("心跳回调 onPing 未被触发")
	}
}

// TestHubMultipleConnections 验证同一用户的多个连接都能收到推送.
func TestHubMultipleConnections(t *testing.T) {
	hub := NewHub()
	srv1 := newTestSSEServer(t, hub, "user-3", nil)
	srv2 := newTestSSEServer(t, hub, "user-3", nil)
	defer srv1.Close()
	defer srv2.Close()

	resp1, err := http.Get(srv1.URL)
	if err != nil {
		t.Fatalf("发起第一个 SSE 请求失败: %v", err)
	}
	defer resp1.Body.Close()
	resp2, err := http.Get(srv2.URL)
	if err != nil {
		t.Fatalf("发起第二个 SSE 请求失败: %v", err)
	}
	defer resp2.Body.Close()

	waitOnline(t, hub, "user-3", true)

	lines1 := readSSELines(resp1)
	lines2 := readSSELines(resp2)

	msg := []byte(`{"type":"new_message","data":{"id":"m2"}}`)
	hub.SendToUser("user-3", msg)

	timeout := time.After(3 * time.Second)
	received := map[int]bool{}
	for len(received) < 2 {
		select {
		case line := <-lines1:
			if line == "data: "+string(msg)+"\n" {
				received[1] = true
			}
		case line := <-lines2:
			if line == "data: "+string(msg)+"\n" {
				received[2] = true
			}
		case <-timeout:
			t.Fatalf("超时未收到全部推送事件, 已收到连接: %v", received)
		}
	}

	// 关闭一个连接后，另一个连接仍保持在线
	resp1.Body.Close()
	waitOnline(t, hub, "user-3", true)
}
