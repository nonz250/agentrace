import { useState, useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuthContext } from '@/App'
import * as authApi from '@/api/auth'

export function ProfileForm() {
  const { user, refetch } = useAuthContext()
  const [displayName, setDisplayName] = useState('')

  useEffect(() => {
    if (user) {
      setDisplayName(user.display_name || '')
    }
  }, [user])

  const updateMutation = useMutation({
    mutationFn: () => authApi.updateMe({ display_name: displayName }),
    onSuccess: () => {
      refetch()
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateMutation.mutate()
  }

  const hasChanges = user && displayName !== (user.display_name || '')

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        label="Display Name"
        placeholder="Enter your display name"
        value={displayName}
        onChange={(e) => setDisplayName(e.target.value)}
        disabled={updateMutation.isPending}
        error={updateMutation.error?.message}
      />
      <Button
        type="submit"
        loading={updateMutation.isPending}
        disabled={!hasChanges}
      >
        Save
      </Button>
      {updateMutation.isSuccess && (
        <p className="text-sm text-green-600">Profile updated successfully.</p>
      )}
    </form>
  )
}
