import { fetchAPI } from './client'
import type { PlanCommentThread, PlanCommentMessage, PlanCommentThreadStatus } from '@/types/plan-document'

interface PlanThreadsResponse {
  threads: PlanCommentThread[]
}

interface CreateThreadRequest {
  target_text: string
  context_before: string
  context_after: string
  content: string
}

interface CreateMessageRequest {
  content: string
}

interface UpdateMessageRequest {
  content: string
}

// Thread operations

export async function getPlanThreads(planId: string, status?: PlanCommentThreadStatus | 'all'): Promise<PlanThreadsResponse> {
  const params = new URLSearchParams()
  if (status) {
    params.set('status', status)
  }
  const query = params.toString()
  const path = `/api/plans/${planId}/comments${query ? `?${query}` : ''}`
  return fetchAPI<PlanThreadsResponse>(path)
}

export async function createPlanThread(planId: string, data: CreateThreadRequest): Promise<PlanCommentThread> {
  return fetchAPI<PlanCommentThread>(`/api/plans/${planId}/comments`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function deletePlanThread(planId: string, threadId: string): Promise<void> {
  return fetchAPI(`/api/plans/${planId}/comments/${threadId}`, {
    method: 'DELETE',
  })
}

export async function resolvePlanThread(planId: string, threadId: string): Promise<PlanCommentThread> {
  return fetchAPI<PlanCommentThread>(`/api/plans/${planId}/comments/${threadId}/resolve`, {
    method: 'POST',
  })
}

// Message operations

export async function addThreadMessage(planId: string, threadId: string, data: CreateMessageRequest): Promise<PlanCommentMessage> {
  return fetchAPI<PlanCommentMessage>(`/api/plans/${planId}/comments/${threadId}/messages`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateThreadMessage(planId: string, threadId: string, messageId: string, data: UpdateMessageRequest): Promise<PlanCommentMessage> {
  return fetchAPI<PlanCommentMessage>(`/api/plans/${planId}/comments/${threadId}/messages/${messageId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export async function deleteThreadMessage(planId: string, threadId: string, messageId: string): Promise<void> {
  return fetchAPI(`/api/plans/${planId}/comments/${threadId}/messages/${messageId}`, {
    method: 'DELETE',
  })
}
