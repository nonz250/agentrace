import { cn } from '@/lib/cn'
import { format } from 'date-fns'
import { User } from 'lucide-react'

export interface UserBlockInfo {
  id: string
  preview: string
  timestamp: string
}

interface UserMessageNavProps {
  userBlocks: UserBlockInfo[]
  activeBlockId: string | null
  onNavigate: (blockId: string) => void
}

export function UserMessageNav({ userBlocks, activeBlockId, onNavigate }: UserMessageNavProps) {
  if (userBlocks.length === 0) {
    return null
  }

  return (
    <nav className="sticky top-20 space-y-1">
      <div className="mb-2 flex items-center gap-1.5 text-xs font-medium text-gray-500">
        <User className="h-3 w-3" />
        <span>User Messages ({userBlocks.length})</span>
      </div>
      <div className="space-y-1">
        {userBlocks.map((block) => (
          <button
            key={block.id}
            onClick={() => onNavigate(block.id)}
            className={cn(
              'w-full rounded-lg px-2 py-1.5 text-left text-xs transition-colors',
              'hover:bg-blue-50',
              activeBlockId === block.id
                ? 'bg-blue-100 text-blue-700'
                : 'bg-gray-50 text-gray-600'
            )}
          >
            <div className="mb-0.5 text-[10px] text-gray-400">
              {format(new Date(block.timestamp), 'HH:mm:ss')}
            </div>
            <div className="line-clamp-2 break-words">
              {block.preview || 'User message'}
            </div>
          </button>
        ))}
      </div>
    </nav>
  )
}
