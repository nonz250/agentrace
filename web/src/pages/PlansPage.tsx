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
import type { PlanDocumentStatus } from '@/types/plan-document'

const PAGE_SIZE = 20

const ALL_STATUSES: PlanDocumentStatus[] = ['scratch', 'draft', 'planning', 'pending', 'implementation', 'complete']
const DEFAULT_SELECTED_STATUSES: PlanDocumentStatus[] = ['scratch', 'draft', 'planning']

const statusConfig: Record<PlanDocumentStatus, { label: string; selectedClass: string; unselectedClass: string }> = {
  scratch: {
    label: 'Scratch',
    selectedClass: 'bg-orange-100 text-orange-700 border-orange-300',
    unselectedClass: 'bg-gray-50 text-gray-400 border-gray-200',
  },
  draft: {
    label: 'Draft',
    selectedClass: 'bg-gray-200 text-gray-700 border-gray-400',
    unselectedClass: 'bg-gray-50 text-gray-400 border-gray-200',
  },
  planning: {
    label: 'Planning',
    selectedClass: 'bg-blue-100 text-blue-700 border-blue-300',
    unselectedClass: 'bg-gray-50 text-gray-400 border-gray-200',
  },
  pending: {
    label: 'Pending',
    selectedClass: 'bg-yellow-100 text-yellow-700 border-yellow-300',
    unselectedClass: 'bg-gray-50 text-gray-400 border-gray-200',
  },
  implementation: {
    label: 'Implementation',
    selectedClass: 'bg-purple-100 text-purple-700 border-purple-300',
    unselectedClass: 'bg-gray-50 text-gray-400 border-gray-200',
  },
  complete: {
    label: 'Complete',
    selectedClass: 'bg-green-100 text-green-700 border-green-300',
    unselectedClass: 'bg-gray-50 text-gray-400 border-gray-200',
  },
}

export function PlansPage() {
  const { user } = useAuth()
  const [searchParams] = useSearchParams()
  const projectId = searchParams.get('project_id')
  const [page, setPage] = useState(1)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedStatuses, setSelectedStatuses] = useState<PlanDocumentStatus[]>(DEFAULT_SELECTED_STATUSES)
  const offset = (page - 1) * PAGE_SIZE

  const toggleStatus = (status: PlanDocumentStatus) => {
    setSelectedStatuses((prev) => {
      if (prev.includes(status)) {
        return prev.filter((s) => s !== status)
      } else {
        return [...prev, status]
      }
    })
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

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <span className="text-sm text-gray-500">Status:</span>
        {ALL_STATUSES.map((status) => {
          const isSelected = selectedStatuses.includes(status)
          const config = statusConfig[status]
          return (
            <button
              key={status}
              onClick={() => toggleStatus(status)}
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
                isSelected ? config.selectedClass : config.unselectedClass
              }`}
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
