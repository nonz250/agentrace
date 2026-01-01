import type { Project } from './project'

export interface Collaborator {
  id: string
  display_name: string
}

export type PlanDocumentStatus = 'draft' | 'planning' | 'pending' | 'implementation' | 'complete'

export interface PlanDocument {
  id: string
  project: Project | null
  description: string
  body: string
  status: PlanDocumentStatus
  collaborators: Collaborator[]
  created_at: string
  updated_at: string
}

export interface PlanDocumentEvent {
  id: string
  plan_document_id: string
  session_id: string | null
  user_id: string | null
  user_name: string | null
  patch: string
  created_at: string
}
