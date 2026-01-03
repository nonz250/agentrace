import { cn } from '@/lib/cn'
import { format } from 'date-fns'
import { Bot, MessageSquare, User } from 'lucide-react'
import type { MessageBlockInfo } from './Timeline'

interface MessageNavProps {
  messageBlocks: MessageBlockInfo[]
  activeBlockId: string | null
  onNavigate: (blockId: string) => void
}

export function MessageNav({ messageBlocks, activeBlockId, onNavigate }: MessageNavProps) {
  if (messageBlocks.length === 0) {
    return null
  }

  // Count user and assistant messages
  const userCount = messageBlocks.filter(b => b.role === 'user').length
  const assistantCount = messageBlocks.filter(b => b.role === 'assistant').length

  return (
    <nav className="sticky top-20 flex max-h-[calc(100vh-6rem)] flex-col">
      <div className="mb-2 flex flex-shrink-0 items-center gap-2 text-xs font-medium text-gray-500">
        <MessageSquare className="h-3 w-3" />
        <span>Messages</span>
        <span className="flex items-center gap-0.5">
          <User className="h-3 w-3" />
          {userCount}
        </span>
        <span className="flex items-center gap-0.5">
          <Bot className="h-3 w-3" />
          {assistantCount}
        </span>
      </div>
      <div className="space-y-1 overflow-y-auto">
        {messageBlocks.map((block) => {
          const isUser = block.role === 'user'
          return (
            <button
              key={block.id}
              onClick={() => onNavigate(block.id)}
              className={cn(
                'w-full rounded-lg px-2 py-1.5 text-left text-xs transition-colors',
                activeBlockId === block.id
                  ? isUser
                    ? 'bg-blue-100 text-blue-700'
                    : 'bg-emerald-100 text-emerald-700'
                  : isUser
                    ? 'bg-gray-50 text-gray-600 hover:bg-blue-50'
                    : 'bg-gray-50 text-gray-600 hover:bg-emerald-50'
              )}
            >
              <div className="mb-0.5 flex items-center gap-1 text-[10px] text-gray-400">
                {isUser ? (
                  <User className="h-2.5 w-2.5" />
                ) : (
                  <Bot className="h-2.5 w-2.5" />
                )}
                {format(new Date(block.timestamp), 'HH:mm:ss')}
              </div>
              <div className="line-clamp-2 break-words">
                {block.preview || (isUser ? 'User message' : 'Assistant message')}
              </div>
            </button>
          )
        })}
      </div>
    </nav>
  )
}
