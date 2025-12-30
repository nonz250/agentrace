import { ContentBlockCard } from './ContentBlockCard'
import type { Event } from '@/types/event'

interface TimelineProps {
  events: Event[]
  projectPath?: string
}

// Label structure for display blocks
export interface BlockLabel {
  text: string       // Main label text like 'User', 'Thinking', 'Tool: Edit'
  params?: string    // Optional parameter to display (e.g., file path, command)
}

// Expanded block for display
export interface DisplayBlock {
  id: string
  eventType: 'user' | 'assistant' | 'tool_use' | 'tool_result'
  blockType: string // 'text', 'thinking', 'tool_use', 'tool_result', 'tool_group', 'local_command', 'local_command_output', 'local_command_group', etc.
  label: BlockLabel
  timestamp: string
  content: unknown
  originalEvent: Event
  // For grouped local commands and tool groups
  childBlocks?: DisplayBlock[]
  // For tool_group: the result block
  toolResultBlock?: DisplayBlock
}

// Extract command name from local command content
function extractCommandName(content: string): string | null {
  const match = content.match(/<command-name>\/([\w-]+)<\/command-name>/)
  return match ? match[1] : null
}

// Tool params extractor configuration
// Each tool can define how to extract display params from its input
type ToolParamsExtractor = (
  input: Record<string, unknown>,
  context: { projectPath?: string; cwd?: string }
) => string | undefined

const TOOL_PARAMS_EXTRACTORS: Record<string, ToolParamsExtractor> = {
  // File-based tools: show relative file path
  Read: (input, ctx) => {
    const filePath = input.file_path as string | undefined
    return filePath ? getDisplayPath(filePath, ctx.projectPath || ctx.cwd) : undefined
  },
  Edit: (input, ctx) => {
    const filePath = input.file_path as string | undefined
    return filePath ? getDisplayPath(filePath, ctx.projectPath || ctx.cwd) : undefined
  },
  Write: (input, ctx) => {
    const filePath = input.file_path as string | undefined
    return filePath ? getDisplayPath(filePath, ctx.projectPath || ctx.cwd) : undefined
  },
  // Task tool: show description
  Task: (input) => input.description as string | undefined,
  // Bash tool: show truncated command
  Bash: (input) => {
    const command = input.command as string | undefined
    if (!command) return undefined
    const maxLength = 50
    return command.length > maxLength ? command.slice(0, maxLength) + '...' : command
  },
  // Glob tool: show pattern
  Glob: (input) => input.pattern as string | undefined,
  // Grep tool: show pattern
  Grep: (input) => input.pattern as string | undefined,
}

function extractToolParams(
  toolName: string,
  input: Record<string, unknown> | undefined,
  context: { projectPath?: string; cwd?: string }
): string | undefined {
  if (!input) return undefined
  const extractor = TOOL_PARAMS_EXTRACTORS[toolName]
  return extractor ? extractor(input, context) : undefined
}

// Extract display path from absolute file path
// Returns relative path from cwd if possible, otherwise just the filename
function getDisplayPath(filePath: string, cwd: string | undefined): string {
  if (!filePath) return ''

  // If no cwd, just return filename
  if (!cwd) {
    return filePath.split('/').pop() || filePath
  }

  // Normalize paths (remove trailing slashes)
  const normalizedCwd = cwd.replace(/\/+$/, '')
  const normalizedPath = filePath.replace(/\/+$/, '')

  // Case 1: file_path is under cwd
  if (normalizedPath.startsWith(normalizedCwd + '/')) {
    return normalizedPath.slice(normalizedCwd.length + 1)
  }

  // Case 2: file_path is in parent directories of cwd
  // Find common ancestor and build relative path
  const cwdParts = normalizedCwd.split('/')
  const pathParts = normalizedPath.split('/')

  let commonLength = 0
  for (let i = 0; i < Math.min(cwdParts.length, pathParts.length); i++) {
    if (cwdParts[i] === pathParts[i]) {
      commonLength = i + 1
    } else {
      break
    }
  }

  if (commonLength > 0) {
    const upCount = cwdParts.length - commonLength
    const remainingPath = pathParts.slice(commonLength).join('/')
    if (upCount <= 3) { // Only use ../ if not too deep
      return '../'.repeat(upCount) + remainingPath
    }
  }

  // Case 3: Fallback to just filename
  return filePath.split('/').pop() || filePath
}

// Check if content is a local command input
// Must start with <command-name>/ to avoid false positives from summaries that mention this pattern
function isLocalCommand(content: unknown): boolean {
  if (typeof content !== 'string') return false
  return content.trimStart().startsWith('<command-name>/')
}

// Check if content is a local command output
function isLocalCommandOutput(content: unknown): boolean {
  if (typeof content !== 'string') return false
  return content.includes('<local-command-stdout>')
}

