import { fetchAPI } from './client'
import type { UserFavorite, UserFavoriteTargetType } from '@/types/user-favorite'

interface GetUserFavoritesParams {
  target_type?: UserFavoriteTargetType
}

export async function getUserFavorites(params?: GetUserFavoritesParams): Promise<{ favorites: UserFavorite[] }> {
  const searchParams = new URLSearchParams()
  if (params?.target_type) searchParams.set('target_type', params.target_type)
  const query = searchParams.toString()
  return fetchAPI(`/api/user-favorites${query ? `?${query}` : ''}`)
}

interface CreateUserFavoriteParams {
  target_type: UserFavoriteTargetType
  target_id: string
}

export async function createUserFavorite(params: CreateUserFavoriteParams): Promise<UserFavorite> {
  return fetchAPI('/api/user-favorites', {
    method: 'POST',
    body: JSON.stringify(params),
  })
}

interface DeleteUserFavoriteParams {
  target_type: UserFavoriteTargetType
  target_id: string
}

export async function deleteUserFavorite(params: DeleteUserFavoriteParams): Promise<void> {
  const searchParams = new URLSearchParams()
  searchParams.set('target_type', params.target_type)
  searchParams.set('target_id', params.target_id)
  return fetchAPI(`/api/user-favorites?${searchParams.toString()}`, {
    method: 'DELETE',
  })
}
