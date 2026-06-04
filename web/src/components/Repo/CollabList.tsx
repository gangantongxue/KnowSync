import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { listCollaborators, updateCollaborator, removeCollaborator } from '../../lib/nodes'
import { getUser } from '../../lib/auth'
import { friendApi } from '../../lib/chat-api'
import type { Collaborator } from '../../lib/nodes'
import type { UserInfo } from '../../lib/auth'
import type { Friend } from '../../lib/chat-api'

interface CollabListProps {
  repoId: string
}

interface CollabWithUser {
  collab: Collaborator
  user: UserInfo | null
  remark: string
}

export default function CollabList({ repoId }: CollabListProps) {
  const navigate = useNavigate()
  const [items, setItems] = useState<CollabWithUser[]>([])

  const load = async () => {
    const [list, friendRes] = await Promise.all([
      listCollaborators(repoId),
      friendApi.getList().catch(() => ({ data: { friends: [] as Friend[] } })),
    ])
    const friendMap = new Map<string, string>()
    for (const f of friendRes.data.friends) {
      friendMap.set(f.friend_id, f.remark)
    }
    const withUsers = await Promise.all(
      list.map(async (c) => {
        const remark = friendMap.get(c.user_id) || ''
        try {
          const u = await getUser(c.user_id)
          return { collab: c, user: u, remark }
        } catch {
          return { collab: c, user: null, remark }
        }
      })
    )
    setItems(withUsers)
  }

  useEffect(() => { load() }, [repoId])

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

  const getDisplayName = (item: CollabWithUser) => item.remark || item.user?.name || item.collab.user_id.slice(0, 8)

  return (
    <div className="mt-6">
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-medium text-gray-700">协作者</h4>
      </div>

      {items.map(item => (
        <div key={item.collab.user_id} className="flex items-center gap-2 py-1.5 border-b border-gray-100 last:border-0">
          <div className="w-7 h-7 rounded-full bg-emerald-100 text-emerald-700 text-xs flex items-center justify-center shrink-0 overflow-hidden">
            {item.user?.avatar ? (
              <img src={item.user.avatar} alt="" className="w-full h-full object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }} />
            ) : (
              getDisplayName(item).charAt(0)
            )}
          </div>

          <button
            onClick={() => navigate(`/users/${item.collab.user_id}`)}
            className="flex-1 min-w-0 text-left text-sm text-gray-700 truncate hover:text-emerald-600"
          >
            {getDisplayName(item)}
          </button>

          <div className="flex items-center gap-1">
            <select
              value={item.collab.role}
              onChange={e => handleUpdateRole(item.collab.user_id, e.target.value)}
              className="text-xs border border-gray-200 rounded px-1 py-0.5 bg-white"
            >
              <option value="ADMIN">管理员</option>
              <option value="DEVELOPER">开发者</option>
              <option value="VIEWER">浏览者</option>
            </select>
            <button onClick={() => handleRemove(item.collab.user_id)} className="text-gray-400 hover:text-red-500 text-xs">✕</button>
          </div>
        </div>
      ))}

      {items.length === 0 && <p className="text-xs text-gray-400">暂无协作者</p>}
    </div>
  )
}
