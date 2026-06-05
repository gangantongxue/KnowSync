import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './store/auth-context'
import { ChatProvider } from './store/chat-context'
import { MessageProvider } from './store/message-store'
import { RepoProvider } from './store/repo-context'
import PublicLayout from './components/Layout/PublicLayout'
import AppLayout from './components/Layout/AppLayout'
import Login from './pages/Login'
import Register from './pages/Register'
import Chat from './pages/Chat'
import RepoList from './pages/RepoList'
import RepoDetail from './pages/RepoDetail'
import ArticleView from './pages/ArticleView'
import ArticleEditor from './pages/ArticleEditor'
import SearchResult from './pages/SearchResult'
import Messages from './pages/Messages'
import UserProfile from './pages/UserProfile'
import FriendDetail from './pages/FriendDetail'
import NotFound from './pages/NotFound'

export default function App() {
  return (
    <AuthProvider>
      <ChatProvider>
        <MessageProvider>
          <RepoProvider>
          <Routes>
            {/* 公开页面 — PublicLayout */}
            <Route element={<PublicLayout />}>
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
            </Route>

            {/* 登录页面入口 */}
            <Route path="/" element={<RootRedirect />} />

            {/* 登录后页面 — AppLayout */}
            <Route element={<AppLayout />}>
              <Route path="/repos" element={<RepoList />} />
              {/* 仓库页面包含文件树 + 设置面板，子路由用于文章查看/编辑 */}
              <Route path="/repos/:repoId" element={<RepoDetail />}>
                <Route path="view/*" element={<ArticleView />} />
                <Route path="edit/*" element={<ArticleEditor />} />
              </Route>
              <Route path="/chat" element={<Chat />} />
              <Route path="/chat/:sessionId" element={<Chat />} />
              <Route path="/search" element={<SearchResult />} />
              <Route path="/messages" element={<Messages />} />
              <Route path="/messages/:conversationType/:conversationId" element={<Messages />} />
              <Route path="/users/:userId" element={<UserProfile />} />
              <Route path="/friends/:friendId" element={<FriendDetail />} />
            </Route>

            <Route path="*" element={<NotFound />} />
          </Routes>
        </RepoProvider>
        </MessageProvider>
      </ChatProvider>
    </AuthProvider>
  )
}

function RootRedirect() {
  const token = localStorage.getItem('access_token')
  if (token) return <Navigate to="/chat" replace />
  return <Navigate to="/login" replace />
}
