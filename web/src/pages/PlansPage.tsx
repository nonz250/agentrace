import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams, Link } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Plus, X } from 'lucide-react'
import { PlanList } from '@/components/plans/PlanList'
import { CreatePlanModal } from '@/components/plans/CreatePlanModal'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { useAuth } from '@/hooks/useAuth'
import * as plansApi from '@/api/plan-documents'
import * as projectsApi from '@/api/projects'
import { getProjectDisplayName } from '@/lib/project-utils'

const PAGE_SIZE = 20

export function PlansPage() {
  const { user } = useAuth()
  const [searchParams] = useSearchParams()
  const projectId = searchParams.get('project_id')
  const [page, setPage] = useState(1)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const offset = (page - 1) * PAGE_SIZE

  const { data: projectData } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getProject(projectId!),
    enabled: !!projectId,
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['plans', 'list', page, projectId],
    queryFn: () => plansApi.getPlans({ projectId: projectId || undefined, limit: PAGE_SIZE, offset }),
  })

  const plans = data?.plans || []
  const hasMore = plans.length === PAGE_SIZE

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
        Failed to load plans: {error.message}
      </div>
    )
  }

  const projectDisplayName = projectData ? getProjectDisplayName(projectData) : null

  // Build breadcrumb items
  const breadcrumbItems: BreadcrumbItem[] = []
  if (projectId && projectDisplayName) {
    breadcrumbItems.push({ label: projectDisplayName, href: `/projects/${projectId}` })
  }
  breadcrumbItems.push({ label: 'Plans' })

  return (
    <div>
      <Breadcrumb items={breadcrumbItems} />

      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-semibold text-gray-900">
            {projectId ? 'Plans' : 'All Plans'}
          </h1>
          {projectId && projectDisplayName && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">
                Filtered by: <span className="font-medium">{projectDisplayName}</span>
              </span>
              <Link
                to="/plans"
                className="flex items-center gap-1 rounded-full bg-gray-100 px-2 py-1 text-xs text-gray-600 hover:bg-gray-200"
              >
                <X className="h-3 w-3" />
                Clear
              </Link>
            </div>
          )}
        </div>
        {user && (
          <Button onClick={() => setShowCreateModal(true)}>
            <Plus className="mr-1 h-4 w-4" />
            Create Plan
          </Button>
        )}
      </div>
      <PlanList plans={plans} />

      <CreatePlanModal
        open={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        defaultProjectId={projectId}
      />

      {(page > 1 || hasMore) && (
        <div className="mt-6 flex items-center justify-between">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="mr-1 h-4 w-4" />
            Previous
          </Button>
          <span className="text-sm text-gray-500">Page {page}</span>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => p + 1)}
            disabled={!hasMore}
          >
            Next
            <ChevronRight className="ml-1 h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}
