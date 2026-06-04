export interface WsEvent {
  type: 'new_message' | 'message_recalled' | 'friend_request' | 'friend_accepted'
  data: Record<string, unknown>
}

type WsEventHandler = (event: WsEvent) => void

const WS_URL = 'ws://localhost:50055/ws'
const MAX_RECONNECT_DELAY = 30000
const INITIAL_RECONNECT_DELAY = 1000

interface WsClient {
  connect: (token: string, onEvent: WsEventHandler) => void
  disconnect: () => void
  isConnected: () => boolean
}

function createWsClient(): WsClient {
  let ws: WebSocket | null = null
  let currentToken: string | null = null
  let currentHandler: WsEventHandler | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempt = 0
  let explicitlyDisconnected = false

  function getReconnectDelay(): number {
    const delay = INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttempt)
    return Math.min(delay, MAX_RECONNECT_DELAY)
  }

  function scheduleReconnect(): void {
    if (explicitlyDisconnected || !currentToken) return
    const delay = getReconnectDelay()
    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${reconnectAttempt + 1})`)
    reconnectTimer = setTimeout(() => {
      reconnectAttempt++
      doConnect()
    }, delay)
  }

  function doConnect(): void {
    if (!currentToken || !currentHandler) return

    if (ws) {
      ws.onopen = null
      ws.onmessage = null
      ws.onclose = null
      ws.onerror = null
      ws.close()
    }

    const url = `${WS_URL}?token=${currentToken}`
    ws = new WebSocket(url)

    ws.onopen = () => {
      console.log('[WS] Connected')
      reconnectAttempt = 0
    }

    ws.onmessage = (event: MessageEvent) => {
      try {
        const parsed = JSON.parse(event.data as string) as { type: string; data: Record<string, unknown> }
        currentHandler?.({ type: parsed.type as WsEvent['type'], data: parsed.data })
      } catch {
        console.error('[WS] Failed to parse message:', event.data)
      }
    }

    ws.onclose = () => {
      console.log('[WS] Disconnected')
      ws = null
      scheduleReconnect()
    }

    ws.onerror = () => {
      console.error('[WS] Error occurred')
    }
  }

  function connect(token: string, handler: WsEventHandler): void {
    explicitlyDisconnected = false
    currentToken = token
    currentHandler = handler
    reconnectAttempt = 0
    doConnect()
  }

  function disconnect(): void {
    explicitlyDisconnected = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.onopen = null
      ws.onmessage = null
      ws.onclose = null
      ws.onerror = null
      ws.close()
      ws = null
    }
    currentToken = null
    currentHandler = null
    reconnectAttempt = 0
  }

  function isConnected(): boolean {
    return ws !== null && ws.readyState === WebSocket.OPEN
  }

  return { connect, disconnect, isConnected }
}

const wsClient = createWsClient()
export default wsClient