// Check if event is a compact summary
function isCompactSummary(event: Event): boolean {
  return event.payload?.isCompactSummary === true
}

// Check if event is a meta message for local commands
function isMetaMessage(event: Event): boolean {
  return event.payload?.isMeta === true
}

// Build a map to associate related events with their local command
// Returns: Map<event.id, localCommandEvent.id>
function buildLocalCommandGroups(events: Event[]): Map<string, string> {
  const groups = new Map<string, string>()

  // Filter to user events only
  const userEvents = events.filter(e => e.event_type === 'user')

  // Sort events by timestamp
  const sortedEvents = [...userEvents].sort((a, b) => {
    const tsA = (a.payload?.timestamp as string) || a.created_at
    const tsB = (b.payload?.timestamp as string) || b.created_at
    return tsA.localeCompare(tsB)
  })

  // First pass: find all local commands and their timestamps
  const localCommandTimestamps = new Map<string, Event>() // timestamp -> command event
  const localCommandIds = new Set<string>()
  for (const event of sortedEvents) {
    const message = event.payload?.message as Record<string, unknown> | undefined
    const content = message?.content
    if (isLocalCommand(content)) {
      const ts = (event.payload?.timestamp as string) || event.created_at
      localCommandTimestamps.set(ts, event)
      localCommandIds.add(event.id)
    }
  }

  // Second pass: group related events
  let currentLocalCommand: Event | null = null

  for (const event of sortedEvents) {
    // Skip the local command itself
    if (localCommandIds.has(event.id)) {
      currentLocalCommand = event
      continue
    }

    const message = event.payload?.message as Record<string, unknown> | undefined
    const content = message?.content
    const eventTs = (event.payload?.timestamp as string) || event.created_at

    // Check if this is a meta message with same timestamp as a local command
    // (These come before the command in the array but should be grouped)
    if (isMetaMessage(event)) {
      const matchingCommand = localCommandTimestamps.get(eventTs)
      if (matchingCommand) {
        groups.set(event.id, matchingCommand.id)
        continue
      }
    }

    // Check if this event should be grouped with the current local command
    if (currentLocalCommand) {
      if (isCompactSummary(event) || isMetaMessage(event) || isLocalCommandOutput(content)) {
        groups.set(event.id, currentLocalCommand.id)
      } else {
        // Regular user message, end the current group
        currentLocalCommand = null
      }
    }
  }

  return groups
}

// Build a map of tool_use_id -> tool_result content block
function buildToolResultMap(events: Event[]): Map<string, { content: unknown; timestamp: string; event: Event }> {
  const map = new Map<string, { content: unknown; timestamp: string; event: Event }>()

  for (const event of events) {
    if (event.event_type !== 'user') continue

    const message = event.payload?.message as Record<string, unknown> | undefined
    const content = message?.content
    const timestamp = (event.payload?.timestamp as string) || event.created_at

    if (Array.isArray(content)) {
      for (const block of content) {
        const blockObj = block as Record<string, unknown>
        if (blockObj?.type === 'tool_result' && typeof blockObj?.tool_use_id === 'string') {
          map.set(blockObj.tool_use_id, { content: block, timestamp, event })
        }
      }
    }
  }

  return map
}

