import { createContext, useContext, useReducer, useCallback, type ReactNode } from 'react'
import * as repoApi from '../lib/repos'

interface RepoState {
  repos: repoApi.Repo[]
  followedRepos: repoApi.Repo[]
  currentRepoId: string | null
  isLoading: boolean
}

type RepoAction =
  | { type: 'SET_REPOS'; repos: repoApi.Repo[] }
  | { type: 'SET_FOLLOWED_REPOS'; repos: repoApi.Repo[] }
  | { type: 'ADD_REPO'; repo: repoApi.Repo }
  | { type: 'UPDATE_REPO'; repo: repoApi.Repo }
  | { type: 'REMOVE_REPO'; repoId: string }
  | { type: 'SET_CURRENT_REPO'; repoId: string | null }
  | { type: 'SET_LOADING'; loading: boolean }

function repoReducer(state: RepoState, action: RepoAction): RepoState {
  switch (action.type) {
    case 'SET_REPOS':
      return { ...state, repos: action.repos }
    case 'SET_FOLLOWED_REPOS':
      return { ...state, followedRepos: action.repos }
    case 'ADD_REPO':
      return { ...state, repos: [action.repo, ...state.repos] }
    case 'UPDATE_REPO':
      return { ...state, repos: state.repos.map(r => r.id === action.repo.id ? action.repo : r) }
    case 'REMOVE_REPO':
      return { ...state, repos: state.repos.filter(r => r.id !== action.repoId) }
    case 'SET_CURRENT_REPO':
      return { ...state, currentRepoId: action.repoId }
    case 'SET_LOADING':
      return { ...state, isLoading: action.loading }
  }
}

interface RepoContextValue extends RepoState {
  loadRepos: () => Promise<void>
  loadFollowedRepos: () => Promise<void>
  selectRepo: (repoId: string | null) => void
}

const RepoContext = createContext<RepoContextValue | null>(null)

export function RepoProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(repoReducer, {
    repos: [],
    followedRepos: [],
    currentRepoId: null,
    isLoading: false,
  })

  const loadRepos = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', loading: true })
    try {
      const repos = await repoApi.listRepos()
      dispatch({ type: 'SET_REPOS', repos })
    } catch {
      // ignore
    }
    dispatch({ type: 'SET_LOADING', loading: false })
  }, [])

  const loadFollowedRepos = useCallback(async () => {
    try {
      const repos = await repoApi.listFollowedRepos()
      dispatch({ type: 'SET_FOLLOWED_REPOS', repos })
    } catch {
      // ignore
    }
  }, [])

  const selectRepo = useCallback((repoId: string | null) => {
    dispatch({ type: 'SET_CURRENT_REPO', repoId })
  }, [])

  return (
    <RepoContext.Provider value={{ ...state, loadRepos, loadFollowedRepos, selectRepo }}>
      {children}
    </RepoContext.Provider>
  )
}

export function useRepo() {
  const ctx = useContext(RepoContext)
  if (!ctx) throw new Error('useRepo must be used within RepoProvider')
  return ctx
}
