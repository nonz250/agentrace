import { cn } from '@/lib/cn'
import { User, Bot, Wrench, ChevronDown, ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { format } from 'date-fns'
import type { Event } from '@/types/event'
import { UserMessage } from './UserMessage'
import { AssistantMessage } from './AssistantMessage'
import { ToolUse } from './ToolUse'

interface EventCardProps {
  event: Event
}

export function EventCard({ event }: EventCardProps) {
  const [expanded, setExpanded] = useState(true)

  // Use payload.timestamp if available, fallback to created_at
  const timestamp = (event.payload?.timestamp as string) || event.created_at

  const iconMap: Record<string, React.ReactNode> = {
    user: <User className="h-4 w-4" />,
    assistant: <Bot className="h-4 w-4" />,
    tool_use: <Wrench className="h-4 w-4" />,
    tool_result: <Wrench className="h-4 w-4" />,
  }

  const labelMap: Record<string, string> = {
    user: 'User',
    assistant: 'Assistant',
    tool_use: `Tool: ${String((event.payload as Record<string, unknown>)?.name || 'Unknown')}`,
    tool_result: 'Tool Result',
  }

  const icon = iconMap[event.event_type] || null
  const label = labelMap[event.event_type] || event.event_type

  return (
    <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
      <button
        className={cn(
          'flex w-full items-center justify-between px-4 py-3',
          'text-left transition-colors hover:bg-gray-50'
        )}
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-2">
          <span
            className={cn(
              'flex h-6 w-6 items-center justify-center rounded-full',
              event.event_type === 'user' && 'bg-blue-100 text-blue-600',
              event.event_type === 'assistant' && 'bg-green-100 text-green-600',
              (event.event_type === 'tool_use' ||
                event.event_type === 'tool_result') &&
                'bg-orange-100 text-orange-600'
            )}
          >
            {icon}
          </span>
          <span className="font-medium text-gray-900">{label}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <span>{format(new Date(timestamp), 'HH:mm:ss')}</span>
          {expanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </div>
      </button>

      {expanded && (
        <div className="border-t border-gray-100 px-4 pb-4">
          {event.event_type === 'user' && (
            <UserMessage payload={event.payload} />
          )}
          {event.event_type === 'assistant' && (
            <AssistantMessage payload={event.payload} />
          )}
          {(event.event_type === 'tool_use' ||
            event.event_type === 'tool_result') && (
            <ToolUse
              payload={event.payload}
              isResult={event.event_type === 'tool_result'}
            />
          )}
        </div>
      )}
    </div>
  )
}
