import { createContext, useContext, useReducer, useEffect, type ReactNode } from 'react'
import { getToken, clearTokens, getRefreshToken } from '../lib/client'
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
      // 用任意需要认证的请求验证 token，这里先不做具体用户请求
      // 路由守卫中会做具体处理
      dispatch({ type: 'SET_LOADING', loading: false })
    } catch {
      dispatch({ type: 'CLEAR_USER' })
    }
  }

  useEffect(() => {
    const token = getToken()
    if (token) {
      dispatch({ type: 'SET_LOADING', loading: false })
      // 实际用户信息在页面加载时通过具体接口获取
    } else {
      dispatch({ type: 'CLEAR_USER' })
    }
  }, [])

  const login = async (email: string, password: string) => {
    await authApi.login(email, password)
    dispatch({ type: 'SET_LOADING', loading: false })
  }

  const logout = async () => {
    try {
      const refresh = getRefreshToken()
      if (refresh) await authApi.logout(refresh)
    } catch {
      // ignore logout errors
    }
    clearTokens()
    dispatch({ type: 'CLEAR_USER' })
  }

  return (
    <AuthContext.Provider value={{ ...state, login, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
