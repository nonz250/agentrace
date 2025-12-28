import { useQuery } from '@tanstack/react-query'
import { SessionList } from '@/components/sessions/SessionList'
import { Spinner } from '@/components/ui/Spinner'
import * as sessionsApi from '@/api/sessions'

export function SessionListPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['sessions'],
    queryFn: sessionsApi.getSessions,
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
      <h1 className="mb-6 text-2xl font-semibold text-gray-900">Sessions</h1>
      <SessionList sessions={data?.sessions || []} />
    </div>
  )
}
