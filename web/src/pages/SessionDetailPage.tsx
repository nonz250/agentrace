import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, Folder, User, Clock } from 'lucide-react'
import { format } from 'date-fns'
import { Timeline } from '@/components/timeline/Timeline'
import { Spinner } from '@/components/ui/Spinner'
import * as sessionsApi from '@/api/sessions'

export function SessionDetailPage() {
  const { id } = useParams<{ id: string }>()

  const { data: session, isLoading, error } = useQuery({
    queryKey: ['session', id],
    queryFn: () => sessionsApi.getSession(id!),
    enabled: !!id,
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
        Failed to load session: {error.message}
      </div>
    )
  }

  if (!session) {
    return (
      <div className="rounded-xl border border-gray-200 bg-white p-8 text-center">
        <p className="text-gray-500">Session not found.</p>
      </div>
    )
  }

  return (
    <div>
      <Link
        to="/"
        className="mb-6 inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
      >
        <ArrowLeft className="h-4 w-4" />
        Sessions
      </Link>

      <div className="mb-6">
        <div className="flex items-start gap-3">
          <Folder className="mt-1 h-5 w-5 flex-shrink-0 text-gray-400" />
          <div>
            <h1 className="font-mono text-lg text-gray-900">
              {session.project_path}
            </h1>
            <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-500">
              <span className="flex items-center gap-1">
                <User className="h-4 w-4" />
                {session.user_name || 'Unknown'}
              </span>
              <span className="flex items-center gap-1">
                <Clock className="h-4 w-4" />
                Started {format(new Date(session.started_at), 'yyyy-MM-dd HH:mm')}
              </span>
            </div>
          </div>
        </div>
      </div>

      <h2 className="mb-4 text-lg font-semibold text-gray-900">Timeline</h2>
      <Timeline events={session.events || []} />
    </div>
  )
}
