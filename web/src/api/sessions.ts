import { fetchAPI } from './client'
import type { Session, SessionDetail } from '@/types/session'

export async function getSessions(): Promise<{ sessions: Session[] }> {
  return fetchAPI('/api/sessions')
}

export async function getSession(id: string): Promise<SessionDetail> {
  return fetchAPI(`/api/sessions/${id}`)
}
