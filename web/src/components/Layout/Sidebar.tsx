import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRepo } from '../../store/repo-context'
import { useAuth } from '../../store/auth-context'
import * as repoApi from '../../lib/repos'

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const { repos, followedRepos, loadRepos, loadFollowedRepos } = useRepo()
  const { user } = useAuth()

  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [newVisibility, setNewVisibility] = useState<'PUBLIC' | 'PRIVATE'>('PUBLIC')
  const [creating, setCreating] = useState(false)
  const [followedExpanded, setFollowedExpanded] = useState(false)

  const myRepos = repos.filter(r => r.owner_id === user?.id)
  const collabRepos = repos.filter(r => r.owner_id !== user?.id)

  const resetCreateForm = () => {
    setNewName('')
    setNewDesc('')
    setNewVisibility('PUBLIC')
  }

  const handleCreate = async () => {
    if (!newName.trim()) return
    setCreating(true)
    try {
      const repo = await repoApi.createRepo(newName.trim(), newDesc.trim() || undefined, newVisibility)
      setShowCreate(false)
      resetCreateForm()
      await loadRepos()
      navigate(`/repos/${repo.id}`)
    } catch {
      // ignore
    }
    setCreating(false)
  }

  const handleFollowedToggle = () => {
    if (!followedExpanded) {
      loadFollowedRepos()
    }
    setFollowedExpanded(!followedExpanded)
  }

  const repoItem = (r: typeof repos[0]) => (
    <button
      key={r.id}
      onClick={() => navigate(`/repos/${r.id}`)}
      className="flex items-center gap-2 w-full px-2 py-1.5 rounded text-sm text-gray-700 hover:bg-gray-100 mb-0.5"
    >
      <span className="w-5 h-5 rounded bg-emerald-100 text-emerald-700 text-xs flex items-center justify-center shrink-0">
        {r.name.charAt(0)}
      </span>
      <span className="truncate">{r.name}</span>
    </button>
  )

  if (collapsed) {
    return (
      <div className="w-12 border-r border-gray-200 bg-gray-50 flex flex-col items-center py-2 gap-3 shrink-0">
        <button onClick={() => setCollapsed(false)} className="text-gray-400 hover:text-gray-600 text-lg" title="展开侧栏">☰</button>
        {myRepos.slice(0, 3).map(r => (
          <button key={r.id} onClick={() => navigate(`/repos/${r.id}`)} className="w-8 h-8 rounded bg-gray-200 text-xs text-gray-600 hover:bg-gray-300" title={r.name}>
            {r.name.charAt(0)}
          </button>
        ))}
      </div>
    )
  }

  return (
    <div className="w-56 border-r border-gray-200 bg-gray-50 flex flex-col shrink-0 overflow-hidden">
      <div className="flex items-center justify-end p-2 border-b border-gray-200">
        <button
          onClick={() => setCollapsed(true)}
          className="text-xs text-gray-400 hover:text-gray-600"
          title="折叠侧栏"
        >
          ◀ 折叠
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {/* 新建按钮 */}
        <div className="flex items-center justify-between px-2 py-1">
          <span className="text-xs text-gray-400 font-medium">知识库</span>
          <button
            onClick={() => setShowCreate(true)}
            className="text-xs text-emerald-500 hover:text-emerald-600"
          >
            + 新建
          </button>
        </div>

        {/* 我的知识库 */}
        {myRepos.map(repoItem)}

        {/* 分隔线 + 协作知识库 */}
        {collabRepos.length > 0 && (
          <>
            <hr className="my-2 border-gray-200" />
            {collabRepos.map(repoItem)}
          </>
        )}

        {/* 分隔线 + 我的关注（可折叠） */}
        <hr className="my-2 border-gray-200" />
        <div>
          <button
            onClick={handleFollowedToggle}
            className="flex items-center gap-1 w-full px-2 py-1 text-xs text-gray-400 font-medium hover:text-gray-600"
          >
            <span className={`transition-transform ${followedExpanded ? 'rotate-90' : ''}`}>▶</span>
            <span>我的关注</span>
          </button>
          {followedExpanded && followedRepos.length > 0 && (
            <div className="mt-1">
              {followedRepos.map(repoItem)}
            </div>
          )}
          {followedExpanded && followedRepos.length === 0 && (
            <div className="px-2 py-2 text-xs text-gray-400">暂无关注</div>
          )}
        </div>
      </div>

      {/* 新建知识库弹窗 */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
          <div className="bg-white rounded-xl p-6 w-80 mx-4 shadow-xl" onClick={e => e.stopPropagation()}>
            <h3 className="text-sm font-medium text-gray-800 mb-4">新建知识库</h3>
            <input
              value={newName}
              onChange={e => setNewName(e.target.value)}
              placeholder="知识库名称"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-3 focus:outline-none focus:border-emerald-400"
              autoFocus
            />
            <input
              value={newDesc}
              onChange={e => setNewDesc(e.target.value)}
              placeholder="描述（可选）"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-3 focus:outline-none focus:border-emerald-400"
            />
            <div className="flex items-center gap-2 mb-4">
              <span className="text-xs text-gray-500">可见性</span>
              <button
                onClick={() => setNewVisibility('PUBLIC')}
                className={`px-3 py-1 rounded text-xs ${newVisibility === 'PUBLIC' ? 'bg-emerald-500 text-white' : 'bg-gray-100 text-gray-500 hover:bg-gray-200'}`}
              >
                公开
              </button>
              <button
                onClick={() => setNewVisibility('PRIVATE')}
                className={`px-3 py-1 rounded text-xs ${newVisibility === 'PRIVATE' ? 'bg-emerald-500 text-white' : 'bg-gray-100 text-gray-500 hover:bg-gray-200'}`}
              >
                私有
              </button>
            </div>
            <div className="flex gap-2 justify-end">
              <button onClick={() => setShowCreate(false)} className="px-3 py-1.5 text-sm text-gray-500 hover:text-gray-700">取消</button>
              <button
                onClick={handleCreate}
                disabled={!newName.trim() || creating}
                className="px-3 py-1.5 text-sm bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 disabled:opacity-50"
              >
                {creating ? '创建中...' : '创建'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
