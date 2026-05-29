import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getArticleSignedUrl, getNode } from '../lib/nodes'
import { Streamdown } from '../components/Chat/Streamdown'

export default function ArticleView() {
  const { repoId, nodeId } = useParams<{ repoId: string; nodeId: string }>()
  const navigate = useNavigate()
  const [content, setContent] = useState('')
  const [title, setTitle] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      if (!repoId || !nodeId) return
      try {
        const node = await getNode(repoId, nodeId)
        setTitle(node.name)

        const signedUrl = await getArticleSignedUrl(repoId, nodeId)
        const res = await fetch(signedUrl)
        const text = await res.text()
        setContent(text)
      } catch {
        setContent('*文章加载失败*')
      }
      setLoading(false)
    }
    load()
  }, [repoId, nodeId])

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white">
        <h1 className="text-sm font-medium text-gray-800">{title}</h1>
        <button
          onClick={() => navigate(`/repos/${repoId}/nodes/${nodeId}/edit`)}
          className="px-3 py-1 text-xs bg-emerald-500 text-white rounded hover:bg-emerald-600"
        >
          编辑
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-6 max-w-3xl mx-auto w-full">
        {loading ? (
          <div className="text-gray-400 text-sm">加载中...</div>
        ) : (
          <Streamdown content={content} />
        )}
      </div>
    </div>
  )
}
