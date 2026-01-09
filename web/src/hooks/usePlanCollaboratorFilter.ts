import { useState, useEffect, useCallback } from 'react'

const STORAGE_KEY = 'agentrace:plan-collaborator-filter'

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

function saveToStorage(collaboratorIds: string[]): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(collaboratorIds))
  } catch {
    // localStorage unavailable - ignore
  }
}

export function usePlanCollaboratorFilter() {
  const [selectedCollaboratorIds, setSelectedCollaboratorIds] = useState<string[]>(() => {
    const stored = loadFromStorage()
    return stored ?? []
  })

  useEffect(() => {
    saveToStorage(selectedCollaboratorIds)
  }, [selectedCollaboratorIds])

  const setCollaboratorIds = useCallback((ids: string[]) => {
    setSelectedCollaboratorIds(ids)
  }, [])

  const toggleCollaborator = useCallback((collaboratorId: string) => {
    setSelectedCollaboratorIds((prev) => {
      if (prev.includes(collaboratorId)) {
        return prev.filter((id) => id !== collaboratorId)
      } else {
        return [...prev, collaboratorId]
      }
    })
  }, [])

  const clearAll = useCallback(() => {
    setSelectedCollaboratorIds([])
  }, [])

  return {
    selectedCollaboratorIds,
    setCollaboratorIds,
    toggleCollaborator,
    clearAll,
  }
}
