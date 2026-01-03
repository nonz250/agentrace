import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Spinner } from '@/components/ui/Spinner'
import * as sessionsApi from '@/api/sessions'

export function SessionRedirectPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: session, isLoading, error } = useQuery({
    queryKey: ['session', id],
    queryFn: () => sessionsApi.getSession(id!),
    enabled: !!id,
  })

  useEffect(() => {
    if (session) {
      const projectId = session.project?.id
      if (projectId) {
        navigate(`/projects/${projectId}/sessions/${id}`, { replace: true })
      } else {
        // Should not happen if default project is always set
        navigate('/', { replace: true })
      }
    }
  }, [session, id, navigate])

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

  return (
    <div className="flex justify-center py-12">
      <Spinner size="lg" />
    </div>
  )
}
