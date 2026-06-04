import { useState, useEffect } from 'react'
import { Input, message } from 'antd'
import { useMessageStore } from '../../store/message-store'

interface CreateGroupModalProps {
  onClose: () => void
}

export default function CreateGroupModal({ onClose }: CreateGroupModalProps) {
  const { friends, loadFriends, createGroup, loadConversations } = useMessageStore()
  const [name, setName] = useState('')
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    loadFriends()
  }, [])

  const getFriendName = (friendId: string): string => {
    const friend = friends.find(f => f.friend_id === friendId)
    return friend?.remark || friendId
  }

  const handleCreate = async () => {
    if (!name.trim()) {
      message.warning('请输入群名称')
      return
    }
    setCreating(true)
    try {
      await createGroup({ name: name.trim(), member_ids: selectedIds })
      message.success('群聊创建成功')
      loadConversations()
      onClose()
    } catch {
      message.error('创建失败')
    } finally {
      setCreating(false)
    }
  }

  const toggleMember = (friendId: string) => {
    setSelectedIds(prev =>
      prev.includes(friendId) ? prev.filter(id => id !== friendId) : [...prev, friendId]
    )
  }

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-4 border-b border-gray-100">
          <h2 className="text-base font-medium text-gray-800">创建群聊</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg leading-none">✕</button>
        </div>

        <div className="p-4 border-b border-gray-100">
          <label className="text-xs text-gray-500 mb-1 block">群名称 *</label>
          <Input
            placeholder="请输入群名称"
            value={name}
            onChange={e => setName(e.target.value)}
          />
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          <label className="text-xs text-gray-500 mb-2 block">选择成员</label>
          {friends.length === 0 ? (
            <div className="text-center py-6 text-sm text-gray-400">暂无好友</div>
          ) : (
            <div className="space-y-1">
              {friends.map(friend => {
                const displayName = getFriendName(friend.friend_id)
                const isSelected = selectedIds.includes(friend.friend_id)
                return (
                  <label
                    key={friend.friend_id}
                    className="flex items-center gap-3 py-2 px-2 rounded cursor-pointer hover:bg-gray-50"
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggleMember(friend.friend_id)}
                      className="rounded border-gray-300 text-blue-500"
                    />
                    <div className="w-8 h-8 rounded-full bg-gray-100 text-gray-600 flex items-center justify-center text-xs font-medium shrink-0">
                      {displayName.charAt(0).toUpperCase()}
                    </div>
                    <div className="min-w-0">
                      <span className="text-sm text-gray-800 truncate block">{displayName}</span>
                    </div>
                  </label>
                )
              })}
            </div>
          )}
        </div>

        <div className="p-4 border-t border-gray-100 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-1.5 text-sm border border-gray-200 text-gray-500 rounded-lg hover:bg-gray-50"
          >
            取消
          </button>
          <button
            onClick={handleCreate}
            disabled={creating || !name.trim()}
            className="px-4 py-1.5 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
          >
            {creating ? '创建中...' : '创建'}
          </button>
        </div>
      </div>
    </div>
  )
}
