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

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    let currentEvent = ''
    for (const line of lines) {
      if (line.startsWith('event: ')) {
        currentEvent = line.slice(7).trim()
      } else if (line.startsWith('data: ')) {
        const jsonStr = line.slice(6)
        try {
          const data = JSON.parse(jsonStr)
          onEvent({ event: currentEvent as SSEEventType, data })
        } catch {
          // skip malformed JSON
        }
      }
    }
  }
}
