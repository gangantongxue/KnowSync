import { useState, useEffect } from 'react'
import { Input, message } from 'antd'
import { useMessageStore } from '../../store/message-store'
import { useAuth } from '../../store/auth-context'
import { groupApi } from '../../lib/chat-api'
import type { Group, GroupMember } from '../../lib/chat-api'

interface GroupSettingsModalProps {
  groupId: string
  onClose: () => void
}

const ROLE_LABEL: Record<string, string> = {
  owner: '群主',
  admin: '管理员',
  member: '成员',
}

const ROLE_COLOR: Record<string, string> = {
  owner: 'bg-yellow-100 text-yellow-700',
  admin: 'bg-blue-100 text-blue-700',
  member: 'bg-gray-100 text-gray-500',
}

export default function GroupSettingsModal({ groupId, onClose }: GroupSettingsModalProps) {
  const { user } = useAuth()
  const { friends, loadFriends, loadUserProfiles, getUserDisplayName, getUserAvatar } = useMessageStore()
  const [group, setGroup] = useState<Group | null>(null)
  const [members, setMembers] = useState<GroupMember[]>([])
  const [loading, setLoading] = useState(true)
  const [editingName, setEditingName] = useState(false)
  const [newName, setNewName] = useState('')
  const [saving, setSaving] = useState(false)
  const [showAddMembers, setShowAddMembers] = useState(false)
  const [selectedMemberIds, setSelectedMemberIds] = useState<string[]>([])
  const [confirmAction, setConfirmAction] = useState<{ type: string; userId?: string } | null>(null)

  const currentUserId = user?.id || ''
  const currentMember = members.find(m => m.user_id === currentUserId)
  const currentRole = currentMember?.role || ''
  const isOwner = currentRole === 'owner'
  const isAdmin = currentRole === 'admin'

  useEffect(() => {
    loadGroupData()
    loadFriends()
  }, [])

  useEffect(() => {
    if (members.length > 0) {
      loadUserProfiles(members.map(m => m.user_id))
    }
  }, [members, loadUserProfiles])

  const getMemberName = (userId: string): string => getUserDisplayName(userId)

  const getMemberAvatar = (userId: string): string => getUserAvatar(userId)

  const loadGroupData = async () => {
    setLoading(true)
    try {
      const [infoRes, membersRes] = await Promise.all([
        groupApi.getInfo(groupId),
        groupApi.getMembers(groupId),
      ])
      setGroup(infoRes.data.group)
      setMembers(membersRes.data.members)
      setNewName(infoRes.data.group.name)
    } catch {
      message.error('加载群组信息失败')
      onClose()
    } finally {
      setLoading(false)
    }
  }

  const handleSaveName = async () => {
    if (!newName.trim()) {
      message.warning('群名称不能为空')
      return
    }
    setSaving(true)
    try {
      await groupApi.update(groupId, { name: newName.trim() })
      message.success('群名称已更新')
      setEditingName(false)
      loadGroupData()
    } catch {
      message.error('更新失败')
    } finally {
      setSaving(false)
    }
  }

  const handleRemoveMember = async (userId: string) => {
    try {
      await groupApi.removeMember(groupId, userId)
      message.success('已移除成员')
      setConfirmAction(null)
      loadGroupData()
    } catch {
      message.error('移除失败')
    }
  }

  const handleTransferOwnership = async (newOwnerId: string) => {
    try {
      await groupApi.transferOwnership(groupId, newOwnerId)
      message.success('群主已转让')
      setConfirmAction(null)
      loadGroupData()
    } catch {
      message.error('转让失败')
    }
  }

  const handleLeave = async () => {
    try {
      await groupApi.leave(groupId)
      message.success('已退出群聊')
      setConfirmAction(null)
      onClose()
    } catch {
      message.error('退出失败')
    }
  }

  const handleAddMembers = async () => {
    if (selectedMemberIds.length === 0) {
      message.warning('请选择要添加的成员')
      return
    }
    setSaving(true)
    try {
      await groupApi.addMembers(groupId, selectedMemberIds)
      message.success('已添加成员')
      setSelectedMemberIds([])
      setShowAddMembers(false)
      loadGroupData()
    } catch {
      message.error('添加失败')
    } finally {
      setSaving(false)
    }
  }

  const getFriendDisplayName = (friendId: string): string => getUserDisplayName(friendId)

  const notInGroupFriends = friends.filter(
    f => !members.some(m => m.user_id === f.friend_id)
  )

  const getConfirmContent = () => {
    if (!confirmAction) return null
    switch (confirmAction.type) {
      case 'remove':
        return { title: '确认移除', content: '确定要移除此成员吗？', onConfirm: () => handleRemoveMember(confirmAction.userId!) }
      case 'transfer':
        return { title: '确认转让', content: '确定要将群主转让给此成员吗？此操作不可撤销。', onConfirm: () => handleTransferOwnership(confirmAction.userId!) }
      case 'leave':
        return { title: '确认退出', content: '确定要退出群聊吗？', onConfirm: () => handleLeave() }
      default:
        return null
    }
  }

  if (loading) {
    return (
      <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
        <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl p-8 text-center text-sm text-gray-400" onClick={e => e.stopPropagation()}>
          加载中...
        </div>
      </div>
    )
  }

  if (!group) return null

  const sortedMembers = [...members].sort((a, b) => {
    const roleOrder = { owner: 0, admin: 1, member: 2 }
    return (roleOrder[a.role as keyof typeof roleOrder] ?? 3) - (roleOrder[b.role as keyof typeof roleOrder] ?? 3)
  })

  const confirmContent = getConfirmContent()

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-4 border-b border-gray-100">
          <h2 className="text-base font-medium text-gray-800">群组设置</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg leading-none">✕</button>
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          {/* Group Name */}
          <div className="mb-4">
            <label className="text-xs text-gray-500 mb-1 block">群名称</label>
            {editingName ? (
              <div className="flex items-center gap-2">
                <Input
                  value={newName}
                  onChange={e => setNewName(e.target.value)}
                  size="small"
                  className="flex-1"
                />
                <button
                  onClick={handleSaveName}
                  disabled={saving}
                  className="px-3 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                >
                  保存
                </button>
                <button
                  onClick={() => { setEditingName(false); setNewName(group.name) }}
                  className="px-3 py-1 text-xs border border-gray-200 text-gray-500 rounded hover:bg-gray-50"
                >
                  取消
                </button>
              </div>
            ) : (
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-800">{group.name}</span>
                {isOwner && (
                  <button
                    onClick={() => setEditingName(true)}
                    className="text-xs text-blue-500 hover:text-blue-600"
                  >
                    编辑
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Members Section */}
          <div className="mb-4">
            <div className="flex items-center justify-between mb-2">
              <label className="text-xs text-gray-500">成员（{members.length}）</label>
              {(isOwner || isAdmin) && (
                <button
                  onClick={() => setShowAddMembers(true)}
                  className="text-xs text-blue-500 hover:text-blue-600"
                >
                  + 添加成员
                </button>
              )}
            </div>

            <div className="space-y-1">
              {sortedMembers.map(member => {
                const isSelf = member.user_id === currentUserId
                const isOwnerMember = member.role === 'owner'
                const canRemove = isOwner && !isOwnerMember && !isSelf
                const canTransfer = isOwner && isOwnerMember && !isSelf

                return (
                  <div key={member.user_id} className="flex items-center justify-between py-2 px-2 rounded hover:bg-gray-50">
                    <div className="flex items-center gap-2 min-w-0">
                      {getMemberAvatar(member.user_id) ? (
                        <img src={getMemberAvatar(member.user_id)} alt="" className="w-8 h-8 rounded-full object-cover shrink-0" />
                      ) : (
                        <div className="w-8 h-8 rounded-full bg-gray-100 text-gray-600 flex items-center justify-center text-xs font-medium shrink-0">
                          {getMemberName(member.user_id).charAt(0).toUpperCase()}
                        </div>
                      )}
                      <div className="min-w-0">
                        <span className="text-sm text-gray-800 truncate block">
                          {getMemberName(member.user_id)}
                          {isSelf && <span className="text-xs text-gray-400 ml-1">（我）</span>}
                        </span>
                        <span className="text-xs text-gray-400 truncate block">ID: {member.user_id}</span>
                      </div>
                      <span className={`text-xs px-1.5 py-0.5 rounded-full shrink-0 ${ROLE_COLOR[member.role] || ''}`}>
                        {ROLE_LABEL[member.role] || member.role}
                      </span>
                    </div>
                    <div className="flex gap-1 shrink-0">
                      {canTransfer && (
                        <button
                          onClick={() => setConfirmAction({ type: 'transfer', userId: member.user_id })}
                          className="text-xs text-orange-500 hover:text-orange-600 px-1.5 py-0.5"
                        >
                          转让
                        </button>
                      )}
                      {canRemove && (
                        <button
                          onClick={() => setConfirmAction({ type: 'remove', userId: member.user_id })}
                          className="text-xs text-red-500 hover:text-red-600 px-1.5 py-0.5"
                        >
                          移除
                        </button>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Leave Group */}
          {!isOwner && (
            <button
              onClick={() => setConfirmAction({ type: 'leave' })}
              className="w-full py-2 text-sm text-red-500 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
            >
              退出群聊
            </button>
          )}
        </div>
      </div>

      {/* Confirm Dialog */}
      {confirmContent && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-[60]" onClick={() => setConfirmAction(null)}>
          <div className="bg-white rounded-xl p-6 w-80 mx-4 shadow-xl" onClick={e => e.stopPropagation()}>
            <h3 className="text-sm font-medium text-gray-800 mb-2">{confirmContent.title}</h3>
            <p className="text-sm text-gray-500 mb-4">{confirmContent.content}</p>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setConfirmAction(null)}
                className="px-4 py-1.5 text-sm border border-gray-200 text-gray-500 rounded-lg hover:bg-gray-50"
              >
                取消
              </button>
              <button
                onClick={confirmContent.onConfirm}
                className="px-4 py-1.5 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600"
              >
                确认
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Members Dialog */}
      {showAddMembers && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-[60]" onClick={() => { setShowAddMembers(false); setSelectedMemberIds([]) }}>
          <div className="bg-white rounded-xl w-full max-w-sm mx-4 shadow-xl max-h-[60vh] flex flex-col" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-gray-100">
              <h3 className="text-sm font-medium text-gray-800">添加成员</h3>
              <button
                onClick={() => { setShowAddMembers(false); setSelectedMemberIds([]) }}
                className="text-gray-400 hover:text-gray-600 text-lg leading-none"
              >
                ✕
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-4">
              {notInGroupFriends.length === 0 ? (
                <div className="text-center py-6 text-sm text-gray-400">没有可添加的好友</div>
              ) : (
                notInGroupFriends.map(friend => {
                  const isSelected = selectedMemberIds.includes(friend.friend_id)
                  const displayName = getFriendDisplayName(friend.friend_id)
                  const avatar = getUserAvatar(friend.friend_id)
                  return (
                    <label
                      key={friend.friend_id}
                      className="flex items-center gap-3 py-2 cursor-pointer hover:bg-gray-50 px-2 rounded"
                    >
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => {
                          setSelectedMemberIds(prev =>
                            isSelected ? prev.filter(id => id !== friend.friend_id) : [...prev, friend.friend_id]
                          )
                        }}
                        className="rounded border-gray-300 text-blue-500"
                      />
                      {avatar ? (
                        <img src={avatar} alt="" className="w-8 h-8 rounded-full object-cover shrink-0" />
                      ) : (
                        <div className="w-8 h-8 rounded-full bg-gray-100 text-gray-600 flex items-center justify-center text-xs font-medium shrink-0">
                          {displayName.charAt(0).toUpperCase()}
                        </div>
                      )}
                      <div className="min-w-0">
                        <span className="text-sm text-gray-800 truncate block">{displayName}</span>
                      </div>
                    </label>
                  )
                })
              )}
            </div>
            <div className="p-4 border-t border-gray-100 flex justify-end gap-2">
              <button
                onClick={() => { setShowAddMembers(false); setSelectedMemberIds([]) }}
                className="px-4 py-1.5 text-sm border border-gray-200 text-gray-500 rounded-lg hover:bg-gray-50"
              >
                取消
              </button>
              <button
                onClick={handleAddMembers}
                disabled={saving || selectedMemberIds.length === 0}
                className="px-4 py-1.5 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
              >
                {saving ? '添加中...' : '添加'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
