import { useState, useEffect } from 'react'
import { groupApi } from '../../lib/chat-api'
import type { GroupMember } from '../../lib/chat-api'

interface MentionDropdownProps {
  groupId: string
  ownerId: string
  onSelect: (username: string) => void
  searchText: string
}

export default function MentionDropdown({ groupId, ownerId, onSelect, searchText }: MentionDropdownProps) {
  const [members, setMembers] = useState<GroupMember[]>([])

  useEffect(() => {
    groupApi.getMembers(groupId).then(res => {
      setMembers(res.data.members)
    }).catch(() => {})
  }, [groupId])

  const lower = searchText.toLowerCase()
  const filtered = members.filter(m => {
    const userId = m.user_id.toLowerCase()
    return userId.includes(lower)
  })

  if (filtered.length === 0) return null

  const handleClick = (member: GroupMember) => {
    const name = member.user_id === ownerId ? '所有人' : `user_${member.user_id.slice(0, 6)}`
    onSelect(name)
  }

  return (
    <div className="absolute bottom-full left-0 mb-1 w-48 bg-white border border-gray-200 rounded-lg shadow-lg max-h-40 overflow-y-auto z-50">
      {filtered.map((member) => {
        const isOwner = member.user_id === ownerId
        return (
          <div
            key={member.id}
            className="px-3 py-2 text-sm cursor-pointer flex items-center gap-2 text-gray-700 hover:bg-gray-50"
            onClick={() => handleClick(member)}
          >
            <div className="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center text-xs text-gray-600 shrink-0">
              {isOwner ? '全' : member.user_id.slice(0, 2)}
            </div>
            <span className="truncate">{isOwner ? '@所有人' : `user_${member.user_id.slice(0, 6)}`}</span>
          </div>
        )
      })}
    </div>
  )
}
