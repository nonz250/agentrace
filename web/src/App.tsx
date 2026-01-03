import { useState, useEffect, createContext, useContext } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { Layout } from '@/components/layout/Layout'
import { WelcomePage } from '@/pages/WelcomePage'
import { RegisterPage } from '@/pages/RegisterPage'
import { LoginPage } from '@/pages/LoginPage'
import { SetupPage } from '@/pages/SetupPage'
import { ProjectsPage } from '@/pages/ProjectsPage'
import { ProjectDetailPage } from '@/pages/ProjectDetailPage'
import { SessionsPage } from '@/pages/SessionsPage'
import { SessionDetailPage } from '@/pages/SessionDetailPage'
import { PlansPage } from '@/pages/PlansPage'
import { PlanDetailPage } from '@/pages/PlanDetailPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { MembersPage } from '@/pages/MembersPage'
import { Spinner } from '@/components/ui/Spinner'
import * as authApi from '@/api/auth'
import type { User } from '@/types/auth'

interface AuthContextType {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  refetch: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | null>(null)

export function useAuthContext() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuthContext must be used within AuthProvider')
  }
  return context
}

function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const queryClient = useQueryClient()

  const fetchUser = async () => {
    try {
      const userData = await authApi.getMe()
      setUser(userData)
      // React Queryのキャッシュにも設定
      queryClient.setQueryData(['me'], userData)
    } catch {
      setUser(null)
      queryClient.setQueryData(['me'], null)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchUser()
  }, [])

  const refetch = async () => {
    setIsLoading(true)
    await fetchUser()
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, isAuthenticated: !!user, refetch }}>
      {children}
    </AuthContext.Provider>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuthContext()

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    )
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuthContext()

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
    <AuthProvider>
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
        {/* Register handles auth internally to support returnTo flow */}
        <Route path="/register" element={<RegisterPage />} />
        {/* Login handles auth internally to support returnTo flow */}
        <Route path="/login" element={<LoginPage />} />

        {/* Setup route - handles auth internally */}
        <Route path="/setup" element={<SetupPage />} />

        {/* Main routes - accessible without auth */}
        <Route path="/" element={<Layout />}>
          <Route index element={<ProjectsPage />} />
          <Route path="projects/:projectId" element={<ProjectDetailPage />} />
          <Route path="projects/:projectId/sessions" element={<SessionsPage />} />
          <Route path="projects/:projectId/plans" element={<PlansPage />} />
          <Route path="sessions" element={<SessionsPage />} />
          <Route path="sessions/:id" element={<SessionDetailPage />} />
          <Route path="plans" element={<PlansPage />} />
          <Route path="plans/:id" element={<PlanDetailPage />} />
          <Route path="members" element={<MembersPage />} />
          {/* Settings requires auth */}
          <Route
            path="settings"
            element={
              <ProtectedRoute>
                <SettingsPage />
              </ProtectedRoute>
            }
          />
        </Route>

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  )
}
