export interface Event {
  id: string
  session_id: string
  event_type: 'user' | 'assistant' | 'tool_use' | 'tool_result' | string
  payload: Record<string, unknown>
  created_at: string
}
