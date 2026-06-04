import { useState, useEffect, useCallback } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { getRepoTree, deleteFile, renameFile, makeDir, uploadFile } from '../../lib/files'
import type { FileEntry } from '../../lib/files'
import type { Repo } from '../../lib/repos'

export const IMAGE_EXTENSIONS = new Set(['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg', '.bmp', '.ico'])
const BINARY_EXTENSIONS = new Set(['.pdf', '.zip', '.tar', '.gz', '.7z', '.rar', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx'])

export function isImageFile(name: string): boolean {
  const ext = name.toLowerCase().slice(name.lastIndexOf('.'))
  return IMAGE_EXTENSIONS.has(ext)
}

function isBinaryFile(name: string): boolean {
  const ext = name.toLowerCase().slice(name.lastIndexOf('.'))
  return BINARY_EXTENSIONS.has(ext)
}

interface RepoTreeProps {
  repoId: string
  repo: Repo | null
  refreshKey?: number
}

interface TreeNode {
  name: string
  type: 'file' | 'dir'
  size: number
  path: string
  depth: number
}

export default function RepoTree({ repoId, repo, refreshKey }: RepoTreeProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const [rootEntries, setRootEntries] = useState<FileEntry[]>([])
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set())
  const [loadedChildren, setLoadedChildren] = useState<Record<string, FileEntry[]>>({})
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; entry?: FileEntry; path?: string } | null>(null)
  const [hoverDir, setHoverDir] = useState<string | null>(null)
  const [dragTarget, setDragTarget] = useState<string | null>(null)

  // 当前打开的文件路径（从 URL 解析）
  const currentFilePath = (() => {
    const m = location.pathname.match(/\/repos\/[^/]+\/(view|edit)\/(.+)/)
    return m ? m[2] : null
  })()

  const loadRoot = useCallback(async () => {
    setLoading(true)
    try {
      const tree = await getRepoTree(repoId, '')
      setRootEntries(tree.entries)
    } catch {
      setRootEntries([])
    }
    setLoading(false)
  }, [repoId])

  useEffect(() => { loadRoot() }, [loadRoot, refreshKey])

  const loadChildren = useCallback(async (dirPath: string) => {
    try {
      const tree = await getRepoTree(repoId, dirPath)
      setLoadedChildren(prev => ({ ...prev, [dirPath]: tree.entries }))
    } catch {
      setLoadedChildren(prev => ({ ...prev, [dirPath]: [] }))
    }
  }, [repoId])

  const toggleDir = async (dirPath: string) => {
    if (expandedDirs.has(dirPath)) {
      setExpandedDirs(prev => { const s = new Set(prev); s.delete(dirPath); return s })
    } else {
      setExpandedDirs(prev => new Set(prev).add(dirPath))
      if (!(dirPath in loadedChildren)) {
        await loadChildren(dirPath)
      }
    }
  }

  const refreshDir = async (dirPath: string) => {
    await loadChildren(dirPath)
    if (!(dirPath in loadedChildren)) return
    // force refresh
    try {
      const tree = await getRepoTree(repoId, dirPath)
      setLoadedChildren(prev => ({ ...prev, [dirPath]: tree.entries }))
    } catch {}
  }

  // --- 构建树 ---
  const buildTree = (entries: FileEntry[], depth: number, parentPath: string = ''): TreeNode[] => {
    return [...entries]
      .sort((a, b) => {
        if (a.type !== b.type) return a.type === 'dir' ? -1 : 1
        return a.name.localeCompare(b.name)
      })
      .flatMap(e => {
        const fullPath = parentPath ? `${parentPath}/${e.name}` : e.name
        const node: TreeNode = { name: e.name, type: e.type, size: e.size, path: fullPath, depth }
        if (e.type === 'dir' && expandedDirs.has(fullPath) && loadedChildren[fullPath]) {
          return [node, ...buildTree(loadedChildren[fullPath], depth + 1, fullPath)]
        }
        return [node]
      })
  }

  const tree = buildTree(rootEntries, 0)

  // --- 操作 ---
  const handleCreateFile = async (targetPath: string) => {
    const name = prompt('请输入文件名称（含 .md 后缀）:')
    if (!name?.trim()) return
    const filePath = targetPath ? `${targetPath}/${name.trim()}` : name.trim()
    navigate(`/repos/${repoId}/edit/${filePath}`)
    setContextMenu(null)
  }

  const handleCreateDir = async (targetPath: string) => {
    const name = prompt('请输入文件夹名称:')
    if (!name?.trim()) return
    const dirPath = targetPath ? `${targetPath}/${name.trim()}` : name.trim()
    try {
      await makeDir(repoId, dirPath)
      await refreshDir(targetPath || '')
      if (targetPath || targetPath === '') loadRoot()
      else refreshDir(targetPath)
    } catch (err: any) {
      alert('创建失败: ' + err.message)
    }
    setContextMenu(null)
  }

  const handleDelete = async (entryPath: string) => {
    const name = entryPath.split('/').pop() || entryPath
    if (!confirm(`确定删除「${name}」？`)) return
    try {
      await deleteFile(repoId, entryPath)
      // 刷新父目录
      const parts = entryPath.split('/')
      parts.pop()
      const parentPath = parts.join('/')
      if (!parentPath) await loadRoot()
      else await refreshDir(parentPath)
    } catch (err: any) {
      alert('删除失败: ' + err.message)
    }
    setContextMenu(null)
  }

  const handleRename = async (entryPath: string) => {
    const name = entryPath.split('/').pop() || entryPath
    const newName = prompt('新名称:', name)
    if (!newName?.trim() || newName === name) return
    const parts = entryPath.split('/')
    parts.pop()
    const parentPath = parts.join('/')
    const newPath = parentPath ? `${parentPath}/${newName.trim()}` : newName.trim()
    try {
      await renameFile(repoId, entryPath, newPath)
      if (!parentPath) await loadRoot()
      else await refreshDir(parentPath)
    } catch (err: any) {
      alert('重命名失败: ' + err.message)
    }
    setContextMenu(null)
  }

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (!files?.length) return
    await uploadFiles(files, '')
    e.target.value = ''
  }

  const uploadFiles = async (files: FileList, targetPath: string) => {
    setUploading(true)
    let errorCount = 0
    for (const file of Array.from(files)) {
      const filePath = targetPath ? `${targetPath}/${file.name}` : file.name
      try {
        await uploadFile(repoId, filePath, file)
      } catch {
        errorCount++
      }
    }
    if (errorCount > 0) {
      alert(`${errorCount} 个文件上传失败`)
    }
    setUploading(false)
    if (!targetPath) await loadRoot()
    else await refreshDir(targetPath)
  }

  // --- 拖拽 ---
  const handleDragOver = (e: React.DragEvent, targetPath: string) => {
    e.preventDefault()
    e.stopPropagation()
    e.dataTransfer.dropEffect = 'copy'
    setDragTarget(targetPath)
  }

  const handleDragLeave = () => setDragTarget(null)

  const handleDrop = async (e: React.DragEvent, targetPath: string) => {
    e.preventDefault()
    e.stopPropagation()
    setDragTarget(null)
    await uploadFiles(e.dataTransfer.files, targetPath)
  }

  // --- 右键 ---
  const handleContextMenu = (e: React.MouseEvent, entry?: FileEntry, entryPath?: string) => {
    e.preventDefault()
    e.stopPropagation()
    setContextMenu({ x: e.clientX, y: e.clientY, entry, path: entryPath })
  }

  return (
    <div className="h-full flex flex-col" onClick={() => setContextMenu(null)}>
      {/* 根目录操作按钮 */}
      <div className="flex items-center gap-1 p-2 border-b border-gray-100 shrink-0">
        <span className="text-xs text-gray-400 font-medium mr-auto">根目录</span>
        <button
          onClick={() => handleCreateDir('')}
          className="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-100 text-gray-400 hover:text-emerald-600"
          title="新建文件夹"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
        </button>
        <button
          onClick={() => handleCreateFile('')}
          className="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-100 text-gray-400 hover:text-emerald-600"
          title="新建文章"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
        </button>
        <label className="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-100 text-gray-400 hover:text-emerald-600 cursor-pointer" title="上传文件">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          <input type="file" onChange={handleFileSelect} className="hidden" multiple />
        </label>
      </div>

      {/* 文件列表（可拖拽投放区） */}
      <div
        className={`flex-1 overflow-y-auto p-1 transition-colors ${uploading ? 'opacity-50' : ''}`}
        onDragOver={(e) => handleDragOver(e, '')}
        onDragLeave={handleDragLeave}
        onDrop={(e) => handleDrop(e, '')}
        onContextMenu={(e) => handleContextMenu(e)}
      >
        {loading ? (
          <div className="text-xs text-gray-400 px-2 py-4 text-center">加载中...</div>
        ) : tree.length === 0 ? (
          <div className={`text-xs px-2 py-8 text-center rounded-lg border-2 border-dashed m-1 transition-colors ${dragTarget === '' ? 'border-emerald-300 text-emerald-500' : 'border-gray-200 text-gray-400'}`}>
            拖拽文件到此处上传
          </div>
        ) : (
          tree.map(node => (
            <TreeNodeRow
              key={node.path + node.depth}
              node={node}
              currentFilePath={currentFilePath}
              hoverDir={hoverDir}
              dragTarget={dragTarget}
              expandedDirs={expandedDirs}
              onToggle={toggleDir}
              onNavigate={(filePath) => {
                if (isBinaryFile(filePath) && repo?.owner_id) {
                  window.open(`/files/${repo.owner_id}/${repoId}/${filePath}`, '_blank')
                } else {
                  navigate(`/repos/${repoId}/view/${filePath}`)
                }
              }}
              onContextMenu={handleContextMenu}
              onHoverDir={setHoverDir}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onCreateFile={handleCreateFile}
              onCreateDir={handleCreateDir}
              onFileSelect={async (e, targetPath) => {
                const files = e.target.files
                if (files?.length) await uploadFiles(files, targetPath)
              }}
            />
          ))
        )}
      </div>

      {/* 右键菜单 */}
      {contextMenu && (
        <div
          className="fixed bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50 w-32"
          style={{ left: contextMenu.x, top: contextMenu.y }}
        >
          {!contextMenu.entry || contextMenu.entry.type === 'dir' ? (
            <>
              <button onClick={() => handleCreateDir(contextMenu.path || '')} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文件夹</button>
              <button onClick={() => handleCreateFile(contextMenu.path || '')} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文章</button>
            </>
          ) : null}
          {contextMenu.entry && (
            <>
              {contextMenu.entry.type !== 'dir' && <div className="border-t border-gray-100" />}
              <button onClick={() => handleRename(contextMenu.path!)} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">重命名</button>
              <button onClick={() => handleDelete(contextMenu.path!)} className="w-full px-3 py-1.5 text-xs text-left text-red-600 hover:bg-gray-50">删除</button>
            </>
          )}
        </div>
      )}
    </div>
  )
}

