import { createContext, useContext, useReducer, useCallback, type ReactNode } from 'react'
import { conversationApi, messageApi, friendApi, groupApi } from '../lib/chat-api'
import type {
  ConversationInfo, Message, FriendRequest, Friend, SearchUserInfo,
} from '../lib/chat-api'

// ===== Helper =====
function getKey(convType: string, convId: string): string {
  return `${convType}_${convId}`
}

// ===== State =====
interface ChatState {
  conversations: ConversationInfo[]
  currentConversationType: string | null
  currentConversationId: string | null
  messages: Record<string, Message[]>
  hasMore: Record<string, boolean>
  loadingMessages: boolean
  sendingMessage: boolean
  friendRequests: FriendRequest[]
  friends: Friend[]
  searchResults: SearchUserInfo[]
  searchQuery: string
  wsConnected: boolean
  unreadCounts: Record<string, number>
  totalUnread: number
  error: string | null
}

// ===== Actions =====
type ChatAction =
  | { type: 'SET_CONVERSATIONS'; conversations: ConversationInfo[] }
  | { type: 'SET_CURRENT_CONVERSATION'; conversationType: string | null; conversationId: string | null }
  | { type: 'SET_MESSAGES'; key: string; messages: Message[]; hasMore: boolean }
  | { type: 'APPEND_MESSAGE'; key: string; message: Message }
  | { type: 'PREPEND_MESSAGES'; key: string; messages: Message[]; hasMore: boolean }
  | { type: 'UPDATE_MESSAGE'; key: string; messageId: string; updates: Partial<Message> }
  | { type: 'SET_LOADING_MESSAGES'; loading: boolean }
  | { type: 'SET_SENDING_MESSAGE'; sending: boolean }
  | { type: 'SET_FRIEND_REQUESTS'; requests: FriendRequest[] }
  | { type: 'SET_FRIENDS'; friends: Friend[] }
  | { type: 'SET_SEARCH_RESULTS'; users: SearchUserInfo[]; query: string }
  | { type: 'SET_WS_CONNECTED'; connected: boolean }
  | { type: 'SET_UNREAD_COUNTS'; counts: Record<string, number> }
  | { type: 'SET_TOTAL_UNREAD'; total: number }
  | { type: 'SET_ERROR'; error: string }
  | { type: 'CLEAR_ERROR' }

// ===== Reducer =====
function chatReducer(state: ChatState, action: ChatAction): ChatState {
  switch (action.type) {
    case 'SET_CONVERSATIONS':
      return { ...state, conversations: action.conversations }
    case 'SET_CURRENT_CONVERSATION':
      return { ...state, currentConversationType: action.conversationType, currentConversationId: action.conversationId }
    case 'SET_MESSAGES':
      return {
        ...state,
        messages: { ...state.messages, [action.key]: action.messages },
        hasMore: { ...state.hasMore, [action.key]: action.hasMore },
      }
    case 'APPEND_MESSAGE': {
      const existing = state.messages[action.key] || []
      return { ...state, messages: { ...state.messages, [action.key]: [...existing, action.message] } }
    }
    case 'PREPEND_MESSAGES':
      return {
        ...state,
        messages: { ...state.messages, [action.key]: [...action.messages, ...(state.messages[action.key] || [])] },
        hasMore: { ...state.hasMore, [action.key]: action.hasMore },
      }
    case 'UPDATE_MESSAGE': {
      const msgs = state.messages[action.key]
      if (!msgs) return state
      return {
        ...state,
        messages: {
          ...state.messages,
          [action.key]: msgs.map(m => m.id === action.messageId ? { ...m, ...action.updates } : m),
        },
      }
    }
    case 'SET_LOADING_MESSAGES':
      return { ...state, loadingMessages: action.loading }
    case 'SET_SENDING_MESSAGE':
      return { ...state, sendingMessage: action.sending }
    case 'SET_FRIEND_REQUESTS':
      return { ...state, friendRequests: action.requests }
    case 'SET_FRIENDS':
      return { ...state, friends: action.friends }
    case 'SET_SEARCH_RESULTS':
      return { ...state, searchResults: action.users, searchQuery: action.query }
    case 'SET_WS_CONNECTED':
      return { ...state, wsConnected: action.connected }
    case 'SET_UNREAD_COUNTS':
      return { ...state, unreadCounts: action.counts }
    case 'SET_TOTAL_UNREAD':
      return { ...state, totalUnread: action.total }
    case 'SET_ERROR':
      return { ...state, error: action.error }
    case 'CLEAR_ERROR':
      return { ...state, error: null }
  }
}

