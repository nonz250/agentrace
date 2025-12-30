import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { Spinner } from '@/components/ui/Spinner'
import * as sessionsApi from '@/api/sessions'

export function HomePage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['sessions', 'recent'],
    queryFn: () => sessionsApi.getSessions({ limit: 5 }),
  })

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
        Failed to load sessions: {error.message}
      </div>
    )
  }

  return (
    <div>
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
      <SessionList sessions={data?.sessions || []} />
    </div>
  )
}
