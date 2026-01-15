import { useState, useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react'
import { PlanList } from '@/components/plans/PlanList'
import { CreatePlanModal } from '@/components/plans/CreatePlanModal'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { MultiSelect } from '@/components/ui/MultiSelect'
import { useAuth } from '@/hooks/useAuth'
import { usePlanStatusFilter } from '@/hooks/usePlanStatusFilter'
import { usePlanCollaboratorFilter } from '@/hooks/usePlanCollaboratorFilter'
import { useSortPreference } from '@/hooks/useSortPreference'
import * as plansApi from '@/api/plan-documents'
import * as projectsApi from '@/api/projects'
import { getProjectDisplayName } from '@/lib/project-utils'
import { statusConfig } from '@/lib/plan-status'
import type { Collaborator, PlanDocumentStatus } from '@/types/plan-document'

const PAGE_SIZE = 20

export function PlansPage() {
  const { user } = useAuth()
  const { projectId } = useParams<{ projectId: string }>()
  const [page, setPage] = useState(1)
  const [cursors, setCursors] = useState<string[]>(['']) // cursors[0] = '' for first page
  const [showCreateModal, setShowCreateModal] = useState(false)
  const { selectedStatuses, setStatuses } = usePlanStatusFilter()
  const { selectedCollaboratorIds, setCollaboratorIds } = usePlanCollaboratorFilter()
  const { sort, updateSort } = useSortPreference('plans')

  const resetPagination = useCallback(() => {
    setPage(1)
    setCursors([''])
  }, [])

  const handleStatusChange = (statuses: string[]) => {
    setStatuses(statuses as PlanDocumentStatus[])
    resetPagination()
  }

  const handleCollaboratorChange = (collaboratorIds: string[]) => {
    setCollaboratorIds(collaboratorIds)
    resetPagination()
  }

  const handleSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    updateSort(e.target.value as 'updated_at' | 'created_at')
    resetPagination()
  }

  // Status options for MultiSelect
  const statusOptions = Object.entries(statusConfig).map(([status, config]) => ({
    value: status,
    label: config.label,
    badgeClassName: config.className,
  }))

  const { data: projectData } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getProject(projectId!),
    enabled: !!projectId,
  })

  // Query to get all collaborators (without collaborator filter)
  const { data: allPlansData } = useQuery({
    queryKey: ['plans', 'all-collaborators', projectId, selectedStatuses],
    queryFn: () =>
      plansApi.getPlans({
        projectId: projectId || undefined,
        statuses: selectedStatuses.length > 0 ? selectedStatuses : undefined,
        limit: 100, // Get enough plans to collect collaborators
      }),
  })

  // Collect unique collaborators from all plans
  const allCollaborators = useMemo(() => {
    const collaboratorMap = new Map<string, Collaborator>()
    for (const plan of allPlansData?.plans || []) {
      for (const collaborator of plan.collaborators || []) {
        if (!collaboratorMap.has(collaborator.id)) {
          collaboratorMap.set(collaborator.id, collaborator)
        }
      }
    }
    return Array.from(collaboratorMap.values()).sort((a, b) =>
      a.display_name.localeCompare(b.display_name)
    )
  }, [allPlansData])

  const cursor = cursors[page - 1] || ''

  const { data, isLoading, isFetching, error } = useQuery({
    queryKey: ['plans', 'list', page, projectId, selectedStatuses, selectedCollaboratorIds, sort, cursor],
    queryFn: () =>
      plansApi.getPlans({
        projectId: projectId || undefined,
        statuses: selectedStatuses.length > 0 ? selectedStatuses : undefined,
        collaboratorIds: selectedCollaboratorIds.length > 0 ? selectedCollaboratorIds : undefined,
        limit: PAGE_SIZE,
        cursor: cursor || undefined,
        sort,
      }),
    placeholderData: (previousData) => previousData, // Keep previous data while fetching
  })

  const plans = data?.plans || []
  const nextCursor = data?.next_cursor
  const hasMore = !!nextCursor

  // Store next cursor when we get it
  const goToNextPage = useCallback(() => {
    if (nextCursor) {
      setCursors(prev => {
        const newCursors = [...prev]
        newCursors[page] = nextCursor
        return newCursors
      })
      setPage(p => p + 1)
    }
  }, [nextCursor, page])

  const goToPrevPage = useCallback(() => {
    setPage(p => Math.max(1, p - 1))
  }, [])

  // Only show full-page loading on initial load (no data yet)
  const showInitialLoading = isLoading && !data

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

      <div className="mb-4 flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500">Status:</span>
          <MultiSelect
            options={statusOptions}
            selectedValues={selectedStatuses}
            onChange={handleStatusChange}
            placeholder="All statuses"
          />
        </div>

        {allCollaborators.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500">Collaborator:</span>
            <MultiSelect
              options={allCollaborators.map((c) => ({ value: c.id, label: c.display_name }))}
              selectedValues={selectedCollaboratorIds}
              onChange={handleCollaboratorChange}
              placeholder="All collaborators"
            />
          </div>
        )}

        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500">Sort:</span>
          <select
            value={sort}
            onChange={handleSortChange}
            className="rounded-lg bg-transparent px-2 py-1 text-sm text-gray-600 hover:bg-gray-100 focus:outline-none"
          >
            <option value="updated_at">Updated</option>
            <option value="created_at">Created</option>
          </select>
        </div>
      </div>

      {showInitialLoading ? (
        <div className="flex justify-center py-12">
          <Spinner size="lg" />
        </div>
      ) : error ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
          Failed to load plans: {error.message}
        </div>
      ) : (
        <div className={isFetching ? 'opacity-50 transition-opacity' : ''}>
          <PlanList plans={plans} />
        </div>
      )}

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
            onClick={goToPrevPage}
            disabled={page === 1}
          >
            <ChevronLeft className="mr-1 h-4 w-4" />
            Previous
          </Button>
          <span className="text-sm text-gray-500">Page {page}</span>
          <Button
            variant="secondary"
            size="sm"
            onClick={goToNextPage}
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
