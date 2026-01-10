export type UserFavoriteTargetType = 'session' | 'plan'

export interface UserFavorite {
  id: string
  user_id: string
  target_type: UserFavoriteTargetType
  target_id: string
  created_at: string
}
