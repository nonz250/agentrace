import { fetchAPI } from './client'
import type { PlanDocument, PlanDocumentEvent, PlanDocumentStatus } from '@/types/plan-document'

interface GetPlansParams {
  gitRemoteUrl?: string
  limit?: number
  offset?: number
}

export async function getPlans(params?: GetPlansParams): Promise<{ plans: PlanDocument[] }> {
  const searchParams = new URLSearchParams()
  if (params?.gitRemoteUrl) searchParams.set('git_remote_url', params.gitRemoteUrl)
  if (params?.limit) searchParams.set('limit', params.limit.toString())
  if (params?.offset) searchParams.set('offset', params.offset.toString())
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
