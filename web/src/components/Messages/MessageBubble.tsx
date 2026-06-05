import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Popover, Button, message as antMessage } from 'antd'
import type { Message } from '../../lib/chat-api'
import InvitationCard from './InvitationCard'
import ForwardModal from './ForwardModal'

interface MessageBubbleProps {
  message: Message
  isOwn: boolean
  senderName: string
  senderAvatar: string
  repliedMessage: Message | null
  currentUserId: string
  onReply: (msg: Message) => void
  onRecall: (msgId: string) => void
  onMention?: (userId: string) => void
  onJumpToMessage?: (messageId: string) => void
  highlight?: boolean
  mentionNames?: Record<string, string>
}

function formatTime(ts: number): string {
  const date = new Date(ts * 1000)
  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

const PASTEL_BG = [
  'bg-red-100', 'bg-blue-100', 'bg-green-100', 'bg-yellow-100',
  'bg-purple-100', 'bg-pink-100', 'bg-indigo-100', 'bg-teal-100',
]

function getAvatarColor(id: string): string {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash)
  }
  return PASTEL_BG[Math.abs(hash) % PASTEL_BG.length]
}

function parseMentions(extra: string | null): string[] {
  if (!extra) return []
  try {
    const data = JSON.parse(extra)
    return Array.isArray(data.mentions) ? data.mentions : []
  } catch {
    return []
  }
}

function renderTextWithMentions(content: string, mentions: string[], currentUserId: string, mentionNames: Record<string, string>, onMention?: (userId: string) => void): React.ReactNode[] {
  const parts: React.ReactNode[] = []
  const regex = /@(\S+)/g
  let lastIndex = 0
  let match: RegExpExecArray | null

  while ((match = regex.exec(content)) !== null) {
    if (match.index > lastIndex) {
      parts.push(<span key={`t${lastIndex}`}>{content.slice(lastIndex, match.index)}</span>)
    }
    const userId = match[1]
    const displayName = mentionNames[userId] || userId
    if (mentions.includes(userId)) {
      const isSelf = userId === currentUserId
      parts.push(
        <span
          key={`m${match.index}`}
          className={`cursor-pointer ${isSelf ? 'bg-yellow-200 text-blue-600 px-0.5 rounded' : 'text-blue-500'}`}
          onClick={() => onMention?.(userId)}
        >
          @{displayName}
        </span>
      )
    } else {
      parts.push(<span key={`p${match.index}`}>@{displayName}</span>)
    }
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < content.length) {
    parts.push(<span key={`t${lastIndex}`}>{content.slice(lastIndex)}</span>)
  }
  return parts.length > 0 ? parts : [<span key="all">{content}</span>]
}

