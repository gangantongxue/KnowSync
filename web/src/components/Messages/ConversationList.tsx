import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Input, Badge, Dropdown, message } from 'antd'
import { UserAddOutlined, TeamOutlined, SearchOutlined } from '@ant-design/icons'
import { useMessageStore } from '../../store/message-store'
import { conversationApi } from '../../lib/chat-api'
import { userApi } from '../../lib/user-api'
import type { ConversationInfo } from '../../lib/chat-api'
import FriendRequestsModal from './FriendRequestsModal'
import SearchUserModal from './SearchUserModal'
import CreateGroupModal from './CreateGroupModal'

function formatRelativeTime(ts: number): string {
  if (!ts) return ''
  const now = Date.now()
  const diff = now - ts * 1000
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  const days = Math.floor(hours / 24)
  if (days < 7) return `${days}天前`
  const date = new Date(ts * 1000)
  return `${date.getMonth() + 1}-${date.getDate()}`
}

const PASTEL_BG = [
  'bg-red-100 text-red-600',
  'bg-blue-100 text-blue-600',
  'bg-green-100 text-green-600',
  'bg-yellow-100 text-yellow-700',
  'bg-purple-100 text-purple-700',
  'bg-pink-100 text-pink-600',
  'bg-indigo-100 text-indigo-600',
  'bg-teal-100 text-teal-600',
]

function getAvatarColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return PASTEL_BG[Math.abs(hash) % PASTEL_BG.length]
}

function getFirstChar(name: string): string {
  return name.charAt(0).toUpperCase() || '?'
}

interface UserInfo {
  name: string
  avatar: string
}

