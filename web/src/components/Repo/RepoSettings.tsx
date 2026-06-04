import { useEffect, useState } from 'react'
import { useAuth } from '../../store/auth-context'
import { updateRepo, deleteRepo } from '../../lib/repos'
import { listCollaborators, updateCollaborator, removeCollaborator } from '../../lib/nodes'
import FriendPickerModal from '../Messages/FriendPickerModal'
import { getUser } from '../../lib/auth'
import type { Repo } from '../../lib/repos'
import type { Collaborator } from '../../lib/nodes'
import type { UserInfo } from '../../lib/auth'

interface RepoSettingsProps {
  repo: Repo
  myRole: string
  onUpdate: () => void
  onClose: () => void
  onDelete: () => void
}

interface CollaboratorWithUser {
  collab: Collaborator
  user: UserInfo | null
}

export default function RepoSettings({ repo, myRole, onUpdate, onClose, onDelete }: RepoSettingsProps) {
  const { user: self } = useAuth()
  const [collaborators, setCollaborators] = useState<CollaboratorWithUser[]>([])
  const [loadingCollabs, setLoadingCollabs] = useState(true)
  const [showFriendPicker, setShowFriendPicker] = useState(false)

  // 编辑状态
  const [editName, setEditName] = useState(repo.name)
  const [editDesc, setEditDesc] = useState(repo.description || '')
  const [editVisibility, setEditVisibility] = useState(repo.visibility)
  const [saving, setSaving] = useState(false)

  // 删除确认状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteConfirmInput, setDeleteConfirmInput] = useState('')
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    setEditName(repo.name)
    setEditDesc(repo.description || '')
    setEditVisibility(repo.visibility)
  }, [repo])

  const hasChanges = editName !== repo.name || editDesc !== (repo.description || '') || editVisibility !== repo.visibility

  const loadCollabs = async () => {
    setLoadingCollabs(true)
    try {
      const list = await listCollaborators(repo.id)
      const withUsers: CollaboratorWithUser[] = []
      for (const c of list) {
        try {
          const u = await getUser(c.user_id)
          withUsers.push({ collab: c, user: u })
        } catch {
          withUsers.push({ collab: c, user: null })
        }
      }
      withUsers.sort((a, b) => {
        if (a.collab.user_id === self?.id) return -1
        if (b.collab.user_id === self?.id) return 1
        return 0
      })
      setCollaborators(withUsers)
    } catch {
      setCollaborators([])
    }
    setLoadingCollabs(false)
  }

  useEffect(() => { loadCollabs() }, [repo.id])

  const handleSave = async () => {
    if (!editName.trim()) return
    setSaving(true)
    try {
      await updateRepo(repo.id, {
        name: editName.trim() !== repo.name ? editName.trim() : undefined,
        description: editDesc !== (repo.description || '') ? editDesc : undefined,
        visibility: editVisibility !== repo.visibility ? editVisibility : undefined,
      })
      onUpdate()
    } catch (err: any) {
      alert('保存失败: ' + err.message)
    }
    setSaving(false)
  }

  const handleUpdateRole = async (userId: string, role: string) => {
    if (userId === self?.id) return
    try {
      await updateCollaborator(repo.id, userId, role)
      loadCollabs()
    } catch (err: any) {
      alert('更新失败: ' + err.message)
    }
  }

  const handleRemove = async (userId: string) => {
    if (userId === self?.id) return
    if (!confirm('确定移除此协作者？')) return
    try {
      await removeCollaborator(repo.id, userId)
      loadCollabs()
    } catch (err: any) {
      alert('移除失败: ' + err.message)
    }
  }

  const handleDelete = async () => {
    if (deleteConfirmInput !== repo.name || deleting) return
    setDeleting(true)
    try {
      await deleteRepo(repo.id)
      onDelete()
    } catch (err: any) {
      alert('删除失败: ' + err.message)
      setDeleting(false)
    }
  }

  const roleLabels: Record<string, string> = {
    ADMIN: '管理员',
    DEVELOPER: '开发者',
    VIEWER: '浏览者',
  }

  return (
    <div>
      {/* 弹窗头部 */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200">
        <span className="text-sm font-medium text-gray-800">知识库设置</span>
        <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-sm">✕</button>
      </div>

      <div className="p-4 space-y-5">
        {/* 仓库基本信息 */}
        <div>
          <div className="text-xs text-gray-400 font-medium mb-3">基本信息</div>

          <div className="space-y-3">
            <div>
              <label className="text-xs text-gray-400 block mb-1">名称</label>
              <input
                value={editName}
                onChange={e => setEditName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400"
              />
            </div>

            <div>
              <label className="text-xs text-gray-400 block mb-1">描述</label>
              <textarea
                value={editDesc}
                onChange={e => setEditDesc(e.target.value)}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400 resize-none"
              />
            </div>

            <div>
              <label className="text-xs text-gray-400 block mb-1">可见性</label>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setEditVisibility('PUBLIC')}
                  className={`px-4 py-1.5 rounded text-sm border transition-colors ${editVisibility === 'PUBLIC' ? 'bg-blue-50 text-blue-600 border-blue-300' : 'bg-gray-50 text-gray-500 border-gray-200 hover:bg-gray-100'}`}
                >
                  公开
                </button>
                <button
                  onClick={() => setEditVisibility('PRIVATE')}
                  className={`px-4 py-1.5 rounded text-sm border transition-colors ${editVisibility === 'PRIVATE' ? 'bg-orange-50 text-orange-600 border-orange-300' : 'bg-gray-50 text-gray-500 border-gray-200 hover:bg-gray-100'}`}
                >
                  私有
                </button>
              </div>
            </div>

            {hasChanges && (
              <div className="flex gap-2 pt-1">
                <button
                  onClick={handleSave}
                  disabled={!editName.trim() || saving}
                  className="px-4 py-1.5 text-sm bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 disabled:opacity-50"
                >
                  {saving ? '保存中...' : '保存'}
                </button>
                <button
                  onClick={() => {
                    setEditName(repo.name)
                    setEditDesc(repo.description || '')
                    setEditVisibility(repo.visibility)
                  }}
                  className="px-4 py-1.5 text-sm text-gray-500 hover:text-gray-700"
                >
                  取消
                </button>
              </div>
            )}
          </div>
        </div>

        <div className="border-t border-gray-100" />

        {/* 协作者 */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-gray-400 font-medium">协作者</span>
            <button onClick={() => setShowFriendPicker(true)} className="text-xs text-emerald-500 hover:text-emerald-600">+ 添加协作者</button>
          </div>

          {loadingCollabs ? (
            <div className="text-xs text-gray-400 py-2">加载中...</div>
          ) : collaborators.length === 0 ? (
            <div className="text-xs text-gray-400 py-2">暂无协作者</div>
          ) : (
            collaborators.map(({ collab, user }) => {
              const isSelf = collab.user_id === self?.id
              return (
                <div key={collab.user_id} className="flex items-center gap-2 py-1.5 border-b border-gray-100 last:border-0">
                  <div className="w-7 h-7 rounded-full bg-emerald-100 text-emerald-700 text-xs flex items-center justify-center shrink-0 overflow-hidden">
                    {user?.avatar ? (
                      <img src={user.avatar} alt="" className="w-full h-full object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }} />
                    ) : (
                      user?.name?.charAt(0) || collab.user_id.charAt(0)
                    )}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="text-sm text-gray-700 truncate">{user?.name || collab.user_id.slice(0, 8) + '...'}</div>
                    {isSelf && <div className="text-xs text-emerald-500">我</div>}
                  </div>

                  {isSelf ? (
                    <span className="text-xs text-gray-500 bg-gray-100 px-1.5 py-0.5 rounded">{roleLabels[collab.role] || collab.role}</span>
                  ) : (
                    <div className="flex items-center gap-1">
                      <select value={collab.role} onChange={e => handleUpdateRole(collab.user_id, e.target.value)} className="text-xs border border-gray-200 rounded px-1 py-0.5 bg-white">
                        <option value="ADMIN">管理员</option>
                        <option value="DEVELOPER">开发者</option>
                        <option value="VIEWER">浏览者</option>
                      </select>
                      <button onClick={() => handleRemove(collab.user_id)} className="text-gray-400 hover:text-red-500 text-xs" title="移除">✕</button>
                    </div>
                  )}
                </div>
              )
            })
          )}

          {showFriendPicker && (
            <FriendPickerModal
              repoId={repo.id}
              repoName={repo.name}
              onClose={() => { setShowFriendPicker(false); loadCollabs() }}
            />
          )}
        </div>

        <div className="border-t border-gray-100" />

        {/* 危险区域 —— 仅管理员可见 */}
        {myRole === 'ADMIN' && (
        <div>
          <div className="text-xs text-gray-400 font-medium mb-3">危险区域</div>
          <div className="p-3 border border-red-200 rounded-lg bg-red-50">
            <div className="text-sm text-red-700 mb-1">删除知识库</div>
            <div className="text-xs text-gray-500 mb-3">
              删除后将无法恢复，包括数据库记录、文件系统和向量存储中的所有关联数据。
            </div>
            {!showDeleteConfirm ? (
              <button
                onClick={() => {
                  setShowDeleteConfirm(true)
                  setDeleteConfirmInput('')
                }}
                className="px-3 py-1.5 text-xs text-red-600 border border-red-300 rounded hover:bg-red-100"
              >
                删除知识库
              </button>
            ) : (
              <div className="space-y-2">
                <div className="text-xs text-gray-600">
                  请输入 <span className="font-medium text-red-600">{repo.name}</span> 以确认删除：
                </div>
                <input
                  value={deleteConfirmInput}
                  onChange={e => setDeleteConfirmInput(e.target.value)}
                  placeholder={repo.name}
                  className="w-full px-3 py-1.5 border border-red-300 rounded text-sm focus:outline-none focus:border-red-500"
                  autoFocus
                />
                <div className="flex gap-2">
                  <button
                    onClick={handleDelete}
                    disabled={deleteConfirmInput !== repo.name || deleting}
                    className={`px-3 py-1.5 text-xs text-white rounded ${
                      deleteConfirmInput !== repo.name || deleting
                        ? 'bg-red-300 cursor-not-allowed'
                        : 'bg-red-500 hover:bg-red-600'
                    }`}
                  >
                    {deleting ? '删除中...' : '确认删除'}
                  </button>
                  <button
                    onClick={() => {
                      setShowDeleteConfirm(false)
                      setDeleteConfirmInput('')
                    }}
                    className="px-3 py-1.5 text-xs text-gray-500 hover:text-gray-700"
                  >
                    取消
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
        )}
      </div>
    </div>
  )
}
