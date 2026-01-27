import { format } from 'date-fns'
import { History } from 'lucide-react'
import type { PlanDocumentEvent } from '@/types/plan-document'

interface VersionSelectorProps {
  events: PlanDocumentEvent[]
  selectedIndex: number | null
  onSelect: (index: number | null) => void
  disabled?: boolean
  variant?: 'default' | 'warning'
}

export function VersionSelector({
  events,
  selectedIndex,
  onSelect,
  disabled = false,
  variant = 'default',
}: VersionSelectorProps) {
  if (events.length === 0) {
    return null
  }

  const effectiveIndex = selectedIndex ?? events.length - 1
  const isLatest = effectiveIndex === events.length - 1

  const formatLabel = (event: PlanDocumentEvent, index: number) => {
    const version = `v${index + 1}`
    const date = format(new Date(event.created_at), 'yyyy/MM/dd HH:mm')
    const suffix =
      index === events.length - 1
        ? ' (latest)'
        : index === 0
          ? ' (initial)'
          : ''
    return `${version}${suffix} - ${date}`
  }

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const index = Number(e.target.value)
    // null means "latest" (no historical view)
    onSelect(index === events.length - 1 ? null : index)
  }

  if (variant === 'warning') {
    return (
      <select
        value={effectiveIndex.toString()}
        onChange={handleChange}
        disabled={disabled}
        className="appearance-none bg-transparent border-none text-xs font-medium text-amber-800 py-0 pr-0 cursor-pointer hover:text-amber-900 hover:underline focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
      >
        {events.map((event, index) => (
          <option key={event.id} value={index.toString()}>
            {formatLabel(event, index)}
          </option>
        ))}
      </select>
    )
  }

  return (
    <div className="flex items-center gap-1.5 text-xs text-gray-400">
      <History className={`h-3 w-3 ${isLatest ? 'text-gray-300' : 'text-amber-400'}`} />
      <select
        value={effectiveIndex.toString()}
        onChange={handleChange}
        disabled={disabled}
        className="bg-transparent border-none text-xs text-gray-400 py-0.5 pr-5 cursor-pointer hover:text-gray-600 focus:outline-none focus:text-gray-600 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {events.map((event, index) => (
          <option key={event.id} value={index.toString()}>
            {formatLabel(event, index)}
          </option>
        ))}
      </select>
    </div>
  )
}
