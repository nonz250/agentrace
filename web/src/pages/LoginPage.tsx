import { useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuth } from '@/hooks/useAuth'

export function LoginPage() {
  const [apiKey, setApiKey] = useState('')
  const { login, loginError, isLoggingIn } = useAuth()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (apiKey.trim()) {
      login(apiKey)
    }
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

          <form onSubmit={handleSubmit}>
            <Input
              label="API Key"
              placeholder="agtr_..."
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              disabled={isLoggingIn}
              error={loginError?.message}
            />

            <Button
              type="submit"
              className="mt-6 w-full"
              size="lg"
              loading={isLoggingIn}
              disabled={!apiKey.trim()}
            >
              Login
            </Button>
          </form>

          <p className="mt-6 text-center text-sm text-gray-600">
            Don't have an account?{' '}
            <Link to="/register" className="text-primary-600 hover:underline">
              Register
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
