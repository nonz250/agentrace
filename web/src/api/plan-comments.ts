import { fetchAPI } from './client'
import type { PlanComment, PlanCommentStatus } from '@/types/plan-document'

interface PlanCommentsResponse {
  comments: PlanComment[]
}

interface CreatePlanCommentRequest {
  target_text: string
  context_before: string
  context_after: string
  content: string
}

interface UpdatePlanCommentRequest {
  content: string
}

export async function getPlanComments(planId: string, status?: PlanCommentStatus | 'all'): Promise<PlanCommentsResponse> {
  const params = new URLSearchParams()
  if (status) {
    params.set('status', status)
  }
  const query = params.toString()
  const path = `/api/plans/${planId}/comments${query ? `?${query}` : ''}`
  return fetchAPI<PlanCommentsResponse>(path)
}

export async function createPlanComment(planId: string, data: CreatePlanCommentRequest): Promise<PlanComment> {
  return fetchAPI<PlanComment>(`/api/plans/${planId}/comments`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updatePlanComment(planId: string, commentId: string, data: UpdatePlanCommentRequest): Promise<PlanComment> {
  return fetchAPI<PlanComment>(`/api/plans/${planId}/comments/${commentId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export async function deletePlanComment(planId: string, commentId: string): Promise<void> {
  return fetchAPI(`/api/plans/${planId}/comments/${commentId}`, {
    method: 'DELETE',
  })
}

export async function resolvePlanComment(planId: string, commentId: string): Promise<PlanComment> {
  return fetchAPI<PlanComment>(`/api/plans/${planId}/comments/${commentId}/resolve`, {
    method: 'POST',
  })
}
