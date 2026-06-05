import { useRef, useState, useEffect, type KeyboardEvent, type ChangeEvent } from 'react'
import { Button, message as antMessage } from 'antd'
import { PictureOutlined, FileOutlined } from '@ant-design/icons'
import { useMessageStore } from '../../store/message-store'
import { messageApi } from '../../lib/chat-api'
import type { Message } from '../../lib/chat-api'
import MentionDropdown from './MentionDropdown'

function extractMentions(content: string): string[] {
  const regex = /@(\S+)/g
  const ids: string[] = []
  let match: RegExpExecArray | null
  while ((match = regex.exec(content)) !== null) {
    if (!ids.includes(match[1])) {
      ids.push(match[1])
    }
  }
  return ids
}

interface MessageInputProps {
  conversationType: string
  conversationId: string
  currentUserId: string
  replyTo: Message | null
  onCancelReply: () => void
  onMention?: string
  mentionTrigger?: number
  groupId?: string
  groupOwnerId?: string
}

export default function MessageInput({ conversationType, conversationId, currentUserId, replyTo, onCancelReply, onMention, mentionTrigger, groupId, groupOwnerId }: MessageInputProps) {
  const { sendMessage, sendingMessage } = useMessageStore()
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const imageInputRef = useRef<HTMLInputElement>(null)
  const [text, setText] = useState('')
  const [showMention, setShowMention] = useState(false)
  const [mentionSearch, setMentionSearch] = useState('')

  useEffect(() => {
    if (replyTo) {
      textareaRef.current?.focus()
    }
  }, [replyTo])

  useEffect(() => {
    if (mentionTrigger && mentionTrigger > 0 && onMention) {
      setText(prev => prev + `@${onMention} `)
      requestAnimationFrame(() => {
        textareaRef.current?.focus()
      })
    }
  }, [mentionTrigger])

  const handleInput = (e: ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value
    setText(value)

    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 150)}px`

    if (groupId && value.includes('@')) {
      const lastAtIndex = value.lastIndexOf('@')
      const afterAt = value.slice(lastAtIndex + 1)
      if (!afterAt.includes(' ') && !afterAt.includes('\n')) {
        setShowMention(true)
        setMentionSearch(afterAt)
        return
      }
    }
    setShowMention(false)
  }

  const handleMentionSelect = (userId: string) => {
    const el = textareaRef.current
    if (!el) return
    const cursorPos = el.selectionStart
    const lastAtIndex = text.lastIndexOf('@', cursorPos - 1)
    if (lastAtIndex === -1) return
    const before = text.slice(0, lastAtIndex)
    const after = text.slice(cursorPos)
    const newText = before + `@${userId} ` + after
    setText(newText)
    setShowMention(false)
    requestAnimationFrame(() => {
      el.focus()
      const newPos = lastAtIndex + userId.length + 3
      el.setSelectionRange(newPos, newPos)
    })
  }

  const handleSend = async () => {
    const trimmed = text.trim()
    if (!trimmed || sendingMessage) return

    if (replyTo) {
      try {
        if (conversationType === 'private') {
          const parts = conversationId.split('_')
          const receiverId = parts.length === 2 ? (parts[0] === currentUserId ? parts[1] : parts[0]) : conversationId
          await messageApi.sendPrivate({
            receiver_id: receiverId,
            content_type: 'text',
            content: trimmed,
            reply_to_id: replyTo.id,
          })
        } else {
          const mentions = extractMentions(trimmed)
          await messageApi.sendGroup({
            group_id: conversationId,
            content_type: 'text',
            content: trimmed,
            reply_to_id: replyTo.id,
            mentions,
          })
        }
      } catch {
        antMessage.error('发送失败')
        return
      }
      onCancelReply()
    } else {
      await sendMessage(conversationType, conversationId, trimmed, 'text', currentUserId)
    }

    setText('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey && !e.nativeEvent.isComposing) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleImageUpload = () => {
    imageInputRef.current?.click()
  }

  const handleFileUpload = () => {
    fileInputRef.current?.click()
  }

  const handleImageChange = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      const base64 = reader.result as string
      sendMessage(conversationType, conversationId, base64, 'image', currentUserId)
    }
    reader.readAsDataURL(file)
    e.target.value = ''
  }

  const handleFileChange = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = () => {
      const base64 = reader.result as string
      const extra = JSON.stringify({ filename: file.name, size: file.size })
      try {
        messageApi.sendGroup({
          group_id: conversationId,
          content_type: 'file',
          content: base64,
          extra,
        }).then(() => {
          // message sent via API, need to refresh
        }).catch(() => antMessage.error('文件发送失败'))
      } catch {
        antMessage.error('文件发送失败')
      }
    }
    reader.readAsDataURL(file)
    e.target.value = ''
  }

  return (
    <div className="border-t border-gray-200 bg-white px-4 py-3">
      {replyTo && (
        <div className="flex items-center gap-2 mb-2 px-3 py-1.5 bg-blue-50 rounded-lg text-sm">
          <span className="text-gray-500">回复:</span>
          <span className="text-gray-700 truncate flex-1">{replyTo.content}</span>
          <button
            onClick={onCancelReply}
            className="text-gray-400 hover:text-gray-600 text-xs shrink-0"
          >
            取消
          </button>
        </div>
      )}

      <div className="flex gap-2 items-end relative">
        <div className="flex-1 relative">
          <textarea
            ref={textareaRef}
            value={text}
            onChange={handleInput}
            onKeyDown={handleKeyDown}
            placeholder="输入消息 (Enter 发送, Shift+Enter 换行)"
            rows={1}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm resize-none focus:outline-none focus:border-blue-400 min-h-[40px] max-h-[150px]"
          />
          {showMention && groupId && groupOwnerId && (
            <MentionDropdown
              groupId={groupId}
              ownerId={groupOwnerId}
              onSelect={handleMentionSelect}
              searchText={mentionSearch}
            />
          )}
        </div>

        <div className="flex items-center gap-1">
          <input
            ref={imageInputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={handleImageChange}
          />
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            onChange={handleFileChange}
          />
          <Button
            type="text"
            icon={<PictureOutlined />}
            onClick={handleImageUpload}
            className="text-gray-500"
          />
          <Button
            type="text"
            icon={<FileOutlined />}
            onClick={handleFileUpload}
            className="text-gray-500"
          />
          {groupId && (
            <Button
              type="text"
              className="text-gray-500 font-bold"
              onClick={() => {
                setText(prev => prev + '@')
                setShowMention(true)
                setMentionSearch('')
                textareaRef.current?.focus()
              }}
            >
              @
            </Button>
          )}
          <Button
            type="primary"
            size="small"
            disabled={!text.trim() || sendingMessage}
            loading={sendingMessage}
            onClick={handleSend}
          >
            发送
          </Button>
        </div>
      </div>
    </div>
  )
}
