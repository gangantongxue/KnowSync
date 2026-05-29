import { Outlet, Navigate } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'

export default function PublicLayout() {
  const { isAuthenticated } = useAuth()
  if (isAuthenticated) return <Navigate to="/chat" replace />
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <Outlet />
    </div>
  )
}
