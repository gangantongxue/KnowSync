import { useState, useEffect } from 'react'
import { message } from 'antd'
import { useAuth } from '../../store/auth-context'
import { useMessageStore } from '../../store/message-store'
import { messageApi } from '../../lib/chat-api'
import type { Message, ConversationInfo } from '../../lib/chat-api'

interface ForwardModalProps {
  message: Message
  onClose: () => void
}

export default function ForwardModal({ message: msg, onClose }: ForwardModalProps) {
  const { user } = useAuth()
  const { conversations, loadUserProfiles, getUserDisplayName, getUserAvatar } = useMessageStore()
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [sending, setSending] = useState(false)

  const currentUserId = user?.id || ''

  useEffect(() => {
    const ids = new Set<string>()
    for (const c of conversations) {
      if (c.conversation_type !== 'private') continue
      const parts = c.conversation_id.split('_')
      if (parts.length !== 2) continue
      const otherId = parts[0] === currentUserId ? parts[1] : parts[0]
      ids.add(otherId)
    }
    if (ids.size > 0) loadUserProfiles([...ids])
  }, [conversations, currentUserId, loadUserProfiles])

  const getConvDisplayName = (conv: ConversationInfo): string => {
    if (conv.conversation_type !== 'private') return conv.name
    const parts = conv.conversation_id.split('_')
    if (parts.length !== 2) return conv.name
    const otherId = parts[0] === currentUserId ? parts[1] : parts[0]
    return getUserDisplayName(otherId)
  }

  const getConvAvatar = (conv: ConversationInfo): string => {
    if (conv.conversation_type !== 'private') return ''
    const parts = conv.conversation_id.split('_')
    if (parts.length !== 2) return ''
    const otherId = parts[0] === currentUserId ? parts[1] : parts[0]
    return getUserAvatar(otherId)
  }

  const isSelected = (conv: ConversationInfo) =>
    selectedIds.has(`${conv.conversation_type}_${conv.conversation_id}`)

  const toggleConversation = (conv: ConversationInfo) => {
    const key = `${conv.conversation_type}_${conv.conversation_id}`
    setSelectedIds(prev => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  const handleSend = async () => {
    if (selectedIds.size === 0) {
      message.warning('请选择要转发的目标')
      return
    }
    setSending(true)
    const entries = Array.from(selectedIds).map(key => {
      const [convType, convId] = [key.substring(0, key.indexOf('_')), key.substring(key.indexOf('_') + 1)]
      return { target_conversation_type: convType, target_conversation_id: convId, message_ids: [msg.id] }
    })
    try {
      await Promise.all(entries.map(data => messageApi.forwardMessage(data)))
      message.success(`已转发至 ${entries.length} 个会话`)
      onClose()
    } catch {
      message.error('转发失败')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-4 border-b border-gray-100">
          <h2 className="text-base font-medium text-gray-800">转发消息</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg leading-none">✕</button>
        </div>

        <div className="flex-1 overflow-y-auto px-4">
          {conversations.length === 0 ? (
            <div className="text-center py-8 text-sm text-gray-400">暂无会话</div>
          ) : (
            conversations.map(conv => {
              const sel = isSelected(conv)
              const displayName = getConvDisplayName(conv)
              const avatar = getConvAvatar(conv)
              return (
                <label
                  key={`${conv.conversation_type}_${conv.conversation_id}`}
                  className="flex items-center gap-3 py-3 border-b border-gray-50 cursor-pointer hover:bg-gray-50 px-2 rounded"
                >
                  <input
                    type="checkbox"
                    checked={sel}
                    onChange={() => toggleConversation(conv)}
                    className="rounded border-gray-300 text-blue-500"
                  />
                  {avatar ? (
                    <img src={avatar} alt="" className="w-10 h-10 rounded-full object-cover shrink-0" />
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-sm font-medium shrink-0">
                      {displayName.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <div className="min-w-0 flex-1">
                    <div className="text-sm text-gray-800 truncate">{displayName}</div>
                    <div className="text-xs text-gray-400">
                      {conv.conversation_type === 'group' ? '群聊' : '好友'}
                    </div>
                  </div>
                </label>
              )
            })
          )}
        </div>

        <div className="p-4 border-t border-gray-100 flex items-center justify-between">
          <span className="text-xs text-gray-400">已选 {selectedIds.size} 个会话</span>
          <div className="flex gap-2">
            <button
              onClick={onClose}
              className="px-4 py-1.5 text-sm border border-gray-200 text-gray-500 rounded-lg hover:bg-gray-50"
            >
              取消
            </button>
            <button
              onClick={handleSend}
              disabled={sending || selectedIds.size === 0}
              className="px-4 py-1.5 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
            >
              {sending ? '发送中...' : '发送'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
