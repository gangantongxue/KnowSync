import { useEffect } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'
import { useRepo } from '../../store/repo-context'
import TopBar from './TopBar'
import Sidebar from './Sidebar'

export default function AppLayout() {
  const { isAuthenticated, isLoading } = useAuth()
  const { loadRepos } = useRepo()
  const location = useLocation()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      navigate('/login', { replace: true })
    }
  }, [isLoading, isAuthenticated, navigate])

  useEffect(() => {
    loadRepos()
  }, [loadRepos])

  const isInsideRepo = /^\/repos\/[^/]/.test(location.pathname)
  const isSearchRoute = location.pathname.startsWith('/search')
  const isMessagesRoute = location.pathname.startsWith('/messages')

  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center text-gray-400">
        加载中...
      </div>
    )
  }

  if (!isAuthenticated) return null

  // 全屏页面 — 无侧边栏
  if (isSearchRoute || isMessagesRoute) {
    return (
      <div className="h-screen flex flex-col">
        <TopBar />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    )
  }

  return (
    <div className="h-screen flex flex-col">
      <TopBar />
      <div className="flex-1 flex overflow-hidden">
        {!isInsideRepo && <Sidebar />}
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
