export interface Collaborator {
  id: string
  display_name: string
}

export interface PlanDocument {
  id: string
  description: string
  body: string
  git_remote_url: string
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
