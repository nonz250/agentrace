import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useAuthContext } from '@/App'
import * as authApi from '@/api/auth'
import type { LoginParams } from '@/api/auth'

export function useAuth() {
  const { user, isLoading, refetch } = useAuthContext()
  const navigate = useNavigate()

  const loginMutation = useMutation({
    mutationFn: (params: LoginParams) => authApi.login(params),
    onSuccess: async () => {
      await refetch()
      navigate('/')
    },
  })

  const loginWithApiKeyMutation = useMutation({
    mutationFn: (apiKey: string) => authApi.loginWithApiKey(apiKey),
    onSuccess: async () => {
      await refetch()
      navigate('/')
    },
  })

  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSuccess: async () => {
      await refetch()
      navigate('/welcome')
    },
  })

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    login: loginMutation.mutate,
    loginError: loginMutation.error,
    isLoggingIn: loginMutation.isPending,
    loginWithApiKey: loginWithApiKeyMutation.mutate,
    loginWithApiKeyError: loginWithApiKeyMutation.error,
    isLoggingInWithApiKey: loginWithApiKeyMutation.isPending,
    logout: logoutMutation.mutate,
    isLoggingOut: logoutMutation.isPending,
  }
}
