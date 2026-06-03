import { useEffect, useState, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAuth } from '../store/auth-context'
import { getRepo } from '../lib/repos'

/** 解析相对路径，处理 ./ 和 ../ */
function resolveRelativePath(baseDir: string, relative: string): string {
  let path = relative.replace(/^\.\/+/, '')
  const parts = baseDir ? baseDir.split('/') : []
  while (path.startsWith('../')) {
    if (parts.length > 0) parts.pop()
    path = path.slice(3)
  }
  if (parts.length > 0) {
    return parts.join('/') + '/' + path
  }
  return path
}

/** 将 markdown 中的相对路径图片转为绝对文件 URL */
function resolveImageUrls(md: string, fileDir: string, baseUrl: string): string {
  return md.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_match, alt, src) => {
    if (/^https?:\/\//.test(src) || src.startsWith('/')) return _match
    return `![${alt}](${baseUrl}/${resolveRelativePath(fileDir, src)})`
  })
}

export default function ArticleView() {
  const { repoId, '*': filePath } = useParams<{ repoId: string; '*': string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (!repoId || !filePath || !user) return

    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller

    setContent('')
    setError('')
    setLoading(true)

    const load = async () => {
      try {
        const repo = await getRepo(repoId)
        if (controller.signal.aborted) return
        const res = await fetch(`/files/${repo.owner_id}/${repoId}/${encodeURIComponent(filePath)}`, {
          signal: controller.signal,
        })
        if (!res.ok) throw new Error('文件不存在')
        const text = await res.text()
        if (!controller.signal.aborted) {
          const baseUrl = `/files/${repo.owner_id}/${repoId}`
          const dir = filePath.includes('/') ? filePath.substring(0, filePath.lastIndexOf('/')) : ''
          const processed = resolveImageUrls(text, dir, baseUrl)
          setContent(processed)
          setLoading(false)
        }
      } catch (err: any) {
        if (err.name === 'AbortError') return
        if (!controller.signal.aborted) {
          setError(err.message || '文章加载失败')
          setLoading(false)
        }
      }
    }
    load()

    return () => {
      controller.abort()
    }
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
