import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { getRepo } from '../lib/repos'
import type { Repo } from '../lib/repos'
import RepoTree from '../components/Repo/RepoTree'
import CollabList from '../components/Repo/CollabList'

export default function RepoDetail() {
  const { repoId } = useParams<{ repoId: string }>()
  const [repo, setRepo] = useState<Repo | null>(null)

  useEffect(() => {
    if (!repoId) return
    getRepo(repoId).then(setRepo)
  }, [repoId])

  return (
    <div className="h-full flex">
      {/* 左侧文件树 */}
      <div className="w-60 border-r border-gray-200 bg-gray-50 overflow-y-auto p-2 shrink-0">
        {repo && repoId && (
          <RepoTree repoId={repoId} />
        )}
      </div>

      {/* 右侧内容区 */}
      <div className="flex-1 overflow-y-auto p-6">
        {repo && (
          <div className="mb-6">
            <h2 className="text-lg font-medium text-gray-800">{repo.name}</h2>
            <p className="text-sm text-gray-500 mt-1">{repo.description}</p>
          </div>
        )}
        <div className="text-sm text-gray-400 text-center py-12">
          选择一篇文章查看或编辑
        </div>

        {repoId && <CollabList repoId={repoId} />}
      </div>
    </div>
  )
}
