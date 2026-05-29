import { request, uploadFile } from './client'

export type NodeType = 'FOLDER' | 'ARTICLE'

export interface Node {
  id: string
  repo_id: string
  parent_id: string
  name: string
  type: NodeType
  file_path: string
  size: number
  created_at: string
  updated_at: string
}

export interface Collaborator {
  repo_id: string
  user_id: string
  role: string
}

export async function listNodes(repoId: string, parentId?: string): Promise<Node[]> {
  const params = parentId ? `?parent_id=${parentId}` : ''
  const res = await request<{ nodes: Node[] }>(`/repos/${repoId}/nodes${params}`)
  return res.data.nodes
}

export async function getNode(repoId: string, nodeId: string): Promise<Node> {
  const res = await request<{ node: Node }>(`/repos/${repoId}/nodes/${nodeId}`)
  return res.data.node
}

export async function createNode(repoId: string, name: string, type: NodeType, parentId?: string): Promise<Node> {
  const res = await request<{ node: Node }>(`/repos/${repoId}/nodes`, {
    method: 'POST',
    body: JSON.stringify({ parent_id: parentId, name, type }),
  })
  return res.data.node
}

export async function updateNode(repoId: string, nodeId: string, data: { name?: string; parent_id?: string }): Promise<Node> {
  const res = await request<{ node: Node }>(`/repos/${repoId}/nodes/${nodeId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  return res.data.node
}

export async function deleteNode(repoId: string, nodeId: string): Promise<{ deleted_node_ids: string[]; deleted_file_paths: string[] }> {
  const res = await request<{ deleted_node_ids: string[]; deleted_file_paths: string[] }>(`/repos/${repoId}/nodes/${nodeId}`, {
    method: 'DELETE',
  })
  return res.data
}

export async function uploadArticleContent(repoId: string, nodeId: string, file: File): Promise<{ node: Node; signed_url: string }> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await uploadFile<{ node: Node; signed_url: string }>(`/repos/${repoId}/nodes/${nodeId}/content`, formData)
  return res.data
}

export async function getArticleSignedUrl(repoId: string, nodeId: string): Promise<string> {
  const res = await request<{ signed_url: string }>(`/repos/${repoId}/nodes/${nodeId}/signed-url`)
  return res.data.signed_url
}

export async function listCollaborators(repoId: string): Promise<Collaborator[]> {
  const res = await request<{ collaborators: Collaborator[] }>(`/repos/${repoId}/collaborators`)
  return res.data.collaborators
}

export async function addCollaborator(repoId: string, userId: string, role: string): Promise<void> {
  await request(`/repos/${repoId}/collaborators`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, role }),
  })
}

export async function updateCollaborator(repoId: string, userId: string, role: string): Promise<void> {
  await request(`/repos/${repoId}/collaborators/${userId}`, {
    method: 'PUT',
    body: JSON.stringify({ role }),
  })
}

export async function removeCollaborator(repoId: string, userId: string): Promise<void> {
  await request(`/repos/${repoId}/collaborators/${userId}`, { method: 'DELETE' })
}
