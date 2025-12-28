import { fetchAPI } from './client'
import type { User } from '@/types/auth'

export async function register(name: string): Promise<{ user: User; api_key: string }> {
  return fetchAPI('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export async function login(apiKey: string): Promise<{ user: User }> {
  return fetchAPI('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ api_key: apiKey }),
  })
}

export async function logout(): Promise<void> {
  return fetchAPI('/api/auth/logout', { method: 'POST' })
}

export async function getMe(): Promise<User> {
  return fetchAPI('/api/me')
}

export async function getUsers(): Promise<{ users: User[] }> {
  return fetchAPI('/api/users')
}
