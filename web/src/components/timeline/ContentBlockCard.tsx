import { cn } from '@/lib/cn'
import { User, Bot, Wrench, Sparkles, ChevronDown, ChevronRight, Terminal, FileText, ExternalLink, Loader2, ArrowRight, Link2, Check } from 'lucide-react'
import { useState } from 'react'
import { format } from 'date-fns'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import ReactMarkdown from 'react-markdown'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getPlan } from '@/api/plan-documents'
import { PlanStatusBadge } from '@/components/plans/PlanStatusBadge'
import type { PlanDocumentStatus } from '@/types/plan-document'
import type { DisplayBlock, PlanLinkInfo } from './Timeline'

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
    'agentrace_tool',
    'local_command',
    'local_command_output',
    'local_command_group',
  ]
  return secondaryBlockTypes.includes(block.blockType)
}

// Check if block should be expanded by default
function shouldExpandByDefault(block: DisplayBlock): boolean {
  // Primary blocks are always expanded
  if (!isSecondaryBlock(block)) return true
  // Agentrace tools should be expanded by default
  if (block.blockType === 'agentrace_tool') return true
  return false
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
  // Tool-related blocks (including agentrace_tool) use Wrench icon
  if (block.blockType === 'tool_use' || block.blockType === 'tool_result' || block.blockType === 'tool_group' || block.blockType === 'agentrace_tool') {
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

// Plan link card component - fetches latest plan data from API
function PlanLinkCard({ plan }: { plan: PlanLinkInfo }) {
  const { data: planData, isLoading, isError } = useQuery({
    queryKey: ['plan', plan.id],
    queryFn: () => getPlan(plan.id),
    staleTime: 30 * 1000, // 30 seconds
  })

  // Truncate plan ID for display
  const shortId = plan.id.length > 8 ? plan.id.slice(0, 8) + '...' : plan.id

  if (isLoading) {
    return (
      <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-3">
        <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
        <span className="text-sm text-gray-500">Loading plan...</span>
      </div>
    )
  }

  if (isError || !planData) {
    return (
      <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-3">
        <FileText className="h-4 w-4 text-gray-400" />
        <span className="text-sm text-gray-500">Plan {shortId}</span>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {plan.changedStatus && (
        <div className="flex items-center gap-2 text-sm text-gray-600">
          <span>Status changed</span>
          <ArrowRight className="h-3 w-3 text-gray-400" />
          <PlanStatusBadge status={plan.changedStatus as PlanDocumentStatus} />
        </div>
      )}
      <Link
        to={`/plans/${plan.id}`}
        className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3 transition-colors hover:border-gray-300 hover:bg-gray-50"
      >
        <div className="flex items-center gap-3 min-w-0">
          <FileText className="h-4 w-4 flex-shrink-0 text-gray-400" />
          <div className="flex items-center gap-2 min-w-0">
            <span className="truncate text-sm font-medium text-gray-900">
              {planData.description}
            </span>
            <PlanStatusBadge status={planData.status} />
          </div>
        </div>
        <ExternalLink className="h-4 w-4 flex-shrink-0 text-gray-400" />
      </Link>
    </div>
  )
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

  // Agentrace tool - show plan cards with links
  if (block.blockType === 'agentrace_tool') {
    const planLinks = block.planLinks || []

    if (planLinks.length === 0) {
      return (
        <div className="text-sm text-gray-500 italic">
          Operation completed
        </div>
      )
    }

    return (
      <div className="space-y-2">
        {planLinks.map((plan) => (
          <PlanLinkCard key={plan.id} plan={plan} />
        ))}
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

// Permalink button component with copy feedback
function PermalinkButton({ blockId }: { blockId: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation()
    const url = `${window.location.origin}${window.location.pathname}#event-${blockId}`
    navigator.clipboard.writeText(url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      onClick={handleCopy}
      className="rounded p-1 text-gray-400 opacity-0 transition-all hover:bg-gray-100 hover:text-gray-600 group-hover:opacity-100"
      title="Copy link to this block"
    >
      {copied ? (
        <Check className="h-3.5 w-3.5 text-green-500" />
      ) : (
        <Link2 className="h-3.5 w-3.5" />
      )}
    </button>
  )
}

export function ContentBlockCard({ block }: ContentBlockCardProps) {
  const isSecondary = isSecondaryBlock(block)
  // Use shouldExpandByDefault for initial state
  const [expanded, setExpanded] = useState(() => shouldExpandByDefault(block))
  const styles = getBlockContainerStyle(block)

  // Primary blocks (User/Assistant) don't have collapse functionality
  if (!isSecondary) {
    return (
      <div className={styles.wrapper}>
        <div className={cn('group overflow-hidden rounded-xl', styles.container)}>
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
            <div className="flex items-center gap-2 text-sm text-gray-500">
              <span>{format(new Date(block.timestamp), 'HH:mm:ss')}</span>
              <PermalinkButton blockId={block.id} />
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
      <div className={cn('group overflow-hidden rounded-xl', styles.container)}>
        <div className={cn('flex w-full items-center justify-between', styles.header)}>
          <button
            className="flex flex-1 items-center gap-2 text-left transition-colors hover:bg-gray-50/50"
            onClick={() => setExpanded(!expanded)}
          >
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
          </button>
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <span>{format(new Date(block.timestamp), 'HH:mm:ss')}</span>
            <PermalinkButton blockId={block.id} />
            <button
              onClick={() => setExpanded(!expanded)}
              className="p-0.5 hover:bg-gray-100 rounded"
            >
              {expanded ? (
                <ChevronDown className="h-3 w-3" />
              ) : (
                <ChevronRight className="h-3 w-3" />
              )}
            </button>
          </div>
        </div>

        {expanded && (
          <div className="border-t border-gray-100 px-3 py-2">
            {renderContent(block)}
          </div>
        )}
      </div>
    </div>
  )
}