// Expand events into individual display blocks
function expandEvents(events: Event[], projectPath?: string): DisplayBlock[] {
  const blocks: DisplayBlock[] = []

  // Build groups: Map<relatedEventId, localCommandId>
  const eventToCommandMap = buildLocalCommandGroups(events)

  // Build tool result map: Map<tool_use_id, tool_result_content>
  const toolResultMap = buildToolResultMap(events)

  // Track which tool_result blocks should be skipped (grouped with tool_use)
  const groupedToolResultIds = new Set<string>()

  // Track which events should be skipped (they'll be grouped with their command)
  const relatedEventIds = new Set(eventToCommandMap.keys())

  for (const event of events) {
    const timestamp = (event.payload?.timestamp as string) || event.created_at
    const message = event.payload?.message as Record<string, unknown> | undefined
    const content = message?.content

    if (event.event_type === 'user') {
      // Skip events related to local commands (they'll be grouped with the command)
      if (relatedEventIds.has(event.id)) {
        continue
      }

      // User events: expand content blocks
      if (Array.isArray(content)) {
        content.forEach((block, i) => {
          const blockObj = block as Record<string, unknown>
          const blockType = blockObj?.type as string || 'text'

          // Skip tool_result blocks that have been grouped with their tool_use
          if (blockType === 'tool_result') {
            const toolUseId = blockObj?.tool_use_id as string
            if (toolUseId && groupedToolResultIds.has(toolUseId)) {
              return // Skip this block
            }
          }

          blocks.push({
            id: `${event.id}-${i}`,
            eventType: 'user',
            blockType,
            label: { text: blockType === 'tool_result' ? 'Tool Result' : 'User' },
            timestamp,
            content: block,
            originalEvent: event,
          })
        })
      } else if (isLocalCommand(content)) {
        // Local command input (e.g., /compact, /clear, /hooks)
        const commandName = extractCommandName(content as string) || 'command'

        // Find related events using the pre-built map
        const childBlocks: DisplayBlock[] = []
        for (const relatedEvent of events) {
          // Check if this event belongs to this local command
          const commandId = eventToCommandMap.get(relatedEvent.id)

          if (commandId === event.id) {
            const childTimestamp = (relatedEvent.payload?.timestamp as string) || relatedEvent.created_at
            const childMessage = relatedEvent.payload?.message as Record<string, unknown> | undefined
            const childContent = childMessage?.content

            if (isCompactSummary(relatedEvent)) {
              // Compact summary
              childBlocks.push({
                id: relatedEvent.id,
                eventType: 'user',
                blockType: 'compact_summary',
                label: { text: 'Summary' },
                timestamp: childTimestamp,
                content: childContent,
                originalEvent: relatedEvent,
              })
            } else if (isMetaMessage(relatedEvent)) {
              // Meta message - skip display (it's just a system note)
              continue
            } else if (isLocalCommandOutput(childContent)) {
              // Command output
              childBlocks.push({
                id: relatedEvent.id,
                eventType: 'user',
                blockType: 'local_command_output',
                label: { text: 'Output' },
                timestamp: childTimestamp,
                content: childContent,
                originalEvent: relatedEvent,
              })
            }
          }
        }

        blocks.push({
          id: event.id,
          eventType: 'user',
          blockType: 'local_command_group',
          label: { text: `/${commandName}` },
          timestamp,
          content: content,
          originalEvent: event,
          childBlocks: childBlocks.length > 0 ? childBlocks : undefined,
        })
      } else if (isLocalCommandOutput(content)) {
        // Standalone local command output (not grouped)
        blocks.push({
          id: event.id,
          eventType: 'user',
          blockType: 'local_command_output',
          label: { text: 'Command Output' },
          timestamp,
          content: content,
          originalEvent: event,
        })
      } else {
        // Simple string content
        blocks.push({
          id: event.id,
          eventType: 'user',
          blockType: 'text',
          label: { text: 'User' },
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
          let label: BlockLabel = { text: 'Assistant' }

          if (blockType === 'thinking') {
            label = { text: 'Thinking' }
          } else if (blockType === 'tool_use') {
            const toolName = blockObj?.name as string || 'Unknown'
            const toolUseId = blockObj?.id as string

            const input = blockObj?.input as Record<string, unknown> | undefined
            const params = extractToolParams(toolName, input, {
              projectPath,
              cwd: event.payload?.cwd as string | undefined,
            })
            label = { text: `Tool: ${toolName}`, params }

            // Check if there's a matching tool_result
            const toolResult = toolUseId ? toolResultMap.get(toolUseId) : undefined

            if (toolResult) {
              // Mark this tool_result as grouped
              groupedToolResultIds.add(toolUseId)

              // Create a tool_group block
              const resultBlock: DisplayBlock = {
                id: `${event.id}-${i}-result`,
                eventType: 'user',
                blockType: 'tool_result',
                label: { text: 'Result' },
                timestamp: toolResult.timestamp,
                content: toolResult.content,
                originalEvent: toolResult.event,
              }

              blocks.push({
                id: `${event.id}-${i}`,
                eventType: 'assistant',
                blockType: 'tool_group',
                label,
                timestamp,
                content: block,
                originalEvent: event,
                toolResultBlock: resultBlock,
              })
              return // Skip the normal push
            }
          } else if (blockType === 'tool_result') {
            label = { text: 'Tool Result' }
          } else if (blockType === 'text') {
            label = { text: 'Assistant' }
          } else {
            label = { text: `Assistant (${blockType})` }
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
          label: { text: 'Assistant' },
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
          label: { text: 'Assistant' },
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
        label: { text: event.event_type === 'tool_use' ? `Tool: ${toolName}` : 'Tool Result' },
        timestamp,
        content: event.payload,
        originalEvent: event,
      })
    }
  }

  return blocks
}

export function Timeline({ events, projectPath }: TimelineProps) {
  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No events yet.</p>
      </div>
    )
  }

  const displayBlocks = expandEvents(events, projectPath)

  return (
    <div className="space-y-3">
      {displayBlocks.map((block) => (
        <ContentBlockCard key={block.id} block={block} />
      ))}
    </div>
  )
}
