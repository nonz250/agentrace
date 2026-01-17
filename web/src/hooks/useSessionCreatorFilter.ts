import { useState, useEffect, useCallback } from 'react'

const STORAGE_KEY = 'agentrace:session-creator-filter'

function loadFromStorage(): string[] | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === null) {
      return null
    }
    const parsed = JSON.parse(stored)
    if (Array.isArray(parsed)) {
      return parsed.filter((s): s is string => typeof s === 'string')
    }
    return null
  } catch {
    return null
  }
}

function saveToStorage(creatorIds: string[]): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(creatorIds))
  } catch {
    // localStorage unavailable - ignore
  }
}

export function useSessionCreatorFilter() {
  const [selectedCreatorIds, setSelectedCreatorIds] = useState<string[]>(() => {
    const stored = loadFromStorage()
    return stored ?? []
  })

  useEffect(() => {
    saveToStorage(selectedCreatorIds)
  }, [selectedCreatorIds])

  const setCreatorIds = useCallback((ids: string[]) => {
    setSelectedCreatorIds(ids)
  }, [])

  const toggleCreator = useCallback((creatorId: string) => {
    setSelectedCreatorIds((prev) => {
      if (prev.includes(creatorId)) {
        return prev.filter((id) => id !== creatorId)
      } else {
        return [...prev, creatorId]
      }
    })
  }, [])

  const clearAll = useCallback(() => {
    setSelectedCreatorIds([])
  }, [])

  return {
    selectedCreatorIds,
    setCreatorIds,
    toggleCreator,
    clearAll,
  }
}
