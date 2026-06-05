import { useState, useEffect } from 'react'
import { message } from 'antd'
import { useMessageStore } from '../../store/message-store'
import { friendApi } from '../../lib/chat-api'
import { userApi } from '../../lib/user-api'
import type { FriendRequest } from '../../lib/chat-api'

interface FriendRequestsModalProps {
  onClose: () => void
}

interface UserInfo {
  name: string
  avatar: string
}

const STATUS_LABEL: Record<string, string> = {
  pending: '待处理',
  accepted: '已同意',
  rejected: '已拒绝',
}

const STATUS_COLOR: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-700',
  accepted: 'bg-green-100 text-green-700',
  rejected: 'bg-gray-100 text-gray-500',
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

export default function FriendRequestsModal({ onClose }: FriendRequestsModalProps) {
  const { acceptFriendRequest, rejectFriendRequest } = useMessageStore()
  const [tab, setTab] = useState<'received' | 'sent'>('received')
  const [receivedRequests, setReceivedRequests] = useState<FriendRequest[]>([])
  const [sentRequests, setSentRequests] = useState<FriendRequest[]>([])
  const [loading, setLoading] = useState(false)
  const [userCache, setUserCache] = useState<Record<string, UserInfo>>({})

  useEffect(() => {
    loadRequests()
  }, [])

  const loadRequests = async () => {
    setLoading(true)
    try {
      const [receivedRes, sentRes] = await Promise.all([
        friendApi.getReceivedRequests(),
        friendApi.getSentRequests(),
      ])
      const received = receivedRes.data.friend_requests
      const sent = sentRes.data.friend_requests
      setReceivedRequests(received)
      setSentRequests(sent)

      const allIds = new Set<string>()
      received.forEach(r => allIds.add(r.sender_id))
      sent.forEach(r => allIds.add(r.receiver_id))
      const unknownIds = [...allIds].filter(id => !userCache[id])

      if (unknownIds.length > 0) {
        const profiles = await Promise.allSettled(
          unknownIds.map(id => userApi.getProfile(id))
        )
        const newCache = { ...userCache }
        profiles.forEach((result, i) => {
          if (result.status === 'fulfilled') {
            newCache[unknownIds[i]] = {
              name: result.value.data.user.name,
              avatar: result.value.data.user.avatar,
            }
          }
        })
        setUserCache(newCache)
      }
    } catch {
      message.error('加载好友请求失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAccept = async (requestId: string) => {
    try {
      await acceptFriendRequest(requestId)
      message.success('已同意好友请求')
      loadRequests()
    } catch {
      message.error('操作失败')
    }
  }

  const handleReject = async (requestId: string) => {
    try {
      await rejectFriendRequest(requestId)
      message.success('已拒绝好友请求')
      loadRequests()
    } catch {
      message.error('操作失败')
    }
  }

  const renderRequest = (req: FriendRequest, isReceived: boolean) => {
    const targetId = isReceived ? req.sender_id : req.receiver_id
    const info = userCache[targetId]
    const displayName = info?.name || targetId
    const avatarUrl = info?.avatar

    return (
      <div key={req.id} className="flex items-center justify-between py-3 border-b border-gray-50">
        <div className="flex items-center gap-3 min-w-0">
          {avatarUrl ? (
            <img src={avatarUrl} alt="" className="w-10 h-10 rounded-full object-cover shrink-0" />
          ) : (
            <div className={`w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium shrink-0 ${getAvatarColor(displayName)}`}>
              {displayName.charAt(0).toUpperCase()}
            </div>
          )}
          <div className="min-w-0">
            <div className="text-sm text-gray-800 truncate font-medium">{displayName}</div>
            <div className="text-xs text-gray-400 truncate">ID: {targetId}</div>
            {req.remark && <div className="text-xs text-gray-500 truncate">备注: {req.remark}</div>}
            <div className="text-xs text-gray-400">{new Date(req.created_at * 1000).toLocaleDateString()}</div>
          </div>
        </div>
        {isReceived && req.status === 'pending' ? (
          <div className="flex gap-2 shrink-0">
            <button
              onClick={() => handleAccept(req.id)}
              className="px-3 py-1 text-xs bg-blue-500 text-white rounded-lg hover:bg-blue-600"
            >
              同意
            </button>
            <button
              onClick={() => handleReject(req.id)}
              className="px-3 py-1 text-xs border border-gray-200 text-gray-500 rounded-lg hover:bg-gray-50"
            >
              拒绝
            </button>
          </div>
        ) : (
          <span className={`text-xs px-2 py-1 rounded-full shrink-0 ${STATUS_COLOR[req.status] || 'bg-gray-100 text-gray-500'}`}>
            {STATUS_LABEL[req.status] || req.status}
          </span>
        )}
      </div>
    )
  }

  const currentList = tab === 'received' ? receivedRequests : sentRequests

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-4 border-b border-gray-100">
          <h2 className="text-base font-medium text-gray-800">好友请求</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg leading-none">✕</button>
        </div>

        <div className="flex border-b border-gray-100">
          <button
            onClick={() => setTab('received')}
            className={`flex-1 py-2.5 text-sm text-center transition-colors ${
              tab === 'received' ? 'text-blue-600 border-b-2 border-blue-500 font-medium' : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            收到的请求
          </button>
          <button
            onClick={() => setTab('sent')}
            className={`flex-1 py-2.5 text-sm text-center transition-colors ${
              tab === 'sent' ? 'text-blue-600 border-b-2 border-blue-500 font-medium' : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            发出的请求
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-4">
          {loading ? (
            <div className="text-center py-8 text-sm text-gray-400">加载中...</div>
          ) : currentList.length === 0 ? (
            <div className="text-center py-8 text-sm text-gray-400">暂无请求</div>
          ) : (
            currentList.map(req => renderRequest(req, tab === 'received'))
          )}
        </div>
      </div>
    </div>
  )
}
