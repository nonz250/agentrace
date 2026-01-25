import type { Project } from './project'

export interface Collaborator {
  id: string
  display_name: string
}

export type PlanDocumentStatus = 'scratch' | 'draft' | 'planning' | 'pending' | 'ready' | 'implementation' | 'complete'

export interface PlanDocument {
  id: string
  project: Project | null
  description: string
  body: string
  status: PlanDocumentStatus
  collaborators: Collaborator[]
  created_at: string
  updated_at: string
  is_favorited: boolean
  active_comments_count: number
}

export type PlanDocumentEventType = 'body_change' | 'status_change'

export interface PlanDocumentEvent {
  id: string
  plan_document_id: string
  session_id: string | null
  claude_session_id: string | null
  tool_use_id: string | null
  user_id: string | null
  user_name: string | null
  event_type: PlanDocumentEventType
  patch: string
  message: string
  created_at: string
}

// Plan Comment Thread types
export type PlanCommentThreadStatus = 'active' | 'resolved' | 'outdated'

export interface CommentPosition {
  start_offset: number
  end_offset: number
  start_line: number
  start_column: number
  end_line: number
  end_column: number
  found: boolean
}

export interface PlanCommentMessage {
  id: string
  thread_id: string
  user_id: string
  user_name: string
  content: string
  created_at: string
  updated_at: string
}

export interface PlanCommentThread {
  id: string
  plan_document_id: string
  target_text: string
  status: PlanCommentThreadStatus
  position: CommentPosition | null
  messages: PlanCommentMessage[]
  created_at: string
  updated_at: string
}
