import { createContext, useContext, useReducer, useEffect, type ReactNode } from 'react'
import { getToken, setToken, clearToken } from '../lib/client'
import * as authApi from '../lib/auth'

interface AuthState {
  user: authApi.UserInfo | null
  isAuthenticated: boolean
  isLoading: boolean
}

type AuthAction =
  | { type: 'SET_USER'; user: authApi.UserInfo }
  | { type: 'CLEAR_USER' }
  | { type: 'SET_LOADING'; loading: boolean }

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'SET_USER':
      return { ...state, user: action.user, isAuthenticated: true, isLoading: false }
    case 'CLEAR_USER':
      return { ...state, user: null, isAuthenticated: false, isLoading: false }
    case 'SET_LOADING':
      return { ...state, isLoading: action.loading }
  }
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  updateUser: (partial: Partial<authApi.UserInfo>) => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, {
    user: null,
    isAuthenticated: false,
    isLoading: true,
  })

  const refreshUser = async () => {
    try {
      const token = getToken()
      if (!token) {
        dispatch({ type: 'CLEAR_USER' })
        return
      }

      // 尝试用 refresh_token cookie 刷新 access_token
      const response = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        credentials: 'same-origin',
      })

      if (response.ok) {
        const data = await response.json()
        if (data.code === 0 || data.code === 200) {
          // 刷新成功，更新 token 并设置用户信息
          setToken(data.data.access_token)
          dispatch({ type: 'SET_USER', user: data.data.user })
        } else {
          // 刷新失败，清除 token
          clearToken()
          dispatch({ type: 'CLEAR_USER' })
        }
      } else {
        clearToken()
        dispatch({ type: 'CLEAR_USER' })
      }
    } catch {
      clearToken()
      dispatch({ type: 'CLEAR_USER' })
    }
  }

  useEffect(() => {
    const token = getToken()
    if (token) {
      // 有 access_token，尝试用 refresh_token 刷新验证
      refreshUser()
    } else {
      // 没有 access_token，直接清除状态
      dispatch({ type: 'CLEAR_USER' })
    }
  }, [])

  const login = async (email: string, password: string) => {
    const data = await authApi.login(email, password)
    dispatch({ type: 'SET_USER', user: data.user })
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } catch {
      // ignore logout errors
    }
    clearToken()
    dispatch({ type: 'CLEAR_USER' })
  }

  const updateUser = (partial: Partial<authApi.UserInfo>) => {
    if (state.user) {
      dispatch({ type: 'SET_USER', user: { ...state.user, ...partial } })
    }
  }

  return (
    <AuthContext.Provider value={{ ...state, login, logout, refreshUser, updateUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
