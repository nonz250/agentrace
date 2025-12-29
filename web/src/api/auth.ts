import { fetchAPI } from './client'
import type { User } from '@/types/auth'

export interface RegisterParams {
  email: string
  password: string
}

export interface LoginParams {
  email: string
  password: string
}

export async function register(params: RegisterParams): Promise<{ user: User; api_key: string }> {
  return fetchAPI('/auth/register', {
    method: 'POST',
    body: JSON.stringify(params),
  })
}

export async function login(params: LoginParams): Promise<{ user: User }> {
  return fetchAPI('/auth/login', {
    method: 'POST',
    body: JSON.stringify(params),
  })
}

export async function loginWithApiKey(apiKey: string): Promise<{ user: User }> {
  return fetchAPI('/auth/login/apikey', {
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
