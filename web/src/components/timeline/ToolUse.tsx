import { useState } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'

interface ToolUseProps {
  payload: Record<string, unknown>
  isResult?: boolean
}

export function ToolUse({ payload, isResult }: ToolUseProps) {
  const [showContent, setShowContent] = useState(false)

  if (isResult) {
    const content = payload?.content
    const isError = payload?.is_error === true

    return (
      <div className="mt-3">
        {isError && (
          <p className="mb-2 text-sm font-medium text-red-600">Error</p>
        )}
        {typeof content === 'string' ? (
          <pre className="whitespace-pre-wrap rounded-lg bg-gray-50 p-3 text-sm text-gray-700">
            {content.length > 500 ? (
              <>
                {showContent ? content : content.slice(0, 500) + '...'}
                <button
                  onClick={() => setShowContent(!showContent)}
                  className="mt-2 flex items-center gap-1 text-primary-600 hover:underline"
                >
                  {showContent ? (
                    <>
                      <ChevronDown className="h-4 w-4" /> Show less
                    </>
                  ) : (
                    <>
                      <ChevronRight className="h-4 w-4" /> Show more
                    </>
                  )}
                </button>
              </>
            ) : (
              content
            )}
          </pre>
        ) : (
          <SyntaxHighlighter
            language="json"
            style={oneLight}
            customStyle={{
              fontSize: '0.875rem',
              borderRadius: '0.5rem',
              margin: 0,
            }}
          >
            {JSON.stringify(content, null, 2)}
          </SyntaxHighlighter>
        )}
      </div>
    )
  }

  const name = payload?.name || 'Unknown'
  const input = payload?.input

  return (
    <div className="mt-3">
      <p className="mb-2 text-sm text-gray-500">
        <span className="font-medium text-gray-700">{String(name)}</span>
      </p>
      <button
        onClick={() => setShowContent(!showContent)}
        className="flex items-center gap-1 text-sm text-primary-600 hover:underline"
      >
        {showContent ? (
          <>
            <ChevronDown className="h-4 w-4" /> Hide input
          </>
        ) : (
          <>
            <ChevronRight className="h-4 w-4" /> Show input
          </>
        )}
      </button>
      {showContent && (
        <div className="mt-2">
          <SyntaxHighlighter
            language="json"
            style={oneLight}
            customStyle={{
              fontSize: '0.875rem',
              borderRadius: '0.5rem',
              margin: 0,
            }}
          >
            {JSON.stringify(input, null, 2)}
          </SyntaxHighlighter>
        </div>
      )}
    </div>
  )
}
