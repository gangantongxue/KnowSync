import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { getRepo } from '../lib/repos'
import { listNodes } from '../lib/nodes'
import type { Repo } from '../lib/repos'
import RepoTree from '../components/Repo/RepoTree'
import CollabList from '../components/Repo/CollabList'
import type { Node } from '../lib/nodes'

export default function RepoDetail() {
  const { repoId } = useParams<{ repoId: string }>()
  const [repo, setRepo] = useState<Repo | null>(null)
  const [nodes, setNodes] = useState<Node[]>([])
  const [currentParentId, setCurrentParentId] = useState<string | null>(null)

  const loadRepo = async () => {
    if (!repoId) return
    const r = await getRepo(repoId)
    setRepo(r)
  }

  const loadAllNodes = async () => {
    if (!repoId) return
    // 当前简化：只加载一层（完整树需要递归加载）
    const all: Node[] = []
    const loadRecursive = async (parentId?: string) => {
      const children = await listNodes(repoId, parentId)
      all.push(...children)
      for (const child of children) {
        if (child.type === 'FOLDER') {
          await loadRecursive(child.id)
        }
      }
    }
    await loadRecursive()
    setNodes(all)
  }

  useEffect(() => { loadRepo(); loadAllNodes() }, [repoId])

  return (
    <div className="h-full flex">
      {/* 左侧文件树 */}
      <div className="w-60 border-r border-gray-200 bg-gray-50 overflow-y-auto p-2 shrink-0">
        <RepoTree
          repoId={repoId!}
          nodes={nodes}
          currentParentId={currentParentId}
          onNavigate={setCurrentParentId}
          onRefresh={loadAllNodes}
        />
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

        <CollabList repoId={repoId!} />
      </div>
    </div>
  )
}
