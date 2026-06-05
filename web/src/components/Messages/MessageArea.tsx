import { useEffect, useRef, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Spin } from 'antd'
import { InfoCircleOutlined } from '@ant-design/icons'
import { useMessageStore } from '../../store/message-store'
import type { Message } from '../../lib/chat-api'
import MessageBubble from './MessageBubble'
import GroupSettingsModal from './GroupSettingsModal'

interface MessageAreaProps {
  conversationType: string
  conversationId: string
  currentUserId: string
  onReply: (msg: Message) => void
  onRecall: (msgId: string) => void
  onMention: (userId: string) => void
}

function formatDateSeparator(ts: number): string {
  const date = new Date(ts * 1000)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const msgDate = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const diffDays = Math.floor((today.getTime() - msgDate.getTime()) / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return '今天'
  if (diffDays === 1) return '昨天'
  if (diffDays < 7) {
    const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
    return weekdays[date.getDay()]
  }
  return `${date.getMonth() + 1}月${date.getDate()}日`
}

function getDateKey(ts: number): string {
  const date = new Date(ts * 1000)
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
}

export default function MessageArea({ conversationType, conversationId, currentUserId, onReply, onRecall, onMention }: MessageAreaProps) {
  const navigate = useNavigate()
  const { messages, loadingMessages, loadMoreMessages, hasMore, conversations, loadUserProfiles, getUserDisplayName, getUserAvatar } = useMessageStore()
  const containerRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const isNearBottomRef = useRef(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [resolvedConvName, setResolvedConvName] = useState('')
  const [resolvedAvatar, setResolvedAvatar] = useState('')

  const [groupSettingsOpen, setGroupSettingsOpen] = useState(false)
  const [highlightMessageId, setHighlightMessageId] = useState('')

  const getSenderName = (senderId: string): string => {
    if (senderId === currentUserId) return '我'
    return getUserDisplayName(senderId)
  }

  const getSenderAvatar = (senderId: string): string => {
    if (senderId === currentUserId) return ''
    return getUserAvatar(senderId)
  }

  const key = `${conversationType}_${conversationId}`
  const msgList = messages[key] || []
  const backendName = conversations.find(
    c => c.conversation_type === conversationType && c.conversation_id === conversationId
  )?.name || conversationId

  const resolveConvName = useCallback(async () => {
    if (conversationType !== 'private') {
      setResolvedConvName(backendName)
      setResolvedAvatar('')
      return
    }
    const parts = conversationId.split('_')
    if (parts.length !== 2) {
      setResolvedConvName(backendName)
      setResolvedAvatar('')
      return
    }
    const otherId = parts[0] === currentUserId ? parts[1] : parts[0]
    setResolvedConvName(getUserDisplayName(otherId))
    setResolvedAvatar(getUserAvatar(otherId))
    loadUserProfiles([otherId])
  }, [conversationType, conversationId, currentUserId, backendName, getUserDisplayName, getUserAvatar, loadUserProfiles])

  useEffect(() => {
    resolveConvName()
  }, [resolveConvName])

  useEffect(() => {
    if (conversationType !== 'group') return
    const ids = new Set<string>()
    for (const msg of msgList) {
      if (msg.sender_id !== currentUserId) ids.add(msg.sender_id)
    }
    if (ids.size > 0) loadUserProfiles([...ids])
  }, [conversationType, msgList, currentUserId, loadUserProfiles])

  const convName = resolvedConvName || backendName

  const mentionNames: Record<string, string> = {}
  for (const msg of msgList) {
    const msgMentions = (() => {
      if (!msg.extra) return []
      try { return JSON.parse(msg.extra).mentions || [] } catch { return [] }
    })()
    for (const userId of msgMentions) {
      if (mentionNames[userId]) continue
      mentionNames[userId] = getUserDisplayName(userId)
    }
  }

  const hasMoreMsgs = hasMore[key] || false

  const scrollToBottom = useCallback((smooth = true) => {
    requestAnimationFrame(() => {
      if (bottomRef.current) {
        bottomRef.current.scrollIntoView({ behavior: smooth ? 'smooth' : 'instant' })
      }
    })
  }, [])

  useEffect(() => {
    if (isNearBottomRef.current && msgList.length > 0) {
      scrollToBottom()
    }
  }, [msgList.length, scrollToBottom])

  const handleJumpToMessage = useCallback((messageId: string) => {
    setHighlightMessageId(messageId)
    const el = document.getElementById(`msg-${messageId}`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
    setTimeout(() => setHighlightMessageId(''), 1500)
  }, [])

  const handleScroll = async () => {
    const container = containerRef.current
    if (!container) return

    isNearBottomRef.current = container.scrollHeight - container.scrollTop - container.clientHeight < 100

    if (container.scrollTop < 100 && !loadingMore && hasMoreMsgs && msgList.length > 0) {
      setLoadingMore(true)
      const prevScrollHeight = container.scrollHeight
      await loadMoreMessages(conversationType, conversationId)
      requestAnimationFrame(() => {
        if (containerRef.current) {
          const newScrollHeight = containerRef.current.scrollHeight
          containerRef.current.scrollTop = newScrollHeight - prevScrollHeight
        }
      })
      setLoadingMore(false)
    }
  }

  const findRepliedMessage = (msg: Message): Message | null => {
    if (!msg.reply_to_id) return null
    return msgList.find(m => m.id === msg.reply_to_id) || null
  }

  const groupedMessages: { dateKey: string; messages: Message[] }[] = []
  const dateGroups = new Map<string, Message[]>()
  for (const msg of msgList) {
    const dk = getDateKey(msg.created_at)
    if (!dateGroups.has(dk)) dateGroups.set(dk, [])
    dateGroups.get(dk)!.push(msg)
  }
  for (const [dateKey, msgs] of dateGroups) {
    groupedMessages.push({ dateKey, messages: msgs })
  }

  return (
    <div className="flex flex-col h-full min-h-0">
      <style>{`
        @keyframes messageFlash {
          0%, 100% { background-color: transparent; }
          15% { background-color: rgba(59, 130, 246, 0.15); }
          30% { background-color: transparent; }
          45% { background-color: rgba(59, 130, 246, 0.15); }
          60% { background-color: transparent; }
        }
        .animate-message-flash {
          animation: messageFlash 1.5s ease-in-out;
        }
      `}</style>
      <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between bg-white shrink-0">
        <div
          className={`flex items-center gap-2 ${conversationType === 'private' ? 'cursor-pointer hover:opacity-80' : ''}`}
          onClick={() => {
            if (conversationType !== 'private') return
            const parts = conversationId.split('_')
            const otherId = parts.length === 2 ? (parts[0] === currentUserId ? parts[1] : parts[0]) : conversationId
            navigate(`/friends/${otherId}`)
          }}
        >
          {conversationType === 'private' && (
            resolvedAvatar ? (
              <img src={resolvedAvatar} alt="" className="w-7 h-7 rounded-full object-cover" />
            ) : (
              <div className="w-7 h-7 rounded-full bg-blue-100 flex items-center justify-center text-xs text-blue-600 font-medium">
                {convName.charAt(0).toUpperCase()}
              </div>
            )
          )}
          <span className="text-sm font-medium text-gray-800">{convName}</span>
        </div>
        {conversationType === 'group' && (
          <button
            onClick={() => setGroupSettingsOpen(true)}
            className="text-gray-400 hover:text-blue-500 p-1"
            title="群组设置"
          >
            <InfoCircleOutlined />
          </button>
        )}
      </div>

      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="flex-1 overflow-y-auto px-4 py-4 min-h-0"
      >
        {loadingMore && (
          <div className="text-center py-2">
            <Spin size="small" />
          </div>
        )}

        {loadingMessages && msgList.length === 0 ? (
          <div className="h-full flex items-center justify-center">
            <Spin />
          </div>
        ) : msgList.length === 0 ? (
          <div className="h-full flex items-center justify-center text-gray-400 text-sm">
            暂无消息，发送第一条消息开始聊天
          </div>
        ) : (
          groupedMessages.map(group => (
            <div key={group.dateKey}>
              <div className="flex justify-center mb-4">
                <span className="text-xs text-gray-400 bg-gray-100 px-3 py-1 rounded-full">
                  {formatDateSeparator(group.messages[0].created_at)}
                </span>
              </div>
              {group.messages.map(msg => {
                const senderName = getSenderName(msg.sender_id)
                const senderAvatar = getSenderAvatar(msg.sender_id)
                return (
                  <MessageBubble
                    key={msg.id}
                    message={msg}
                    isOwn={msg.sender_id === currentUserId}
                    senderName={senderName}
                    senderAvatar={senderAvatar}
                    repliedMessage={findRepliedMessage(msg)}
                    currentUserId={currentUserId}
                    onReply={onReply}
                    onRecall={onRecall}
                    onMention={onMention}
                    onJumpToMessage={handleJumpToMessage}
                    highlight={msg.id === highlightMessageId}
                    mentionNames={mentionNames}
                  />
                )
              })}
            </div>
          ))
        )}
        <div ref={bottomRef} />
      </div>

      {groupSettingsOpen && (
        <GroupSettingsModal groupId={conversationId} onClose={() => setGroupSettingsOpen(false)} />
      )}
    </div>
  )
}
