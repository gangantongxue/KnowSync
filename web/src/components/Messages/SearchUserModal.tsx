import { useState, useEffect, useRef } from 'react'
import { Input, message } from 'antd'
import { useMessageStore } from '../../store/message-store'
import { friendApi } from '../../lib/chat-api'
import type { SearchUserInfo } from '../../lib/chat-api'

interface SearchUserModalProps {
  onClose: () => void
}

export default function SearchUserModal({ onClose }: SearchUserModalProps) {
  const { sendFriendRequest } = useMessageStore()
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchUserInfo[]>([])
  const [sentMap, setSentMap] = useState<Record<string, boolean>>({})
  const [addingId, setAddingId] = useState<string | null>(null)
  const [remarkInput, setRemarkInput] = useState('')
  const [remarkTarget, setRemarkTarget] = useState<SearchUserInfo | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!query.trim()) {
      setResults([])
      return
    }
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(async () => {
      try {
        const res = await friendApi.searchUsers(query.trim())
        setResults(res.data.users)
      } catch {
        message.error('搜索用户失败')
      }
    }, 300)
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [query])

  const handleSendRequest = async () => {
    if (!remarkTarget) return
    setAddingId(remarkTarget.id)
    try {
      await sendFriendRequest(remarkTarget.id, remarkInput)
      message.success('好友请求已发送')
      setSentMap(prev => ({ ...prev, [remarkTarget.id]: true }))
      setRemarkTarget(null)
      setRemarkInput('')
    } catch {
      message.error('发送失败')
    } finally {
      setAddingId(null)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-4 border-b border-gray-100">
          <h2 className="text-base font-medium text-gray-800">搜索用户</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg leading-none">✕</button>
        </div>

        <div className="p-4 border-b border-gray-100">
          <Input
            placeholder="输入用户名或ID搜索"
            value={query}
            onChange={e => setQuery(e.target.value)}
            autoFocus
          />
        </div>

        <div className="flex-1 overflow-y-auto px-4">
          {results.length === 0 ? (
            <div className="text-center py-8 text-sm text-gray-400">
              {query.trim() ? '未找到匹配的用户' : '输入关键词开始搜索'}
            </div>
          ) : (
            results.map(user => (
              <div key={user.id} className="flex items-center justify-between py-3 border-b border-gray-50">
                <div className="flex items-center gap-3 min-w-0">
                  {user.avatar ? (
                    <img src={user.avatar} alt="" className="w-10 h-10 rounded-full object-cover shrink-0" />
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-purple-100 text-purple-600 flex items-center justify-center text-sm font-medium shrink-0">
                      {user.name.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <div className="min-w-0">
                    <div className="text-sm text-gray-800 truncate">{user.name}</div>
                  </div>
                </div>
                {sentMap[user.id] ? (
                  <span className="text-xs text-green-600 shrink-0">已发送</span>
                ) : remarkTarget?.id === user.id ? (
                  <div className="flex items-center gap-2 shrink-0" onClick={e => e.stopPropagation()}>
                    <input
                      value={remarkInput}
                      onChange={e => setRemarkInput(e.target.value)}
                      placeholder="添加备注（可选）"
                      className="w-28 px-2 py-1 text-xs border border-gray-300 rounded focus:outline-none focus:border-blue-400"
                      onKeyDown={e => { if (e.key === 'Enter') handleSendRequest() }}
                      autoFocus
                    />
                    <button
                      onClick={handleSendRequest}
                      disabled={addingId === user.id}
                      className="px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                    >
                      {addingId === user.id ? '发送中' : '发送'}
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => { setRemarkTarget(user); setRemarkInput('') }}
                    className="px-3 py-1 text-xs bg-blue-500 text-white rounded-lg hover:bg-blue-600 shrink-0"
                  >
                    添加好友
                  </button>
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
