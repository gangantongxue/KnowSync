import { useEffect, useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { request } from '../lib/client'
import { getRepo } from '../lib/repos'
import type { Repo } from '../lib/repos'

export default function SearchResult() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q') || ''
  const navigate = useNavigate()
  const [repos, setRepos] = useState<Repo[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)

  useEffect(() => {
    const search = async () => {
      if (!query.trim()) { setLoading(false); return }
      setLoading(true)
      try {
        const res = await request<{ repo_ids: string[]; total_pages: number; has_more: boolean }>('/ai/search', {
          method: 'POST',
          body: JSON.stringify({ query: query.trim(), page, page_size: 20 }),
          skipAuth: false,
        })
        setHasMore(res.data.has_more)

        const details = await Promise.all(
          (res.data.repo_ids || []).map(id => getRepo(id).catch(() => null))
        )
        setRepos(details.filter(Boolean) as Repo[])
      } catch {
        setRepos([])
      }
      setLoading(false)
    }
    search()
  }, [query, page])

  return (
    <div className="max-w-3xl mx-auto px-6 py-8">
      <h2 className="text-base font-medium text-gray-800 mb-1">搜索结果</h2>
      <p className="text-xs text-gray-400 mb-6">关键词: "{query}"</p>

      {loading ? (
        <div className="text-gray-400 text-sm">搜索中...</div>
      ) : repos.length === 0 ? (
        <div className="text-gray-400 text-sm py-8 text-center">未找到匹配的知识库</div>
      ) : (
        <div className="space-y-3">
          {repos.map(repo => (
            <div
              key={repo.id}
              onClick={() => navigate(`/repos/${repo.id}`)}
              className="p-4 border border-gray-200 rounded-lg hover:border-emerald-300 cursor-pointer transition-colors"
            >
              <h3 className="text-sm font-medium text-gray-800">{repo.name}</h3>
              <p className="text-xs text-gray-500 mt-1">{repo.description || '暂无描述'}</p>
              <div className="flex gap-3 mt-2 text-xs text-gray-400">
                <span>{repo.article_count} 篇文章</span>
                <span>{repo.visibility === 'PUBLIC' ? '公开' : '私有'}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {hasMore && (
        <div className="text-center mt-6">
          <button onClick={() => setPage(p => p + 1)} className="text-sm text-emerald-600 hover:underline">
            加载更多
          </button>
        </div>
      )}
    </div>
  )
}
