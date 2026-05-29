import { createContext, useContext, useReducer, useCallback, type ReactNode } from 'react'
import * as chatApi from '../lib/chat'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  thinking: string
  createdAt: number
  isStreaming?: boolean
}

interface ChatState {
  sessions: chatApi.ChatSession[]
  currentSessionId: string | null
  messages: Message[]
  hasMoreMessages: boolean
  isStreaming: boolean
}

type ChatAction =
  | { type: 'SET_SESSIONS'; sessions: chatApi.ChatSession[] }
  | { type: 'SET_CURRENT_SESSION'; sessionId: string | null }
  | { type: 'SET_MESSAGES'; messages: chatApi.ChatMessage[]; hasMore: boolean }
  | { type: 'APPEND_MESSAGE'; message: Message }
  | { type: 'UPDATE_LAST_MESSAGE'; content: string }
  | { type: 'SET_STREAMING'; streaming: boolean }
  | { type: 'ADD_SESSION'; session: chatApi.ChatSession }

function chatReducer(state: ChatState, action: ChatAction): ChatState {
  switch (action.type) {
    case 'SET_SESSIONS':
      return { ...state, sessions: action.sessions }
    case 'SET_CURRENT_SESSION':
      return { ...state, currentSessionId: action.sessionId, messages: [], hasMoreMessages: false }
    case 'SET_MESSAGES':
      return { ...state, messages: action.messages.map(m => ({
        id: m.id, role: m.role, content: m.content, thinking: m.thinking, createdAt: m.created_at,
      })), hasMoreMessages: action.hasMore }
    case 'APPEND_MESSAGE':
      return { ...state, messages: [...state.messages, action.message] }
    case 'UPDATE_LAST_MESSAGE': {
      const msgs = [...state.messages]
      const last = msgs[msgs.length - 1]
      if (last && last.isStreaming) {
        msgs[msgs.length - 1] = { ...last, content: last.content + action.content }
      }
      return { ...state, messages: msgs }
    }
    case 'SET_STREAMING':
      return { ...state, isStreaming: action.streaming }
    case 'ADD_SESSION':
      return { ...state, sessions: [action.session, ...state.sessions] }
  }
}

interface ChatContextValue extends ChatState {
  loadSessions: () => Promise<void>
  loadMessages: (sessionId: string) => Promise<void>
  sendMessage: (message: string) => Promise<void>
  deleteSession: (sessionId: string) => Promise<void>
  setCurrentSession: (sessionId: string | null) => void
}

const ChatContext = createContext<ChatContextValue | null>(null)

export function ChatProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(chatReducer, {
    sessions: [],
    currentSessionId: null,
    messages: [],
    hasMoreMessages: false,
    isStreaming: false,
  })

  const loadSessions = useCallback(async () => {
    try {
      const { sessions } = await chatApi.listSessions()
      dispatch({ type: 'SET_SESSIONS', sessions })
    } catch {
      // ignore
    }
  }, [])

  const loadMessages = useCallback(async (sessionId: string) => {
    try {
      dispatch({ type: 'SET_CURRENT_SESSION', sessionId })
      const { messages, has_more } = await chatApi.getMessages(sessionId)
      dispatch({ type: 'SET_MESSAGES', messages: messages.reverse(), hasMore: has_more })
    } catch {
      // ignore
    }
  }, [])

  const sendMessage = useCallback(async (message: string) => {
    const sessionId = state.currentSessionId || ''

    dispatch({ type: 'SET_STREAMING', streaming: true })

    // 添加用户消息
    dispatch({
      type: 'APPEND_MESSAGE',
      message: { id: `temp-${Date.now()}`, role: 'user', content: message, thinking: '', createdAt: Date.now() },
    })

    // 添加占位 assistant 消息
    const msgId = `stream-${Date.now()}`
    dispatch({
      type: 'APPEND_MESSAGE',
      message: { id: msgId, role: 'assistant', content: '', thinking: '', createdAt: Date.now(), isStreaming: true },
    })

    try {
      await chatApi.createChatStream(sessionId, message, (event) => {
        switch (event.event) {
          case 'thinking':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', content: event.data.content as string })
            break
          case 'content':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', content: event.data.content as string })
            break
          case 'done': {
            const evData = event.data as { session_id: string; title?: string; title_updated?: boolean }
            if (evData.title_updated) {
              // 更新会话标题
            }
            break
          }
          case 'ask_user':
            // 弹窗由 UI 组件处理，这里触发事件
            window.dispatchEvent(new CustomEvent('ask-user', { detail: event.data }))
            break
        }
      })
    } catch (err) {
      console.error('Chat stream error:', err)
    }

    dispatch({ type: 'SET_STREAMING', streaming: false })
    loadSessions()
  }, [state.currentSessionId, loadSessions])

  const deleteSession = useCallback(async (sessionId: string) => {
    await chatApi.deleteSession(sessionId)
    loadSessions()
    if (state.currentSessionId === sessionId) {
      dispatch({ type: 'SET_CURRENT_SESSION', sessionId: null })
    }
  }, [state.currentSessionId, loadSessions])

  const setCurrentSession = useCallback((sessionId: string | null) => {
    dispatch({ type: 'SET_CURRENT_SESSION', sessionId })
  }, [])

  return (
    <ChatContext.Provider value={{ ...state, loadSessions, loadMessages, sendMessage, deleteSession, setCurrentSession }}>
      {children}
    </ChatContext.Provider>
  )
}

export function useChat() {
  const ctx = useContext(ChatContext)
  if (!ctx) throw new Error('useChat must be used within ChatProvider')
  return ctx
}
