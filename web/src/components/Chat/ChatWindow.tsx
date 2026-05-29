import { useState, useRef, useEffect } from 'react'
import { useChat } from '../../store/chat-context'
import MessageBubble from './MessageBubble'
import AskUserModal from './AskUserModal'

interface ChatWindowProps {
}

export default function ChatWindow(_props: ChatWindowProps) {
  const { messages, isStreaming, sendMessage } = useChat()
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSubmit = () => {
    const trimmed = input.trim()
    if (!trimmed || isStreaming) return
    sendMessage(trimmed)
    setInput('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 200)}px`
  }

  const hasMessages = messages.length > 0

  return (
    <>
      <div className="flex-1 overflow-y-auto px-4 py-4">
        {hasMessages ? (
          messages.map(msg => (
            <MessageBubble
              key={msg.id}
              role={msg.role}
              content={msg.content}
              thinking={msg.thinking}
              isStreaming={msg.isStreaming}
            />
          ))
        ) : (
          <div className="h-full flex flex-col items-center justify-center text-center text-gray-400">
            <div className="text-4xl mb-3">💡</div>
            <h2 className="text-lg font-medium text-gray-700 mb-2">开始与 KnowSync 对话</h2>
            <p className="text-sm mb-6 max-w-md">AI 知识助手，帮你快速找到所需信息</p>
            <div className="space-y-2 w-64">
              {['搜索某篇文章', '帮我总结某个知识库', '解释某个概念'].map(q => (
                <button
                  key={q}
                  onClick={() => sendMessage(q)}
                  className="w-full px-4 py-2 text-sm border border-gray-200 rounded-lg hover:border-emerald-300 hover:text-emerald-600 transition-colors"
                >
                  {q}
                </button>
              ))}
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="border-t border-gray-200 p-3">
        <div className="flex gap-2 items-end">
          <textarea
            ref={textareaRef}
            value={input}
            onChange={handleInput}
            onKeyDown={handleKeyDown}
            placeholder="输入消息 (Enter 发送, Shift+Enter 换行)"
            rows={1}
            className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm resize-none focus:outline-none focus:border-emerald-400 max-h-[200px]"
          />
          <button
            onClick={handleSubmit}
            disabled={!input.trim() || isStreaming}
            className="px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600 disabled:opacity-50 shrink-0"
          >
            发送
          </button>
        </div>
      </div>

      <AskUserModal />
    </>
  )
}
