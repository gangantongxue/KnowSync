import { createContext, useContext, useReducer, useCallback, type ReactNode } from 'react'
import * as repoApi from '../lib/repos'
import * as nodesApi from '../lib/nodes'

interface RepoState {
  repos: repoApi.Repo[]
  currentRepoId: string | null
  nodes: nodesApi.Node[]
  isLoading: boolean
}

type RepoAction =
  | { type: 'SET_REPOS'; repos: repoApi.Repo[] }
  | { type: 'ADD_REPO'; repo: repoApi.Repo }
  | { type: 'UPDATE_REPO'; repo: repoApi.Repo }
  | { type: 'REMOVE_REPO'; repoId: string }
  | { type: 'SET_CURRENT_REPO'; repoId: string | null }
  | { type: 'SET_NODES'; nodes: nodesApi.Node[] }
  | { type: 'ADD_NODE'; node: nodesApi.Node }
  | { type: 'UPDATE_NODE'; node: nodesApi.Node }
  | { type: 'REMOVE_NODE'; nodeId: string }
  | { type: 'SET_LOADING'; loading: boolean }

function repoReducer(state: RepoState, action: RepoAction): RepoState {
  switch (action.type) {
    case 'SET_REPOS':
      return { ...state, repos: action.repos }
    case 'ADD_REPO':
      return { ...state, repos: [action.repo, ...state.repos] }
    case 'UPDATE_REPO':
      return { ...state, repos: state.repos.map(r => r.id === action.repo.id ? action.repo : r) }
    case 'REMOVE_REPO':
      return { ...state, repos: state.repos.filter(r => r.id !== action.repoId) }
    case 'SET_CURRENT_REPO':
      return { ...state, currentRepoId: action.repoId, nodes: [] }
    case 'SET_NODES':
      return { ...state, nodes: action.nodes }
    case 'ADD_NODE':
      return { ...state, nodes: [...state.nodes, action.node] }
    case 'UPDATE_NODE':
      return { ...state, nodes: state.nodes.map(n => n.id === action.node.id ? action.node : n) }
    case 'REMOVE_NODE':
      return { ...state, nodes: state.nodes.filter(n => n.id !== action.nodeId) }
    case 'SET_LOADING':
      return { ...state, isLoading: action.loading }
  }
}

interface RepoContextValue extends RepoState {
  loadRepos: () => Promise<void>
  loadNodes: (repoId: string, parentId?: string) => Promise<void>
  selectRepo: (repoId: string | null) => void
}

const RepoContext = createContext<RepoContextValue | null>(null)

export function RepoProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(repoReducer, {
    repos: [],
    currentRepoId: null,
    nodes: [],
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

  const loadNodes = useCallback(async (repoId: string, parentId?: string) => {
    try {
      const nodes = await nodesApi.listNodes(repoId, parentId)
      dispatch({ type: 'SET_NODES', nodes })
    } catch {
      // ignore
    }
  }, [])

  const selectRepo = useCallback((repoId: string | null) => {
    dispatch({ type: 'SET_CURRENT_REPO', repoId })
  }, [])

  return (
    <RepoContext.Provider value={{ ...state, loadRepos, loadNodes, selectRepo }}>
      {children}
    </RepoContext.Provider>
  )
}

export function useRepo() {
  const ctx = useContext(RepoContext)
  if (!ctx) throw new Error('useRepo must be used within RepoProvider')
  return ctx
}
