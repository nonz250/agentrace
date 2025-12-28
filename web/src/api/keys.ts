import { fetchAPI } from './client'
import type { ApiKey } from '@/types/auth'

export async function getKeys(): Promise<{ keys: ApiKey[] }> {
  return fetchAPI('/api/keys')
}

export async function createKey(name: string): Promise<{ key: ApiKey; api_key: string }> {
  return fetchAPI('/api/keys', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export async function deleteKey(id: string): Promise<void> {
  return fetchAPI(`/api/keys/${id}`, { method: 'DELETE' })
}
