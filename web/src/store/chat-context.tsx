import { createContext, useContext, useReducer, useCallback, useEffect, useRef, type ReactNode } from 'react'
import * as chatApi from '../lib/chat'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  thinking: string
  createdAt: number
  isStreaming?: boolean
  hidden?: boolean // ask_user 回复不独立展示，整合到同一轮对话中
}

// CacheEntry 消息缓存条目，同时存储消息列表和 hasMore 状态
interface CacheEntry {
  messages: Message[]
  hasMore: boolean
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
  | { type: 'SET_CURRENT_SESSION'; sessionId: string | null; preserveMessages?: boolean }
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
  | { type: 'HIDE_LAST_ASSISTANT' }

function chatReducer(state: ChatState, action: ChatAction): ChatState {
  switch (action.type) {
    case 'SET_SESSIONS':
      return { ...state, sessions: action.sessions }
    case 'SET_CURRENT_SESSION':
      return { ...state, currentSessionId: action.sessionId, ...(action.preserveMessages ? {} : { messages: [], hasMoreMessages: false }) }
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
    case 'ADD_SESSION': {
      // 去重：如果会话已存在，不重复添加；同时清除该 ID 的虚拟会话
      const exists = state.sessions.some(s => s.id === action.session.id)
      const filtered = state.sessions.filter(s => s.id !== action.session.id)
      return {
        ...state,
        sessions: exists ? state.sessions : [action.session, ...filtered],
        virtualSession: action.session.id === state.virtualSession?.id ? null : state.virtualSession,
      }
    }
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
    case 'HIDE_LAST_ASSISTANT': {
      const msgs = [...state.messages]
      for (let i = msgs.length - 1; i >= 0; i--) {
        if (msgs[i].role === 'assistant') {
          msgs[i] = { ...msgs[i], hidden: true }
          break
        }
      }
      return { ...state, messages: msgs }
    }
  }
}

interface ChatContextValue extends ChatState {
  loadSessions: () => Promise<void>
  loadMessages: (sessionId: string) => Promise<void>
  loadMoreMessages: () => Promise<boolean>
  sendMessage: (message: string) => Promise<void>
  sendAskUserResponse: (answer: string) => Promise<void>
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

  const messagesCacheRef = useRef<Map<string, CacheEntry>>(new Map())
  const isStreamingRef = useRef(false)

  useEffect(() => {
    if (!state.isStreaming && state.currentSessionId && state.messages.length > 0) {
      // 不缓存含临时 ID 的流式消息，防止后续加载时返回不完整数据
      const hasTempMessages = state.messages.some(m => m.id.startsWith('temp-') || m.id.startsWith('stream-'))
      if (!hasTempMessages) {
        messagesCacheRef.current.set(state.currentSessionId, {
          messages: state.messages,
          hasMore: state.hasMoreMessages,
        })
      }
    }
  }, [state.isStreaming, state.currentSessionId, state.messages, state.hasMoreMessages])

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
      dispatch({ type: 'SET_CACHED_MESSAGES', messages: cached.messages, hasMore: cached.hasMore })
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
      messagesCacheRef.current.set(sessionId, { messages: formattedMessages, hasMore: has_more })
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
      const existing = messagesCacheRef.current.get(currentSessionId)
      const existingMessages = existing?.messages || []
      messagesCacheRef.current.set(currentSessionId, {
        messages: [...formattedMessages, ...existingMessages],
        hasMore: has_more,
      })
      
