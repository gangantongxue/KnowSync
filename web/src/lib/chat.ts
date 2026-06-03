import { request, getToken } from './client'

export interface ChatSession {
  id: string
  title: string
  created_at: number
  updated_at: number
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  thinking: string
  created_at: number
}

export async function listSessions(cursor?: number, limit = 20): Promise<{ sessions: ChatSession[]; has_more: boolean }> {
  const params = new URLSearchParams()
  if (cursor) params.set('cursor', String(cursor))
  params.set('limit', String(limit))
  const res = await request<{ sessions: ChatSession[]; has_more: boolean }>(`/ai/sessions?${params}`)
  return res.data
}

export async function getMessages(sessionId: string, cursor?: number, limit = 50): Promise<{ messages: ChatMessage[]; has_more: boolean }> {
  const params = new URLSearchParams()
  if (cursor) params.set('cursor', String(cursor))
  params.set('limit', String(limit))
  const res = await request<{ messages: ChatMessage[]; has_more: boolean }>(`/ai/sessions/${sessionId}/messages?${params}`)
  return res.data
}

export async function deleteSession(sessionId: string): Promise<void> {
  await request(`/ai/sessions/${sessionId}`, { method: 'DELETE' })
}

export type SSEEventType = 'thinking' | 'thinking_finished' | 'content' | 'ask_user' | 'done' | 'error'

export interface SSEEvent {
  event: SSEEventType
  data: Record<string, unknown>
}

export async function createChatStream(
  sessionId: string,
  message: string,
  onEvent: (event: SSEEvent) => void,
  signal?: AbortSignal
): Promise<void> {
  const token = getToken()
  const response = await fetch('/api/v1/ai/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ session_id: sessionId, message }),
    signal,
  })

  if (!response.ok) {
    const err = await response.json().catch(() => ({ message: 'Chat request failed' }))
    throw new Error(err.message || `HTTP ${response.status}`)
  }

  const reader = response.body?.getReader()
  if (!reader) throw new Error('No response body')

  const decoder = new TextDecoder()
  let buffer = ''

  // 流式事件延迟队列 — 按顺序串行发送，每个小内容块间隔 30ms，模拟流式效果
  let processing = Promise.resolve()
  const dispatchEvent = (event: SSEEvent) => {
    const isStreamChunk = event.event === 'content' || event.event === 'thinking' || event.event === 'thinking_finished'
    if (isStreamChunk) {
      const content = event.data.content as string || ''
      // 大块内容拆成小片（每片约 3 个字符），逐片延迟发布
      if (content.length > 3) {
        const chars = [...content] // 按 Unicode 字符拆分（支持中文）
        for (let i = 0; i < chars.length; i += 3) {
          const piece = chars.slice(i, i + 3).join('')
          const pieceEvent = { ...event, data: { ...event.data, content: piece } }
          processing = processing.then(() => new Promise<void>(resolve => {
            onEvent(pieceEvent)
            setTimeout(resolve, 25)
          }))
        }
        return
      }
      processing = processing.then(() => new Promise<void>(resolve => {
        onEvent(event)
        setTimeout(resolve, 25)
      }))
    } else {
      // done/ask_user/error 立即发送
      processing = processing.then(() => {
        onEvent(event)
      })
    }
  }

  // parseSSELines 解析 SSE 行并触发事件回调
  const parseSSELines = (lines: string[]) => {
    let currentEvent = ''
    for (const line of lines) {
      if (line.startsWith('event: ')) {
        currentEvent = line.slice(7).trim()
      } else if (line.startsWith('data: ')) {
        const jsonStr = line.slice(6)
        try {
          const data = JSON.parse(jsonStr)
          dispatchEvent({ event: currentEvent as SSEEventType, data })
        } catch {
          // skip malformed JSON
        }
      }
    }
  }

  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      // 流结束后 flush buffer 中剩余的完整事件
      parseSSELines(buffer.split('\n'))
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    // 最后一行可能不完整，保留在 buffer 中等下一轮拼接
    buffer = lines.pop() || ''
    parseSSELines(lines)
  }
}
