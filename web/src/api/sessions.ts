import { fetchAPI } from './client'
import type { Session, SessionDetail } from '@/types/session'

export type SortBy = 'updated_at' | 'created_at'

interface GetSessionsParams {
  projectId?: string
  limit?: number
  offset?: number
  sort?: SortBy
}

export async function getSessions(params?: GetSessionsParams): Promise<{ sessions: Session[] }> {
  const searchParams = new URLSearchParams()
  if (params?.projectId) searchParams.set('project_id', params.projectId)
  if (params?.limit) searchParams.set('limit', params.limit.toString())
  if (params?.offset) searchParams.set('offset', params.offset.toString())
  if (params?.sort) searchParams.set('sort', params.sort)
  const query = searchParams.toString()
  return fetchAPI(`/api/sessions${query ? `?${query}` : ''}`)
}

export async function getSession(id: string): Promise<SessionDetail> {
  return fetchAPI(`/api/sessions/${id}`)
}

export async function updateSessionTitle(id: string, title: string): Promise<Session> {
  return fetchAPI(`/api/sessions/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ title }),
  })
}

interface UpdateSessionParams {
  title?: string
  project_id?: string
}

export async function updateSession(id: string, params: UpdateSessionParams): Promise<Session> {
  return fetchAPI(`/api/sessions/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(params),
  })
}
