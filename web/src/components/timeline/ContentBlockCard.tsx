import { cn } from '@/lib/cn'
import { User, Bot, Wrench, Sparkles, ChevronDown, ChevronRight, Terminal } from 'lucide-react'
import { useState } from 'react'
import { format } from 'date-fns'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import ReactMarkdown from 'react-markdown'
import type { DisplayBlock } from './Timeline'

interface ContentBlockCardProps {
  block: DisplayBlock
}

// Check if block is secondary (less prominent) - Thinking, Tool, LocalCommand
function isSecondaryBlock(block: DisplayBlock): boolean {
  const secondaryBlockTypes = [
    'thinking',
    'tool_use',
    'tool_result',
    'tool_group',
    'local_command',
    'local_command_output',
    'local_command_group',
  ]
  return secondaryBlockTypes.includes(block.blockType)
}

// Get container style based on block prominence
function getBlockContainerStyle(block: DisplayBlock) {
  if (isSecondaryBlock(block)) {
    return {
      wrapper: 'ml-4',
      container: 'border border-gray-200 bg-gray-50/50',
      header: 'px-3 py-2',
    }
  }
  // Primary blocks (User/Assistant text)
  const borderColor = block.eventType === 'user' ? 'border-blue-200' : 'border-green-200'
  return {
    wrapper: '',
    container: `border-2 ${borderColor} bg-white`,
    header: 'px-4 py-3',
  }
}

