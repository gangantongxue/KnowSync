import { request } from './client'

export interface FileEntry {
  name: string
  type: 'file' | 'dir'
  size: number
}

export interface TreeResponse {
  entries: FileEntry[]
  path: string
}

export async function getRepoTree(repoId: string, dirPath?: string): Promise<TreeResponse> {
  const params = dirPath ? `?path=${encodeURIComponent(dirPath)}` : ''
  const res = await request<TreeResponse>(`/repos/${repoId}/files/tree${params}`)
  return res.data
}

export async function uploadFile(
  repoId: string,
  filePath: string,
  file: File
): Promise<{ path: string; size: number; file_url: string }> {
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

  const data = await res.json()
  return data.data
}

export async function deleteFile(repoId: string, filePath: string): Promise<void> {
  await request(`/repos/${repoId}/files?path=${encodeURIComponent(filePath)}`, { method: 'DELETE' })
}

export async function renameFile(repoId: string, oldPath: string, newPath: string): Promise<{ old_path: string; new_path: string }> {
  const res = await request<{ old_path: string; new_path: string }>(`/repos/${repoId}/files?path=${encodeURIComponent(oldPath)}`, {
    method: 'PUT',
    body: JSON.stringify({ new_path: newPath }),
  })
  return res.data
}

export async function makeDir(repoId: string, dirPath: string): Promise<{ path: string }> {
  const res = await request<{ path: string }>(`/repos/${repoId}/dirs?path=${encodeURIComponent(dirPath)}`, { method: 'POST' })
  return res.data
}

export function getFileUrl(userId: string, repoId: string, filePath: string): string {
  return `/files/${userId}/${repoId}/${filePath}`
}
