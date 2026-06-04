import { useEffect, useState, useCallback, useMemo } from 'react'
import { Outlet, useParams, useLocation, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { getRepo } from '../lib/repos'
import type { Repo } from '../lib/repos'
import RepoTree from '../components/Repo/RepoTree'
import RepoSettings from '../components/Repo/RepoSettings'
import { resolveImageUrls, markdownComponents } from '../lib/markdown'

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
  const [repo, setRepo] = useState<Repo | null>(null)
  const [myRole, setMyRole] = useState('')
  const [showSettings, setShowSettings] = useState(false)
  const [readmeContent, setReadmeContent] = useState('')
  const [readmeLoading, setReadmeLoading] = useState(false)
  const [readmeExists, setReadmeExists] = useState(false)
  const [treeRefreshKey, setTreeRefreshKey] = useState(0)

  const refreshTree = useCallback(() => setTreeRefreshKey(k => k + 1), [])
  const outletContext = useMemo<RepoDetailContext>(() => ({ refreshTree }), [refreshTree])

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
      await loadReadme(r.owner_id)
    } catch { /* ignore */ }
  }, [repoId, loadReadme])

  useEffect(() => { loadRepo() }, [loadRepo])

  if (!repoId) return null

  const isChildRoute = /\/repos\/[^/]+\/(view|edit)\//.test(location.pathname)

  return (
    <div className="h-full flex overflow-hidden">
      {/* 左侧文件树 */}
      <div className="w-60 border-r border-gray-200 bg-gray-50 overflow-y-auto shrink-0 flex flex-col">
        {/* 文件树头部：知识库名称 + 设置按钮 */}
        {repo && (
          <div className="flex items-center justify-between px-3 py-2 border-b border-gray-200 shrink-0">
            <span className="text-sm font-medium text-gray-700 truncate">{repo.name}</span>
            <button
              onClick={() => setShowSettings(true)}
              className="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-200 text-gray-400 hover:text-gray-600"
              title="知识库设置"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
            </button>
          </div>
        )}
        <div className="flex-1 overflow-y-auto">
          <RepoTree repoId={repoId} repo={repo} refreshKey={treeRefreshKey} />
        </div>
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
