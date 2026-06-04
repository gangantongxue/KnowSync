import { useState } from 'react'
import { Button } from 'antd'
import { friendApi } from '../../lib/chat-api'

interface InvitationCardProps {
  repoName: string
  role: string
  extra: string | null
}

export default function InvitationCard({ repoName, role, extra }: InvitationCardProps) {
  const [status, setStatus] = useState<'pending' | 'accepted' | 'rejected'>('pending')
  const [loading, setLoading] = useState(false)

  const roleLabel: Record<string, string> = {
    editor: '编辑者',
    viewer: '查看者',
    admin: '管理员',
  }

  const handleAccept = async () => {
    setLoading(true)
    try {
      if (extra) {
        const data = JSON.parse(extra)
        if (data.request_id) {
          await friendApi.acceptRequest(data.request_id)
        }
      }
      setStatus('accepted')
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  const handleReject = async () => {
    setLoading(true)
    try {
      if (extra) {
        const data = JSON.parse(extra)
        if (data.request_id) {
          await friendApi.rejectRequest(data.request_id)
        }
      }
      setStatus('rejected')
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  if (status === 'accepted') {
    return (
      <div className="bg-green-50 border border-green-200 rounded-lg p-3 text-sm text-green-700">
        已接受邀请
      </div>
    )
  }

  if (status === 'rejected') {
    return (
      <div className="bg-gray-50 border border-gray-200 rounded-lg p-3 text-sm text-gray-500">
        已拒绝邀请
      </div>
    )
  }

  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 max-w-sm">
      <div className="text-sm text-gray-800 mb-2">
        邀请你加入知识库「<span className="font-medium">{repoName}</span>」
      </div>
      <div className="mb-3">
        <span className="inline-block px-2 py-0.5 bg-blue-100 text-blue-700 text-xs rounded-full">
          {roleLabel[role] || role}
        </span>
      </div>
      <div className="flex gap-2">
        <Button size="small" type="primary" loading={loading} onClick={handleAccept}>
          接受
        </Button>
        <Button size="small" loading={loading} onClick={handleReject}>
          拒绝
        </Button>
      </div>
    </div>
  )
}
