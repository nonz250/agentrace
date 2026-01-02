import { Link } from 'react-router-dom'
import type { Project } from '@/types/project'
import { parseRepoName, isDefaultProject, getRepoUrl } from '@/lib/project-utils'
import { ProjectIcon } from './ProjectIcon'

interface ProjectCardProps {
  project: Project & { created_at: string }
}

export function ProjectCard({ project }: ProjectCardProps) {
  const repoName = parseRepoName(project)
  const repoUrl = getRepoUrl(project)
  const hasProject = !isDefaultProject(project)

  return (
    <Link
      to={`/projects/${project.id}`}
      className="block rounded-xl border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md"
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 text-gray-900">
            <ProjectIcon project={project} className="h-5 w-5 flex-shrink-0" />
            {hasProject ? (
              <span className="font-medium truncate">{repoName}</span>
            ) : (
              <span className="font-medium text-gray-500">(no project)</span>
            )}
          </div>
          {hasProject && repoUrl && (
            <p className="mt-1 text-sm text-gray-500 truncate">
              {project.canonical_git_repository}
            </p>
          )}
        </div>
      </div>
    </Link>
  )
}
