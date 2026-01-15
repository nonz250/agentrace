import { fetchAPI } from './client'
import type { Project } from '@/types/project'

interface ProjectListItem extends Project {
  created_at: string
}

interface GetProjectsParams {
  limit?: number
  cursor?: string
}

interface GetProjectsResponse {
  projects: ProjectListItem[]
  next_cursor?: string
}

export async function getProjects(params?: GetProjectsParams): Promise<GetProjectsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.limit) searchParams.set('limit', params.limit.toString())
  if (params?.cursor) searchParams.set('cursor', params.cursor)
  const query = searchParams.toString()
  return fetchAPI(`/api/projects${query ? `?${query}` : ''}`)
}

export async function getProject(id: string): Promise<ProjectListItem> {
  return fetchAPI(`/api/projects/${id}`)
}