export default function MessageBubble({ message, isOwn, senderName, senderAvatar, repliedMessage, currentUserId, onReply, onRecall, onMention, onJumpToMessage, highlight, mentionNames }: MessageBubbleProps) {
  const navigate = useNavigate()
  const [imgError, setImgError] = useState(false)
  const [showForward, setShowForward] = useState(false)

  const mentions = parseMentions(message.extra)
  const isMentionedSelf = mentions.includes(currentUserId)

  if (message.status === 'recalled') {
    return (
      <div className={`flex ${isOwn ? 'justify-end' : 'justify-start'} mb-3`}>
        <div className="text-xs text-gray-400 italic bg-gray-50 rounded-lg px-3 py-2">
          消息已撤回
        </div>
      </div>
    )
  }

  const renderContent = () => {
    switch (message.content_type) {
      case 'text':
        return (
          <p className="whitespace-pre-wrap break-words">
            {renderTextWithMentions(message.content, mentions, currentUserId, mentionNames || {}, onMention)}
          </p>
        )
      case 'image':
        return (
          <div className="max-w-xs">
            {!imgError ? (
              <img
                src={message.content}
                alt=""
                className="max-w-full rounded-lg cursor-pointer"
                onClick={() => window.open(message.content, '_blank')}
                onError={() => setImgError(true)}
              />
            ) : (
              <div className="text-gray-400 text-sm">图片加载失败</div>
            )}
          </div>
        )
      case 'file':
        return (
          <div className="flex items-center gap-2 bg-gray-50 rounded-lg px-3 py-2 cursor-pointer hover:bg-gray-100"
            onClick={() => window.open(message.content, '_blank')}
          >
            <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <div className="text-sm text-blue-600 truncate max-w-[200px]">
              {message.content.split('/').pop() || '文件'}
            </div>
          </div>
        )
      case 'system_invitation': {
        let repoName = '未知知识库'
        let role = 'viewer'
        let extra = message.extra
        try {
          const data = JSON.parse(message.content)
          repoName = data.repo_name || repoName
          role = data.role || role
        } catch {
          // use defaults
        }
        return <InvitationCard repoName={repoName} role={role} extra={extra} />
      }
      default:
        return <p className="whitespace-pre-wrap">{message.content}</p>
    }
  }

  const toolbarItems: { label: string; handler: () => void }[] = [
    { label: '回复', handler: () => onReply(message) },
    { label: '转发', handler: () => setShowForward(true) },
    {
      label: '复制',
      handler: () => {
        navigator.clipboard.writeText(message.content).then(() => {
          antMessage.success('已复制')
        }).catch(() => {})
      },
    },
  ]

  if (isOwn && message.status !== 'recalled') {
    const now = Date.now() / 1000
    const msgTime = message.created_at
    if (now - msgTime < 120) {
      toolbarItems.push({ label: '撤回', handler: () => onRecall(message.id) })
    }
  }

  const toolbar = (
    <div className="flex gap-1">
      {toolbarItems.map(item => (
        <Button key={item.label} size="small" type="text" className="text-xs" onClick={item.handler}>
          {item.label}
        </Button>
      ))}
    </div>
  )

  const bubbleBg = isOwn
    ? 'bg-blue-500 text-white rounded-tr-sm'
    : isMentionedSelf
      ? 'bg-yellow-100 text-gray-800 rounded-tl-sm'
      : 'bg-gray-100 text-gray-800 rounded-tl-sm'

  return (
    <>
      <div id={`msg-${message.id}`} className={`flex mb-3 gap-2 ${isOwn ? 'flex-row-reverse' : ''} ${highlight ? 'animate-message-flash' : ''}`}>
        {!isOwn && (
          <div
            className="cursor-pointer shrink-0"
            onClick={() => navigate(`/friends/${message.sender_id}`)}
          >
            {senderAvatar ? (
              <img src={senderAvatar} alt="" className="w-8 h-8 rounded-full object-cover" />
            ) : (
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm text-white ${getAvatarColor(message.sender_id)}`}>
                {senderName.charAt(0).toUpperCase()}
              </div>
            )}
          </div>
        )}

        <div className={`max-w-[70%] ${isOwn ? 'items-end' : 'items-start'}`}>
          {!isOwn && (
            <div
              className={`text-xs text-gray-500 mb-1 ml-1 ${onMention ? 'cursor-pointer hover:text-blue-500' : ''}`}
              onClick={() => onMention?.(message.sender_id)}
            >
              {senderName}
            </div>
          )}

          {repliedMessage && (
            <div
              className={`text-xs px-3 py-1 rounded-t-lg truncate max-w-full cursor-pointer hover:opacity-80 ${isOwn ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-200 text-gray-500'}`}
              onClick={() => onJumpToMessage?.(repliedMessage.id)}
            >
              回复: {repliedMessage.content}
            </div>
          )}

          <Popover content={toolbar} trigger="hover" placement={isOwn ? 'left' : 'right'}>
            <div className={`px-3 py-2 text-sm leading-relaxed rounded-lg ${bubbleBg}`}>
              {renderContent()}
              <div className={`text-xs mt-1 ${isOwn ? 'text-blue-200' : isMentionedSelf ? 'text-gray-500' : 'text-gray-400'}`}>
                {formatTime(message.created_at)}
              </div>
            </div>
          </Popover>
        </div>
      </div>

      {showForward && <ForwardModal message={message} onClose={() => setShowForward(false)} />}
    </>)
}
