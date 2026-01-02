import { Github, GitBranch, Folder } from 'lucide-react'
import type { Project } from '@/types/project'
import { isDefaultProject } from '@/lib/project-utils'

interface ProjectIconProps {
  project: Project | null
  className?: string
}

function isGitHubUrl(url: string): boolean {
  return url.includes('github.com')
}

export function ProjectIcon({ project, className = 'h-5 w-5' }: ProjectIconProps) {
  const hasProject = !isDefaultProject(project)
  const isGitHub = hasProject && project?.canonical_git_repository && isGitHubUrl(project.canonical_git_repository)

  if (!hasProject) {
    return <Folder className={`${className} text-gray-400`} />
  }

  if (isGitHub) {
    return <Github className={`${className} text-gray-700`} />
  }

  return <GitBranch className={`${className} text-gray-500`} />
}
