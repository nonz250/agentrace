import { ContentBlockCard } from './ContentBlockCard'
import type { Event } from '@/types/event'

interface TimelineProps {
  events: Event[]
}

// Expanded block for display
export interface DisplayBlock {
  id: string
  eventType: 'user' | 'assistant' | 'tool_use' | 'tool_result'
  blockType: string // 'text', 'thinking', 'tool_use', 'tool_result', etc.
  label: string // Display label like 'User', 'Assistant (Thinking)', 'Assistant (Tool: Edit)'
  timestamp: string
  content: unknown
  originalEvent: Event
}

// Expand events into individual display blocks
function expandEvents(events: Event[]): DisplayBlock[] {
  const blocks: DisplayBlock[] = []

  for (const event of events) {
    const timestamp = (event.payload?.timestamp as string) || event.created_at
    const message = event.payload?.message as Record<string, unknown> | undefined
    const content = message?.content

    if (event.event_type === 'user') {
      // User events: expand content blocks
      if (Array.isArray(content)) {
        content.forEach((block, i) => {
          const blockType = (block as Record<string, unknown>)?.type as string || 'text'
          blocks.push({
            id: `${event.id}-${i}`,
            eventType: 'user',
            blockType,
            label: blockType === 'tool_result' ? 'Tool Result' : 'User',
            timestamp,
            content: block,
            originalEvent: event,
          })
        })
      } else {
        // Simple string content
        blocks.push({
          id: event.id,
          eventType: 'user',
          blockType: 'text',
          label: 'User',
          timestamp,
          content: content,
          originalEvent: event,
        })
      }
    } else if (event.event_type === 'assistant') {
      // Assistant events: expand content blocks
      if (Array.isArray(content)) {
        content.forEach((block, i) => {
          const blockObj = block as Record<string, unknown>
          const blockType = blockObj?.type as string || 'text'
          let label = 'Assistant'

          if (blockType === 'thinking') {
            label = 'Assistant (Thinking)'
          } else if (blockType === 'tool_use') {
            const toolName = blockObj?.name as string || 'Unknown'
            label = `Tool: ${toolName}`
          } else if (blockType === 'tool_result') {
            label = 'Tool Result'
          } else if (blockType === 'text') {
            label = 'Assistant'
          } else {
            label = `Assistant (${blockType})`
          }

          blocks.push({
            id: `${event.id}-${i}`,
            eventType: 'assistant',
            blockType,
            label,
            timestamp,
            content: block,
            originalEvent: event,
          })
        })
      } else if (typeof content === 'string') {
        blocks.push({
          id: event.id,
          eventType: 'assistant',
          blockType: 'text',
          label: 'Assistant',
          timestamp,
          content: content,
          originalEvent: event,
        })
      } else {
        // Fallback for unexpected format
        blocks.push({
          id: event.id,
          eventType: 'assistant',
          blockType: 'unknown',
          label: 'Assistant',
          timestamp,
          content: event.payload,
          originalEvent: event,
        })
      }
    } else if (event.event_type === 'tool_use' || event.event_type === 'tool_result') {
      // Tool events: show as single block
      const toolName = (event.payload as Record<string, unknown>)?.name as string || 'Unknown'
      blocks.push({
        id: event.id,
        eventType: event.event_type as 'tool_use' | 'tool_result',
        blockType: event.event_type,
        label: event.event_type === 'tool_use' ? `Tool: ${toolName}` : 'Tool Result',
        timestamp,
        content: event.payload,
        originalEvent: event,
      })
    }
  }

  return blocks
}

export function Timeline({ events }: TimelineProps) {
  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No events yet.</p>
      </div>
    )
  }

  const displayBlocks = expandEvents(events)

  return (
    <div className="space-y-3">
      {displayBlocks.map((block) => (
        <ContentBlockCard key={block.id} block={block} />
      ))}
    </div>
  )
}
