import { useState, useRef, useEffect } from 'react'
import { ChevronDown, X, Check } from 'lucide-react'
import { cn } from '@/lib/cn'

export interface MultiSelectOption {
  value: string
  label: string
  badgeClassName?: string // Custom class for the selected badge
}

interface MultiSelectProps {
  options: MultiSelectOption[]
  selectedValues: string[]
  onChange: (values: string[]) => void
  placeholder?: string
  className?: string
}

export function MultiSelect({
  options,
  selectedValues,
  onChange,
  placeholder = 'Select...',
  className,
}: MultiSelectProps) {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const toggleOption = (value: string, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (selectedValues.includes(value)) {
      onChange(selectedValues.filter((v) => v !== value))
    } else {
      onChange([...selectedValues, value])
    }
  }

  const removeOption = (value: string, e: React.MouseEvent) => {
    e.stopPropagation()
    onChange(selectedValues.filter((v) => v !== value))
  }

  const clearAll = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange([])
  }

  const selectedOptions = selectedValues
    .map((v) => options.find((o) => o.value === v))
    .filter((o): o is MultiSelectOption => o !== undefined)

  return (
    <div ref={containerRef} className={cn('relative inline-block', className)}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'flex items-center gap-2 rounded-lg bg-transparent px-1 py-1 text-left text-sm',
          'hover:bg-gray-100 focus:outline-none',
          isOpen && 'bg-gray-100'
        )}
      >
        <div className="flex items-center gap-1">
          {selectedValues.length === 0 ? (
            <span className="text-gray-500">{placeholder}</span>
          ) : (
            selectedOptions.map((option) => (
              <span
                key={option.value}
                className={cn(
                  'inline-flex items-center gap-1 whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium',
                  option.badgeClassName || 'bg-blue-100 text-blue-800'
                )}
              >
                {option.label}
                <button
                  type="button"
                  onClick={(e) => removeOption(option.value, e)}
                  className="opacity-60 hover:opacity-100"
                >
                  <X className="h-3 w-3" />
                </button>
              </span>
            ))
          )}
        </div>
        <div className="flex items-center gap-1">
          {selectedValues.length > 0 && (
            <button
              type="button"
              onClick={clearAll}
              className="rounded p-0.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            >
              <X className="h-4 w-4" />
            </button>
          )}
          <ChevronDown
            className={cn('h-4 w-4 text-gray-400 transition-transform', isOpen && 'rotate-180')}
          />
        </div>
      </button>

      {isOpen && (
        <div className="absolute z-10 mt-1 max-h-60 min-w-full overflow-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg">
          {options.length === 0 ? (
            <div className="px-3 py-2 text-sm text-gray-500">No options</div>
          ) : (
            options.map((option) => {
              const isSelected = selectedValues.includes(option.value)
              return (
                <button
                  key={option.value}
                  type="button"
                  onClick={(e) => toggleOption(option.value, e)}
                  className={cn(
                    'flex w-full items-center gap-2 whitespace-nowrap px-3 py-2 text-left text-sm',
                    'hover:bg-gray-100',
                    isSelected && 'bg-blue-50'
                  )}
                >
                  <div
                    className={cn(
                      'flex h-4 w-4 shrink-0 items-center justify-center rounded border',
                      isSelected
                        ? 'border-blue-500 bg-blue-500 text-white'
                        : 'border-gray-300 bg-white'
                    )}
                  >
                    {isSelected && <Check className="h-3 w-3" />}
                  </div>
                  <span
                    className={cn(
                      'rounded-full px-2 py-0.5 text-xs font-medium',
                      option.badgeClassName || 'bg-gray-100 text-gray-700'
                    )}
                  >
                    {option.label}
                  </span>
                </button>
              )
            })
          )}
        </div>
      )}
    </div>
  )
}
