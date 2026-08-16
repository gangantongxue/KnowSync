// sse-client 全局 SSE 长连接客户端。
// 页面打开即建立连接（无论是否在聊天页），新消息、好友申请等事件通过 SSE 实时推送，
// 发送消息仍走 HTTP 请求。EventSource 自带指数退避自动重连，无需手动实现。
export interface SseEvent {
  type: 'new_message' | 'message_recalled' | 'friend_request' | 'friend_accepted'
  data: Record<string, unknown>
}

type SseEventHandler = (event: SseEvent) => void

// SSE_URL 与后端 /sse 端点同源（经 Caddy 反向代理到 chat-server）
const SSE_URL = `${window.location.protocol}//${window.location.host}/sse`

interface SseClient {
  connect: (token: string, onEvent: SseEventHandler) => void
  disconnect: () => void
  isConnected: () => boolean
}

function createSseClient(): SseClient {
  let es: EventSource | null = null
  let currentToken: string | null = null
  let currentHandler: SseEventHandler | null = null
  let explicitlyDisconnected = false

  function doConnect(): void {
    if (!currentToken || !currentHandler) return

    const url = `${SSE_URL}?token=${encodeURIComponent(currentToken)}`
    es = new EventSource(url)

    es.onopen = () => {
      console.log('[SSE] Connected')
    }

    // 后端以默认事件（message）下发 data: {json}，type 字段在 JSON 内，
    // 解析逻辑与原 WebSocket 完全一致
    es.onmessage = (event: MessageEvent) => {
      try {
        const parsed = JSON.parse(event.data as string) as { type: string; data: Record<string, unknown> }
        currentHandler?.({ type: parsed.type as SseEvent['type'], data: parsed.data })
      } catch {
        console.error('[SSE] Failed to parse message:', event.data)
      }
    }

    es.onerror = () => {
      // 主动断开时不重连；其余情况由 EventSource 自动重连
      if (explicitlyDisconnected && es) {
        es.close()
        es = null
      }
    }
  }

  function connect(token: string, handler: SseEventHandler): void {
    explicitlyDisconnected = false
    currentToken = token
    currentHandler = handler
    // 若已有旧连接（如 token 刷新后重连），先关闭再新建
    if (es) {
      es.close()
      es = null
    }
    doConnect()
  }

  function disconnect(): void {
    explicitlyDisconnected = true
    if (es) {
      es.close()
      es = null
    }
    currentToken = null
    currentHandler = null
  }

  function isConnected(): boolean {
    return es !== null && es.readyState === EventSource.OPEN
  }

  return { connect, disconnect, isConnected }
}

const sseClient = createSseClient()
export default sseClient
