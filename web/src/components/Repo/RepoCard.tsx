import { useState } from 'react'
import type { Repo } from '../../lib/repos'
import { updateRepo, deleteRepo } from '../../lib/repos'

interface RepoCardProps {
  repo: Repo
  onSelect: (repo: Repo) => void
  onUpdate: () => void
}

export default function RepoCard({ repo, onSelect, onUpdate }: RepoCardProps) {
  const [showMenu, setShowMenu] = useState(false)
  const [editing, setEditing] = useState(false)
  const [newName, setNewName] = useState(repo.name)

  const handleRename = async () => {
    if (newName.trim() && newName !== repo.name) {
      await updateRepo(repo.id, { name: newName.trim() })
      onUpdate()
    }
    setEditing(false)
  }

  const handleDelete = async () => {
    if (confirm(`确定删除「${repo.name}」？`)) {
      await deleteRepo(repo.id)
      onUpdate()
    }
  }

  const handleToggleVisibility = async () => {
    const newVis = repo.visibility === 'PUBLIC' ? 'PRIVATE' : 'PUBLIC'
    await updateRepo(repo.id, { visibility: newVis })
    onUpdate()
  }

  return (
    <div className="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg cursor-pointer group" onClick={() => onSelect(repo)}>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-400 shrink-0" />
          {editing ? (
            <input
              value={newName}
              onChange={e => setNewName(e.target.value)}
              onBlur={handleRename}
              onKeyDown={e => e.key === 'Enter' && handleRename()}
              className="text-sm font-medium border border-gray-300 rounded px-1"
              autoFocus
              onClick={e => e.stopPropagation()}
            />
          ) : (
            <span className="text-sm font-medium text-gray-800">{repo.name}</span>
          )}
          <span className={`text-xs px-1.5 py-0.5 rounded ${repo.visibility === 'PUBLIC' ? 'bg-blue-50 text-blue-600' : 'bg-gray-100 text-gray-500'}`}>
            {repo.visibility === 'PUBLIC' ? '公开' : '私有'}
          </span>
        </div>
        <p className="text-xs text-gray-400 mt-0.5 truncate">{repo.description || `${repo.article_count} 篇文章`}</p>
      </div>

      <div className="relative" onClick={e => e.stopPropagation()}>
        <button onClick={() => setShowMenu(!showMenu)} className="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-gray-600">⋮</button>
        {showMenu && (
          <div className="absolute right-0 top-6 w-28 bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-10">
            <button onClick={() => { setEditing(true); setShowMenu(false) }} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">重命名</button>
            <button onClick={() => { handleToggleVisibility(); setShowMenu(false) }} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">
              {repo.visibility === 'PUBLIC' ? '设为私有' : '设为公开'}
            </button>
            <button onClick={() => { handleDelete(); setShowMenu(false) }} className="w-full px-3 py-1.5 text-xs text-left text-red-600 hover:bg-gray-50">删除</button>
          </div>
        )}
      </div>
    </div>
  )
}
