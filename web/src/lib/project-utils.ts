import type { Project } from '@/types/project'

// Parse canonical git repository URL to get repo name (e.g., "owner/repo")
export function parseRepoName(project: Project | null): string | null {
  if (!project || !project.canonical_git_repository) return null

  const url = project.canonical_git_repository

  // Handle HTTPS format: https://github.com/owner/repo
  const match = url.match(/https?:\/\/[^/]+\/(.+?)(?:\.git)?$/)
  if (match) return match[1]

  return null
}

// Get the display name for a project
export function getProjectDisplayName(project: Project | null): string | null {
  if (!project) return null
  if (!project.canonical_git_repository) return '(no project)'

  const repoName = parseRepoName(project)
  return repoName || project.canonical_git_repository
}

// Check if this is the default "no project" project
export function isDefaultProject(project: Project | null): boolean {
  if (!project) return true
  return project.canonical_git_repository === ''
}

// Get repository URL for linking
export function getRepoUrl(project: Project | null): string | null {
  if (!project || !project.canonical_git_repository) return null

  // The canonical URL is already in HTTPS format
  const url = project.canonical_git_repository

  // Check if it looks like a valid HTTPS URL
  if (url.startsWith('https://')) {
    return url
  }

  return null
}
