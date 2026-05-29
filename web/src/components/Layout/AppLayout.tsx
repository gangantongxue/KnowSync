import { useEffect } from 'react'
import { Outlet, useParams, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'
import { useRepo } from '../../store/repo-context'
import TopBar from './TopBar'
import Sidebar from './Sidebar'

export default function AppLayout() {
  const { isAuthenticated, isLoading } = useAuth()
  const { loadRepos } = useRepo()
  const { repoId } = useParams()
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

  const isRepoRoute = location.pathname.startsWith('/repos')
  const isSearchRoute = location.pathname.startsWith('/search')

  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center text-gray-400">
        加载中...
      </div>
    )
  }

  if (!isAuthenticated) return null

  // 搜索页 — 全屏，无侧边栏
  if (isSearchRoute) {
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
        <Sidebar
          isFileTree={isRepoRoute}
          repoId={repoId}
          onBack={() => navigate('/chat')}
        />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
