import { useEffect } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'
import { useMessageStore } from '../../store/message-store'
import { useRepo } from '../../store/repo-context'
import sseClient from '../../lib/sse-client'
import TopBar from './TopBar'
import Sidebar from './Sidebar'

export default function AppLayout() {
  const { isAuthenticated, isLoading, user } = useAuth()
  const { loadRepos } = useRepo()
  const { handlePushEvent, setSseConnected, loadConversations, loadFriendRequests } = useMessageStore()
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

  // 全局 SSE 连接：登录后立即建立（不限页面），token/用户变化时重建，退出登录时断开。
  // 新消息、撤回、好友申请等事件全部经由该连接推送。
  useEffect(() => {
    if (!isAuthenticated || isLoading) return
    const token = localStorage.getItem('access_token')
    if (!token) return

    sseClient.connect(token, (event) => {
      handlePushEvent(event)
    })
    setSseConnected(true)
    // 全局初始化未读数与好友申请，支撑 TopBar 红点的初始状态
    loadConversations()
    loadFriendRequests()

    return () => {
      sseClient.disconnect()
      setSseConnected(false)
    }
  }, [isAuthenticated, isLoading, user, handlePushEvent, setSseConnected, loadConversations, loadFriendRequests])

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