function getIcon(block: DisplayBlock) {
  // Local command uses Terminal icon
  if (block.blockType === 'local_command' || block.blockType === 'local_command_output' || block.blockType === 'local_command_group') {
    return <Terminal className="h-4 w-4" />
  }
  // Tool-related blocks use Wrench icon
  if (block.blockType === 'tool_use' || block.blockType === 'tool_result' || block.blockType === 'tool_group') {
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
  // Secondary blocks use muted gray
  if (isSecondaryBlock(block)) {
    return 'bg-gray-200 text-gray-500'
  }
  // User (non-tool) uses blue
  if (block.eventType === 'user') {
    return 'bg-blue-100 text-blue-600'
  }
  // Default: green for assistant
  return 'bg-green-100 text-green-600'
}

// Extract content from local command output tags
function extractCommandOutput(content: string): string {
  const match = content.match(/<local-command-stdout>([\s\S]*?)<\/local-command-stdout>/)
  return match ? match[1].trim() : content
}

function renderContent(block: DisplayBlock) {
  const content = block.content as Record<string, unknown>

  // Local command group - show command and its children together
  if (block.blockType === 'local_command_group') {
    return (
      <div className="space-y-3">
        <div className="text-sm text-gray-500 italic">
          Command executed
        </div>
        {block.childBlocks && block.childBlocks.length > 0 && (
          <div className="space-y-2 border-l-2 border-gray-200 pl-3">
            {block.childBlocks.map((child) => (
              <div key={child.id}>
                <div className="mb-1 text-xs font-medium text-gray-400">
                  {child.label.text}
                </div>
                {child.blockType === 'local_command_output' ? (
                  <pre className="max-h-[200px] overflow-auto whitespace-pre-wrap rounded-lg bg-gray-50 p-2 font-mono text-xs text-gray-600">
                    {typeof child.content === 'string'
                      ? extractCommandOutput(child.content)
                      : ''}
                  </pre>
                ) : child.blockType === 'compact_summary' ? (
                  <div className="max-h-[300px] overflow-auto rounded-lg bg-amber-50 p-3 text-xs text-gray-700">
                    <pre className="whitespace-pre-wrap font-mono">
                      {typeof child.content === 'string' ? child.content : ''}
                    </pre>
                  </div>
                ) : (
                  <div className="prose prose-sm max-w-none text-gray-600">
                    {typeof child.content === 'string' ? child.content : ''}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }

  // Tool group - show tool input and result together
  if (block.blockType === 'tool_group') {
    const input = content?.input
    const resultBlock = block.toolResultBlock
    const resultContent = resultBlock?.content as Record<string, unknown> | undefined
    const resultData = resultContent?.content

    let displayResult: string
    if (typeof resultData === 'string') {
      displayResult = resultData
    } else if (Array.isArray(resultData)) {
      displayResult = resultData
        .map((c) => {
          if (typeof c === 'string') return c
          if (c?.type === 'text' && typeof c.text === 'string') return c.text
          return JSON.stringify(c, null, 2)
        })
        .join('\n')
    } else {
      displayResult = JSON.stringify(resultData, null, 2)
    }

    return (
      <div className="space-y-3">
        <div>
          <div className="mb-1 text-xs font-medium text-gray-400">Input</div>
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
            {JSON.stringify(input, null, 2)}
          </SyntaxHighlighter>
        </div>
        {resultBlock && (
          <div>
            <div className="mb-1 text-xs font-medium text-gray-400">Result</div>
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
              {displayResult}
            </SyntaxHighlighter>
          </div>
        )}
      </div>
    )
  }

  // Local command input - show command name only (content is minimal)
  if (block.blockType === 'local_command') {
    return (
      <div className="text-sm text-gray-500 italic">
        Command executed
      </div>
    )
  }

  // Local command output - extract and display the output
  if (block.blockType === 'local_command_output') {
    const output = typeof block.content === 'string'
      ? extractCommandOutput(block.content)
      : ''

    if (!output) {
      return (
        <div className="text-sm text-gray-400 italic">
          (no output)
        </div>
      )
    }

    return (
      <pre className="max-h-[300px] overflow-auto whitespace-pre-wrap rounded-lg bg-gray-50 p-3 font-mono text-xs text-gray-600">
        {output}
      </pre>
    )
  }

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
  const isSecondary = isSecondaryBlock(block)
  // Secondary blocks default to collapsed, primary blocks are always expanded
  const [expanded, setExpanded] = useState(!isSecondary)
  const styles = getBlockContainerStyle(block)

  // Primary blocks (User/Assistant) don't have collapse functionality
  if (!isSecondary) {
    return (
      <div className={styles.wrapper}>
        <div className={cn('overflow-hidden rounded-xl', styles.container)}>
          <div className={cn('flex items-center justify-between', styles.header)}>
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  'flex h-6 w-6 items-center justify-center rounded-full',
                  getIconStyle(block)
                )}
              >
                {getIcon(block)}
              </span>
              <span className="font-medium text-gray-900">{block.label.text}</span>
            </div>
            <div className="text-sm text-gray-500">
              <span>{format(new Date(block.timestamp), 'HH:mm:ss')}</span>
            </div>
          </div>
          <div className="border-t border-gray-100 px-4 py-3">
            {renderContent(block)}
          </div>
        </div>
      </div>
    )
  }

  // Secondary blocks have collapse functionality
  return (
    <div className={styles.wrapper}>
      <div className={cn('overflow-hidden rounded-xl', styles.container)}>
        <button
          className={cn(
            'flex w-full items-center justify-between',
            styles.header,
            'text-left transition-colors hover:bg-gray-50/50'
          )}
          onClick={() => setExpanded(!expanded)}
        >
          <div className="flex items-center gap-2">
            <span
              className={cn(
                'flex h-5 w-5 items-center justify-center rounded-full',
                getIconStyle(block)
              )}
            >
              {getIcon(block)}
            </span>
            <span className="text-sm font-medium text-gray-600">
              {block.label.text}
            </span>
            {block.label.params && (
              <code className="rounded bg-gray-100 px-1 py-0.5 text-xs font-normal text-gray-700">
                {block.label.params}
              </code>
            )}
          </div>
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <span>{format(new Date(block.timestamp), 'HH:mm:ss')}</span>
            {expanded ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronRight className="h-3 w-3" />
            )}
          </div>
        </button>

        {expanded && (
          <div className="border-t border-gray-100 px-3 py-2">
            {renderContent(block)}
          </div>
        )}
      </div>
    </div>
  )
}