// ===== Context Value =====
interface ChatContextValue extends ChatState {
  loadConversations: () => Promise<void>
  loadMessages: (convType: string, convId: string) => Promise<void>
  loadMoreMessages: (convType: string, convId: string) => Promise<boolean>
  sendMessage: (convType: string, convId: string, content: string, contentType: string) => Promise<void>
  recallMessage: (messageId: string) => Promise<void>
  sendFriendRequest: (receiverId: string, remark: string) => Promise<void>
  acceptFriendRequest: (requestId: string) => Promise<void>
  rejectFriendRequest: (requestId: string) => Promise<void>
  loadFriendRequests: () => Promise<void>
  loadFriends: () => Promise<void>
  searchUsers: (query: string) => Promise<void>
  createGroup: (data: { name: string; avatar?: string; member_ids?: string[] }) => Promise<void>
  addMembers: (groupId: string, memberIds: string[]) => Promise<void>
  removeMember: (groupId: string, userId: string) => Promise<void>
  loadConversation: (convType: string, convId: string) => void
  markConversationRead: (convType: string, convId: string) => Promise<void>
  togglePin: (convType: string, convId: string) => Promise<void>
  handleWsMessage: (event: { type: string; data: any }) => void
  setWsConnected: (connected: boolean) => void
}

const ChatContext = createContext<ChatContextValue | null>(null)

