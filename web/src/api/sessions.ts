import { fetchAPI } from './client'
import type { Session, SessionDetail } from '@/types/session'

interface GetSessionsParams {
  limit?: number
  offset?: number
}

export async function getSessions(params?: GetSessionsParams): Promise<{ sessions: Session[] }> {
  const searchParams = new URLSearchParams()
  if (params?.limit) searchParams.set('limit', params.limit.toString())
  if (params?.offset) searchParams.set('offset', params.offset.toString())
  const query = searchParams.toString()
  return fetchAPI(`/api/sessions${query ? `?${query}` : ''}`)
}

export async function getSession(id: string): Promise<SessionDetail> {
  return fetchAPI(`/api/sessions/${id}`)
}
