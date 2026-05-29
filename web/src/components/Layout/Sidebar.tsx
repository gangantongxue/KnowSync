import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useRepo } from '../../store/repo-context'

interface SidebarProps {
  /** 如果为 true，显示文件树模式（带返回按钮） */
  isFileTree?: boolean
  repoId?: string
  onBack?: () => void
}

export default function Sidebar({ isFileTree, repoId, onBack }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { repos } = useRepo()

  if (collapsed) {
    return (
      <div className="w-12 border-r border-gray-200 bg-gray-50 flex flex-col items-center py-2 gap-3 shrink-0">
        <button onClick={() => setCollapsed(false)} className="text-gray-400 hover:text-gray-600 text-lg">☰</button>
        <button onClick={() => navigate('/chat')} className={`p-2 rounded-lg ${location.pathname.startsWith('/chat') ? 'bg-emerald-100 text-emerald-600' : 'text-gray-400 hover:text-gray-600'}`} title="AI 对话">💬</button>
        {repos.slice(0, 3).map(r => (
          <button key={r.id} onClick={() => navigate(`/repos/${r.id}`)} className="w-8 h-8 rounded bg-gray-200 text-xs text-gray-600 hover:bg-gray-300" title={r.name}>
            {r.name.charAt(0)}
          </button>
        ))}
      </div>
    )
  }

  return (
    <div className="w-56 border-r border-gray-200 bg-gray-50 flex flex-col shrink-0 overflow-hidden">
      {/* 文件树模式：显示返回按钮 */}
      {isFileTree ? (
        <div className="p-3 border-b border-gray-200">
          <button
            onClick={onBack}
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-emerald-600"
          >
            ← 返回所有知识库
          </button>
        </div>
      ) : (
        <div className="p-3 border-b border-gray-200">
          <button
            onClick={() => navigate('/chat')}
            className={`flex items-center gap-2 w-full px-2 py-1.5 rounded text-sm ${location.pathname.startsWith('/chat') ? 'bg-emerald-100 text-emerald-700 font-medium' : 'text-gray-600 hover:bg-gray-100'}`}
          >
            💬 AI 对话
          </button>
        </div>
      )}

      {/* 仓库列表或文件树 */}
      <div className="flex-1 overflow-y-auto p-2">
        {isFileTree ? (
          <FileTree repoId={repoId!} />
        ) : (
          <>
            <div className="text-xs text-gray-400 font-medium px-2 py-1">知识库</div>
            {repos.map(repo => (
              <button
                key={repo.id}
                onClick={() => navigate(`/repos/${repo.id}`)}
                className="flex items-center gap-2 w-full px-2 py-1.5 rounded text-sm text-gray-700 hover:bg-gray-100 mb-0.5"
              >
                <span className="w-5 h-5 rounded bg-emerald-100 text-emerald-700 text-xs flex items-center justify-center shrink-0">
                  {repo.name.charAt(0)}
                </span>
                <span className="truncate">{repo.name}</span>
              </button>
            ))}
          </>
        )}
      </div>

      {/* 折叠按钮 */}
      <div className="p-2 border-t border-gray-200">
        <button
          onClick={() => setCollapsed(true)}
          className="text-xs text-gray-400 hover:text-gray-600 w-full text-left"
        >
          ◀ 折叠
        </button>
      </div>
    </div>
  )
}

// 简化的文件树组件，后续 Task 会完善
function FileTree(_props: { repoId: string }) {
  return (
    <div className="text-sm text-gray-500 px-2 py-4 text-center">
      文件树加载中...
      <br />
      <span className="text-xs">(Task 7 完善)</span>
    </div>
  )
}
