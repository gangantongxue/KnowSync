import { request, setTokens } from './client'

export interface LoginData {
  access_token: string
  refresh_token: string
  user: UserInfo
}

export interface UserInfo {
  id: string
  name: string
  email: string
  avatar: string
}

export async function login(email: string, password: string): Promise<LoginData> {
  const res = await request<LoginData>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
    skipAuth: true,
  })
  setTokens(res.data.access_token, res.data.refresh_token)
  return res.data
}

export async function register(
  name: string,
  email: string,
  password: string,
  verifyCode: string,
  avatar?: File
): Promise<UserInfo> {
  const formData = new FormData()
  formData.append('name', name)
  formData.append('email', email)
  formData.append('password', password)
  formData.append('verify_code', verifyCode)
  if (avatar) formData.append('avatar', avatar)

  const res = await fetch('/api/v1/auth/register', {
    method: 'POST',
    body: formData,
  })
  const data = await res.json()
  if (!res.ok || (data.code !== 0 && data.code !== 200)) {
    throw new Error(data.message || '注册失败')
  }
  return data.data
}

export async function logout(refreshToken: string): Promise<void> {
  await request('/auth/logout', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function sendVerifyCode(email: string): Promise<void> {
  await request('/verify-codes', {
    method: 'POST',
    body: JSON.stringify({ email }),
    skipAuth: true,
  })
}

export async function getUser(userId: string): Promise<UserInfo> {
  const res = await request<UserInfo>(`/users/${userId}`)
  return res.data
}

export async function updateUser(userId: string, data: Partial<{ name: string; email: string }>): Promise<UserInfo> {
  const res = await request<UserInfo>(`/users/${userId}`, {
    method: 'PUT',
    body: JSON.stringify({ user: data }),
  })
  return res.data
}

export async function uploadAvatar(userId: string, file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await fetch(`/api/v1/users/${userId}/avatar`, {
    method: 'PUT',
    headers: { 'Authorization': `Bearer ${localStorage.getItem('access_token')}` },
    body: formData,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.message || '头像上传失败')
  return data.data.avatar
}

export async function changePassword(userId: string, oldPwd: string, newPwd: string): Promise<void> {
  await request(`/users/${userId}/password`, {
    method: 'PUT',
    body: JSON.stringify({ old_password: oldPwd, new_password: newPwd }),
  })
}

export async function forgetPassword(email: string, password: string, verifyCode: string): Promise<void> {
  await request('/password/forget', {
    method: 'POST',
    body: JSON.stringify({ email, password, verify_code: verifyCode }),
    skipAuth: true,
  })
}
