import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { PlanList } from '@/components/plans/PlanList'
import { Spinner } from '@/components/ui/Spinner'
import * as sessionsApi from '@/api/sessions'
import * as plansApi from '@/api/plan-documents'

export function HomePage() {
  const { data: sessionsData, isLoading: isSessionsLoading, error: sessionsError } = useQuery({
    queryKey: ['sessions', 'recent'],
    queryFn: () => sessionsApi.getSessions({ limit: 5 }),
  })

  const { data: plansData, isLoading: isPlansLoading, error: plansError } = useQuery({
    queryKey: ['plans', 'recent'],
    queryFn: () => plansApi.getPlans({ limit: 5 }),
  })

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
