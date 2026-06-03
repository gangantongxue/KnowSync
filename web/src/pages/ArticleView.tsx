import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAuth } from '../store/auth-context'
import { getRepo } from '../lib/repos'

export default function ArticleView() {
  const { repoId, '*': filePath } = useParams<{ repoId: string; '*': string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const load = async () => {
      if (!repoId || !filePath || !user) return
      try {
        const repo = await getRepo(repoId)
        const res = await fetch(`/files/${repo.owner_id}/${repoId}/${filePath}`)
        if (!res.ok) throw new Error('文件不存在')
        const text = await res.text()
        setContent(text)
      } catch (err: any) {
        setError(err.message || '文章加载失败')
      }
      setLoading(false)
    }
    load()
  }, [repoId, filePath, user])

  const fileName = filePath?.split('/').pop() || ''

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white shrink-0">
        <h1 className="text-sm font-medium text-gray-800 truncate">{fileName}</h1>
        <button
          onClick={() => navigate(`/repos/${repoId}/edit/${filePath}`)}
          className="px-3 py-1 text-xs bg-emerald-500 text-white rounded hover:bg-emerald-600"
        >
          编辑
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-6 max-w-3xl mx-auto w-full">
        {loading ? (
          <div className="text-gray-400 text-sm">加载中...</div>
        ) : error ? (
          <div className="text-red-400 text-sm">{error}</div>
        ) : (
          <div className="prose prose-sm max-w-none">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
          </div>
        )}
      </div>
    </div>
  )
}
