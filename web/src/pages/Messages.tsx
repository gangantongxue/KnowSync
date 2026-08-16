import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useMessageStore } from '../store/message-store'
import { useAuth } from '../store/auth-context'
import ConversationList from '../components/Messages/ConversationList'
import MessageArea from '../components/Messages/MessageArea'
import MessageInput from '../components/Messages/MessageInput'
import type { Message } from '../lib/chat-api'

export default function Messages() {
  const { conversationType, conversationId } = useParams()
  const { user } = useAuth()
  const {
    loadMessages, recallMessage, markConversationRead,
  } = useMessageStore()
  const [replyTo, setReplyTo] = useState<Message | null>(null)
  const [mentionTrigger, setMentionTrigger] = useState(0)
  const [mentionUserId, setMentionUserId] = useState('')

  // 全局 SSE 连接由 AppLayout 维护，此处无需再建立 WS

  useEffect(() => {
    if (conversationType && conversationId) {
      loadMessages(conversationType, conversationId)
      markConversationRead(conversationType, conversationId)
    }
  }, [conversationType, conversationId, loadMessages, markConversationRead])

  useEffect(() => {
    setReplyTo(null)
  }, [conversationType, conversationId])

  const hasConversation = !!(conversationType && conversationId)
  const currentUserId = user?.id || ''

  return (
    <div className="h-full flex">
      <ConversationList />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {hasConversation ? (
          <>
            <MessageArea
              conversationType={conversationType!}
              conversationId={conversationId!}
              currentUserId={currentUserId}
              onReply={setReplyTo}
              onRecall={recallMessage}
              onMention={(userId) => {
                setMentionUserId(userId)
                setMentionTrigger(t => t + 1)
              }}
            />
            <MessageInput
              conversationType={conversationType!}
              conversationId={conversationId!}
              currentUserId={currentUserId}
              replyTo={replyTo}
              onCancelReply={() => setReplyTo(null)}
              onMention={mentionUserId}
              mentionTrigger={mentionTrigger}
              groupId={conversationType === 'group' ? conversationId : undefined}
            />
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-400">
            <p className="text-lg">选择一个会话开始聊天</p>
          </div>
        )}
      </div>
    </div>
  )
}
