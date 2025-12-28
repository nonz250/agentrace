interface UserMessageProps {
  payload: Record<string, unknown>
}

export function UserMessage({ payload }: UserMessageProps) {
  const message = payload?.message as Record<string, unknown> | undefined
  const content = message?.content

  if (typeof content === 'string') {
    return (
      <p className="mt-3 whitespace-pre-wrap text-gray-700">{content}</p>
    )
  }

  if (Array.isArray(content)) {
    return (
      <div className="mt-3 space-y-2">
        {content.map((block, i) => {
          if (typeof block === 'string') {
            return (
              <p key={i} className="whitespace-pre-wrap text-gray-700">
                {block}
              </p>
            )
          }
          if (block?.type === 'text' && typeof block.text === 'string') {
            return (
              <p key={i} className="whitespace-pre-wrap text-gray-700">
                {block.text}
              </p>
            )
          }
          return null
        })}
      </div>
    )
  }

  return (
    <pre className="mt-3 whitespace-pre-wrap text-sm text-gray-600">
      {JSON.stringify(payload, null, 2)}
    </pre>
  )
}
