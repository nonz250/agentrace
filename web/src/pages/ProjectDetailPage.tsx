import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { ArrowRight } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { PlanList } from '@/components/plans/PlanList'
import { Breadcrumb } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { usePlanStatusFilter } from '@/hooks/usePlanStatusFilter'
import { ALL_STATUSES, statusConfig, getFilterButtonClass } from '@/lib/plan-status'
import * as projectsApi from '@/api/projects'
import * as sessionsApi from '@/api/sessions'
import * as plansApi from '@/api/plan-documents'
import { getProjectDisplayName } from '@/lib/project-utils'

export function ProjectDetailPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const { selectedStatuses, toggleStatus } = usePlanStatusFilter()

  const { data: project, isLoading: isProjectLoading, error: projectError } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getProject(projectId!),
    enabled: !!projectId,
  })

  const { data: sessionsData, isLoading: isSessionsLoading, error: sessionsError } = useQuery({
    queryKey: ['sessions', 'project', projectId],
    queryFn: () => sessionsApi.getSessions({ projectId: projectId!, limit: 5 }),
    enabled: !!projectId,
  })

  const { data: plansData, isLoading: isPlansLoading, error: plansError } = useQuery({
    queryKey: ['plans', 'project', projectId, selectedStatuses],
    queryFn: () =>
      plansApi.getPlans({
        projectId: projectId!,
        statuses: selectedStatuses.length > 0 ? selectedStatuses : undefined,
        limit: 5,
      }),
    enabled: !!projectId,
  })

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
          <Link
            to={`/projects/${projectId}/plans`}
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
        <div className="mb-4 flex flex-wrap items-center gap-2">
          <span className="text-sm text-gray-500">Filter by status:</span>
          {ALL_STATUSES.map((status) => {
            const isSelected = selectedStatuses.includes(status)
            return (
              <button
                key={status}
                onClick={() => toggleStatus(status)}
                className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${getFilterButtonClass(status, isSelected)}`}
              >
                {statusConfig[status].label}
              </button>
            )
          })}
        </div>
        {isPlansLoading ? (
          <div className="flex justify-center py-12">
            <Spinner size="lg" />
          </div>
        ) : plansError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            Failed to load plans: {plansError.message}
          </div>
        ) : (
          <PlanList plans={plansData?.plans || []} />
        )}
      </section>

      {/* Recent Sessions */}
      <section>
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Recent Sessions</h2>
          <Link
            to={`/projects/${projectId}/sessions`}
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
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
      </section>
    </div>
  )
}
