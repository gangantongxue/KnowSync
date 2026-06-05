import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth-context'
import { useMessageStore } from '../store/message-store'
import { friendApi } from '../lib/chat-api'
import { userApi } from '../lib/user-api'

export default function FriendDetail() {
  const { friendId } = useParams<{ friendId: string }>()
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const { friends, loadFriends } = useMessageStore()

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [profile, setProfile] = useState<{ name: string; avatar: string } | null>(null)
  const [remark, setRemark] = useState('')
  const [editingRemark, setEditingRemark] = useState(false)
  const [editValue, setEditValue] = useState('')
  const [saving, setSaving] = useState(false)

  const currentUserId = currentUser?.id || ''

  useEffect(() => {
    if (!friendId) return
    setLoading(true)
    loadFriends()
    userApi.getProfile(friendId)
      .then(res => {
        setProfile({ name: res.data.user.name, avatar: res.data.user.avatar })
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [friendId, loadFriends])

  useEffect(() => {
    if (!friendId) return
    const friend = friends.find(f => f.friend_id === friendId)
    setRemark(friend?.remark || '')
  }, [friendId, friends])

  const handleStartEdit = () => {
    setEditValue(remark || profile?.name || '')
    setEditingRemark(true)
  }

  const handleSaveRemark = async () => {
    if (!friendId) return
    setSaving(true)
    try {
      await friendApi.updateRemark(friendId, editValue)
      setRemark(editValue)
      setEditingRemark(false)
      loadFriends()
    } catch (err: any) {
      setError(err.message)
    }
    setSaving(false)
  }

  const handleClearRemark = async () => {
    if (!friendId) return
    setSaving(true)
    try {
      await friendApi.updateRemark(friendId, '')
      setRemark('')
      loadFriends()
    } catch (err: any) {
      setError(err.message)
    }
    setSaving(false)
  }

  const handleSendMessage = () => {
    if (!friendId || !currentUserId) return
    const ids = [currentUserId, friendId].sort()
    const conversationId = ids.join('_')
    navigate(`/messages/private/${conversationId}`)
  }

  const handleDeleteFriend = async () => {
    if (!friendId) return
    if (!window.confirm('确定删除该好友？')) return
    try {
      await friendApi.deleteFriend(friendId)
      navigate('/messages')
    } catch (err: any) {
      setError(err.message)
    }
  }

  const displayRemark = remark || profile?.name || friendId || ''
  const displayName = profile?.name || ''
  const displayAvatar = profile?.avatar || ''

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-gray-400 text-sm">加载中...</div>
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto bg-gray-50">
      {error && (
        <div className="bg-red-50 text-red-600 text-sm text-center py-2">{error}</div>
      )}

      {/* Avatar */}
      <div className="flex flex-col items-center pt-8 pb-4">
        {displayAvatar ? (
          <img src={displayAvatar} alt="" className="w-20 h-20 rounded-full object-cover" />
        ) : (
          <div className="w-20 h-20 rounded-full bg-blue-100 flex items-center justify-center text-2xl text-blue-600 font-medium">
            {(displayRemark || displayName).charAt(0).toUpperCase() || '?'}
          </div>
        )}
      </div>

      {/* Remark */}
      <div className="flex items-center justify-center gap-2 px-4 mb-1">
        {editingRemark ? (
          <div className="flex items-center gap-2">
            <input
              value={editValue}
              onChange={e => setEditValue(e.target.value)}
              className="w-48 px-3 py-1.5 border border-gray-300 rounded-lg text-base font-medium text-center text-gray-800 focus:outline-none focus:border-blue-400"
              autoFocus
              onKeyDown={e => {
                if (e.key === 'Enter') handleSaveRemark()
                if (e.key === 'Escape') setEditingRemark(false)
              }}
            />
            <button
              onClick={handleSaveRemark}
              disabled={saving}
              className="text-sm text-blue-500 hover:text-blue-600 shrink-0"
            >
              {saving ? '保存中' : '保存'}
            </button>
            <button
              onClick={() => setEditingRemark(false)}
              className="text-sm text-gray-400 hover:text-gray-600 shrink-0"
            >
              取消
            </button>
          </div>
        ) : (
          <>
            <span className="text-lg font-semibold text-gray-800">{displayRemark}</span>
            <button
              onClick={handleStartEdit}
              className="text-gray-400 hover:text-blue-500 p-1"
              title="修改备注"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
              </svg>
            </button>
          </>
        )}
      </div>

      {remark && (
        <div className="text-center mb-1">
          <button
            onClick={handleClearRemark}
            className="text-xs text-gray-400 hover:text-red-500"
          >
            清除备注
          </button>
        </div>
      )}

      {/* Nickname & ID */}
      <div className="text-center mb-6">
        {displayName && (
          <p className="text-sm text-gray-500">昵称: {displayName}</p>
        )}
        <p className="text-xs text-gray-400 mt-0.5">ID: {friendId}</p>
      </div>

      {/* Action Buttons */}
      <div className="px-6 space-y-3">
        <button
          onClick={handleSendMessage}
          className="w-full py-2.5 bg-blue-500 text-white rounded-lg text-sm font-medium hover:bg-blue-600 transition-colors"
        >
          发送消息
        </button>
        <button
          onClick={handleDeleteFriend}
          className="w-full py-2.5 bg-white text-red-500 border border-red-200 rounded-lg text-sm font-medium hover:bg-red-50 transition-colors"
        >
          删除好友
        </button>
      </div>
    </div>
  )
}
