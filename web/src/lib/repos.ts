import { request } from './client'

export interface Repo {
  id: string
  owner_id: string
  name: string
  visibility: string
  description: string
  article_count: number
  created_at: string
  updated_at: string
}

export async function listRepos(): Promise<Repo[]> {
  const res = await request<{ repos: Repo[] }>('/repos')
  return res.data.repos
}

export async function getRepo(repoId: string): Promise<Repo & { my_role: string }> {
  const res = await request<Repo & { my_role: string }>(`/repos/${repoId}`)
  return res.data
}

export async function createRepo(name: string, description?: string, visibility?: string): Promise<Repo> {
  const res = await request<{ repo: Repo }>('/repos', {
    method: 'POST',
    body: JSON.stringify({ name, description, visibility }),
  })
  return res.data.repo
}

export async function updateRepo(repoId: string, data: { name?: string; description?: string; visibility?: string }): Promise<Repo> {
  const res = await request<{ repo: Repo }>(`/repos/${repoId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  return res.data.repo
}

export async function deleteRepo(repoId: string): Promise<void> {
  await request(`/repos/${repoId}`, { method: 'DELETE' })
}
