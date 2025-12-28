import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { CopyButton } from '@/components/ui/CopyButton'
import * as authApi from '@/api/auth'

export function RegisterPage() {
  const [name, setName] = useState('')
  const [apiKey, setApiKey] = useState<string | null>(null)
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const registerMutation = useMutation({
    mutationFn: () => authApi.register(name),
    onSuccess: (data) => {
      setApiKey(data.api_key)
      queryClient.invalidateQueries({ queryKey: ['me'] })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (name.trim()) {
      registerMutation.mutate()
    }
  }

  if (apiKey) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4">
        <div className="w-full max-w-md">
          <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
            <div className="mb-6 text-center">
              <div className="mb-2 inline-flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
                <span className="text-2xl text-green-600">&#10003;</span>
              </div>
              <h1 className="text-xl font-semibold text-gray-900">
                Account Created!
              </h1>
            </div>

            <div className="mb-6">
              <label className="mb-2 block text-sm font-medium text-gray-700">
                Your API Key
              </label>
              <div className="flex items-center gap-2 rounded-lg border border-gray-300 bg-gray-50 px-3 py-2">
                <code className="flex-1 break-all font-mono text-sm text-gray-900">
                  {apiKey}
                </code>
                <CopyButton text={apiKey} />
              </div>
              <p className="mt-2 text-sm text-amber-600">
                Save this key - it won't be shown again.
              </p>
            </div>

            <div className="mb-6 border-t border-gray-200 pt-6">
              <p className="mb-2 text-sm font-medium text-gray-700">
                Set up CLI:
              </p>
              <div className="rounded-lg bg-gray-900 px-4 py-3">
                <code className="font-mono text-sm text-gray-100">
                  $ npx agentrace init
                </code>
              </div>
            </div>

            <Button
              onClick={() => navigate('/')}
              className="w-full"
              size="lg"
            >
              Go to Dashboard
            </Button>
          </div>
        </div>
      </div>
    )
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
            Create Account
          </h1>

          <form onSubmit={handleSubmit}>
            <Input
              label="Your Name"
              placeholder="Enter your name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={registerMutation.isPending}
              error={registerMutation.error?.message}
            />

            <Button
              type="submit"
              className="mt-6 w-full"
              size="lg"
              loading={registerMutation.isPending}
              disabled={!name.trim()}
            >
              Create Account
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}
