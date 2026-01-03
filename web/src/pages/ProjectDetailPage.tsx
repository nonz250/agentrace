import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { ArrowRight, Plus } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { PlanList } from '@/components/plans/PlanList'
import { CreatePlanModal } from '@/components/plans/CreatePlanModal'
import { Breadcrumb } from '@/components/ui/Breadcrumb'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { useAuth } from '@/hooks/useAuth'
import * as projectsApi from '@/api/projects'
import * as sessionsApi from '@/api/sessions'
import * as plansApi from '@/api/plan-documents'
import { getProjectDisplayName } from '@/lib/project-utils'

export function ProjectDetailPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const { user } = useAuth()
  const [showCreateModal, setShowCreateModal] = useState(false)

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
    queryKey: ['plans', 'project', projectId],
    queryFn: () => plansApi.getPlans({ projectId: projectId!, limit: 5 }),
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
    <div className="space-y-10">
      <Breadcrumb items={[{ label: projectDisplayName }]} project={project} />

      {/* Recent Plans */}
      <section>
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Recent Plans</h2>
          <div className="flex items-center gap-3">
            {user && (
              <Button size="sm" onClick={() => setShowCreateModal(true)}>
                <Plus className="mr-1 h-4 w-4" />
                Create Plan
              </Button>
            )}
            <Link
              to={`/projects/${projectId}/plans`}
              className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
            >
              View all
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
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

      <CreatePlanModal
        open={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        defaultProjectId={projectId}
      />
    </div>
  )
}
