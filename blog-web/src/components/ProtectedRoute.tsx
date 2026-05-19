import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

interface Props {
  children: React.ReactNode
  /** What action requires login — used for the redirect message */
  action?: string
}

/**
 * Route guard component — redirects to /login if user is not authenticated.
 * Preserves the current URL as ?redirect= param so user returns after login.
 */
export default function ProtectedRoute({ children, action }: Props) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  if (!isAuthenticated) {
    const currentPath = window.location.pathname + window.location.search
    const redirectParam = encodeURIComponent(currentPath)
    return <Navigate to={`/login?redirect=${redirectParam}`} replace />
  }

  return <>{children}</>
}
