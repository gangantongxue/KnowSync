import { createContext, useContext, useReducer, useCallback, useEffect, useRef, type ReactNode } from 'react'
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
  virtualSession: chatApi.ChatSession | null
  isLoadingSessions: boolean
  isLoadingMessages: boolean
}

type ChatAction =
  | { type: 'SET_SESSIONS'; sessions: chatApi.ChatSession[] }
  | { type: 'SET_CURRENT_SESSION'; sessionId: string | null }
  | { type: 'SET_MESSAGES'; messages: chatApi.ChatMessage[]; hasMore: boolean }
  | { type: 'SET_CACHED_MESSAGES'; messages: Message[]; hasMore: boolean }
  | { type: 'LOAD_MORE_MESSAGES'; messages: Message[]; hasMore: boolean }
  | { type: 'APPEND_MESSAGE'; message: Message }
  | { type: 'UPDATE_LAST_MESSAGE'; content?: string; thinking?: string }
  | { type: 'SET_STREAMING'; streaming: boolean }
  | { type: 'ADD_SESSION'; session: chatApi.ChatSession }
  | { type: 'SET_VIRTUAL_SESSION'; session: chatApi.ChatSession | null }
  | { type: 'UPDATE_SESSION_TITLE'; sessionId: string; title: string }
  | { type: 'SET_LOADING_SESSIONS'; loading: boolean }
  | { type: 'SET_LOADING_MESSAGES'; loading: boolean }

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
    case 'SET_CACHED_MESSAGES':
      return { ...state, messages: action.messages, hasMoreMessages: action.hasMore }
    case 'LOAD_MORE_MESSAGES':
      return { ...state, messages: [...action.messages, ...state.messages], hasMoreMessages: action.hasMore }
    case 'APPEND_MESSAGE':
      return { ...state, messages: [...state.messages, action.message] }
    case 'UPDATE_LAST_MESSAGE': {
      const msgs = [...state.messages]
      const last = msgs[msgs.length - 1]
      if (last && last.isStreaming) {
        msgs[msgs.length - 1] = {
          ...last,
          content: action.content !== undefined ? last.content + action.content : last.content,
          thinking: action.thinking !== undefined ? last.thinking + action.thinking : last.thinking,
        }
      }
      return { ...state, messages: msgs }
    }
    case 'SET_STREAMING':
      return { ...state, isStreaming: action.streaming }
    case 'ADD_SESSION':
      return { ...state, sessions: [action.session, ...state.sessions] }
    case 'SET_VIRTUAL_SESSION':
      return { ...state, virtualSession: action.session }
    case 'UPDATE_SESSION_TITLE':
      return {
        ...state,
        sessions: state.sessions.map(s =>
          s.id === action.sessionId ? { ...s, title: action.title } : s
        ),
      }
    case 'SET_LOADING_SESSIONS':
      return { ...state, isLoadingSessions: action.loading }
    case 'SET_LOADING_MESSAGES':
      return { ...state, isLoadingMessages: action.loading }
  }
}

interface ChatContextValue extends ChatState {
  loadSessions: () => Promise<void>
  loadMessages: (sessionId: string) => Promise<void>
  loadMoreMessages: () => Promise<boolean>
  sendMessage: (message: string) => Promise<void>
  deleteSession: (sessionId: string) => Promise<void>
  setCurrentSession: (sessionId: string | null) => void
  createNewSession: () => void
}

const ChatContext = createContext<ChatContextValue | null>(null)

