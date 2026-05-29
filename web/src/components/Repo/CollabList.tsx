import { useState, useEffect } from 'react'
import { listCollaborators, addCollaborator, updateCollaborator, removeCollaborator } from '../../lib/nodes'
import type { Collaborator } from '../../lib/nodes'

interface CollabListProps {
  repoId: string
}

export default function CollabList({ repoId }: CollabListProps) {
  const [collabs, setCollabs] = useState<Collaborator[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [newUserId, setNewUserId] = useState('')
  const [newRole, setNewRole] = useState('DEVELOPER')

  const load = async () => {
    const list = await listCollaborators(repoId)
    setCollabs(list)
  }

  useEffect(() => { load() }, [repoId])

  const handleAdd = async () => {
    if (!newUserId.trim()) return
    await addCollaborator(repoId, newUserId.trim(), newRole)
    setShowAdd(false)
    setNewUserId('')
    load()
  }

  const handleUpdateRole = async (userId: string, role: string) => {
    await updateCollaborator(repoId, userId, role)
    load()
  }

  const handleRemove = async (userId: string) => {
    if (confirm('确定移除此协作者？')) {
      await removeCollaborator(repoId, userId)
      load()
    }
  }

  return (
    <div className="mt-6">
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-medium text-gray-700">协作者</h4>
        <button onClick={() => setShowAdd(true)} className="text-xs text-emerald-600 hover:underline">+ 添加</button>
      </div>

      {collabs.map(c => (
        <div key={c.user_id} className="flex items-center justify-between py-1.5 text-sm">
          <span className="text-gray-600 text-xs">{c.user_id.slice(0, 8)}...</span>
          <div className="flex items-center gap-1">
            <select
              value={c.role}
              onChange={e => handleUpdateRole(c.user_id, e.target.value)}
              className="text-xs border border-gray-200 rounded px-1 py-0.5"
            >
              <option value="ADMIN">管理员</option>
              <option value="DEVELOPER">开发者</option>
              <option value="VIEWER">查看者</option>
            </select>
            <button onClick={() => handleRemove(c.user_id)} className="text-gray-400 hover:text-red-500 text-xs">✕</button>
          </div>
        </div>
      ))}

      {collabs.length === 0 && <p className="text-xs text-gray-400">暂无协作者</p>}

      {showAdd && (
        <div className="mt-2 space-y-2">
          <input value={newUserId} onChange={e => setNewUserId(e.target.value)} placeholder="用户 ID" className="w-full px-2 py-1 border border-gray-300 rounded text-xs" />
          <select value={newRole} onChange={e => setNewRole(e.target.value)} className="w-full px-2 py-1 border border-gray-300 rounded text-xs">
            <option value="DEVELOPER">开发者</option>
            <option value="VIEWER">查看者</option>
          </select>
          <div className="flex gap-2">
            <button onClick={handleAdd} className="px-2 py-1 bg-emerald-500 text-white rounded text-xs">添加</button>
            <button onClick={() => setShowAdd(false)} className="px-2 py-1 text-gray-500 text-xs">取消</button>
          </div>
        </div>
      )}
    </div>
  )
}
