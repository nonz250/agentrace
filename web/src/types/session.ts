import type { Event } from './event'

export interface Session {
  id: string
  user_id: string | null
  user_name: string | null
  claude_session_id: string
  project_path: string
  started_at: string
  ended_at: string | null
  event_count: number
}

export interface SessionDetail extends Session {
  events: Event[]
}