export function ChatProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(chatReducer, {
    sessions: [],
    currentSessionId: null,
    messages: [],
    hasMoreMessages: false,
    isStreaming: false,
    virtualSession: null,
    isLoadingSessions: false,
    isLoadingMessages: false,
  })

  const messagesCacheRef = useRef<Map<string, Message[]>>(new Map())

  useEffect(() => {
    if (!state.isStreaming && state.currentSessionId && state.messages.length > 0) {
      messagesCacheRef.current.set(state.currentSessionId, state.messages)
    }
  }, [state.isStreaming, state.currentSessionId, state.messages])

  const loadSessions = useCallback(async () => {
    try {
      dispatch({ type: 'SET_LOADING_SESSIONS', loading: true })
      const { sessions } = await chatApi.listSessions()
      dispatch({ type: 'SET_SESSIONS', sessions })
    } catch {
      // ignore
    } finally {
      dispatch({ type: 'SET_LOADING_SESSIONS', loading: false })
    }
  }, [])

  const createNewSession = useCallback(() => {
    const virtualSession: chatApi.ChatSession = {
      id: `virtual_${Date.now()}`,
      title: '新对话',
      created_at: Math.floor(Date.now() / 1000),
      updated_at: Math.floor(Date.now() / 1000),
    }
    dispatch({ type: 'SET_VIRTUAL_SESSION', session: virtualSession })
    dispatch({ type: 'SET_CURRENT_SESSION', sessionId: virtualSession.id })
    // Removed redundant SET_MESSAGES - SET_CURRENT_SESSION already clears messages
  }, [])

  const loadMessages = useCallback(async (sessionId: string) => {
    // Check cache first
    const cached = messagesCacheRef.current.get(sessionId)
    if (cached) {
      dispatch({ type: 'SET_CURRENT_SESSION', sessionId })
      dispatch({ type: 'SET_CACHED_MESSAGES', messages: cached, hasMore: false })
      return
    }

    try {
      dispatch({ type: 'SET_CURRENT_SESSION', sessionId })
      dispatch({ type: 'SET_LOADING_MESSAGES', loading: true })
      const { messages, has_more } = await chatApi.getMessages(sessionId)
      const formattedMessages = messages.reverse().map(m => ({
        id: m.id, role: m.role, content: m.content, thinking: m.thinking, createdAt: m.created_at,
      }))
      // Cache the messages
      messagesCacheRef.current.set(sessionId, formattedMessages)
      dispatch({ type: 'SET_CACHED_MESSAGES', messages: formattedMessages, hasMore: has_more })
    } catch {
      // ignore
    } finally {
      dispatch({ type: 'SET_LOADING_MESSAGES', loading: false })
    }
  }, [])

  const loadMoreMessages = useCallback(async (): Promise<boolean> => {
    const { currentSessionId, messages, hasMoreMessages } = state
    if (!currentSessionId || !hasMoreMessages) {
      return false
    }

    // Use the oldest message's createdAt as cursor (backend expects timestamp, not offset)
    const oldestMessage = messages[0]
    const cursor = oldestMessage?.createdAt || 0

    try {
      const { messages: newMessages, has_more } = await chatApi.getMessages(
        currentSessionId,
        cursor
      )
      const formattedMessages = newMessages.reverse().map(m => ({
        id: m.id, role: m.role, content: m.content, thinking: m.thinking, createdAt: m.created_at,
      }))
      
      dispatch({ type: 'LOAD_MORE_MESSAGES', messages: formattedMessages, hasMore: has_more })
      
      // Update cache
      const cached = messagesCacheRef.current.get(currentSessionId) || []
      messagesCacheRef.current.set(currentSessionId, [...formattedMessages, ...cached])
      
      return formattedMessages.length > 0
    } catch {
      return false
    }
  }, [state.currentSessionId, state.messages, state.hasMoreMessages])

  const sendMessage = useCallback(async (message: string) => {
    // Guard against concurrent streams
    if (state.isStreaming) return
    
    let sessionId = state.currentSessionId || ''
    const isVirtual = sessionId.startsWith('virtual_')
    
    // If virtual session, send without session_id to create new one
    const sessionIdToSend = isVirtual ? '' : sessionId

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
      await chatApi.createChatStream(sessionIdToSend, message, (event) => {
        switch (event.event) {
          case 'thinking':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', thinking: event.data.content as string })
            break
          case 'content':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', content: event.data.content as string })
            break
          case 'done': {
            const evData = event.data as { session_id: string; title?: string; title_updated?: boolean }
            // Handle new session creation (virtual -> real)
            if (isVirtual && evData.session_id) {
              sessionId = evData.session_id
              dispatch({ type: 'SET_CURRENT_SESSION', sessionId: evData.session_id })
              dispatch({ type: 'SET_VIRTUAL_SESSION', session: null })
              // Add new session to list
              dispatch({
                type: 'ADD_SESSION',
                session: {
                  id: evData.session_id,
                  title: evData.title || '新对话',
                  created_at: Math.floor(Date.now() / 1000),
                  updated_at: Math.floor(Date.now() / 1000),
                },
              })
            }
            // Handle title update
            if (evData.title_updated && evData.title) {
              dispatch({ type: 'UPDATE_SESSION_TITLE', sessionId: evData.session_id, title: evData.title })
            }
            break
          }
          case 'ask_user':
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
    messagesCacheRef.current.delete(sessionId)  // Clear cache
    loadSessions()
    if (state.currentSessionId === sessionId) {
      dispatch({ type: 'SET_CURRENT_SESSION', sessionId: null })
    }
  }, [state.currentSessionId, loadSessions])

  const setCurrentSession = useCallback((sessionId: string | null) => {
    dispatch({ type: 'SET_CURRENT_SESSION', sessionId })
    // Clear virtual session when switching to a real session
    if (sessionId && !sessionId.startsWith('virtual_')) {
      dispatch({ type: 'SET_VIRTUAL_SESSION', session: null })
    }
  }, [])

  return (
    <ChatContext.Provider value={{ ...state, loadSessions, loadMessages, loadMoreMessages, sendMessage, deleteSession, setCurrentSession, createNewSession }}>
      {children}
    </ChatContext.Provider>
  )
}

export function useChat() {
  const ctx = useContext(ChatContext)
  if (!ctx) throw new Error('useChat must be used within ChatProvider')
  return ctx
}
