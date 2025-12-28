import { useState } from 'react'
import { ChevronDown, ChevronRight, Brain } from 'lucide-react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'

interface ContentBlock {
  type?: string
  text?: string
  thinking?: string
  signature?: string
  name?: string
  input?: unknown
  [key: string]: unknown
}

interface AssistantMessageProps {
  payload: Record<string, unknown>
}

function ThinkingBlock({ thinking }: { thinking: string }) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="rounded-lg border border-purple-200 bg-purple-50">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-purple-700 hover:bg-purple-100"
      >
        <Brain className="h-4 w-4" />
        <span>Thinking</span>
        {expanded ? (
          <ChevronDown className="ml-auto h-4 w-4" />
        ) : (
          <ChevronRight className="ml-auto h-4 w-4" />
        )}
      </button>
      {expanded && (
        <div className="border-t border-purple-200 px-3 py-2">
          <p className="whitespace-pre-wrap text-sm text-purple-900">
            {thinking}
          </p>
        </div>
      )}
    </div>
  )
}

function renderContentBlock(block: ContentBlock, index: number) {
  // Text block
  if (block.type === 'text' && typeof block.text === 'string') {
    return (
      <p key={index} className="whitespace-pre-wrap text-gray-700">
        {block.text}
      </p>
    )
  }

  // Thinking block
  if (block.type === 'thinking' && typeof block.thinking === 'string') {
    return <ThinkingBlock key={index} thinking={block.thinking} />
  }

  // Tool use block
  if (block.type === 'tool_use') {
    return (
      <div key={index} className="rounded-lg bg-gray-50 p-3">
        <p className="mb-2 text-sm font-medium text-gray-600">
          Tool: {String(block.name || 'Unknown')}
        </p>
        <SyntaxHighlighter
          language="json"
          style={oneLight}
          customStyle={{
            fontSize: '0.875rem',
            borderRadius: '0.5rem',
            margin: 0,
          }}
        >
          {JSON.stringify(block.input, null, 2)}
        </SyntaxHighlighter>
      </div>
    )
  }

  // Tool result block
  if (block.type === 'tool_result') {
    return (
      <div key={index} className="rounded-lg bg-gray-50 p-3">
        <p className="mb-2 text-sm font-medium text-gray-600">Tool Result</p>
        <pre className="whitespace-pre-wrap text-sm text-gray-700">
          {typeof block.content === 'string'
            ? block.content
            : JSON.stringify(block.content, null, 2)}
        </pre>
      </div>
    )
  }

  // Unknown block type - show as JSON
  if (block.type) {
    return (
      <div key={index} className="rounded-lg bg-gray-100 p-3">
        <p className="mb-2 text-xs font-medium text-gray-500">
          {block.type}
        </p>
        <pre className="whitespace-pre-wrap text-sm text-gray-600">
          {JSON.stringify(block, null, 2)}
        </pre>
      </div>
    )
  }

  return null
}

export function AssistantMessage({ payload }: AssistantMessageProps) {
  const message = payload?.message as Record<string, unknown> | undefined
  const content = message?.content

  if (!content) {
    return (
      <pre className="mt-3 whitespace-pre-wrap text-sm text-gray-600">
        {JSON.stringify(payload, null, 2)}
      </pre>
    )
  }

  if (typeof content === 'string') {
    return (
      <p className="mt-3 whitespace-pre-wrap text-gray-700">{content}</p>
    )
  }

  if (Array.isArray(content)) {
    const renderedBlocks = content
      .map((block, i) => {
        if (typeof block === 'string') {
          return (
            <p key={i} className="whitespace-pre-wrap text-gray-700">
              {block}
            </p>
          )
        }
        return renderContentBlock(block as ContentBlock, i)
      })
      .filter(Boolean)

    if (renderedBlocks.length === 0) {
      return (
        <pre className="mt-3 whitespace-pre-wrap text-sm text-gray-600">
          {JSON.stringify(payload, null, 2)}
        </pre>
      )
    }

    return <div className="mt-3 space-y-3">{renderedBlocks}</div>
  }

  return (
    <pre className="mt-3 whitespace-pre-wrap text-sm text-gray-600">
      {JSON.stringify(payload, null, 2)}
    </pre>
  )
}
