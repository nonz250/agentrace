import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Laptop, Trash2 } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import type { ApiKey } from '@/types/auth'
import * as keysApi from '@/api/keys'

interface ApiKeyListProps {
  keys: ApiKey[]
}

export function ApiKeyList({ keys }: ApiKeyListProps) {
  const queryClient = useQueryClient()

  const deleteMutation = useMutation({
    mutationFn: keysApi.deleteKey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keys'] })
    },
  })

  if (keys.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-6 text-center">
        <p className="text-gray-500">No API keys.</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {keys.map((key) => (
        <Card key={key.id}>
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-start gap-3">
              <Laptop className="mt-0.5 h-5 w-5 flex-shrink-0 text-gray-400" />
              <div>
                <p className="font-medium text-gray-900">{key.name}</p>
                <p className="mt-1 text-sm text-gray-500">
                  <code className="font-mono">{key.key_prefix}...</code>
                  {' · '}
                  {key.last_used_at
                    ? `Last used ${formatDistanceToNow(new Date(key.last_used_at), { addSuffix: true })}`
                    : 'Never used'}
                </p>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                if (confirm('Are you sure you want to delete this API key?')) {
                  deleteMutation.mutate(key.id)
                }
              }}
              disabled={deleteMutation.isPending}
            >
              <Trash2 className="h-4 w-4 text-gray-400" />
            </Button>
          </div>
        </Card>
      ))}
    </div>
  )
}
