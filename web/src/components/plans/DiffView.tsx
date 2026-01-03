interface DiffViewProps {
  patch: string
}

interface DiffLine {
  type: 'header' | 'context' | 'addition' | 'deletion'
  content: string
}

function safeDecode(text: string): string {
  try {
    return decodeURIComponent(text)
  } catch {
    return text
  }
}

function parsePatch(patch: string): DiffLine[] {
  const lines: DiffLine[] = []

  // Split by lines, handling different line endings
  const rawLines = patch.split(/\r?\n/)

  for (const line of rawLines) {
    if (line.startsWith('@@')) {
      // Hunk header: @@ -start,len +start,len @@
      lines.push({ type: 'header', content: line })
    } else if (line.startsWith('+')) {
      // Addition - decode URL-encoded content from diff-match-patch
      lines.push({ type: 'addition', content: safeDecode(line.slice(1)) })
    } else if (line.startsWith('-')) {
      // Deletion - decode URL-encoded content from diff-match-patch
      lines.push({ type: 'deletion', content: safeDecode(line.slice(1)) })
    } else if (line.startsWith(' ')) {
      // Context (unchanged) - decode URL-encoded content
      lines.push({ type: 'context', content: safeDecode(line.slice(1)) })
    } else if (line.trim() !== '') {
      // Other content (treat as context)
      lines.push({ type: 'context', content: safeDecode(line) })
    }
  }

  return lines
}

function getLineStyle(type: DiffLine['type']): string {
  switch (type) {
    case 'header':
      return 'bg-blue-50 text-blue-700 font-medium'
    case 'addition':
      return 'bg-green-50 text-green-800'
    case 'deletion':
      return 'bg-red-50 text-red-800'
    case 'context':
    default:
      return 'text-gray-600'
  }
}

function getLinePrefix(type: DiffLine['type']): string {
  switch (type) {
    case 'header':
      return ''
    case 'addition':
      return '+'
    case 'deletion':
      return '-'
    case 'context':
    default:
      return ' '
  }
}

export function DiffView({ patch }: DiffViewProps) {
  const lines = parsePatch(patch)

  if (lines.length === 0) {
    return (
      <pre className="mt-2 overflow-x-auto rounded bg-gray-50 p-2 text-gray-600 font-mono text-xs">
        {patch}
      </pre>
    )
  }

  return (
    <div className="mt-2 overflow-x-auto rounded border border-gray-200 font-mono text-xs">
      {lines.map((line, index) => (
        <div
          key={index}
          className={`px-2 py-0.5 whitespace-pre ${getLineStyle(line.type)}`}
        >
          {line.type !== 'header' && (
            <span className="inline-block w-4 select-none opacity-50">
              {getLinePrefix(line.type)}
            </span>
          )}
          {line.content || '\u00A0'}
        </div>
      ))}
    </div>
  )
}
