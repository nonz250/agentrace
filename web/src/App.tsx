import { Routes, Route, Navigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Layout } from '@/components/layout/Layout'
import { WelcomePage } from '@/pages/WelcomePage'
import { RegisterPage } from '@/pages/RegisterPage'
import { LoginPage } from '@/pages/LoginPage'
import { SessionListPage } from '@/pages/SessionListPage'
import { SessionDetailPage } from '@/pages/SessionDetailPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { Spinner } from '@/components/ui/Spinner'
import * as authApi from '@/api/auth'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading, error } = useQuery({
    queryKey: ['me'],
    queryFn: authApi.getMe,
    retry: false,
  })

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    )
  }

  if (error || !user) {
    return <Navigate to="/welcome" replace />
  }

  return <>{children}</>
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: authApi.getMe,
    retry: false,
  })

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    )
  }

  if (user) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route
        path="/welcome"
        element={
          <PublicRoute>
            <WelcomePage />
          </PublicRoute>
        }
      />
      <Route
        path="/register"
        element={
          <PublicRoute>
            <RegisterPage />
          </PublicRoute>
        }
      />
      <Route
        path="/login"
        element={
          <PublicRoute>
            <LoginPage />
          </PublicRoute>
        }
      />

      {/* Protected routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<SessionListPage />} />
        <Route path="sessions/:id" element={<SessionDetailPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>

      {/* Fallback */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
