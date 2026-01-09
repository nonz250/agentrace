import { useState, useEffect, useCallback } from 'react'
import type { PlanDocumentStatus } from '@/types/plan-document'
import { statusConfig } from '@/lib/plan-status'

const STORAGE_KEY = 'agentrace:plan-status-filter'

function loadFromStorage(): PlanDocumentStatus[] | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === null) {
      return null
    }
    const parsed = JSON.parse(stored)
    if (Array.isArray(parsed)) {
      return parsed.filter((s): s is PlanDocumentStatus => s in statusConfig)
    }
    return null
  } catch {
    return null
  }
}

function saveToStorage(statuses: PlanDocumentStatus[]): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(statuses))
  } catch {
    // localStorage が使えない環境では無視
  }
}

export function usePlanStatusFilter() {
  const [selectedStatuses, setSelectedStatuses] = useState<PlanDocumentStatus[]>(() => {
    const stored = loadFromStorage()
    // localStorageに値がない場合は空（フィルターなし = 全件表示）
    return stored ?? []
  })

  useEffect(() => {
    saveToStorage(selectedStatuses)
  }, [selectedStatuses])

  const setStatuses = useCallback((statuses: PlanDocumentStatus[]) => {
    setSelectedStatuses(statuses)
  }, [])

  const toggleStatus = useCallback((status: PlanDocumentStatus) => {
    setSelectedStatuses((prev) => {
      if (prev.includes(status)) {
        return prev.filter((s) => s !== status)
      } else {
        return [...prev, status]
      }
    })
  }, [])

  const selectAll = useCallback(() => {
    setSelectedStatuses(Object.keys(statusConfig) as PlanDocumentStatus[])
  }, [])

  const clearAll = useCallback(() => {
    setSelectedStatuses([])
  }, [])

  return {
    selectedStatuses,
    setStatuses,
    toggleStatus,
    selectAll,
    clearAll,
  }
}
