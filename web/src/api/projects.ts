import { fetchAPI } from './client'
import type { Project } from '@/types/project'

interface ProjectListItem extends Project {
  created_at: string
}

interface GetProjectsParams {
  limit?: number
  offset?: number
}

export async function getProjects(params?: GetProjectsParams): Promise<{ projects: ProjectListItem[] }> {
  const searchParams = new URLSearchParams()
  if (params?.limit) searchParams.set('limit', params.limit.toString())
  if (params?.offset) searchParams.set('offset', params.offset.toString())
  const query = searchParams.toString()
  return fetchAPI(`/api/projects${query ? `?${query}` : ''}`)
}

export async function getProject(id: string): Promise<ProjectListItem> {
  return fetchAPI(`/api/projects/${id}`)
}
