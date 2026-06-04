import { useEffect, useRef, useState, useCallback } from 'react'
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

export default function MessageArea({ conversationType, conversationId, currentUserId, onReply, onRecall }: MessageAreaProps) {
  const { messages, loadingMessages, loadMoreMessages, hasMore, conversations } = useMessageStore()
  const containerRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const isNearBottomRef = useRef(true)
  const [loadingMore, setLoadingMore] = useState(false)

  const [groupSettingsOpen, setGroupSettingsOpen] = useState(false)

  const key = `${conversationType}_${conversationId}`
  const msgList = messages[key] || []
  const convName = conversations.find(
    c => c.conversation_type === conversationType && c.conversation_id === conversationId
  )?.name || conversationId

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
    <div className="flex flex-col h-full">
      <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between bg-white shrink-0">
        <span className="text-sm font-medium text-gray-800">{convName}</span>
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
        className="flex-1 overflow-y-auto px-4 py-4"
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
              {group.messages.map(msg => (
                <MessageBubble
                  key={msg.id}
                  message={msg}
                  isOwn={msg.sender_id === currentUserId}
                  senderName={msg.sender_id === currentUserId ? '我' : convName}
                  repliedMessage={findRepliedMessage(msg)}
                  currentUserId={currentUserId}
                  onReply={onReply}
                  onRecall={onRecall}
                />
              ))}
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
