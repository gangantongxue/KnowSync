import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth-context'
import { friendApi } from '../lib/chat-api'
import { userApi, type UserProfileData } from '../lib/user-api'
import { sendVerifyCode } from '../lib/auth'

export default function UserProfile() {
  const { userId } = useParams<{ userId: string }>()
  const navigate = useNavigate()
  const { user: currentUser, logout } = useAuth()

  const [profile, setProfile] = useState<UserProfileData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [showEditModal, setShowEditModal] = useState(false)
  const [editName, setEditName] = useState('')
  const [saving, setSaving] = useState(false)

  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [deleteEmail, setDeleteEmail] = useState('')
  const [deleteCode, setDeleteCode] = useState('')
  const [deletePassword, setDeletePassword] = useState('')
  const [codeSending, setCodeSending] = useState(false)
  const [codeSent, setCodeSent] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const fileInputRef = useRef<HTMLInputElement>(null)

  const isOwner = currentUser?.id === userId

  const loadProfile = () => {
    if (!userId) return
    setLoading(true)
    setError('')
    userApi.getProfile(userId)
      .then(res => setProfile(res.data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    loadProfile()
  }, [userId])

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !userId) return
    try {
      await userApi.uploadAvatar(userId, file)
      const res = await userApi.getProfile(userId)
      setProfile(res.data)
    } catch (err: any) {
      setError(err.message)
    }
  }

  const handleSaveProfile = async () => {
    if (!userId) return
    setSaving(true)
    try {
      await userApi.updateProfile(userId, { name: editName })
      const res = await userApi.getProfile(userId)
      setProfile(res.data)
      setShowEditModal(false)
    } catch (err: any) {
      setError(err.message)
    }
    setSaving(false)
  }

  const handleSendCode = async () => {
    if (!deleteEmail) return
    setCodeSending(true)
    try {
      await sendVerifyCode(deleteEmail)
      setCodeSent(true)
    } catch (err: any) {
      setError(err.message)
    }
    setCodeSending(false)
  }

  const handleDeleteAccount = async () => {
    if (!userId || !deletePassword || !deleteCode) return
    setDeleting(true)
    try {
      await userApi.deleteAccount(userId)
      localStorage.clear()
      window.location.href = '/login'
    } catch (err: any) {
      setError(err.message)
    }
    setDeleting(false)
  }

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const handleSendMessage = () => {
    if (!userId || !currentUser) return
    const ids = [currentUser.id, userId].sort()
    const conversationId = ids.join('_')
    navigate(`/messages/private/${conversationId}`)
  }

  const handleAddFriend = async () => {
    if (!userId) return
    const remark = prompt('请输入好友备注（可选）')
    if (remark === null) return
    try {
      await friendApi.sendRequest(userId, remark || '')
      loadProfile()
    } catch (err: any) {
      setError(err.message)
    }
  }

  const handleAcceptFriend = async () => {
    if (!userId) return
    try {
      const res = await friendApi.getReceivedRequests()
      const request = res.data.friend_requests.find(r => r.sender_id === userId)
      if (request) {
        await friendApi.acceptRequest(request.id)
        loadProfile()
      }
    } catch (err: any) {
      setError(err.message)
    }
  }

  const handleRejectFriend = async () => {
    if (!userId) return
    try {
      const res = await friendApi.getReceivedRequests()
      const request = res.data.friend_requests.find(r => r.sender_id === userId)
      if (request) {
        await friendApi.rejectRequest(request.id)
        loadProfile()
      }
    } catch (err: any) {
      setError(err.message)
    }
  }

  const handleDeleteFriend = async () => {
    if (!userId) return
    if (!confirm('确定删除好友？')) return
    try {
      const res = await friendApi.getList()
      const friend = res.data.friends.find(f => f.friend_id === userId)
      if (friend) {
        await friendApi.deleteFriend(friend.id)
        loadProfile()
      }
    } catch (err: any) {
      setError(err.message)
    }
  }

  const renderActionButtons = () => {
    if (isOwner) {
      return (
        <div className="flex justify-center gap-3 flex-wrap">
          <button onClick={() => { setEditName(profile?.user.name || ''); setShowEditModal(true) }}
            className="px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600">
            编辑资料
          </button>
          <button onClick={handleLogout}
            className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-50">
            退出登录
          </button>
          <button onClick={() => setShowDeleteModal(true)}
            className="px-4 py-2 bg-red-50 text-red-600 rounded-lg text-sm hover:bg-red-100">
            注销账号
          </button>
        </div>
      )
    }

    switch (profile?.friend_status) {
      case 'none':
        return (
          <div className="flex justify-center">
            <button onClick={handleAddFriend} className="px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600">
              添加好友
            </button>
          </div>
        )
      case 'friends':
        return (
          <div className="flex justify-center gap-3">
            <button onClick={handleSendMessage} className="px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600">
              发送消息
            </button>
            <button onClick={handleDeleteFriend} className="px-4 py-2 bg-red-50 text-red-600 rounded-lg text-sm hover:bg-red-100">
              删除好友
            </button>
          </div>
        )
      case 'pending_sent':
        return (
          <div className="flex justify-center">
            <span className="px-4 py-2 bg-gray-100 text-gray-500 rounded-lg text-sm">
              已发送申请
            </span>
          </div>
        )
      case 'pending_received':
        return (
          <div className="flex justify-center gap-3">
            <button onClick={handleAcceptFriend} className="px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600">
              接受
            </button>
            <button onClick={handleRejectFriend} className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg text-sm hover:bg-gray-200">
              拒绝
            </button>
          </div>
        )
      default:
        return null
    }
  }

  if (loading) {
    return (
      <div className="min-h-full flex items-center justify-center text-gray-400">
        加载中...
      </div>
    )
  }

  if (error && !profile) {
    return (
      <div className="min-h-full flex flex-col items-center justify-center text-gray-500">
        <p className="text-red-500 mb-4">{error}</p>
        <button onClick={() => navigate(-1)} className="text-emerald-500 hover:text-emerald-600 text-sm">
          ← 返回
        </button>
      </div>
    )
  }

  if (!profile) return null

  return (
    <div className="min-h-full">
      <div className="p-4">
        <button onClick={() => navigate(-1)} className="text-gray-400 hover:text-gray-600 text-sm">
          ← 返回
        </button>
      </div>

      <div className="max-w-2xl mx-auto px-4 pb-8">
        {/* Profile Header */}
        <div className="flex flex-col items-center mb-6">
          <div className="relative">
            <div className="w-24 h-24 rounded-full bg-gray-100 overflow-hidden">
              {profile.user.avatar ? (
                <img src={profile.user.avatar} alt="" className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-3xl text-gray-400 font-medium">
                  {profile.user.name?.charAt(0) || 'U'}
                </div>
              )}
            </div>
          </div>
          <h1 className="text-xl font-bold text-gray-800 mt-4">{profile.user.name}</h1>
          {profile.user.bio && <p className="text-gray-500 text-sm mt-1">{profile.user.bio}</p>}
          <p className="text-xs text-gray-400 mt-1">ID: {profile.user.id}</p>
        </div>

        {/* Action Buttons */}
        <div className="mb-8">
          {renderActionButtons()}
        </div>

        {/* Repos Tab */}
        <div>
          <h2 className="text-sm font-medium text-gray-700 mb-3">知识库</h2>
          {profile.repos.length === 0 ? (
            <p className="text-gray-400 text-sm">暂无知识库</p>
          ) : (
            <div className="space-y-1">
              {profile.repos.map((repo: any) => (
                <div
                  key={repo.id}
                  onClick={() => navigate(`/repos/${repo.id}`)}
                  className="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg cursor-pointer"
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-emerald-400 shrink-0" />
                      <span className="text-sm font-medium text-gray-800">{repo.name}</span>
                    </div>
                    {repo.description && <p className="text-xs text-gray-400 mt-0.5 truncate">{repo.description}</p>}
                    {repo.follower_count > 0 && (
                      <p className="text-xs text-gray-400 mt-0.5">{repo.follower_count} 关注</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Edit Profile Modal */}
      {showEditModal && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setShowEditModal(false)}>
          <div className="bg-white rounded-xl w-full max-w-sm p-6 mx-4" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-800 mb-4">编辑资料</h3>

            {/* Avatar */}
            <div className="flex justify-center mb-4">
              <div
                className="w-20 h-20 rounded-full bg-gray-100 overflow-hidden cursor-pointer ring-2 ring-emerald-200 hover:ring-emerald-400"
                onClick={() => fileInputRef.current?.click()}
              >
                {profile?.user.avatar ? (
                  <img src={profile.user.avatar} alt="" className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-2xl text-gray-400 font-medium">
                    {profile?.user.name?.charAt(0) || 'U'}
                  </div>
                )}
              </div>
              <input ref={fileInputRef} type="file" accept="image/*" onChange={handleAvatarUpload} className="hidden" />
            </div>

            {/* Name */}
            <label className="text-sm text-gray-600 mb-1 block">昵称</label>
            <input
              value={editName}
              onChange={e => setEditName(e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-400 mb-4"
              autoFocus
            />

            <div className="flex gap-2 justify-end">
              <button onClick={() => setShowEditModal(false)}
                className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700">
                取消
              </button>
              <button onClick={handleSaveProfile} disabled={saving || !editName.trim()}
                className="px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600 disabled:opacity-50">
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Account Modal */}
      {showDeleteModal && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setShowDeleteModal(false)}>
          <div className="bg-white rounded-xl w-full max-w-sm p-6 mx-4" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-800 mb-2">注销账号</h3>
            <p className="text-sm text-gray-500 mb-5">此操作不可撤销。请验证身份后继续。</p>

            <label className="text-sm text-gray-600 mb-1 block">邮箱</label>
            <input
              value={deleteEmail}
              onChange={e => setDeleteEmail(e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-400 mb-3"
              placeholder="请输入注册邮箱"
            />

            <div className="flex gap-2 mb-3">
              <div className="flex-1">
                <label className="text-sm text-gray-600 mb-1 block">验证码</label>
                <input
                  value={deleteCode}
                  onChange={e => setDeleteCode(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-400"
                  placeholder="请输入验证码"
                />
              </div>
              <button
                onClick={handleSendCode}
                disabled={codeSending || !deleteEmail}
                className="self-end px-3 py-2 border border-emerald-500 text-emerald-600 rounded-lg text-sm hover:bg-emerald-50 disabled:opacity-50 shrink-0"
              >
                {codeSending ? '发送中...' : codeSent ? '已发送' : '发送验证码'}
              </button>
            </div>

            <label className="text-sm text-gray-600 mb-1 block">密码</label>
            <input
              type="password"
              value={deletePassword}
              onChange={e => setDeletePassword(e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-400 mb-5"
              placeholder="请输入密码"
            />

            <div className="flex gap-2 justify-end">
              <button onClick={() => setShowDeleteModal(false)}
                className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700">
                取消
              </button>
              <button
                onClick={handleDeleteAccount}
                disabled={deleting || !deleteEmail || !deleteCode || !deletePassword}
                className="px-4 py-2 bg-red-500 text-white rounded-lg text-sm hover:bg-red-600 disabled:opacity-50"
              >
                {deleting ? '处理中...' : '确认注销'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
