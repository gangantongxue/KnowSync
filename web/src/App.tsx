import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './store/auth-context'
import { ChatProvider } from './store/chat-context'
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
import NotFound from './pages/NotFound'

export default function App() {
  return (
    <AuthProvider>
      <ChatProvider>
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
              <Route path="/repos/:repoId" element={<RepoDetail />} />
              <Route path="/repos/:repoId/nodes/:nodeId" element={<ArticleView />} />
              <Route path="/repos/:repoId/nodes/:nodeId/edit" element={<ArticleEditor />} />
              <Route path="/chat" element={<Chat />} />
              <Route path="/chat/:sessionId" element={<Chat />} />
              <Route path="/search" element={<SearchResult />} />
            </Route>

            <Route path="*" element={<NotFound />} />
          </Routes>
        </RepoProvider>
      </ChatProvider>
    </AuthProvider>
  )
}

function RootRedirect() {
  const token = localStorage.getItem('access_token')
  if (token) return <Navigate to="/chat" replace />
  return <Navigate to="/login" replace />
}