      return formattedMessages.length > 0
    } catch {
      return false
    }
  }, [state.currentSessionId, state.messages, state.hasMoreMessages])

  const sendMessage = useCallback(async (message: string) => {
    // 使用 ref 做并发保护，避免闭包过期问题
    if (isStreamingRef.current) return
    isStreamingRef.current = true

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
            // 首次创建会话时（虚拟会话或无当前会话），替换为服务端返回的真实会话
            if (evData.session_id && (!sessionId || isVirtual)) {
              sessionId = evData.session_id
              dispatch({ type: 'SET_CURRENT_SESSION', sessionId: evData.session_id, preserveMessages: true })
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
              dispatch({ type: 'UPDATE_SESSION_TITLE', sessionId: evData.session_id || sessionId, title: evData.title })
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
    isStreamingRef.current = false

    // 流式结束后清除缓存并重新从 API 拉取真实消息，防止临时消息污染缓存
    if (sessionId && !sessionId.startsWith('virtual_')) {
      messagesCacheRef.current.delete(sessionId)
      try {
        const { messages, has_more } = await chatApi.getMessages(sessionId)
        const formattedMessages = messages.reverse().map(m => ({
          id: m.id, role: m.role as 'user' | 'assistant', content: m.content, thinking: m.thinking, createdAt: m.created_at,
        }))
        messagesCacheRef.current.set(sessionId, { messages: formattedMessages, hasMore: has_more })
        dispatch({ type: 'SET_CACHED_MESSAGES', messages: formattedMessages, hasMore: has_more })
      } catch {
        // 拉取失败时不清除已有缓存，避免数据丢失
      }
    }

    loadSessions()
  }, [state.currentSessionId, loadSessions])

  // sendAskUserResponse ask_user 反问的回复，整合到同一轮对话中
  const sendAskUserResponse = useCallback(async (answer: string) => {
    if (isStreamingRef.current) return
    isStreamingRef.current = true

    const sessionId = state.currentSessionId
    if (!sessionId) return

    // 隐藏上一轮的系统消息 "[系统消息] 已向用户提问"
    dispatch({ type: 'HIDE_LAST_ASSISTANT' })

    dispatch({ type: 'SET_STREAMING', streaming: true })

    // 用户回答不独立展示（hidden）
    dispatch({
      type: 'APPEND_MESSAGE',
      message: { id: `temp-${Date.now()}`, role: 'user', content: answer, thinking: '', createdAt: Date.now(), hidden: true },
    })

    const msgId = `stream-${Date.now()}`
    dispatch({
      type: 'APPEND_MESSAGE',
      message: { id: msgId, role: 'assistant', content: '', thinking: '', createdAt: Date.now(), isStreaming: true },
    })

    try {
      await chatApi.createChatStream(sessionId, answer, (event) => {
        switch (event.event) {
          case 'thinking':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', thinking: event.data.content as string })
            break
          case 'content':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', content: event.data.content as string })
            break
          case 'done':
            break
          case 'ask_user':
            window.dispatchEvent(new CustomEvent('ask-user', { detail: event.data }))
            break
        }
      })
    } catch (err) {
      console.error('Chat stream error:', err)
    }

    dispatch({ type: 'SET_STREAMING', streaming: false })
    isStreamingRef.current = false

    // 拉取真实消息更新
    if (sessionId) {
      messagesCacheRef.current.delete(sessionId)
      try {
        const { messages, has_more } = await chatApi.getMessages(sessionId)
        const formattedMessages = messages.reverse().map(m => ({
          id: m.id, role: m.role as 'user' | 'assistant', content: m.content, thinking: m.thinking, createdAt: m.created_at,
        }))
        messagesCacheRef.current.set(sessionId, { messages: formattedMessages, hasMore: has_more })
        dispatch({ type: 'SET_CACHED_MESSAGES', messages: formattedMessages, hasMore: has_more })
      } catch { /* ignore */ }
    }

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
    <ChatContext.Provider value={{ ...state, loadSessions, loadMessages, loadMoreMessages, sendMessage, sendAskUserResponse, deleteSession, setCurrentSession, createNewSession }}>
      {children}
    </ChatContext.Provider>
  )
}

export function useChat() {
  const ctx = useContext(ChatContext)
  if (!ctx) throw new Error('useChat must be used within ChatProvider')
  return ctx
}
