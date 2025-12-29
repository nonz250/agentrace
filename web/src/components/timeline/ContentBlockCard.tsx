import { cn } from '@/lib/cn'
import { User, Bot, Wrench, Sparkles, ChevronDown, ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { format } from 'date-fns'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import ReactMarkdown from 'react-markdown'
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
      return (
        <div className="prose prose-sm max-w-none text-gray-700 prose-headings:text-gray-900 prose-code:rounded prose-code:bg-gray-100 prose-code:px-1 prose-code:py-0.5 prose-code:text-gray-800 prose-code:before:content-none prose-code:after:content-none prose-pre:bg-gray-100 prose-pre:text-gray-800">
          <ReactMarkdown
            components={{
              code({ className, children, ...props }) {
                const match = /language-(\w+)/.exec(className || '')
                const code = String(children).replace(/\n$/, '')
                return match ? (
                  <SyntaxHighlighter
                    language={match[1]}
                    style={oneLight}
                    customStyle={{
                      fontSize: '0.75rem',
                      borderRadius: '0.5rem',
                      margin: 0,
                    }}
                  >
                    {code}
                  </SyntaxHighlighter>
                ) : (
                  <code className={className} {...props}>
                    {children}
                  </code>
                )
              },
            }}
          >
            {text}
          </ReactMarkdown>
        </div>
      )
    }
  }

  // Thinking block
  if (block.blockType === 'thinking') {
    const thinking = content?.thinking as string
    return (
      <div className="prose prose-sm max-w-none text-purple-900 prose-headings:text-purple-900 prose-code:rounded prose-code:bg-purple-100 prose-code:px-1 prose-code:py-0.5 prose-code:text-purple-800 prose-code:before:content-none prose-code:after:content-none prose-pre:bg-purple-50 prose-pre:text-purple-900">
        <ReactMarkdown
          components={{
            code({ className, children, ...props }) {
              const match = /language-(\w+)/.exec(className || '')
              const code = String(children).replace(/\n$/, '')
              return match ? (
                <SyntaxHighlighter
                  language={match[1]}
                  style={oneLight}
                  customStyle={{
                    fontSize: '0.75rem',
                    borderRadius: '0.5rem',
                    margin: 0,
                  }}
                >
                  {code}
                </SyntaxHighlighter>
              ) : (
                <code className={className} {...props}>
                  {children}
                </code>
              )
            },
          }}
        >
          {thinking}
        </ReactMarkdown>
      </div>
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
        {displayContent}
      </SyntaxHighlighter>
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
  // Thinking, Tool Use, Tool Result blocks default to collapsed
  const [expanded, setExpanded] = useState(
    block.blockType !== 'thinking' &&
      block.blockType !== 'tool_use' &&
      block.blockType !== 'tool_result'
  )

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
