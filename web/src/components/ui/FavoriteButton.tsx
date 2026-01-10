import { Star } from 'lucide-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { cn } from '@/lib/cn'
import { createUserFavorite, deleteUserFavorite } from '@/api/user-favorites'
import type { UserFavoriteTargetType } from '@/types/user-favorite'

interface FavoriteButtonProps {
  targetType: UserFavoriteTargetType
  targetId: string
  isFavorited: boolean
  className?: string
  size?: 'sm' | 'md'
}

export function FavoriteButton({
  targetType,
  targetId,
  isFavorited,
  className,
  size = 'md',
}: FavoriteButtonProps) {
  const queryClient = useQueryClient()

  const addFavorite = useMutation({
    mutationFn: () => createUserFavorite({ target_type: targetType, target_id: targetId }),
    onSuccess: () => {
      // Invalidate relevant queries
      if (targetType === 'session') {
        queryClient.invalidateQueries({ queryKey: ['sessions'] })
        queryClient.invalidateQueries({ queryKey: ['session', targetId] })
      } else {
        queryClient.invalidateQueries({ queryKey: ['plans'] })
        queryClient.invalidateQueries({ queryKey: ['plan', targetId] })
      }
    },
  })

  const removeFavorite = useMutation({
    mutationFn: () => deleteUserFavorite({ target_type: targetType, target_id: targetId }),
    onSuccess: () => {
      // Invalidate relevant queries
      if (targetType === 'session') {
        queryClient.invalidateQueries({ queryKey: ['sessions'] })
        queryClient.invalidateQueries({ queryKey: ['session', targetId] })
      } else {
        queryClient.invalidateQueries({ queryKey: ['plans'] })
        queryClient.invalidateQueries({ queryKey: ['plan', targetId] })
      }
    },
  })

  const isLoading = addFavorite.isPending || removeFavorite.isPending

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (isLoading) return

    if (isFavorited) {
      removeFavorite.mutate()
    } else {
      addFavorite.mutate()
    }
  }

  const iconSize = size === 'sm' ? 'h-4 w-4' : 'h-5 w-5'

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={isLoading}
      className={cn(
        'inline-flex items-center justify-center rounded-md transition-colors',
        'hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-1 focus:ring-yellow-400',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        size === 'sm' ? 'p-1' : 'p-1.5',
        className
      )}
      aria-label={isFavorited ? 'Remove from favorites' : 'Add to favorites'}
    >
      <Star
        className={cn(
          iconSize,
          'transition-colors',
          isFavorited
            ? 'fill-yellow-400 text-yellow-400'
            : 'text-gray-400 hover:text-yellow-400'
        )}
      />
    </button>
  )
}
