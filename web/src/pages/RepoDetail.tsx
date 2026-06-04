import { useEffect, useState, useCallback, useMemo } from 'react'
import { Outlet, useParams, useLocation, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { getRepo, followRepo, unfollowRepo } from '../lib/repos'
import type { Repo } from '../lib/repos'
import { useAuth } from '../store/auth-context'
import RepoTree from '../components/Repo/RepoTree'
import RepoSettings from '../components/Repo/RepoSettings'
import { resolveImageUrls, markdownComponents } from '../lib/markdown'
import { userApi } from '../lib/user-api'

/** 仓库详情页向外暴露的上下文 */
export interface RepoDetailContext {
  refreshTree: () => void
}

/** 默认帮助信息（Markdown 格式） */
const DEFAULT_HELP = `# 快速上手指南

此知识库暂无 README.md 文件。

## 创建文章

- 点击左侧文件树上方的 **📄+** 按钮，或在文件夹上右键选择「新建文章」
- 输入文件名称（含 \`.md\` 后缀）

## 管理文件

- 右键单击文件或文件夹进行**重命名**、**删除**等操作
- 文件夹支持**展开/折叠**，可将文件拖拽到目标文件夹

## 编辑文章

- 点击文件树中的文件即可**查看**
- 在查看页面点击「编辑」进入编辑模式，支持实时预览
`

export default function RepoDetail() {
  const { repoId } = useParams<{ repoId: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [repo, setRepo] = useState<Repo | null>(null)
  const [myRole, setMyRole] = useState('')
  const [isFollowing, setIsFollowing] = useState(false)
  const [followLoading, setFollowLoading] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [readmeContent, setReadmeContent] = useState('')
  const [readmeLoading, setReadmeLoading] = useState(false)
  const [readmeExists, setReadmeExists] = useState(false)
  const [treeRefreshKey, setTreeRefreshKey] = useState(0)
  const [ownerInfo, setOwnerInfo] = useState<{ id: string; name: string; avatar: string } | null>(null)

  const refreshTree = useCallback(() => setTreeRefreshKey(k => k + 1), [])
  const outletContext = useMemo<RepoDetailContext>(() => ({ refreshTree }), [refreshTree])

  const readOnly = myRole === '' || myRole === 'COLLABORATOR_ROLE_UNSPECIFIED'
  const isOwner = user?.id === repo?.owner_id

  const loadReadme = useCallback(async (ownerId: string) => {
    if (!repoId) return
    setReadmeLoading(true)
    try {
      const res = await fetch(`/files/${ownerId}/${repoId}/README.md`)
      if (res.ok) {
        const text = await res.text()
        const baseUrl = `/files/${ownerId}/${repoId}`
        setReadmeContent(resolveImageUrls(text, '', baseUrl))
        setReadmeExists(true)
      } else {
        setReadmeExists(false)
      }
    } catch {
      setReadmeExists(false)
    }
    setReadmeLoading(false)
  }, [repoId])

  const loadRepo = useCallback(async () => {
    if (!repoId) return
    try {
      const r = await getRepo(repoId)
      setRepo(r)
      setMyRole(r.my_role)
      setIsFollowing(r.is_following)
      await loadReadme(r.owner_id)
      // 获取拥有者信息
      try {
        const profileRes = await userApi.getProfile(r.owner_id)
        setOwnerInfo({ id: profileRes.data.user.id, name: profileRes.data.user.name, avatar: profileRes.data.user.avatar })
      } catch { /* ignore */ }
    } catch { /* ignore */ }
  }, [repoId, loadReadme])

  useEffect(() => { loadRepo() }, [loadRepo])

  const handleFollow = async () => {
    if (!repoId) return
    setFollowLoading(true)
    try {
      await followRepo(repoId)
      setIsFollowing(true)
      if (repo) {
        setRepo({ ...repo, follower_count: repo.follower_count + 1 })
      }
    } catch { /* ignore */ }
    setFollowLoading(false)
  }

  const handleUnfollow = async () => {
    if (!repoId) return
    setFollowLoading(true)
    try {
      await unfollowRepo(repoId)
      setIsFollowing(false)
      if (repo && repo.follower_count > 0) {
        setRepo({ ...repo, follower_count: repo.follower_count - 1 })
      }
    } catch { /* ignore */ }
    setFollowLoading(false)
  }

  if (!repoId) return null

  const isChildRoute = /\/repos\/[^/]+\/(view|edit)\//.test(location.pathname)

  return (
    <div className="h-full flex overflow-hidden">
      {/* 左侧文件树 */}
      <div className="w-60 border-r border-gray-200 bg-gray-50 overflow-y-auto shrink-0 flex flex-col">
        {/* 文件树头部：知识库名称 + 关注按钮 + 设置按钮 */}
        {repo && (
          <div className="px-3 py-2 border-b border-gray-200 shrink-0">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-700 truncate">{repo.name}</span>
              <div className="flex items-center gap-1 shrink-0 ml-1">
                {!isOwner && (isFollowing ? (
                  <button
                    onClick={handleUnfollow}
                    disabled={followLoading}
                    className="text-xs px-2 py-0.5 rounded border border-gray-300 text-gray-500 hover:text-red-500 hover:border-red-300 disabled:opacity-50"
                  >
                    {followLoading ? '...' : '取消关注'}
                  </button>
                ) : (
                  <button
                    onClick={handleFollow}
                    disabled={followLoading}
                    className="text-xs px-2 py-0.5 rounded border border-emerald-300 text-emerald-600 hover:bg-emerald-50 disabled:opacity-50"
                  >
                    {followLoading ? '...' : '关注'}
                  </button>
                ))}
                {/* 被关注数 */}
                {repo.follower_count > 0 && (
                  <span className="text-xs text-gray-400">{repo.follower_count} 关注</span>
                )}
                {/* 设置按钮仅协作者可见 */}
                {!readOnly && (
                  <button
                    onClick={() => setShowSettings(true)}
                    className="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-200 text-gray-400 hover:text-gray-600"
                    title="知识库设置"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
                  </button>
                )}
              </div>
            </div>
          </div>
        )}
        <div className="flex-1 overflow-y-auto">
          <RepoTree repoId={repoId} repo={repo} refreshKey={treeRefreshKey} readOnly={readOnly} />
        </div>
        {/* 拥有者信息 */}
        {ownerInfo && (
          <div
            onClick={() => navigate(`/users/${ownerInfo.id}`)}
            className="flex items-center gap-2 px-3 py-2 border-t border-gray-200 shrink-0 cursor-pointer hover:bg-gray-100"
          >
            <div className="w-6 h-6 rounded-full bg-emerald-100 text-emerald-700 text-xs flex items-center justify-center overflow-hidden shrink-0">
              {ownerInfo.avatar ? (
                <img src={ownerInfo.avatar} alt="" className="w-full h-full object-cover" />
              ) : (
                ownerInfo.name?.charAt(0) || '?'
              )}
            </div>
            <span className="text-xs text-gray-500 truncate">{ownerInfo.name}</span>
          </div>
        )}
      </div>

      {/* 中间内容区 */}
      <div className="flex-1 overflow-y-auto min-w-0">
        <Outlet context={outletContext} />
        {!isChildRoute && (
          <div className="p-6 max-w-3xl mx-auto">
            {readmeLoading ? (
              <div className="text-gray-400 text-sm text-center py-12">加载中...</div>
            ) : readmeExists ? (
              <div>
                <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                  {readmeContent}
                </ReactMarkdown>
              </div>
            ) : (
              <div>
                <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                  {DEFAULT_HELP}
                </ReactMarkdown>
              </div>
            )}
          </div>
        )}
      </div>

      {/* 设置弹窗 */}
      {showSettings && repo && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setShowSettings(false)}>
          <div className="bg-white rounded-xl shadow-xl w-[480px] max-h-[85vh] overflow-y-auto mx-4" onClick={e => e.stopPropagation()}>
            <RepoSettings repo={repo} myRole={myRole} onUpdate={loadRepo} onClose={() => setShowSettings(false)} onDelete={() => { setShowSettings(false); navigate('/repos') }} />
          </div>
        </div>
      )}
    </div>
  )
}
