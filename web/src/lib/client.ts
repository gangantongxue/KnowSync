const BASE_URL = ''

let refreshPromise: Promise<boolean> | null = null

function getToken(): string | null {
  return localStorage.getItem('access_token')
}

function setToken(token: string) {
  localStorage.setItem('access_token', token)
}

function clearToken() {
  localStorage.removeItem('access_token')
}

async function refreshAccessToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise

  refreshPromise = doRefresh()
  const result = await refreshPromise

  if (result) {
    // Keep promise for 1 second to batch concurrent requests
    setTimeout(() => { refreshPromise = null }, 1000)
  } else {
    refreshPromise = null
  }

  return result
}

// 刷新 access_token（refresh_token 通过 httpOnly cookie 自动携带）
async function doRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      credentials: 'same-origin', // 携带 cookie
    })
    if (!res.ok) return false
    const data = await res.json()
    if (data.code === 0 || data.code === 200) {
      setToken(data.data.access_token)
      return true
    }
    return false
  } catch {
    return false
  }
}

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export async function request<T>(
  path: string,
  options?: RequestInit & { skipAuth?: boolean }
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }
  if (!options?.skipAuth) {
    const token = getToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
  }

  let res = await fetch(`${BASE_URL}/api/v1${path}`, {
    ...options,
    headers,
  })

  if (res.status === 401 && !options?.skipAuth) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      headers['Authorization'] = `Bearer ${getToken()}`
      res = await fetch(`${BASE_URL}/api/v1${path}`, {
        ...options,
        headers,
      })
    } else {
      clearToken()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

export async function uploadFile<T>(
  path: string,
  formData: FormData
): Promise<ApiResponse<T>> {
  const token = getToken()
  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`

  // Clone FormData before first request (it's consumed after use)
  // Using cast because FormData.clone() isn't in TypeScript's DOM lib types yet
  const formDataClone = (formData as any).clone() as FormData

  let res = await fetch(`${BASE_URL}/api/v1${path}`, {
    method: 'PUT',
    headers,
    body: formData,
  })

  if (res.status === 401) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      headers['Authorization'] = `Bearer ${getToken()}`
      res = await fetch(`${BASE_URL}/api/v1${path}`, {
        method: 'PUT',
        headers,
        body: formDataClone,
      })
    } else {
      clearToken()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Upload failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

export { getToken, setToken, clearToken }
