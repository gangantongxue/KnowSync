import { useState, useEffect, useRef } from 'react'
import { useChat } from '../../store/chat-context'

interface AskUserData {
  session_id: string
  question: string
  type: 'single' | 'multiple'
  options?: string[]
  has_other?: boolean
}

export default function AskUserModal() {
  const [data, setData] = useState<AskUserData | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [customText, setCustomText] = useState('')
  const [customSelected, setCustomSelected] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const { sendAskUserResponse, isStreaming } = useChat()

  useEffect(() => {
    const handler = (e: Event) => {
      setData((e as CustomEvent).detail as AskUserData)
      setSelected([])
      setCustomText('')
      setCustomSelected(false)
    }
    window.addEventListener('ask-user', handler)
    return () => window.removeEventListener('ask-user', handler)
  }, [])

  useEffect(() => {
    if (customSelected && inputRef.current) {
      inputRef.current.focus()
    }
  }, [customSelected])

  if (!data) return null

  const isMultiple = data.type === 'multiple'
  const options = data.options || []

  const toggleOption = (opt: string) => {
    setSelected(prev => {
      if (prev.includes(opt)) return prev.filter(o => o !== opt)
      if (isMultiple) return [...prev, opt]
      return [opt]
    })
    // 单选时选了普通选项，取消自定义选中
    if (!isMultiple) setCustomSelected(false)
  }

  const toggleCustom = () => {
    if (customSelected) {
      setCustomSelected(false)
    } else {
      setCustomSelected(true)
      if (!isMultiple) setSelected([])
    }
  }

  const handleSubmit = () => {
    if (isStreaming) return
    const parts: string[] = []
    if (customSelected && customText.trim()) {
      parts.push(customText.trim())
    }
    parts.push(...selected)
    if (parts.length === 0) return
    sendAskUserResponse(parts.join('\n'))
    setData(null)
  }

  const handleCancel = () => {
    if (!isStreaming) {
      sendAskUserResponse('用户取消了选择')
    }
    setData(null)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.nativeEvent.isComposing) {
      e.preventDefault()
      handleSubmit()
    }
  }

  const hasSelection = selected.length > 0 || (customSelected && customText.trim().length > 0)

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 max-w-md w-full mx-4 shadow-xl">
        <h3 className="text-sm font-medium text-gray-800 mb-1">{data.question}</h3>
        <p className="text-xs text-gray-400 mb-4">
          {isMultiple ? '可多选' : '单选'}，选择后点击确定发送
        </p>

        <div className="space-y-2 mb-4">
          {options.map((opt, i) => {
            const isSel = selected.includes(opt)
            return (
              <button
                key={opt + i}
                onClick={() => toggleOption(opt)}
                className={`w-full px-3 py-2 text-sm text-left border rounded-lg transition-colors ${
                  isSel
                    ? 'border-emerald-400 bg-emerald-50 text-emerald-700'
                    : 'border-gray-200 hover:bg-gray-50 hover:border-emerald-300'
                }`}
              >
                <span className="flex items-center gap-2">
                  {isMultiple && (
                    <span className={`w-4 h-4 rounded-sm border flex items-center justify-center shrink-0 ${
                      isSel ? 'bg-emerald-400 border-emerald-400 text-white text-xs' : 'border-gray-300'
                    }`}>
                      {isSel ? '✓' : ''}
                    </span>
                  )}
                  {opt}
                </span>
              </button>
            )
          })}

          {/* 自由输入行 */}
          {data.has_other && (
            <div>
              <button
                onClick={toggleCustom}
                className={`w-full px-3 py-2 text-sm text-left border rounded-lg transition-colors ${
                  customSelected
                    ? 'border-emerald-400 bg-emerald-50 text-emerald-700'
                    : 'border-gray-200 hover:bg-gray-50 hover:border-emerald-300'
                }`}
              >
                <span className="flex items-center gap-2">
                  {isMultiple && (
                    <span className={`w-4 h-4 rounded-sm border flex items-center justify-center shrink-0 ${
                      customSelected ? 'bg-emerald-400 border-emerald-400 text-white text-xs' : 'border-gray-300'
                    }`}>
                      {customSelected ? '✓' : ''}
                    </span>
                  )}
                  其他（自由输入）
                </span>
              </button>
              {customSelected && (
                <input
                  ref={inputRef}
                  type="text"
                  value={customText}
                  onChange={e => setCustomText(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="请输入你的回答"
                  className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400"
                />
              )}
            </div>
          )}
        </div>

        <div className="flex gap-2">
          <button
            onClick={handleSubmit}
            disabled={!hasSelection || isStreaming}
            className="flex-1 px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600 disabled:opacity-50 transition-colors"
          >
            确定
          </button>
          <button onClick={handleCancel} className="px-4 py-2 border border-gray-200 text-gray-500 rounded-lg text-sm hover:bg-gray-50">
            取消
          </button>
        </div>
      </div>
    </div>
  )
}
