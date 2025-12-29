import { cn } from '@/lib/cn'
import { User, Bot, Wrench, Sparkles, ChevronDown, ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { format } from 'date-fns'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import type { DisplayBlock } from './Timeline'

interface ContentBlockCardProps {
  block: DisplayBlock
}

function getIcon(block: DisplayBlock) {
  // Tool-related blocks use Wrench icon
  if (block.blockType === 'tool_use' || block.blockType === 'tool_result') {
    return <Wrench className="h-4 w-4" />
  }
  if (block.eventType === 'tool_use' || block.eventType === 'tool_result') {
    return <Wrench className="h-4 w-4" />
  }
  // Thinking uses Sparkles
  if (block.blockType === 'thinking') {
    return <Sparkles className="h-4 w-4" />
  }
  // User (non-tool) uses User icon
  if (block.eventType === 'user') {
    return <User className="h-4 w-4" />
  }
  // Default: Bot icon for assistant
  return <Bot className="h-4 w-4" />
}

function getIconStyle(block: DisplayBlock) {
  // Tool-related blocks use orange
  if (block.blockType === 'tool_use' || block.blockType === 'tool_result') {
    return 'bg-orange-100 text-orange-600'
  }
  if (block.eventType === 'tool_use' || block.eventType === 'tool_result') {
    return 'bg-orange-100 text-orange-600'
  }
  // Thinking uses purple
  if (block.blockType === 'thinking') {
    return 'bg-purple-100 text-purple-600'
  }
  // User (non-tool) uses blue
  if (block.eventType === 'user') {
    return 'bg-blue-100 text-blue-600'
  }
  // Default: green for assistant
  return 'bg-green-100 text-green-600'
}

function renderContent(block: DisplayBlock) {
  const content = block.content as Record<string, unknown>

  // Text content (user or assistant)
  if (block.blockType === 'text') {
    const text = typeof content === 'string' ? content : content?.text
    if (typeof text === 'string') {
      return <p className="whitespace-pre-wrap text-gray-700">{text}</p>
    }
  }

  // Thinking block
  if (block.blockType === 'thinking') {
    const thinking = content?.thinking as string
    return (
      <p className="whitespace-pre-wrap text-sm text-purple-900">{thinking}</p>
    )
  }

  // Tool use block
  if (block.blockType === 'tool_use') {
    const input = content?.input
    return (
      <SyntaxHighlighter
        language="json"
        style={oneLight}
        customStyle={{
          fontSize: '0.75rem',
          borderRadius: '0.5rem',
          margin: 0,
          maxHeight: '400px',
          overflow: 'auto',
        }}
      >
        {JSON.stringify(input, null, 2)}
      </SyntaxHighlighter>
    )
  }

  // Tool result in user message
  if (block.blockType === 'tool_result') {
    const resultContent = content?.content
    let displayContent: string

    if (typeof resultContent === 'string') {
      displayContent = resultContent
    } else if (Array.isArray(resultContent)) {
      displayContent = resultContent
        .map((c) => {
          if (typeof c === 'string') return c
          if (c?.type === 'text' && typeof c.text === 'string') return c.text
          return JSON.stringify(c, null, 2)
        })
        .join('\n')
    } else {
      displayContent = JSON.stringify(resultContent, null, 2)
    }

    const isJSON =
      displayContent.trim().startsWith('{') ||
      displayContent.trim().startsWith('[')

    return isJSON ? (
      <SyntaxHighlighter
        language="json"
        style={oneLight}
        customStyle={{
          fontSize: '0.75rem',
          borderRadius: '0.5rem',
          margin: 0,
          maxHeight: '300px',
          overflow: 'auto',
        }}
      >
        {displayContent}
      </SyntaxHighlighter>
    ) : (
      <pre className="max-h-[300px] overflow-auto whitespace-pre-wrap text-xs text-gray-700">
        {displayContent}
      </pre>
    )
  }

  // Standalone tool_use/tool_result events
  if (block.eventType === 'tool_use' || block.eventType === 'tool_result') {
    const payload = block.content as Record<string, unknown>
    const input = payload?.input || payload?.result || payload

    return (
      <SyntaxHighlighter
        language="json"
        style={oneLight}
        customStyle={{
          fontSize: '0.75rem',
          borderRadius: '0.5rem',
          margin: 0,
          maxHeight: '400px',
          overflow: 'auto',
        }}
      >
        {JSON.stringify(input, null, 2)}
      </SyntaxHighlighter>
    )
  }

  // Fallback: show as JSON
  return (
    <pre className="max-h-[300px] overflow-auto whitespace-pre-wrap text-xs text-gray-600">
      {JSON.stringify(content, null, 2)}
    </pre>
  )
}

export function ContentBlockCard({ block }: ContentBlockCardProps) {
  // Thinking blocks default to collapsed
  const [expanded, setExpanded] = useState(block.blockType !== 'thinking')

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
              getIconStyle(block)
            )}
          >
            {getIcon(block)}
          </span>
          <span className="font-medium text-gray-900">{block.label}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <span>{format(new Date(block.timestamp), 'HH:mm:ss')}</span>
          {expanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </div>
      </button>

      {expanded && (
        <div className="border-t border-gray-100 px-4 py-3">
          {renderContent(block)}
        </div>
      )}
    </div>
  )
}
