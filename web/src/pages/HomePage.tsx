import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { ArrowRight, Terminal, Copy, Check } from 'lucide-react'
import { useState } from 'react'
import { SessionList } from '@/components/sessions/SessionList'
import { PlanList } from '@/components/plans/PlanList'
import { Spinner } from '@/components/ui/Spinner'
import * as sessionsApi from '@/api/sessions'
import * as plansApi from '@/api/plan-documents'

function SetupGuide() {
  const [copied, setCopied] = useState(false)
  const serverUrl = window.location.origin
  const command = `npx agentrace init --url ${serverUrl}`

  const handleCopy = async () => {
    await navigator.clipboard.writeText(command)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
      <div className="mb-6 text-center">
        <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-primary-100">
          <Terminal className="h-8 w-8 text-primary-600" />
        </div>
        <h1 className="text-xl font-semibold text-gray-900">Welcome to Agentrace</h1>
        <p className="mt-2 text-gray-600">
          Get started by connecting Claude Code to this server.
        </p>
      </div>

      <div className="space-y-4">
        <div>
          <p className="mb-2 text-sm font-medium text-gray-700">
            Run this command in your terminal:
          </p>
          <div className="flex items-center gap-2 rounded-lg bg-gray-900 p-3">
            <code className="flex-1 font-mono text-sm text-gray-100">
              {command}
            </code>
            <button
              onClick={handleCopy}
              className="rounded p-1.5 text-gray-400 hover:bg-gray-800 hover:text-gray-200"
              title="Copy to clipboard"
            >
              {copied ? (
                <Check className="h-4 w-4 text-green-400" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>

        <p className="text-center text-sm text-gray-500">
          After setup, your Claude Code sessions will appear here.
        </p>
      </div>
    </div>
  )
}

export function HomePage() {
  const { data: sessionsData, isLoading: isSessionsLoading, error: sessionsError } = useQuery({
    queryKey: ['sessions', 'recent'],
    queryFn: () => sessionsApi.getSessions({ limit: 5 }),
  })

  const { data: plansData, isLoading: isPlansLoading, error: plansError } = useQuery({
    queryKey: ['plans', 'recent'],
    queryFn: () => plansApi.getPlans({ limit: 5 }),
  })

  const isLoading = isSessionsLoading || isPlansLoading
  const hasError = sessionsError || plansError
  const sessionsCount = sessionsData?.sessions?.length ?? 0
  const plansCount = plansData?.plans?.length ?? 0
  const hasNoData = !isLoading && !hasError && sessionsCount === 0 && plansCount === 0

  // Show setup guide when there's no data (and no errors)
  if (hasNoData) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="w-full max-w-md">
          <SetupGuide />
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-10">
      {/* Recent Sessions */}
      <section>
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-2xl font-semibold text-gray-900">Recent Sessions</h1>
          <Link
            to="/sessions"
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
        {isSessionsLoading ? (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        ) : sessionsError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            Failed to load sessions: {sessionsError.message}
          </div>
        ) : (
          <SessionList sessions={sessionsData?.sessions || []} />
        )}
      </section>

      {/* Recent Plans */}
      <section>
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-2xl font-semibold text-gray-900">Recent Plans</h1>
          <Link
            to="/plans"
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
        {isPlansLoading ? (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        ) : plansError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            Failed to load plans: {plansError.message}
          </div>
        ) : (
          <PlanList plans={plansData?.plans || []} />
        )}
      </section>
    </div>
  )
}