// --- TreeNodeRow ---
function TreeNodeRow({
  node,
  currentFilePath,
  hoverDir,
  dragTarget,
  expandedDirs,
  onToggle,
  onNavigate,
  onContextMenu,
  onHoverDir,
  onDragOver,
  onDragLeave,
  onDrop,
  onCreateFile,
  onCreateDir,
  onFileSelect,
}: {
  node: TreeNode
  currentFilePath: string | null
  hoverDir: string | null
  dragTarget: string | null
  expandedDirs: Set<string>
  onToggle: (path: string) => void
  onNavigate: (path: string) => void
  onContextMenu: (e: React.MouseEvent, entry?: FileEntry, path?: string) => void
  onHoverDir: (path: string | null) => void
  onDragOver: (e: React.DragEvent, path: string) => void
  onDragLeave: () => void
  onDrop: (e: React.DragEvent, path: string) => void
  onCreateFile: (path: string) => void
  onCreateDir: (path: string) => void
  onFileSelect: (e: React.ChangeEvent<HTMLInputElement>, targetPath: string) => void
}) {
  const isSelected = currentFilePath === node.path
  const isDir = node.type === 'dir'
  const isExpanded = isDir && expandedDirs.has(node.path)
  const isDragHover = dragTarget === node.path

  return (
    <div
      className={`flex items-center gap-1 px-1 py-1 rounded text-sm text-gray-700 cursor-pointer group transition-colors mb-0.5 ${isSelected ? 'bg-emerald-100 text-emerald-800' : 'hover:bg-gray-100'} ${isDragHover ? 'ring-2 ring-emerald-300 bg-emerald-50' : ''}`}
      style={{ paddingLeft: `${node.depth * 16 + 4}px` }}
      onClick={() => {
        if (isDir) {
          onToggle(node.path)
        } else {
          onNavigate(node.path)
        }
      }}
      onContextMenu={(e) => onContextMenu(e, { name: node.name, type: node.type, size: node.size }, node.path)}
      onMouseEnter={() => isDir && onHoverDir(node.path)}
      onMouseLeave={() => isDir && onHoverDir(null)}
      {...(isDir ? {
        onDragOver: (e: React.DragEvent) => onDragOver(e, node.path),
        onDragLeave: onDragLeave,
        onDrop: (e: React.DragEvent) => onDrop(e, node.path),
      } : {})}
    >
      {/* 展开/折叠箭头 */}
      {isDir ? (
        <span className="w-4 h-4 flex items-center justify-center text-gray-400 shrink-0">
          {isExpanded ? (
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="6 9 12 15 18 9"/></svg>
          ) : (
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
          )}
        </span>
      ) : (
        <span className="w-4 shrink-0" />
      )}

      {/* 图标 */}
      <span className="shrink-0">{isDir ? (isExpanded ? '📂' : '📁') : '📄'}</span>

      {/* 名称 */}
      <span className="truncate text-xs">{node.name}</span>

      {/* 右侧：文件大小 / 文件夹悬停操作 */}
      {isDir ? (
        <div className={`ml-auto flex items-center gap-0.5 shrink-0 ${hoverDir === node.path ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'} transition-opacity`}>
          <button
            onClick={(e) => { e.stopPropagation(); onCreateDir(node.path) }}
            className="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-200 text-gray-400 hover:text-emerald-600"
            title="新建文件夹"
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
          </button>
          <button
            onClick={(e) => { e.stopPropagation(); onCreateFile(node.path) }}
            className="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-200 text-gray-400 hover:text-emerald-600"
            title="新建文章"
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
          </button>
          <label
            className="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-200 text-gray-400 hover:text-emerald-600 cursor-pointer"
            title="上传"
            onClick={(e) => e.stopPropagation()}
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            <input type="file" onChange={(e) => onFileSelect(e, node.path)} className="hidden" multiple />
          </label>
        </div>
      ) : (
        <span className="text-xs text-gray-400 ml-auto shrink-0">{formatSize(node.size)}</span>
      )}
    </div>
  )
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`
}
