import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import * as authApi from '@/api/auth'

export function useAuth() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const { data: user, isLoading, error } = useQuery({
    queryKey: ['me'],
    queryFn: authApi.getMe,
    retry: false,
  })

  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] })
      navigate('/')
    },
  })

  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      queryClient.clear()
      navigate('/welcome')
    },
  })

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    error,
    login: loginMutation.mutate,
    loginError: loginMutation.error,
    isLoggingIn: loginMutation.isPending,
    logout: logoutMutation.mutate,
    isLoggingOut: logoutMutation.isPending,
  }
}
