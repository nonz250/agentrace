import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { CopyButton } from '@/components/ui/CopyButton'
import * as keysApi from '@/api/keys'

export function ApiKeyForm() {
  const [name, setName] = useState('')
  const [newApiKey, setNewApiKey] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: () => keysApi.createKey(name),
    onSuccess: (data) => {
      setNewApiKey(data.api_key)
      setName('')
      queryClient.invalidateQueries({ queryKey: ['keys'] })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (name.trim()) {
      createMutation.mutate()
    }
  }

  const handleCloseModal = () => {
    setNewApiKey(null)
  }

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Name"
          placeholder="e.g. Work Laptop"
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={createMutation.isPending}
          error={createMutation.error?.message}
        />
        <Button
          type="submit"
          loading={createMutation.isPending}
          disabled={!name.trim()}
        >
          Create API Key
        </Button>
      </form>

      <Modal
        open={!!newApiKey}
        onClose={handleCloseModal}
        title="New API Key Created!"
      >
        <div className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium text-gray-700">
              Your API Key
            </label>
            <div className="flex items-center gap-2 rounded-lg border border-gray-300 bg-gray-50 px-3 py-2">
              <code className="flex-1 break-all font-mono text-sm text-gray-900">
                {newApiKey}
              </code>
              <CopyButton text={newApiKey || ''} />
            </div>
          </div>

          <p className="text-sm text-amber-600">
            Save this key - it won't be shown again.
          </p>

          <Button onClick={handleCloseModal} className="w-full">
            Done
          </Button>
        </div>
      </Modal>
    </>
  )
}