export default function ConversationList() {
  const navigate = useNavigate()
  const { conversationType, conversationId } = useParams()
  const {
    conversations, togglePin, friendRequests, loadFriendRequests,
    loadConversations, loadMessages, friends,
  } = useMessageStore()
  const [searchText, setSearchText] = useState('')
  const [showFriendRequests, setShowFriendRequests] = useState(false)
  const [showSearchUser, setShowSearchUser] = useState(false)
  const [showCreateGroup, setShowCreateGroup] = useState(false)
  const [userCache, setUserCache] = useState<Record<string, UserInfo>>({})
  const pendingCount = friendRequests.filter(r => r.status === 'pending').length

  useEffect(() => {
    loadFriendRequests()
  }, [loadFriendRequests])

  // 解析私聊会话中的好友 ID 为用户昵称
  useEffect(() => {
    const unknownIds = new Set<string>()
    for (const c of conversations) {
      if (c.conversation_type !== 'private') continue
      const parts = c.conversation_id.split('_')
      if (parts.length !== 2) continue
      for (const id of parts) {
        if (userCache[id]) continue
        const friend = friends.find(f => f.friend_id === id)
        if (friend?.remark) {
          setUserCache(prev => ({ ...prev, [id]: { name: friend.remark!, avatar: '' } }))
          continue
        }
        unknownIds.add(id)
      }
    }
    if (unknownIds.size === 0) return
    const ids = [...unknownIds]
    Promise.allSettled(ids.map(id => userApi.getProfile(id))).then(results => {
      const newCache: Record<string, UserInfo> = {}
      results.forEach((r, i) => {
        if (r.status === 'fulfilled') {
          newCache[ids[i]] = { name: r.value.data.user.name, avatar: r.value.data.user.avatar }
        }
      })
      if (Object.keys(newCache).length > 0) {
        setUserCache(prev => ({ ...prev, ...newCache }))
      }
    })
  }, [conversations, friends, userCache])

  const getDisplayName = (conv: ConversationInfo): string => {
    if (conv.conversation_type !== 'private') return conv.name
    const parts = conv.conversation_id.split('_')
    if (parts.length !== 2) return conv.name
    for (const id of parts) {
      const cached = userCache[id]
      if (cached?.name) return cached.name
      const friend = friends.find(f => f.friend_id === id)
      if (friend?.remark) return friend.remark
    }
    return conv.name
  }

  const getAvatar = (conv: ConversationInfo): string | null => {
    if (conv.conversation_type !== 'private') return null
    const parts = conv.conversation_id.split('_')
    if (parts.length !== 2) return null
    for (const id of parts) {
      const avatar = userCache[id]?.avatar
      if (avatar) return avatar
    }
    return null
  }

  const sorted = [...conversations].sort((a, b) => {
    if (a.pinned !== b.pinned) return a.pinned ? -1 : 1
    return (b.last_message_at || 0) - (a.last_message_at || 0)
  })

  const filtered = searchText
    ? sorted.filter(c => c.name.toLowerCase().includes(searchText.toLowerCase()))
    : sorted

  const isActive = (conv: ConversationInfo) =>
    conversationType === conv.conversation_type && conversationId === conv.conversation_id

  const handleClick = (conv: ConversationInfo) => {
    navigate(`/messages/${conv.conversation_type}/${conv.conversation_id}`)
    loadMessages(conv.conversation_type, conv.conversation_id)
  }

  const handleDelete = async (conv: ConversationInfo) => {
    try {
      await conversationApi.deleteConversation(conv.conversation_type, conv.conversation_id)
      message.success('已删除会话')
      loadConversations()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <div className="w-[280px] shrink-0 border-r border-gray-200 bg-white flex flex-col">
      <div className="p-3 border-b border-gray-100">
        <Input.Search
          placeholder="搜索会话"
          value={searchText}
          onChange={e => setSearchText(e.target.value)}
          size="small"
        />
      </div>

      <div className="px-3 py-2 border-b border-gray-100">
        <div className="flex items-center justify-between mb-2">
          <button
            onClick={() => { loadFriendRequests(); setShowFriendRequests(true) }}
            className="flex items-center gap-2 text-sm text-gray-600 hover:text-blue-600"
          >
            <Badge count={pendingCount} size="small" offset={[4, -4]}>
              <UserAddOutlined className="text-base" />
            </Badge>
            <span>好友申请</span>
          </button>
          <button
            onClick={() => setShowCreateGroup(true)}
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-blue-600"
          >
            <TeamOutlined className="text-base" />
            <span>创建群聊</span>
          </button>
        </div>
        <button
          onClick={() => setShowSearchUser(true)}
          className="flex items-center gap-1 text-xs text-gray-400 hover:text-blue-500 w-full"
        >
          <SearchOutlined />
          <span>搜索用户添加好友</span>
        </button>
      </div>

      {showFriendRequests && <FriendRequestsModal onClose={() => setShowFriendRequests(false)} />}
      {showSearchUser && <SearchUserModal onClose={() => setShowSearchUser(false)} />}
      {showCreateGroup && <CreateGroupModal onClose={() => setShowCreateGroup(false)} />}

      <div className="flex-1 overflow-y-auto">
        {filtered.map(conv => {
          const active = isActive(conv)
          const previewContent = conv.last_message?.content || ''
          const displayName = getDisplayName(conv)
          const avatarUrl = getAvatar(conv)

          const contextMenuItems = [
            {
              key: 'pin',
              label: conv.pinned ? '取消置顶' : '置顶',
              onClick: () => togglePin(conv.conversation_type, conv.conversation_id),
            },
            { type: 'divider' as const },
            {
              key: 'delete',
              label: '删除会话',
              danger: true,
              onClick: () => handleDelete(conv),
            },
          ]

          return (
            <Dropdown key={`${conv.conversation_type}_${conv.conversation_id}`} menu={{ items: contextMenuItems }} trigger={['contextMenu']}>
              <div
                onClick={() => handleClick(conv)}
                className={`flex items-start gap-3 px-3 py-3 cursor-pointer border-b border-gray-50 transition-colors ${
                  active ? 'bg-blue-50' : 'hover:bg-gray-50'
                }`}
              >
                {avatarUrl ? (
                  <img src={avatarUrl} alt="" className="w-10 h-10 rounded-full object-cover shrink-0" />
                ) : (
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium shrink-0 ${getAvatarColor(displayName)}`}>
                    {getFirstChar(displayName)}
                  </div>
                )}

                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <span className={`text-sm font-medium truncate ${active ? 'text-blue-600' : 'text-gray-800'}`}>
                      {displayName}
                    </span>
                    <span className="text-xs text-gray-400 shrink-0 ml-1">
                      {formatRelativeTime(conv.last_message_at)}
                    </span>
                  </div>
                  <div className="flex items-center justify-between mt-0.5">
                    <span className="text-xs text-gray-500 truncate flex-1">
                      {previewContent || '暂无消息'}
                    </span>
                    {conv.unread_count > 0 && (
                      <Badge count={conv.unread_count} size="small" className="shrink-0 ml-1" />
                    )}
                  </div>
                </div>
              </div>
            </Dropdown>
          )
        })}
      </div>
    </div>
  )
}