export function MessageProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(chatReducer, {
    conversations: [],
    currentConversationType: null,
    currentConversationId: null,
    messages: {},
    hasMore: {},
    loadingMessages: false,
    sendingMessage: false,
    friendRequests: [],
    friends: [],
    searchResults: [],
    searchQuery: '',
    wsConnected: false,
    unreadCounts: {},
    totalUnread: 0,
    error: null,
  })

  const loadConversations = useCallback(async () => {
    try {
      const res = await conversationApi.getList()
      const conversations = res.data.conversations
      dispatch({ type: 'SET_CONVERSATIONS', conversations })
      const total = conversations.reduce((sum, c) => sum + c.unread_count, 0)
      dispatch({ type: 'SET_TOTAL_UNREAD', total })
      // 加载会话后同步加载好友列表（用于解析昵称）
      const friendRes = await friendApi.getList()
      dispatch({ type: 'SET_FRIENDS', friends: friendRes.data.friends })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const loadMessages = useCallback(async (convType: string, convId: string) => {
    const key = getKey(convType, convId)
    dispatch({ type: 'SET_CURRENT_CONVERSATION', conversationType: convType, conversationId: convId })
    dispatch({ type: 'SET_LOADING_MESSAGES', loading: true })
    try {
      const res = await messageApi.getMessages(convType, convId)
      dispatch({ type: 'SET_MESSAGES', key, messages: res.data.messages, hasMore: res.data.messages.length >= 50 })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    } finally {
      dispatch({ type: 'SET_LOADING_MESSAGES', loading: false })
    }
  }, [])

  const loadMoreMessages = useCallback(async (convType: string, convId: string): Promise<boolean> => {
    const key = getKey(convType, convId)
    if (!state.hasMore[key]) return false
    const msgs = state.messages[key]
    if (!msgs || msgs.length === 0) return false
    const oldestSeqId = msgs[0].seq_id

    try {
      const res = await messageApi.getMessages(convType, convId, oldestSeqId)
      dispatch({ type: 'PREPEND_MESSAGES', key, messages: res.data.messages, hasMore: res.data.messages.length >= 50 })
      return res.data.messages.length > 0
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
      return false
    }
  }, [state.hasMore, state.messages])

  const sendMessage = useCallback(async (convType: string, convId: string, content: string, contentType: string) => {
    dispatch({ type: 'SET_SENDING_MESSAGE', sending: true })
    try {
      let res: { data: { message: Message } }
      if (convType === 'private') {
        res = await messageApi.sendPrivate({ receiver_id: convId, content_type: contentType, content })
      } else {
        res = await messageApi.sendGroup({ group_id: convId, content_type: contentType, content })
      }
      const key = getKey(convType, convId)
      dispatch({ type: 'APPEND_MESSAGE', key, message: res.data.message })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    } finally {
      dispatch({ type: 'SET_SENDING_MESSAGE', sending: false })
    }
  }, [])

  const recallMessage = useCallback(async (messageId: string) => {
    try {
      await messageApi.recallMessage(messageId)
      for (const [key, msgs] of Object.entries(state.messages)) {
        if (msgs.some(m => m.id === messageId)) {
          dispatch({ type: 'UPDATE_MESSAGE', key, messageId, updates: { status: 'recalled' } })
          break
        }
      }
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [state.messages])

  const sendFriendRequest = useCallback(async (receiverId: string, remark: string) => {
    try {
      await friendApi.sendRequest(receiverId, remark)
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const acceptFriendRequest = useCallback(async (requestId: string) => {
    try {
      await friendApi.acceptRequest(requestId)
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const rejectFriendRequest = useCallback(async (requestId: string) => {
    try {
      await friendApi.rejectRequest(requestId)
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const loadFriendRequests = useCallback(async () => {
    try {
      const res = await friendApi.getReceivedRequests()
      dispatch({ type: 'SET_FRIEND_REQUESTS', requests: res.data.friend_requests })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const loadFriends = useCallback(async () => {
    try {
      const res = await friendApi.getList()
      dispatch({ type: 'SET_FRIENDS', friends: res.data.friends })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const searchUsersFn = useCallback(async (query: string) => {
    try {
      const res = await friendApi.searchUsers(query)
      dispatch({ type: 'SET_SEARCH_RESULTS', users: res.data.users, query })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const createGroup = useCallback(async (data: { name: string; avatar?: string; member_ids?: string[] }) => {
    try {
      await groupApi.create(data)
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const addMembers = useCallback(async (groupId: string, memberIds: string[]) => {
    try {
      await groupApi.addMembers(groupId, memberIds)
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const removeMember = useCallback(async (groupId: string, userId: string) => {
    try {
      await groupApi.removeMember(groupId, userId)
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [])

  const loadConversation = useCallback((convType: string, convId: string) => {
    dispatch({ type: 'SET_CURRENT_CONVERSATION', conversationType: convType, conversationId: convId })
  }, [])

  const markConversationRead = useCallback(async (convType: string, convId: string) => {
    try {
      await conversationApi.markRead(convType, convId)
      const conversations = state.conversations.map(c =>
        c.conversation_type === convType && c.conversation_id === convId ? { ...c, unread_count: 0 } : c
      )
      dispatch({ type: 'SET_CONVERSATIONS', conversations })
      const total = conversations.reduce((sum, c) => sum + c.unread_count, 0)
      dispatch({ type: 'SET_TOTAL_UNREAD', total })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [state.conversations])

  const togglePinFn = useCallback(async (convType: string, convId: string) => {
    try {
      const res = await conversationApi.togglePin(convType, convId)
      const conversations = state.conversations.map(c =>
        c.conversation_type === convType && c.conversation_id === convId ? { ...c, pinned: res.data.pinned } : c
      )
      dispatch({ type: 'SET_CONVERSATIONS', conversations })
    } catch (err) {
      dispatch({ type: 'SET_ERROR', error: (err as Error).message })
    }
  }, [state.conversations])

  const handleWsMessage = useCallback((event: { type: string; data: any }) => {
    switch (event.type) {
      case 'new_message': {
        const msg = event.data as Message
        const key = getKey(msg.conversation_type, msg.conversation_id)
        if (state.messages[key]) {
          dispatch({ type: 'APPEND_MESSAGE', key, message: msg })
        }
        const conversations = state.conversations.map(c =>
          c.conversation_type === msg.conversation_type && c.conversation_id === msg.conversation_id
            ? { ...c, last_message: msg, last_message_at: msg.created_at, unread_count: c.unread_count + 1 }
            : c
        )
        dispatch({ type: 'SET_CONVERSATIONS', conversations })
        const total = conversations.reduce((sum, c) => sum + c.unread_count, 0)
        dispatch({ type: 'SET_TOTAL_UNREAD', total })
        break
      }
      case 'message_recalled': {
        const { message_id, conversation_type, conversation_id } = event.data
        const key = getKey(conversation_type, conversation_id)
        dispatch({ type: 'UPDATE_MESSAGE', key, messageId: message_id, updates: { status: 'recalled' } })
        break
      }
      case 'friend_request':
        loadFriendRequests()
        break
      case 'friend_accepted':
        loadFriends()
        break
    }
  }, [state.messages, state.conversations, loadFriendRequests, loadFriends])

  const setWsConnected = useCallback((connected: boolean) => {
    dispatch({ type: 'SET_WS_CONNECTED', connected })
  }, [])

  return (
    <ChatContext.Provider value={{
      ...state,
      loadConversations,
      loadMessages,
      loadMoreMessages,
      sendMessage,
      recallMessage,
      sendFriendRequest,
      acceptFriendRequest,
      rejectFriendRequest,
      loadFriendRequests,
      loadFriends,
      searchUsers: searchUsersFn,
      createGroup,
      addMembers,
      removeMember,
      loadConversation,
      markConversationRead,
      togglePin: togglePinFn,
      handleWsMessage,
      setWsConnected,
    }}>
      {children}
    </ChatContext.Provider>
  )
}

export function useMessageStore() {
  const ctx = useContext(ChatContext)
  if (!ctx) throw new Error('useMessageStore must be used within MessageProvider')
  return ctx
}
