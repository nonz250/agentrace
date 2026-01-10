import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { ArrowRight, Plus } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { PlanList } from '@/components/plans/PlanList'
import { CreatePlanModal } from '@/components/plans/CreatePlanModal'
import { Breadcrumb } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { MultiSelect } from '@/components/ui/MultiSelect'
import { useAuth } from '@/hooks/useAuth'
import { usePlanStatusFilter } from '@/hooks/usePlanStatusFilter'
import { usePlanCollaboratorFilter } from '@/hooks/usePlanCollaboratorFilter'
import { useSortPreference } from '@/hooks/useSortPreference'
import { statusConfig } from '@/lib/plan-status'
import * as projectsApi from '@/api/projects'
import * as sessionsApi from '@/api/sessions'
import * as plansApi from '@/api/plan-documents'
import { getProjectDisplayName } from '@/lib/project-utils'
import type { Collaborator, PlanDocumentStatus } from '@/types/plan-document'

export function ProjectDetailPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const { user } = useAuth()
  const [showCreateModal, setShowCreateModal] = useState(false)
  const { selectedStatuses, setStatuses } = usePlanStatusFilter()
  const { selectedCollaboratorIds, setCollaboratorIds } = usePlanCollaboratorFilter()
  const { sort: plansSort, updateSort: updatePlansSort } = useSortPreference('plans')
  const { sort: sessionsSort, updateSort: updateSessionsSort } = useSortPreference('sessions')

  const handleStatusChange = (statuses: string[]) => {
    setStatuses(statuses as PlanDocumentStatus[])
  }

  const handleCollaboratorChange = (collaboratorIds: string[]) => {
    setCollaboratorIds(collaboratorIds)
  }

  const handlePlansSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    updatePlansSort(e.target.value as 'updated_at' | 'created_at')
  }

  const handleSessionsSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    updateSessionsSort(e.target.value as 'updated_at' | 'created_at')
  }

  // Status options for MultiSelect
  const statusOptions = Object.entries(statusConfig).map(([status, config]) => ({
    value: status,
    label: config.label,
    badgeClassName: config.className,
  }))

  const { data: project, isLoading: isProjectLoading, error: projectError } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getProject(projectId!),
    enabled: !!projectId,
  })

  const { data: sessionsData, isLoading: isSessionsLoading, error: sessionsError } = useQuery({
    queryKey: ['sessions', 'project', projectId, sessionsSort],
    queryFn: () => sessionsApi.getSessions({ projectId: projectId!, limit: 5, sort: sessionsSort }),
    enabled: !!projectId,
  })

  // Query to get all collaborators (without collaborator filter)
  const { data: allPlansData } = useQuery({
    queryKey: ['plans', 'all-collaborators', projectId, selectedStatuses],
    queryFn: () =>
      plansApi.getPlans({
        projectId: projectId!,
        statuses: selectedStatuses.length > 0 ? selectedStatuses : undefined,
        limit: 100,
      }),
    enabled: !!projectId,
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

  const { data: plansData, isLoading: isPlansLoading, isFetching: isPlansFetching, error: plansError } = useQuery({
    queryKey: ['plans', 'project', projectId, selectedStatuses, selectedCollaboratorIds, plansSort],
    queryFn: () =>
      plansApi.getPlans({
        projectId: projectId!,
        statuses: selectedStatuses.length > 0 ? selectedStatuses : undefined,
        collaboratorIds: selectedCollaboratorIds.length > 0 ? selectedCollaboratorIds : undefined,
        limit: 5,
        sort: plansSort,
      }),
    enabled: !!projectId,
    placeholderData: (previousData) => previousData,
  })

  const showInitialPlansLoading = isPlansLoading && !plansData

  if (isProjectLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  if (projectError) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
        Failed to load project: {projectError.message}
      </div>
    )
  }

  if (!project) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
        Project not found
      </div>
    )
  }

  const projectDisplayName = getProjectDisplayName(project) || '(no project)'

  return (
    <div>
      <Breadcrumb items={[{ label: projectDisplayName }]} project={project} />

      {/* Recent Plans */}
      <section className="mb-10">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Recent Plans</h2>
          {user && (
            <Button size="sm" onClick={() => setShowCreateModal(true)}>
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
              value={plansSort}
              onChange={handlePlansSortChange}
              className="rounded-lg bg-transparent px-2 py-1 text-sm text-gray-600 hover:bg-gray-100 focus:outline-none"
            >
              <option value="updated_at">Updated</option>
              <option value="created_at">Created</option>
            </select>
          </div>
        </div>
        {showInitialPlansLoading ? (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        ) : plansError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            Failed to load plans: {plansError.message}
          </div>
        ) : (
          <div className={isPlansFetching ? 'opacity-50 transition-opacity' : ''}>
            <PlanList plans={plansData?.plans || []} />
          </div>
        )}
        <div className="mt-4 text-right">
          <Link
            to={`/projects/${projectId}/plans`}
            className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </section>

      <CreatePlanModal
        open={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        defaultProjectId={projectId}
      />

      {/* Recent Sessions */}
      <section>
        <div className="mb-4">
          <h2 className="text-xl font-semibold text-gray-900">Recent Sessions</h2>
        </div>
        <div className="mb-4 flex flex-wrap items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500">Sort:</span>
            <select
              value={sessionsSort}
              onChange={handleSessionsSortChange}
              className="rounded-lg bg-transparent px-2 py-1 text-sm text-gray-600 hover:bg-gray-100 focus:outline-none"
            >
              <option value="updated_at">Updated</option>
              <option value="created_at">Created</option>
            </select>
          </div>
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
        <div className="mt-4 text-right">
          <Link
            to={`/projects/${projectId}/sessions`}
            className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </section>
    </div>
  )
}
