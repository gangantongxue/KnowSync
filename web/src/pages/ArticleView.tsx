import { useEffect, useState, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAuth } from '../store/auth-context'
import { getRepo } from '../lib/repos'
import { resolveImageUrls, markdownComponents } from '../lib/markdown'
import { isImageFile } from '../components/Repo/RepoTree'

const MARKDOWN_EXTENSIONS = new Set(['.md', '.markdown', '.mdown', '.mkd', '.mkdn', '.mdwn'])
const TEXT_EXTENSIONS = new Set(['.txt', '.text', '.log', '.csv', '.json', '.xml', '.yml', '.yaml', '.toml', '.ini', '.cfg', '.conf'])

function isMarkdown(name: string): boolean {
  const ext = name.toLowerCase().slice(name.lastIndexOf('.'))
  return MARKDOWN_EXTENSIONS.has(ext)
}

function isTextViewable(name: string): boolean {
  const ext = name.toLowerCase().slice(name.lastIndexOf('.'))
  return TEXT_EXTENSIONS.has(ext)
}

export default function ArticleView() {
  const { repoId, '*': filePath } = useParams<{ repoId: string; '*': string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [content, setContent] = useState('')
  const [imageUrl, setImageUrl] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const abortRef = useRef<AbortController | null>(null)

  const fileName = filePath?.split('/').pop() || ''
  const showImage = filePath ? isImageFile(filePath) : false

  useEffect(() => {
    if (!repoId || !filePath || !user) return

    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller

    setContent('')
    setImageUrl('')
    setError('')
    setLoading(true)

    const load = async () => {
      try {
        const repo = await getRepo(repoId)
        if (controller.signal.aborted) return

        const baseUrl = `/files/${repo.owner_id}/${repoId}`

        if (isImageFile(filePath)) {
          setImageUrl(`${baseUrl}/${encodeURIComponent(filePath)}`)
          setLoading(false)
          return
        }

        const res = await fetch(`/files/${repo.owner_id}/${repoId}/${encodeURIComponent(filePath)}`, {
          signal: controller.signal,
        })
        if (!res.ok) throw new Error('文件不存在')
        const text = await res.text()
        if (!controller.signal.aborted) {
          if (isMarkdown(filePath)) {
            const dir = filePath.includes('/') ? filePath.substring(0, filePath.lastIndexOf('/')) : ''
            const processed = resolveImageUrls(text, dir, baseUrl)
            setContent(processed)
          } else {
            setContent(text)
          }
          setLoading(false)
        }
      } catch (err: any) {
        if (err.name === 'AbortError') return
        if (!controller.signal.aborted) {
          setError(err.message || '文件加载失败')
          setLoading(false)
        }
      }
    }
    load()

    return () => {
      controller.abort()
    }
  }, [repoId, filePath, user])

  const isMd = filePath ? isMarkdown(filePath) : false
  const isTxt = filePath ? isTextViewable(filePath) : false

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white shrink-0">
        <div className="flex items-center gap-2">
          <button onClick={() => navigate(`/repos/${repoId}`)} className="text-gray-400 hover:text-gray-600 text-sm">
            ← 返回
          </button>
          <h1 className="text-sm font-medium text-gray-800 truncate">{fileName}</h1>
        </div>
        {(isMd || isTxt) && (
          <button
            onClick={() => navigate(`/repos/${repoId}/edit/${filePath}`)}
            className="px-3 py-1 text-xs bg-emerald-500 text-white rounded hover:bg-emerald-600"
          >
            编辑
          </button>
        )}
      </div>
      <div className="flex-1 overflow-y-auto p-6 max-w-3xl mx-auto w-full">
        {loading ? (
          <div className="text-gray-400 text-sm text-center py-12">加载中...</div>
        ) : error ? (
          <div className="text-red-400 text-sm text-center py-12">{error}</div>
        ) : showImage ? (
          <div className="flex items-center justify-center min-h-[200px]">
            <img
              src={imageUrl}
              alt={fileName}
              className="max-w-full max-h-[80vh] h-auto rounded-lg shadow-md"
            />
          </div>
        ) : isMd ? (
          <div>
            <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>{content}</ReactMarkdown>
          </div>
        ) : isTxt ? (
          <pre className="text-sm text-gray-700 font-mono whitespace-pre-wrap break-words bg-gray-50 rounded-lg p-4 border border-gray-200">
            {content}
          </pre>
        ) : (
          <div className="text-sm text-gray-400 text-center py-12">
            不支持预览此类型文件
          </div>
        )}
      </div>
    </div>
  )
}
