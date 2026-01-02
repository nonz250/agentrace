import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { ArrowRight, ExternalLink } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { PlanList } from '@/components/plans/PlanList'
import { ProjectIcon } from '@/components/projects/ProjectIcon'
import { Breadcrumb } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import * as projectsApi from '@/api/projects'
import * as sessionsApi from '@/api/sessions'
import * as plansApi from '@/api/plan-documents'
import { parseRepoName, isDefaultProject, getRepoUrl, getProjectDisplayName } from '@/lib/project-utils'

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>()

  const { data: project, isLoading: isProjectLoading, error: projectError } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectsApi.getProject(id!),
    enabled: !!id,
  })

  const { data: sessionsData, isLoading: isSessionsLoading, error: sessionsError } = useQuery({
    queryKey: ['sessions', 'project', id],
    queryFn: () => sessionsApi.getSessions({ projectId: id!, limit: 5 }),
    enabled: !!id,
  })

  const { data: plansData, isLoading: isPlansLoading, error: plansError } = useQuery({
    queryKey: ['plans', 'project', id],
    queryFn: () => plansApi.getPlans({ projectId: id!, limit: 5 }),
    enabled: !!id,
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

  const repoName = parseRepoName(project)
  const repoUrl = getRepoUrl(project)
  const hasProject = !isDefaultProject(project)
  const projectDisplayName = getProjectDisplayName(project) || '(no project)'

  return (
    <div className="space-y-10">
      <Breadcrumb items={[{ label: projectDisplayName }]} />

      {/* Project Header */}
      <div className="border-b border-gray-200 pb-6">
        <div className="flex items-center gap-3">
          <ProjectIcon project={project} className="h-8 w-8" />
          <div>
            <h1 className="text-2xl font-semibold text-gray-900">
              {hasProject ? repoName : '(no project)'}
            </h1>
            {hasProject && repoUrl && (
              <a
                href={repoUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="mt-1 flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800"
              >
                {project.canonical_git_repository}
                <ExternalLink className="h-3 w-3" />
              </a>
            )}
          </div>
        </div>
      </div>

      {/* Recent Sessions */}
      <section>
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Recent Sessions</h2>
          <Link
            to={`/sessions?project_id=${id}`}
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

      {/* Recent Plans */}
      <section>
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Recent Plans</h2>
          <Link
            to={`/plans?project_id=${id}`}
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
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
    </div>
  )
}
