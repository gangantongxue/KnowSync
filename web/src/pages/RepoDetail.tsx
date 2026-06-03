import { useEffect, useState } from 'react'
import { Outlet, useParams, useLocation } from 'react-router-dom'
import { getRepo } from '../lib/repos'
import type { Repo } from '../lib/repos'
import RepoTree from '../components/Repo/RepoTree'
import RepoSettings from '../components/Repo/RepoSettings'

export default function RepoDetail() {
  const { repoId } = useParams<{ repoId: string }>()
  const location = useLocation()
  const [repo, setRepo] = useState<Repo | null>(null)
  const [showSettings, setShowSettings] = useState(false)

  const loadRepo = () => {
    if (!repoId) return
    getRepo(repoId).then(setRepo).catch(() => {})
  }

  useEffect(() => { loadRepo() }, [repoId])

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
          <RepoTree repoId={repoId} repo={repo} />
        </div>
      </div>

      {/* 中间内容区 */}
      <div className="flex-1 overflow-y-auto min-w-0">
        <Outlet />
        {!isChildRoute && (
          <div className="p-6">
            {repo && (
              <div className="mb-6">
                <h2 className="text-lg font-medium text-gray-800">{repo.name}</h2>
                <p className="text-sm text-gray-500 mt-1">{repo.description || '暂无描述'}</p>
              </div>
            )}
            <div className="text-sm text-gray-400 text-center py-12">
              选择一篇文章查看或编辑
            </div>
          </div>
        )}
      </div>

      {/* 设置弹窗 */}
      {showSettings && repo && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setShowSettings(false)}>
          <div className="bg-white rounded-xl shadow-xl w-[480px] max-h-[85vh] overflow-y-auto mx-4" onClick={e => e.stopPropagation()}>
            <RepoSettings repo={repo} onUpdate={loadRepo} onClose={() => setShowSettings(false)} />
          </div>
        </div>
      )}
    </div>
  )
}
