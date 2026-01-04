import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react'
import { PlanList } from '@/components/plans/PlanList'
import { CreatePlanModal } from '@/components/plans/CreatePlanModal'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { useAuth } from '@/hooks/useAuth'
import { usePlanStatusFilter } from '@/hooks/usePlanStatusFilter'
import * as plansApi from '@/api/plan-documents'
import * as projectsApi from '@/api/projects'
import { getProjectDisplayName } from '@/lib/project-utils'
import { statusConfig, getFilterButtonClass } from '@/lib/plan-status'

const PAGE_SIZE = 20

export function PlansPage() {
  const { user } = useAuth()
  const { projectId } = useParams<{ projectId: string }>()
  const [page, setPage] = useState(1)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const { selectedStatuses, toggleStatus: baseToggleStatus } = usePlanStatusFilter()
  const offset = (page - 1) * PAGE_SIZE

  const toggleStatus = (status: typeof selectedStatuses[number]) => {
    baseToggleStatus(status)
    setPage(1) // Reset to first page when filter changes
  }

  const { data: projectData } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getProject(projectId!),
    enabled: !!projectId,
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['plans', 'list', page, projectId, selectedStatuses],
    queryFn: () =>
      plansApi.getPlans({
        projectId: projectId || undefined,
        statuses: selectedStatuses.length > 0 ? selectedStatuses : undefined,
        limit: PAGE_SIZE,
        offset,
      }),
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
      <Breadcrumb items={breadcrumbItems} project={projectData} />

      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-gray-900">Plans</h1>
        {user && (
          <Button onClick={() => setShowCreateModal(true)}>
            <Plus className="mr-1 h-4 w-4" />
            Create Plan
          </Button>
        )}
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <span className="text-sm text-gray-500">Filter by status:</span>
        {Object.entries(statusConfig).map(([status, config]) => {
          const isSelected = selectedStatuses.includes(status as keyof typeof statusConfig)
          return (
            <button
              key={status}
              onClick={() => toggleStatus(status as keyof typeof statusConfig)}
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${getFilterButtonClass(status as keyof typeof statusConfig, isSelected)}`}
            >
              {config.label}
            </button>
          )
        })}
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
