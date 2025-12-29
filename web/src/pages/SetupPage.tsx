import { useState, useEffect } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Terminal, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useAuthContext } from '@/App'
import * as keysApi from '@/api/keys'

type SetupState = 'idle' | 'generating' | 'sending' | 'success' | 'error'

export function SetupPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { user, isLoading } = useAuthContext()

  const [state, setState] = useState<SetupState>('idle')
  const [error, setError] = useState<string | null>(null)

  const token = searchParams.get('token')
  const callback = searchParams.get('callback')

  // Validate required parameters
  const isValidSetup = token && callback

  // Validate callback URL is localhost
  const isValidCallback = (() => {
    if (!callback) return false
    try {
      const url = new URL(callback)
      return url.hostname === 'localhost' || url.hostname === '127.0.0.1'
    } catch {
      return false
    }
  })()

  // If user is not authenticated, redirect to login with returnTo
  useEffect(() => {
    if (!isLoading && !user && isValidSetup) {
      const returnTo = `/setup?${searchParams.toString()}`
      navigate(`/login?returnTo=${encodeURIComponent(returnTo)}`, { replace: true })
    }
  }, [isLoading, user, isValidSetup, searchParams, navigate])

  const handleSetup = async () => {
    if (!token || !callback) return

    try {
      setState('generating')
      setError(null)

      // Generate API key
      const hostname = window.location.hostname || 'CLI'
      const keyName = `CLI Setup - ${hostname}`
      const result = await keysApi.createKey(keyName)

      setState('sending')

      // Send to CLI callback
      const response = await fetch(callback, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          api_key: result.api_key,
          token: token,
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to send API key to CLI')
      }

      setState('success')
    } catch (err) {
      setState('error')
      setError(err instanceof Error ? err.message : 'An error occurred')
    }
  }

  // Show loading while checking auth
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
      </div>
    )
  }

  // If not authenticated, the useEffect will redirect
  if (!user) {
    return null
  }

  // Invalid setup parameters
  if (!isValidSetup) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4">
        <div className="w-full max-w-md">
          <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
            <div className="mb-6 text-center">
              <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
                <XCircle className="h-8 w-8 text-red-600" />
              </div>
              <h1 className="text-xl font-semibold text-gray-900">Invalid Setup Link</h1>
              <p className="mt-2 text-gray-600">
                This setup link is missing required parameters.
              </p>
            </div>
            <Button onClick={() => navigate('/')} className="w-full" size="lg">
              Go to Dashboard
            </Button>
          </div>
        </div>
      </div>
    )
  }

  // Invalid callback URL (not localhost)
  if (!isValidCallback) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4">
        <div className="w-full max-w-md">
          <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
            <div className="mb-6 text-center">
              <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
                <XCircle className="h-8 w-8 text-red-600" />
              </div>
              <h1 className="text-xl font-semibold text-gray-900">Invalid Callback URL</h1>
              <p className="mt-2 text-gray-600">
                For security, the callback URL must be localhost.
              </p>
            </div>
            <Button onClick={() => navigate('/')} className="w-full" size="lg">
              Go to Dashboard
            </Button>
          </div>
        </div>
      </div>
    )
  }

  // Success state
  if (state === 'success') {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4">
        <div className="w-full max-w-md">
          <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
            <div className="mb-6 text-center">
              <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
                <CheckCircle className="h-8 w-8 text-green-600" />
              </div>
              <h1 className="text-xl font-semibold text-gray-900">Setup Complete!</h1>
              <p className="mt-2 text-gray-600">
                Your CLI has been configured successfully.
                <br />
                You can close this tab.
              </p>
            </div>
            <Button onClick={() => navigate('/')} className="w-full" variant="secondary" size="lg">
              Go to Dashboard
            </Button>
          </div>
        </div>
      </div>
    )
  }

  // Main setup form
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md">
        <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
          <div className="mb-6 text-center">
            <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-primary-100">
              <Terminal className="h-8 w-8 text-primary-600" />
            </div>
            <h1 className="text-xl font-semibold text-gray-900">CLI Setup</h1>
            <p className="mt-2 text-gray-600">
              Click the button below to complete the CLI setup.
              <br />
              This will generate an API key for your CLI.
            </p>
          </div>

          {error && (
            <div className="mb-4 rounded-lg bg-red-50 p-4 text-sm text-red-600">
              {error}
              <button
                onClick={handleSetup}
                className="mt-2 block font-medium underline hover:no-underline"
              >
                Try again
              </button>
            </div>
          )}

          <Button
            onClick={handleSetup}
            className="w-full"
            size="lg"
            disabled={state === 'generating' || state === 'sending'}
          >
            {state === 'generating' && (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Generating API Key...
              </>
            )}
            {state === 'sending' && (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Sending to CLI...
              </>
            )}
            {(state === 'idle' || state === 'error') && 'Complete Setup'}
          </Button>

          <p className="mt-4 text-center text-xs text-gray-500">
            Logged in as {user.email}
          </p>
        </div>
      </div>
    </div>
  )
}
