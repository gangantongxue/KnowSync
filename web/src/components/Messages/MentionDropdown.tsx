import { useState, useEffect, useMemo } from 'react'
import { groupApi } from '../../lib/chat-api'
import { useMessageStore } from '../../store/message-store'
import type { GroupMember } from '../../lib/chat-api'

interface MentionDropdownProps {
  groupId: string
  ownerId: string
  onSelect: (userId: string) => void
  searchText: string
}

export default function MentionDropdown({ groupId, ownerId: _ownerId, onSelect, searchText }: MentionDropdownProps) {
  const { loadUserProfiles, getUserDisplayName, getUserAvatar } = useMessageStore()
  const [members, setMembers] = useState<GroupMember[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    groupApi.getMembers(groupId).then(res => {
      setMembers(res.data.members)
      const ids = res.data.members.map(m => m.user_id)
      loadUserProfiles(ids)
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [groupId, loadUserProfiles])

  const lower = searchText.toLowerCase()
  const filtered = useMemo(() => members.filter(m => {
    if (!lower) return true
    const displayName = getUserDisplayName(m.user_id).toLowerCase()
    return m.user_id.toLowerCase().includes(lower) || displayName.includes(lower)
  }), [members, lower, getUserDisplayName])

  if (loading) {
    return (
      <div className="absolute bottom-full left-0 mb-1 w-56 bg-white border border-gray-200 rounded-lg shadow-lg z-50">
        <div className="px-3 py-2 text-sm text-gray-400">加载中...</div>
      </div>
    )
  }

  if (filtered.length === 0) return null

  return (
    <div className="absolute bottom-full left-0 mb-1 w-56 bg-white border border-gray-200 rounded-lg shadow-lg max-h-52 overflow-y-auto z-50">
      {filtered.map((member) => {
        const displayName = getUserDisplayName(member.user_id)
        const avatar = getUserAvatar(member.user_id)
        return (
          <div
            key={member.user_id}
            className="px-3 py-2 text-sm cursor-pointer flex items-center gap-2 text-gray-700 hover:bg-gray-50"
            onClick={() => onSelect(member.user_id)}
          >
            {avatar ? (
              <img src={avatar} alt="" className="w-7 h-7 rounded-full object-cover shrink-0" />
            ) : (
              <div className="w-7 h-7 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-xs font-medium shrink-0">
                {displayName.charAt(0).toUpperCase()}
              </div>
            )}
            <div className="min-w-0">
              <span className="truncate block">{displayName}</span>
            </div>
          </div>
        )
      })}
    </div>
  )
}
