import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { getRepoTree, deleteFile, renameFile, makeDir } from '../../lib/files'
import type { FileEntry } from '../../lib/files'

interface RepoTreeProps {
  repoId: string
}

export default function RepoTree({ repoId }: RepoTreeProps) {
  const [currentPath, setCurrentPath] = useState('')
  const [entries, setEntries] = useState<FileEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; entry?: FileEntry } | null>(null)
  const navigate = useNavigate()

  const loadDir = async (dirPath?: string) => {
    setLoading(true)
    try {
      const tree = await getRepoTree(repoId, dirPath)
      setEntries(tree.entries)
      setCurrentPath(tree.path || '')
    } catch {
      setEntries([])
    }
    setLoading(false)
  }

  useEffect(() => { loadDir('') }, [repoId])

  const handleCreateFile = async () => {
    const name = prompt('请输入文件名称（含 .md 后缀）:')
    if (!name?.trim()) return
    const filePath = currentPath ? `${currentPath}/${name.trim()}` : name.trim()
    navigate(`/repos/${repoId}/edit/${filePath}`)
    setContextMenu(null)
  }

  const handleCreateDir = async () => {
    const name = prompt('请输入文件夹名称:')
    if (!name?.trim()) return
    const dirPath = currentPath ? `${currentPath}/${name.trim()}` : name.trim()
    try {
      await makeDir(repoId, dirPath)
      loadDir(currentPath)
    } catch (err: any) {
      alert('创建失败: ' + err.message)
    }
    setContextMenu(null)
  }

  const handleDelete = async (entry: FileEntry) => {
    if (!confirm(`确定删除「${entry.name}」？`)) return
    const filePath = currentPath ? `${currentPath}/${entry.name}` : entry.name
    try {
      await deleteFile(repoId, filePath)
      loadDir(currentPath)
    } catch (err: any) {
      alert('删除失败: ' + err.message)
    }
    setContextMenu(null)
  }

  const handleRename = async (entry: FileEntry) => {
    const newName = prompt('新名称:', entry.name)
    if (!newName?.trim() || newName === entry.name) return
    const oldPath = currentPath ? `${currentPath}/${entry.name}` : entry.name
    const newPath = currentPath ? `${currentPath}/${newName.trim()}` : newName.trim()
    try {
      await renameFile(repoId, oldPath, newPath)
      loadDir(currentPath)
    } catch (err: any) {
      alert('重命名失败: ' + err.message)
    }
    setContextMenu(null)
  }

  const handleUploadFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const filePath = currentPath ? `${currentPath}/${file.name}` : file.name
    try {
      // Upload via fetch directly
      const token = localStorage.getItem('access_token')
      const formData = new FormData()
      formData.append('file', file)

      const res = await fetch(`/api/v1/repos/${repoId}/files?path=${encodeURIComponent(filePath)}`, {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: formData,
      })

      if (!res.ok) {
        const err = await res.json().catch(() => ({ message: 'Upload failed' }))
        throw new Error(err.message || `HTTP ${res.status}`)
      }

      loadDir(currentPath)
    } catch (err: any) {
      alert('上传失败: ' + err.message)
    }
  }

  const sorted = [...entries].sort((a, b) => {
    if (a.type !== b.type) return a.type === 'dir' ? -1 : 1
    return a.name.localeCompare(b.name)
  })

  const pathParts = currentPath ? currentPath.split('/').filter(Boolean) : []

  return (
    <div onClick={() => setContextMenu(null)}>
      {/* 面包屑 */}
      <div className="flex items-center gap-1 text-xs text-gray-500 mb-2 px-1 flex-wrap">
        <button onClick={() => loadDir('')} className="hover:text-emerald-600">根目录</button>
        {pathParts.map((part, i) => {
          const fullPath = pathParts.slice(0, i + 1).join('/')
          return (
            <span key={i} className="flex items-center gap-1">
              <span>/</span>
              <button onClick={() => loadDir(fullPath)} className="hover:text-emerald-600">{part}</button>
            </span>
          )
        })}
      </div>

      {/* 操作按钮 */}
      <div className="flex items-center gap-1 mb-1" onContextMenu={e => { e.preventDefault(); setContextMenu({ x: e.clientX, y: e.clientY }) }}>
        <button onClick={handleCreateDir} className="px-2 py-1 text-xs text-gray-400 hover:text-emerald-600 hover:bg-gray-100 rounded">
          + 文件夹
        </button>
        <button onClick={handleCreateFile} className="px-2 py-1 text-xs text-gray-400 hover:text-emerald-600 hover:bg-gray-100 rounded">
          + 文章
        </button>
        <label className="px-2 py-1 text-xs text-gray-400 hover:text-emerald-600 hover:bg-gray-100 rounded cursor-pointer">
          + 上传
          <input type="file" onChange={handleUploadFile} className="hidden" />
        </label>
      </div>

      {/* 目录内容 */}
      {loading ? (
        <div className="text-xs text-gray-400 px-2 py-4 text-center">加载中...</div>
      ) : sorted.length === 0 ? (
        <div className="text-xs text-gray-400 px-2 py-4 text-center">空目录</div>
      ) : (
        sorted.map(entry => (
          <div
            key={entry.name}
            className="flex items-center gap-2 px-2 py-1.5 rounded text-sm cursor-pointer hover:bg-gray-100 text-gray-700 mb-0.5"
            onClick={() => {
              if (entry.type === 'dir') {
                const newPath = currentPath ? `${currentPath}/${entry.name}` : entry.name
                loadDir(newPath)
              } else {
                const filePath = currentPath ? `${currentPath}/${entry.name}` : entry.name
                navigate(`/repos/${repoId}/view/${filePath}`)
              }
            }}
            onContextMenu={e => {
              e.preventDefault()
              e.stopPropagation()
              setContextMenu({ x: e.clientX, y: e.clientY, entry })
            }}
          >
            <span className="shrink-0">{entry.type === 'dir' ? '📁' : '📄'}</span>
            <span className="truncate">{entry.name}</span>
            {entry.type === 'file' && (
              <span className="text-xs text-gray-400 ml-auto shrink-0">{formatSize(entry.size)}</span>
            )}
          </div>
        ))
      )}

      {/* 右键菜单 */}
      {contextMenu && (
        <div
          className="fixed bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50 w-32"
          style={{ left: contextMenu.x, top: contextMenu.y }}
        >
          {!contextMenu.entry ? (
            <>
              <button onClick={handleCreateDir} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文件夹</button>
              <button onClick={handleCreateFile} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文章</button>
            </>
          ) : (
            <>
              <button onClick={() => handleRename(contextMenu.entry!)} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">重命名</button>
              <button onClick={() => handleDelete(contextMenu.entry!)} className="w-full px-3 py-1.5 text-xs text-left text-red-600 hover:bg-gray-50">删除</button>
            </>
          )}
        </div>
      )}
    </div>
  )
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`
}
