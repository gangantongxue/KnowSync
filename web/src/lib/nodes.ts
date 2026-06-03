import { request } from './client'

export interface Collaborator {
  repo_id: string
  user_id: string
  role: string
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
