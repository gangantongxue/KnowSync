import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getArticleSignedUrl, getNode, uploadArticleContent } from '../lib/nodes'
import MarkdownEditor from '../components/Editor/MarkdownEditor'

export default function ArticleEditor() {
  const { repoId, nodeId } = useParams<{ repoId: string; nodeId: string }>()
  const navigate = useNavigate()
  const [content, setContent] = useState('')
  const [title, setTitle] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    const load = async () => {
      if (!repoId || !nodeId) return
      try {
        const node = await getNode(repoId, nodeId)
        setTitle(node.name)

        if (node.file_path) {
          const signedUrl = await getArticleSignedUrl(repoId, nodeId)
          const res = await fetch(signedUrl)
          const text = await res.text()
          setContent(text)
        }
      } catch {
        // 新文章，内容为空
      }
      setLoading(false)
    }
    load()
  }, [repoId, nodeId])

  const handleSave = async () => {
    if (!repoId || !nodeId || !content.trim()) return
    setSaving(true)
    try {
      const blob = new Blob([content], { type: 'text/markdown' })
      const file = new File([blob], `${title || 'article'}.md`, { type: 'text/markdown' })
      await uploadArticleContent(repoId, nodeId, file)
      setDirty(false)
    } catch (err: any) {
      alert('保存失败: ' + err.message)
    }
    setSaving(false)
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.name.endsWith('.md') && !file.name.endsWith('.markdown')) {
      alert('仅支持 .md / .markdown 文件')
      return
    }
    const text = await file.text()
    setContent(text)
    setTitle(file.name.replace(/\.(md|markdown)$/, ''))
    setDirty(true)
  }

  if (loading) {
    return <div className="h-full flex items-center justify-center text-gray-400">加载中...</div>
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white">
        <div className="flex items-center gap-2">
          <button onClick={() => navigate(`/repos/${repoId}/nodes/${nodeId}`)} className="text-gray-400 hover:text-gray-600 text-sm">
            ← 返回
          </button>
          <span className="text-sm font-medium text-gray-800">{title || '新文章'}</span>
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
        <MarkdownEditor value={content} onChange={(v) => { setContent(v); setDirty(true) }} />
      </div>
    </div>
  )
}
