import { useState, useEffect } from 'react'
import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuth } from '@/hooks/useAuth'
import { useAuthContext } from '@/App'
import * as authApi from '@/api/auth'

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [githubEnabled, setGithubEnabled] = useState(false)
  const { login, loginError, isLoggingIn } = useAuth()
  const { user, isLoading } = useAuthContext()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const returnTo = searchParams.get('returnTo')

  // If already logged in and no returnTo, redirect to dashboard
  useEffect(() => {
    if (!isLoading && user && !returnTo) {
      navigate('/', { replace: true })
    }
  }, [isLoading, user, returnTo, navigate])

  useEffect(() => {
    authApi.getAuthConfig().then((config) => {
      setGithubEnabled(config.github_enabled)
    }).catch(() => {
      // Ignore errors, just don't show GitHub button
    })
  }, [])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (email.trim() && password.trim()) {
      login({ email, password })
    }
  }

  const handleGitHubLogin = () => {
    const githubUrl = new URL('/auth/github', window.location.origin)
    if (returnTo) {
      githubUrl.searchParams.set('returnTo', returnTo)
    }
    window.location.href = githubUrl.toString()
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md">
        <Link
          to="/welcome"
          className="mb-6 inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </Link>

        <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
          <h1 className="mb-6 text-center text-xl font-semibold text-gray-900">
            Login
          </h1>

          {githubEnabled && (
            <>
              <Button
                type="button"
                variant="secondary"
                className="w-full"
                size="lg"
                onClick={handleGitHubLogin}
              >
                <svg className="mr-2 h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                </svg>
                Continue with GitHub
              </Button>

              <div className="my-6 flex items-center">
                <div className="flex-1 border-t border-gray-200" />
                <span className="px-4 text-sm text-gray-500">or</span>
                <div className="flex-1 border-t border-gray-200" />
              </div>
            </>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Email"
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoggingIn}
            />

            <Input
              label="Password"
              type="password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoggingIn}
            />

            {loginError && (
              <p className="text-sm text-red-600">{loginError.message}</p>
            )}

            <Button
              type="submit"
              className="w-full"
              size="lg"
              loading={isLoggingIn}
              disabled={!email.trim() || !password.trim()}
            >
              Login
            </Button>
          </form>

          <p className="mt-6 text-center text-sm text-gray-600">
            Don't have an account?{' '}
            <Link
              to={returnTo ? `/register?returnTo=${encodeURIComponent(returnTo)}` : '/register'}
              className="text-primary-600 hover:underline"
            >
              Register
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
