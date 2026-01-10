import { fetchAPI } from './client'
import type { PlanDocument, PlanDocumentEvent, PlanDocumentStatus } from '@/types/plan-document'

export type SortBy = 'updated_at' | 'created_at'

interface GetPlansParams {
  projectId?: string
  gitRemoteUrl?: string // For backward compatibility
  statuses?: string[]
  collaboratorIds?: string[]
  limit?: number
  offset?: number
  sort?: SortBy
}

export async function getPlans(params?: GetPlansParams): Promise<{ plans: PlanDocument[] }> {
  const searchParams = new URLSearchParams()
  if (params?.projectId) searchParams.set('project_id', params.projectId)
  if (params?.gitRemoteUrl) searchParams.set('git_remote_url', params.gitRemoteUrl)
  if (params?.statuses && params.statuses.length > 0) {
    searchParams.set('status', params.statuses.join(','))
  }
  if (params?.collaboratorIds && params.collaboratorIds.length > 0) {
    searchParams.set('collaborator', params.collaboratorIds.join(','))
  }
  if (params?.limit) searchParams.set('limit', params.limit.toString())
  if (params?.offset) searchParams.set('offset', params.offset.toString())
  if (params?.sort) searchParams.set('sort', params.sort)
  const query = searchParams.toString()
  return fetchAPI(`/api/plans${query ? `?${query}` : ''}`)
}

export async function getPlan(id: string): Promise<PlanDocument> {
  return fetchAPI(`/api/plans/${id}`)
}

export async function getPlanEvents(id: string): Promise<{ events: PlanDocumentEvent[] }> {
  return fetchAPI(`/api/plans/${id}/events`)
}

export async function setPlanStatus(id: string, status: PlanDocumentStatus): Promise<PlanDocument> {
  return fetchAPI(`/api/plans/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}

interface CreatePlanParams {
  description: string
  body: string
  project_id?: string
  status?: string
}

export async function createPlan(params: CreatePlanParams): Promise<PlanDocument> {
  return fetchAPI('/api/plans', {
    method: 'POST',
    body: JSON.stringify(params),
  })
}

interface UpdatePlanParams {
  description?: string
  body?: string
  patch?: string
  project_id?: string
}

export async function updatePlan(id: string, params: UpdatePlanParams): Promise<PlanDocument> {
  return fetchAPI(`/api/plans/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(params),
  })
}

export async function deletePlan(id: string): Promise<void> {
  return fetchAPI(`/api/plans/${id}`, {
    method: 'DELETE',
  })
}
