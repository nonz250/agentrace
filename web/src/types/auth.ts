export interface User {
  id: string
  email: string
  display_name: string
  created_at: string
}

export interface ApiKey {
  id: string
  name: string
  key_prefix: string
  last_used_at: string | null
  created_at: string
}
