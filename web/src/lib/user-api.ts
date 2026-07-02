import { request, uploadFile } from './client'

export interface UserProfile {
  id: string
  name: string
  email: string
  avatar: string
  bio: string
}

export interface UserProfileData {
  user: UserProfile
  repos: any[]
  friend_status: string
}

export const userApi = {
  getProfile: (userId: string) =>
    request<UserProfileData>(`/users/${userId}/profile`),
  getUserRepos: (userId: string) =>
    request<{ repos: any[] }>(`/users/${userId}/repos`),
  updateProfile: (userId: string, data: { name?: string }) =>
    request(`/users/${userId}`, { method: 'PUT', body: JSON.stringify({ user: data }) }),
  uploadAvatar: (userId: string, file: File) => {
    const formData = new FormData()
    formData.append('avatar', file)
    return uploadFile(`/users/${userId}/avatar`, formData)
  },
  deleteAccount: (userId: string) =>
    request(`/users/${userId}`, { method: 'DELETE' }),
}
