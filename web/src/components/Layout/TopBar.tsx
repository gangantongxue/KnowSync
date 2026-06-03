import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'
import SettingsModal from '../Settings/SettingsModal'

export default function TopBar() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const [query, setQuery] = useState('')
  const [showDropdown, setShowDropdown] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setShowDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      navigate(`/search?q=${encodeURIComponent(query.trim())}`)
    }
  }

  const handleLogout = async () => {
    setShowDropdown(false)
    await logout()
    navigate('/login')
  }

  return (
    <>
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

        <div className="relative shrink-0" ref={dropdownRef}>
          <button
            onClick={() => setShowDropdown(!showDropdown)}
            className="w-8 h-8 rounded-full bg-emerald-100 text-emerald-700 font-medium text-sm flex items-center justify-center hover:bg-emerald-200 overflow-hidden"
          >
            {user?.avatar ? (
              <img src={user.avatar} alt="" className="w-full h-full object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }} />
            ) : (
              user?.name?.charAt(0) || 'U'
            )}
          </button>

          {showDropdown && (
            <div className="absolute right-0 top-10 w-36 bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50">
              <button
                onClick={() => { setShowDropdown(false); setShowSettings(true) }}
                className="w-full px-3 py-2 text-sm text-left hover:bg-gray-50"
              >
                设置
              </button>
              <button
                onClick={handleLogout}
                className="w-full px-3 py-2 text-sm text-left text-red-600 hover:bg-gray-50"
              >
                退出登录
              </button>
            </div>
          )}
        </div>
      </header>

      {showSettings && <SettingsModal onClose={() => setShowSettings(false)} />}
    </>
  )
}
