import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import ImgCrop from 'antd-img-crop'
import { Upload } from 'antd'
import { register, sendVerifyCode } from '../lib/auth'

export default function Register() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [verifyCode, setVerifyCode] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [codeSent, setCodeSent] = useState(false)
  const [codeSending, setCodeSending] = useState(false)
  const [avatarFile, setAvatarFile] = useState<File | null>(null)
  const [avatarPreview, setAvatarPreview] = useState<string>('/img/default_avatar.jpg')

  const handleSendCode = async () => {
    if (!email) return
    setCodeSending(true)
    try {
      await sendVerifyCode(email)
      setCodeSent(true)
    } catch (err: any) {
      setError(err.message || '发送验证码失败')
    }
    setCodeSending(false)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await register(name, email, password, verifyCode, avatarFile || undefined)
      navigate('/login')
    } catch (err: any) {
      setError(err.message || '注册失败')
    }
    setLoading(false)
  }

  return (
    <div className="w-full max-w-sm mx-auto">
      <h1 className="text-2xl font-bold text-center text-gray-800 mb-2">注册 KnowSync</h1>
      <p className="text-sm text-gray-500 text-center mb-6">创建你的账号</p>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* 头像 */}
        <div className="flex flex-col items-center">
          <label className="block text-sm font-medium text-gray-700 mb-3 self-start">头像</label>
          <div className="relative w-20 h-20 rounded-full overflow-hidden border-2 border-gray-200">
            <img src={avatarPreview} alt="avatar" className="w-full h-full object-cover"
              onError={() => setAvatarPreview('/img/default_avatar.jpg')} />
          </div>
          <div className="mt-3">
            <ImgCrop aspect={1} quality={1} modalTitle="裁剪头像">
              <Upload
                showUploadList={false}
                beforeUpload={(file) => {
                  setAvatarFile(file)
                  setAvatarPreview(URL.createObjectURL(file))
                  return false
                }}
              >
                <span className="text-sm text-emerald-600 cursor-pointer hover:text-emerald-500">
                  选择图片并裁剪
                </span>
              </Upload>
            </ImgCrop>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">用户名</label>
          <input type="text" value={name} onChange={e => setName(e.target.value)} required
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
          <input type="email" value={email} onChange={e => setEmail(e.target.value)} required
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">密码</label>
          <input type="password" value={password} onChange={e => setPassword(e.target.value)} required minLength={6}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">验证码</label>
          <div className="flex gap-2">
            <input type="text" value={verifyCode} onChange={e => setVerifyCode(e.target.value)} required
              className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            <button type="button" onClick={handleSendCode} disabled={codeSending || !email}
              className="px-3 py-2 bg-gray-100 text-gray-600 rounded-lg text-sm hover:bg-gray-200 disabled:opacity-50 shrink-0">
              {codeSending ? '发送中' : codeSent ? '已发送' : '发送'}
            </button>
          </div>
        </div>

        {error && <p className="text-red-500 text-sm">{error}</p>}

        <button type="submit" disabled={loading}
          className="w-full py-2 bg-emerald-500 text-white rounded-lg font-medium text-sm hover:bg-emerald-600 disabled:opacity-50">
          {loading ? '注册中...' : '注册'}
        </button>
      </form>

      <p className="text-sm text-gray-500 text-center mt-6">
        已有账号？ <Link to="/login" className="text-emerald-600 hover:underline">登录</Link>
      </p>
    </div>
  )
}
