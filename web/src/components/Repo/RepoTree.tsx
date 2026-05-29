import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Node, NodeType } from '../../lib/nodes'
import { createNode, updateNode, deleteNode } from '../../lib/nodes'

interface RepoTreeProps {
  repoId: string
  nodes: Node[]
  currentParentId: string | null
  onNavigate: (parentId: string | null) => void
  onRefresh: () => void
}

export default function RepoTree({ repoId, nodes, currentParentId, onNavigate, onRefresh }: RepoTreeProps) {
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; node?: Node } | null>(null)
  const navigate = useNavigate()

  const handleContextMenu = (e: React.MouseEvent, node?: Node) => {
    e.preventDefault()
    setContextMenu({ x: e.clientX, y: e.clientY, node })
  }

  const handleCreate = async (type: NodeType) => {
    const name = prompt(`请输入${type === 'FOLDER' ? '文件夹' : '文章'}名称`)
    if (!name?.trim()) return
    const n = await createNode(repoId, name.trim(), type, currentParentId || undefined)
    if (type === 'ARTICLE') {
      navigate(`/repos/${repoId}/nodes/${n.id}/edit`)
    }
    onRefresh()
    setContextMenu(null)
  }

  const handleRename = async (node: Node) => {
    const name = prompt('新名称:', node.name)
    if (name?.trim() && name !== node.name) {
      await updateNode(repoId, node.id, { name: name.trim() })
      onRefresh()
    }
    setContextMenu(null)
  }

  const handleMove = async (node: Node) => {
    const newParentId = prompt('目标文件夹 ID:')
    if (newParentId) {
      await updateNode(repoId, node.id, { parent_id: newParentId })
      onRefresh()
    }
    setContextMenu(null)
  }

  const handleDelete = async (node: Node) => {
    if (confirm(`确定删除「${node.name}」？`)) {
      await deleteNode(repoId, node.id)
      onRefresh()
    }
    setContextMenu(null)
  }

  // 获取当前层级下的子节点
  const childNodes = nodes.filter(n => n.parent_id === (currentParentId || ''))

  // 开始拖拽
  const handleDragStart = (e: React.DragEvent, node: Node) => {
    e.dataTransfer.setData('text/plain', node.id)
    e.dataTransfer.effectAllowed = 'move'
  }

  // 拖拽到文件夹上
  const handleDrop = async (e: React.DragEvent, targetNode: Node) => {
    e.preventDefault()
    if (targetNode.type !== 'FOLDER') return
    const nodeId = e.dataTransfer.getData('text/plain')
    if (nodeId === targetNode.id) return
    try {
      await updateNode(repoId, nodeId, { parent_id: targetNode.id })
      onRefresh()
    } catch {
      // ignore
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }

  // 点击空白区域创建
  const handleCanvasCreate = async () => {
    const type = confirm('创建文件夹点确定，创建文章点取消') ? 'FOLDER' : 'ARTICLE'
    const name = prompt(`请输入${type === 'FOLDER' ? '文件夹' : '文章'}名称`)
    if (!name?.trim()) return
    const n = await createNode(repoId, name.trim(), type as NodeType, currentParentId || undefined)
    if (type === 'ARTICLE') {
      navigate(`/repos/${repoId}/nodes/${n.id}/edit`)
    }
    onRefresh()
  }

  return (
    <div onContextMenu={e => handleContextMenu(e)} onClick={() => setContextMenu(null)}>
      {/* 面包屑导航 */}
      <div className="flex items-center gap-1 text-xs text-gray-500 mb-2 px-1">
        <button onClick={() => onNavigate(null)} className="hover:text-emerald-600">根目录</button>
        {currentParentId && <span>/</span>}
      </div>

      {/* 新建按钮 */}
      <button onClick={handleCanvasCreate} className="w-full px-2 py-1 text-xs text-gray-400 hover:text-emerald-600 hover:bg-gray-100 rounded text-left mb-1">
        + 新建
      </button>

      {/* 节点列表 */}
      {childNodes.map(node => (
        <NodeItem
          key={node.id}
          node={node}
          repoId={repoId}
          onContextMenu={handleContextMenu}
          onNavigate={onNavigate}
          onDragStart={handleDragStart}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
        />
      ))}

      {childNodes.length === 0 && (
        <div className="text-xs text-gray-400 px-2 py-4 text-center">
          暂无内容，点击上方"+ 新建"创建
        </div>
      )}

      {/* 右键菜单 */}
      {contextMenu && (
        <div
          className="fixed bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50 w-36"
          style={{ left: contextMenu.x, top: contextMenu.y }}
        >
          {!contextMenu.node && (
            <>
              <button onClick={() => handleCreate('FOLDER')} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文件夹</button>
              <button onClick={() => handleCreate('ARTICLE')} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文章</button>
            </>
          )}
          {contextMenu.node && (
            <>
              <button onClick={() => handleRename(contextMenu.node!)} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">重命名</button>
              <button onClick={() => handleMove(contextMenu.node!)} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">移动到</button>
              <button onClick={() => handleDelete(contextMenu.node!)} className="w-full px-3 py-1.5 text-xs text-left text-red-600 hover:bg-gray-50">删除</button>
            </>
          )}
        </div>
      )}
    </div>
  )
}

interface NodeItemProps {
  node: Node
  repoId: string
  onContextMenu: (e: React.MouseEvent, node: Node) => void
  onNavigate: (parentId: string | null) => void
  onDragStart: (e: React.DragEvent, node: Node) => void
  onDrop: (e: React.DragEvent, targetNode: Node) => void
  onDragOver: (e: React.DragEvent) => void
}

function NodeItem({ node, repoId, onContextMenu, onNavigate, onDragStart, onDrop, onDragOver }: NodeItemProps) {
  const navigate = useNavigate()

  return (
    <div
      draggable
      onContextMenu={e => onContextMenu(e, node)}
      onDragStart={e => onDragStart(e, node)}
      onDragOver={onDragOver}
      onDrop={e => onDrop(e, node)}
      onClick={() => {
        if (node.type === 'FOLDER') {
          onNavigate(node.id)
        } else {
          navigate(`/repos/${repoId}/nodes/${node.id}`)
        }
      }}
      className="flex items-center gap-2 px-2 py-1.5 rounded text-sm cursor-pointer hover:bg-gray-100 text-gray-700 mb-0.5"
    >
      <span>{node.type === 'FOLDER' ? '📁' : '📄'}</span>
      <span className="truncate">{node.name}</span>
    </div>
  )
}
