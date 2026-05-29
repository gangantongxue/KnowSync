import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRepo } from '../store/repo-context'
import { createRepo } from '../lib/repos'
import RepoCard from '../components/Repo/RepoCard'
import type { Repo } from '../lib/repos'

export default function RepoList() {
  const navigate = useNavigate()
  const { repos, loadRepos } = useRepo()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [selectedRepo, setSelectedRepo] = useState<Repo | null>(null)

  useEffect(() => {
    loadRepos()
  }, [loadRepos])

  const handleCreate = async () => {
    if (!newName.trim()) return
    await createRepo(newName.trim(), newDesc.trim() || undefined)
    setShowCreate(false)
    setNewName('')
    setNewDesc('')
    loadRepos()
  }

  const handleSelect = (repo: Repo) => {
    setSelectedRepo(repo)
  }

  return (
    <div className="h-full flex">
      {/* 左侧列表 */}
      <div className="w-72 border-r border-gray-200 overflow-y-auto">
        <div className="p-3 border-b border-gray-200 flex items-center justify-between">
          <span className="text-sm font-medium text-gray-700">所有知识库</span>
          <button onClick={() => setShowCreate(true)} className="text-xs px-2 py-1 bg-emerald-500 text-white rounded hover:bg-emerald-600">+ 新建</button>
        </div>
        {repos.map(repo => (
          <RepoCard key={repo.id} repo={repo} onSelect={handleSelect} onUpdate={loadRepos} />
        ))}
      </div>

      {/* 右侧详情 */}
      <div className="flex-1 p-6 overflow-y-auto">
        {selectedRepo ? (
          <div>
            <h2 className="text-lg font-medium text-gray-800">{selectedRepo.name}</h2>
            <p className="text-sm text-gray-500 mt-1">{selectedRepo.description || '暂无描述'}</p>
            <div className="mt-4 flex gap-2 text-xs text-gray-400">
              <span>文章: {selectedRepo.article_count}</span>
              <span>可见性: {selectedRepo.visibility === 'PUBLIC' ? '公开' : '私有'}</span>
            </div>
            <button
              onClick={() => navigate(`/repos/${selectedRepo.id}`)}
              className="mt-4 px-4 py-1.5 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600"
            >
              进入知识库
            </button>
          </div>
        ) : (
          <div className="text-gray-400 text-sm flex items-center justify-center h-full">选择一个知识库查看详情</div>
        )}
      </div>

      {/* 新建弹窗 */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
          <div className="bg-white rounded-xl p-6 w-80 mx-4 shadow-xl" onClick={e => e.stopPropagation()}>
            <h3 className="text-sm font-medium text-gray-800 mb-4">新建知识库</h3>
            <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="知识库名称" className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-3 focus:outline-none focus:border-emerald-400" autoFocus />
            <input value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="描述（可选）" className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-4 focus:outline-none focus:border-emerald-400" />
            <div className="flex gap-2 justify-end">
              <button onClick={() => setShowCreate(false)} className="px-3 py-1.5 text-sm text-gray-500 hover:text-gray-700">取消</button>
              <button onClick={handleCreate} className="px-3 py-1.5 text-sm bg-emerald-500 text-white rounded-lg hover:bg-emerald-600">创建</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
