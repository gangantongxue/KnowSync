import { useState, useEffect, useRef } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'
import { messageApi, friendApi } from '../../lib/chat-api'

export default function TopBar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user } = useAuth()
  const [query, setQuery] = useState('')
  const [hasNotification, setHasNotification] = useState(false)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    const isMessagesPage = location.pathname.startsWith('/messages')
    if (isMessagesPage) return

    const fetchStatus = async () => {
      try {
        const [unreadResp, friendResp] = await Promise.all([
          messageApi.getUnreadCount(),
          friendApi.getReceivedRequests(),
        ])
        const counts = unreadResp.data?.counts || {}
        const totalUnread = Object.values(counts).reduce((a, b) => a + b, 0)
        const pendingRequests = (friendResp.data?.friend_requests || []).filter(r => r.status === 'pending').length
        const total = totalUnread + pendingRequests

        if (total > 0) {
          setHasNotification(true)
          if (intervalRef.current) {
            clearInterval(intervalRef.current)
            intervalRef.current = null
          }
        }
      } catch { /* ignore */ }
    }

    setHasNotification(false)
    fetchStatus()
    intervalRef.current = setInterval(fetchStatus, 30000)
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }
  }, [location.pathname])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      navigate(`/search?q=${encodeURIComponent(query.trim())}`)
    }
  }

  return (
    <header className="h-14 border-b border-gray-200 bg-white flex items-center px-4 gap-4 shrink-0">
      <button onClick={() => navigate('/chat')} className="shrink-0">
        <img src="/img/logo-wide.jpg" alt="KnowSync" className="h-8 w-auto" />
      </button>

      <form onSubmit={handleSearch} className="flex-1 flex justify-center">
        <div className="flex items-center max-w-xl w-full">
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="搜索知识库..."
            className="flex-1 px-3 py-1.5 border border-gray-300 rounded-l-lg text-sm focus:outline-none focus:border-emerald-400 bg-gray-50"
          />
          <button
            type="submit"
            className="px-4 py-1.5 bg-emerald-500 text-white rounded-r-lg text-sm hover:bg-emerald-600 shrink-0"
          >
            搜索
          </button>
        </div>
      </form>

      <div className="flex items-center gap-2 shrink-0">
        <button
          onClick={() => navigate('/messages')}
          className="relative w-9 h-9 rounded-lg hover:bg-gray-100 flex items-center justify-center"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-5 h-5 text-gray-600">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
            <polyline points="22,6 12,13 2,6" />
          </svg>
          {hasNotification && (
            <span className="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-red-500 rounded-full" />
          )}
        </button>

        <button
          onClick={() => navigate(`/users/${user?.id}`)}
          className="w-8 h-8 rounded-full bg-emerald-100 text-emerald-700 font-medium text-sm flex items-center justify-center hover:bg-emerald-200 overflow-hidden"
        >
          {user?.avatar ? (
            <img src={user.avatar} alt="" className="w-full h-full object-cover" />
          ) : (
            user?.name?.charAt(0) || 'U'
          )}
        </button>
      </div>
    </header>
  )
}
