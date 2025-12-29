import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'

interface ContentBlock {
  type?: string
  text?: string
  tool_use_id?: string
  content?: unknown
  [key: string]: unknown
}

interface UserMessageProps {
  payload: Record<string, unknown>
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

  // Tool result block
  if (block.type === 'tool_result') {
    const content = block.content

    // Handle different content formats
    let displayContent: string
    if (typeof content === 'string') {
      displayContent = content
    } else if (Array.isArray(content)) {
      // Content can be array of text blocks
      displayContent = content
        .map((c) => {
          if (typeof c === 'string') return c
          if (c?.type === 'text' && typeof c.text === 'string') return c.text
          return JSON.stringify(c, null, 2)
        })
        .join('\n')
    } else {
      displayContent = JSON.stringify(content, null, 2)
    }

    // Check if content looks like JSON
    const isJSON =
      displayContent.trim().startsWith('{') ||
      displayContent.trim().startsWith('[')

    return (
      <div key={index} className="rounded-lg border border-gray-200 bg-gray-50 p-3">
        <p className="mb-2 text-xs font-medium text-gray-500">Tool Result</p>
        {isJSON ? (
          <SyntaxHighlighter
            language="json"
            style={oneLight}
            customStyle={{
              fontSize: '0.75rem',
              borderRadius: '0.375rem',
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
        )}
      </div>
    )
  }

  // Image block (sometimes tool results include images)
  if (block.type === 'image') {
    return (
      <div key={index} className="rounded-lg border border-gray-200 bg-gray-50 p-3">
        <p className="text-xs font-medium text-gray-500">[Image]</p>
      </div>
    )
  }

  // Unknown block type - show as JSON
  if (block.type) {
    return (
      <div key={index} className="rounded-lg bg-gray-100 p-3">
        <p className="mb-2 text-xs font-medium text-gray-500">{block.type}</p>
        <pre className="max-h-[200px] overflow-auto whitespace-pre-wrap text-xs text-gray-600">
          {JSON.stringify(block, null, 2)}
        </pre>
      </div>
    )
  }

  return null
}

export function UserMessage({ payload }: UserMessageProps) {
  const message = payload?.message as Record<string, unknown> | undefined
  const content = message?.content

  // Simple string content
  if (typeof content === 'string') {
    return (
      <p className="mt-3 whitespace-pre-wrap text-gray-700">{content}</p>
    )
  }

  // Array of content blocks
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

  // Fallback: show raw JSON
  return (
    <pre className="mt-3 whitespace-pre-wrap text-sm text-gray-600">
      {JSON.stringify(payload, null, 2)}
    </pre>
  )
}
