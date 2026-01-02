// Use VITE_API_URL if set, otherwise use relative path (same origin)
const BASE_URL = import.meta.env.VITE_API_URL || ''

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export async function fetchAPI<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    mode: 'cors',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!res.ok) {
    const message = await res.text().catch(() => 'Unknown error')
    throw new ApiError(res.status, message)
  }

  // No content
  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}
