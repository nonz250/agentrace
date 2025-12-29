import { useState, useEffect, createContext, useContext } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { Layout } from '@/components/layout/Layout'
import { WelcomePage } from '@/pages/WelcomePage'
import { RegisterPage } from '@/pages/RegisterPage'
import { LoginPage } from '@/pages/LoginPage'
import { SessionListPage } from '@/pages/SessionListPage'
import { SessionDetailPage } from '@/pages/SessionDetailPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { Spinner } from '@/components/ui/Spinner'
import * as authApi from '@/api/auth'
import type { User } from '@/types/auth'

interface AuthContextType {
  user: User | null
  isLoading: boolean
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
    <AuthContext.Provider value={{ user, isLoading, refetch }}>
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
    return <Navigate to="/welcome" replace />
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
    </AuthProvider>
  )
}
