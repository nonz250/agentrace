import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Spinner } from '@/components/ui/Spinner'
import * as plansApi from '@/api/plan-documents'

export function PlanRedirectPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: plan, isLoading, error } = useQuery({
    queryKey: ['plan', id],
    queryFn: () => plansApi.getPlan(id!),
    enabled: !!id,
  })

  useEffect(() => {
    if (plan) {
      const projectId = plan.project?.id
      if (projectId) {
        navigate(`/projects/${projectId}/plans/${id}`, { replace: true })
      } else {
        // Should not happen if default project is always set
        navigate('/', { replace: true })
      }
    }
  }, [plan, id, navigate])

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
        Failed to load plan: {error.message}
      </div>
    )
  }

  return (
    <div className="flex justify-center py-12">
      <Spinner size="lg" />
    </div>
  )
}
