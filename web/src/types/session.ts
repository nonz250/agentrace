import type { Event } from './event'
import type { Project } from './project'

export interface Session {
  id: string
  user_id: string | null
  user_name: string | null
  project: Project | null
  claude_session_id: string
  project_path: string
  git_branch: string
  started_at: string
  ended_at: string | null
  event_count: number
}

export interface SessionDetail extends Session {
  events: Event[]
}
