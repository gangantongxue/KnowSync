import { useState, useEffect } from 'react'
import { message } from 'antd'
import { friendApi } from '../../lib/chat-api'
import { getUser } from '../../lib/auth'
import { request } from '../../lib/client'
import type { Friend } from '../../lib/chat-api'
import type { UserInfo } from '../../lib/auth'

interface FriendPickerModalProps {
  repoId: string
  repoName: string
  onClose: () => void
}

interface FriendWithUser {
  friend: Friend
  user: UserInfo | null
}

export default function FriendPickerModal({ repoId, repoName, onClose }: FriendPickerModalProps) {
  const [friends, setFriends] = useState<FriendWithUser[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [selectedFriendId, setSelectedFriendId] = useState('')
  const [role, setRole] = useState('DEVELOPER')
  const [sending, setSending] = useState(false)

  useEffect(() => {
    const load = async () => {
      try {
        const res = await friendApi.getList()
        const friendList = res.data.friends
        const withUsers = await Promise.all(
          friendList.map(async (f) => {
            try {
              const u = await getUser(f.friend_id)
              return { friend: f, user: u }
            } catch {
              return { friend: f, user: null }
            }
          })
        )
        setFriends(withUsers)
      } catch {
        // silent
      }
      setLoading(false)
    }
    load()
  }, [])

  const filtered = friends.filter(({ friend, user }) => {
    if (!search) return true
    const q = search.toLowerCase()
    const name = user?.name || ''
    const remark = friend.remark || ''
    return name.toLowerCase().includes(q) || remark.toLowerCase().includes(q)
  })

  const handleInvite = async () => {
    if (!selectedFriendId || sending) return
    setSending(true)
    try {
      await request('/collaborators/invite', {
        method: 'POST',
        body: JSON.stringify({ friend_id: selectedFriendId, repo_id: repoId, role }),
      })
      message.success('邀请已发送')
      onClose()
    } catch (err: any) {
      message.error(err.message || '邀请失败')
    } finally {
      setSending(false)
    }
  }

  const getDisplayName = (fw: FriendWithUser) => fw.friend.remark || fw.user?.name || fw.friend.friend_id.slice(0, 8)

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-4 border-b border-gray-100">
          <div>
            <h2 className="text-base font-medium text-gray-800">邀请协作者</h2>
            <p className="text-xs text-gray-400 mt-0.5">知识库：{repoName}</p>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg leading-none">✕</button>
        </div>

        <div className="p-4 border-b border-gray-100">
          <input
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="搜索好友..."
            className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-emerald-400"
          />
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          {loading ? (
            <div className="text-center py-6 text-sm text-gray-400">加载中...</div>
          ) : filtered.length === 0 ? (
            <div className="text-center py-6 text-sm text-gray-400">{search ? '无匹配好友' : '暂无好友'}</div>
          ) : (
            <div className="space-y-1">
              {filtered.map(fw => (
                <label
                  key={fw.friend.friend_id}
                  className="flex items-center gap-3 py-2 px-2 rounded cursor-pointer hover:bg-gray-50"
                >
                  <input
                    type="radio"
                    name="friend"
                    checked={selectedFriendId === fw.friend.friend_id}
                    onChange={() => setSelectedFriendId(fw.friend.friend_id)}
                    className="border-gray-300 text-emerald-500"
                  />
                  <div className="w-8 h-8 rounded-full bg-emerald-100 text-emerald-700 flex items-center justify-center text-xs font-medium shrink-0 overflow-hidden">
                    {fw.user?.avatar ? (
                      <img src={fw.user.avatar} alt="" className="w-full h-full object-cover" />
                    ) : (
                      getDisplayName(fw).charAt(0) || '?'
                    )}
                  </div>
                  <div className="min-w-0">
                    <span className="text-sm text-gray-800 truncate block">{getDisplayName(fw)}</span>
                  </div>
                </label>
              ))}
            </div>
          )}
        </div>

        <div className="p-4 border-t border-gray-100 space-y-3">
          <div className="flex items-center gap-3">
            <span className="text-xs text-gray-500">角色：</span>
            <select
              value={role}
              onChange={e => setRole(e.target.value)}
              className="text-xs border border-gray-200 rounded px-2 py-1 bg-white"
            >
              <option value="DEVELOPER">开发者</option>
              <option value="VIEWER">浏览者</option>
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <button
              onClick={onClose}
              className="px-4 py-1.5 text-sm border border-gray-200 text-gray-500 rounded-lg hover:bg-gray-50"
            >
              取消
            </button>
            <button
              onClick={handleInvite}
              disabled={sending || !selectedFriendId}
              className="px-4 py-1.5 text-sm bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 disabled:opacity-50"
            >
              {sending ? '邀请中...' : '确认邀请'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
