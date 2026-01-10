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
}

export type PlanDocumentEventType = 'body_change' | 'status_change'

export interface PlanDocumentEvent {
  id: string
  plan_document_id: string
  session_id: string | null
  user_id: string | null
  user_name: string | null
  event_type: PlanDocumentEventType
  patch: string
  created_at: string
}
