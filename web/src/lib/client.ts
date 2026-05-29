const BASE_URL = ''

function getToken(): string | null {
  return localStorage.getItem('access_token')
}

function getRefreshToken(): string | null {
  return localStorage.getItem('refresh_token')
}

function setTokens(access: string, refresh: string) {
  localStorage.setItem('access_token', access)
  localStorage.setItem('refresh_token', refresh)
}

function clearTokens() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

async function refreshAccessToken(): Promise<boolean> {
  const refresh = getRefreshToken()
  if (!refresh) return false
  try {
    const res = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh }),
    })
    if (!res.ok) return false
    const data = await res.json()
    if (data.code === 0 || data.code === 200) {
      setTokens(data.data.access_token, data.data.refresh_token)
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
      clearTokens()
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

  const res = await fetch(`${BASE_URL}/api/v1${path}`, {
    method: 'PUT',
    headers,
    body: formData,
  })

  if (res.status === 401) {
    clearTokens()
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Upload failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

export { getToken, setTokens, clearTokens, getRefreshToken }
