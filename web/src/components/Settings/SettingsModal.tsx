import { useState, useEffect } from 'react'
import { useAuth } from '../../store/auth-context'
import { getUser, updateUser, uploadAvatar, changePassword } from '../../lib/auth'
import { Upload } from 'antd'
import ImgCrop from 'antd-img-crop'
import { UserOutlined } from '@ant-design/icons'

interface SettingsModalProps {
  onClose: () => void
}

export default function SettingsModal({ onClose }: SettingsModalProps) {
  const { user, refreshUser } = useAuth()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [avatarUrl, setAvatarUrl] = useState('')
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    const load = async () => {
      if (!user?.id) return
      const u = await getUser(user.id)
      setName(u.name)
      setEmail(u.email)
      setAvatarUrl(u.avatar)
    }
    load()
  }, [user])

  const handleSaveProfile = async () => {
    if (!user?.id) return
    setSaving(true)
    try {
      await updateUser(user.id, { name, email })
      await refreshUser()
      setMessage('个人信息已更新')
    } catch (err: any) {
      setMessage(err.message || '更新失败')
    }
    setSaving(false)
  }

  const handleAvatarUpload = async (file: File) => {
    if (!user?.id) return
    try {
      const url = await uploadAvatar(user.id, file)
      setAvatarUrl(url)
      await refreshUser()
      setMessage('头像已更新')
    } catch (err: any) {
      setMessage(err.message || '头像上传失败')
    }
  }

  const handleChangePassword = async () => {
    if (!user?.id || !oldPassword || !newPassword) return
    setSaving(true)
    try {
      await changePassword(user.id, oldPassword, newPassword)
      setMessage('密码已修改')
      setOldPassword('')
      setNewPassword('')
    } catch (err: any) {
      setMessage(err.message || '密码修改失败')
    }
    setSaving(false)
  }

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-base font-medium text-gray-800">设置</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">✕</button>
          </div>

          {/* Avatar */}
          <div className="flex flex-col items-center mb-6">
            <ImgCrop rotationSlider>
              <Upload
                showUploadList={false}
                beforeUpload={(file) => { handleAvatarUpload(file as File); return false }}
                accept=".jpg,.jpeg,.png,.webp"
              >
                <div className="w-14 h-14 rounded-full bg-gray-100 flex items-center justify-center cursor-pointer overflow-hidden border-2 border-dashed border-gray-300 hover:border-emerald-400">
                  {avatarUrl ? (
                    <img src={avatarUrl} className="w-full h-full object-cover" />
                  ) : (
                    <UserOutlined className="text-gray-400 text-xl" />
                  )}
                </div>
              </Upload>
            </ImgCrop>
            <div className="mt-2 text-center">
              <div className="text-sm font-medium text-gray-800">{name}</div>
              <div className="text-xs text-gray-400">点击头像更换</div>
            </div>
          </div>

          {/* Info */}
          <div className="space-y-3 mb-6">
            <div>
              <label className="block text-xs text-gray-500 mb-1">用户名</label>
              <input value={name} onChange={e => setName(e.target.value)}
                className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            </div>
            <div>
              <label className="block text-xs text-gray-500 mb-1">邮箱</label>
              <input value={email} type="email" readOnly
                className="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm bg-gray-50 text-gray-500 cursor-not-allowed focus:outline-none" />
            </div>
            <button onClick={handleSaveProfile} disabled={saving}
              className="px-4 py-1.5 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600 disabled:opacity-50">
              保存信息
            </button>
          </div>

          <hr className="my-4" />

          {/* Password */}
          <div className="space-y-3">
            <h3 className="text-sm font-medium text-gray-700">修改密码</h3>
            <input value={oldPassword} onChange={e => setOldPassword(e.target.value)} type="password" placeholder="当前密码"
              className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            <input value={newPassword} onChange={e => setNewPassword(e.target.value)} type="password" placeholder="新密码"
              className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            <button onClick={handleChangePassword} disabled={saving || !oldPassword || !newPassword}
              className="px-4 py-1.5 bg-gray-100 text-gray-700 rounded-lg text-sm hover:bg-gray-200 disabled:opacity-50">
              修改密码
            </button>
          </div>

          {message && <p className="text-xs text-emerald-600 mt-4">{message}</p>}
        </div>
      </div>
    </div>
  )
}
