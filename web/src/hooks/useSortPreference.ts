import { useState, useCallback, useEffect } from 'react'
import type { SortBy } from '@/api/sessions'

type SortPreferenceKey = 'sessions' | 'plans'

function getStorageKey(key: SortPreferenceKey): string {
  return `agentrace:${key}:sort`
}

export function useSortPreference(key: SortPreferenceKey) {
  const storageKey = getStorageKey(key)

  const [sort, setSort] = useState<SortBy>(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(storageKey)
      if (stored === 'created_at') {
        return 'created_at'
      }
    }
    return 'updated_at'
  })

  // Sync with localStorage on mount
  useEffect(() => {
    const stored = localStorage.getItem(storageKey)
    if (stored === 'created_at') {
      setSort('created_at')
    }
  }, [storageKey])

  const updateSort = useCallback((newSort: SortBy) => {
    setSort(newSort)
    localStorage.setItem(storageKey, newSort)
  }, [storageKey])

  return { sort, updateSort }
}
