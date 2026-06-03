import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import MDEditor from '@uiw/react-md-editor'
import { useAuth } from '../store/auth-context'
import { getRepo } from '../lib/repos'
import { uploadFile } from '../lib/files'

export default function ArticleEditor() {
  const { repoId, '*': filePath } = useParams<{ repoId: string; '*': string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (!repoId || !user) return

    const load = async () => {
      try {
        const repo = await getRepo(repoId)

        if (filePath) {
          const res = await fetch(`/files/${repo.owner_id}/${repoId}/${filePath}`)
          if (res.ok) {
            const text = await res.text()
            setContent(text)
          }
        }
      } catch { /* ignore */ }
      setLoading(false)
    }
    load()
  }, [repoId, filePath, user])

  const handleSave = async () => {
    if (!repoId || !user) return
    const path = filePath || prompt('请输入文件名（含 .md 后缀）:')
    if (!path) return

    setSaving(true)
    try {
      const blob = new Blob([content], { type: 'text/markdown' })
      const file = new File([blob], path.split('/').pop() || 'article.md', { type: 'text/markdown' })
      await uploadFile(repoId, path, file)
      setDirty(false)
    } catch (err: any) {
      alert('保存失败: ' + err.message)
    }
    setSaving(false)
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const text = await file.text()
    setContent(text)
    setDirty(true)
  }

  if (loading) {
    return <div className="h-full flex items-center justify-center text-gray-400">加载中...</div>
  }

  return (
    <div className="h-full flex flex-col" data-color-mode="light">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white shrink-0">
        <div className="flex items-center gap-2">
          <button onClick={() => navigate(`/repos/${repoId}`)} className="text-gray-400 hover:text-gray-600 text-sm">
            ← 返回
          </button>
          <span className="text-sm font-medium text-gray-800">{filePath?.split('/').pop() || '新文章'}</span>
          {dirty && <span className="text-xs text-orange-500">未保存</span>}
        </div>
        <div className="flex items-center gap-2">
          <label className="px-3 py-1 text-xs border border-gray-200 rounded cursor-pointer hover:bg-gray-50">
            上传 .md 文件
            <input type="file" accept=".md,.markdown" onChange={handleFileUpload} className="hidden" />
          </label>
          <button
            onClick={handleSave}
            disabled={saving || !dirty}
            className="px-3 py-1 text-xs bg-emerald-500 text-white rounded hover:bg-emerald-600 disabled:opacity-50"
          >
            {saving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>
      <div className="flex-1">
        <MDEditor
          value={content}
          onChange={(v) => { setContent(v || ''); setDirty(true) }}
          height="100%"
          preview="live"
        />
      </div>
    </div>
  )
}
