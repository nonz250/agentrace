import { useState, useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { MultiSelect } from '@/components/ui/MultiSelect'
import { useSortPreference } from '@/hooks/useSortPreference'
import { useSessionCreatorFilter } from '@/hooks/useSessionCreatorFilter'
import * as sessionsApi from '@/api/sessions'
import * as projectsApi from '@/api/projects'
import { getProjectDisplayName } from '@/lib/project-utils'

const PAGE_SIZE = 20

interface Creator {
  id: string
  display_name: string
}

export function SessionsPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const [page, setPage] = useState(1)
  const [cursors, setCursors] = useState<string[]>(['']) // cursors[0] = '' for first page
  const { sort, updateSort } = useSortPreference('sessions')
  const { selectedCreatorIds, setCreatorIds } = useSessionCreatorFilter()

  const resetPagination = useCallback(() => {
    setPage(1)
    setCursors([''])
  }, [])

  const handleSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    updateSort(e.target.value as 'updated_at' | 'created_at')
    resetPagination()
  }

  const handleCreatorChange = (creatorIds: string[]) => {
    setCreatorIds(creatorIds)
    resetPagination()
  }

  const { data: projectData } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getProject(projectId!),
    enabled: !!projectId,
  })

  // Query to get all creators (without creator filter)
  const { data: allSessionsData } = useQuery({
    queryKey: ['sessions', 'all-creators', projectId],
    queryFn: () =>
      sessionsApi.getSessions({
        projectId: projectId || undefined,
        limit: 100, // Get enough sessions to collect creators
      }),
  })

  // Collect unique creators from all sessions
  const allCreators = useMemo(() => {
    const creatorMap = new Map<string, Creator>()
    for (const session of allSessionsData?.sessions || []) {
      if (session.user_id && session.user_name) {
        if (!creatorMap.has(session.user_id)) {
          creatorMap.set(session.user_id, {
            id: session.user_id,
            display_name: session.user_name,
          })
        }
      }
    }
    return Array.from(creatorMap.values()).sort((a, b) =>
      a.display_name.localeCompare(b.display_name)
    )
  }, [allSessionsData])

  const cursor = cursors[page - 1] || ''

  const { data, isLoading, isFetching, error } = useQuery({
    queryKey: ['sessions', 'list', page, projectId, selectedCreatorIds, sort, cursor],
    queryFn: () => sessionsApi.getSessions({
      projectId: projectId || undefined,
      userIds: selectedCreatorIds.length > 0 ? selectedCreatorIds : undefined,
      limit: PAGE_SIZE,
      cursor: cursor || undefined,
      sort,
    }),
    placeholderData: (previousData) => previousData, // Keep previous data while fetching
  })

  const sessions = data?.sessions || []
  const nextCursor = data?.next_cursor
  const hasMore = !!nextCursor

  // Store next cursor when we get it
  const goToNextPage = useCallback(() => {
    if (nextCursor) {
      setCursors(prev => {
        const newCursors = [...prev]
        newCursors[page] = nextCursor
        return newCursors
      })
      setPage(p => p + 1)
    }
  }, [nextCursor, page])

  const goToPrevPage = useCallback(() => {
    setPage(p => Math.max(1, p - 1))
  }, [])

  // Only show full-page loading on initial load (no data yet)
  const showInitialLoading = isLoading && !data

  const projectDisplayName = projectData ? getProjectDisplayName(projectData) : null

  // Build breadcrumb items
  const breadcrumbItems: BreadcrumbItem[] = []
  if (projectId && projectDisplayName) {
    breadcrumbItems.push({ label: projectDisplayName, href: `/projects/${projectId}` })
  }
  breadcrumbItems.push({ label: 'Sessions' })

  return (
    <div>
      <Breadcrumb items={breadcrumbItems} project={projectData} />

      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-gray-900">Sessions</h1>
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-4">
        {allCreators.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500">Creator:</span>
            <MultiSelect
              options={allCreators.map((c) => ({ value: c.id, label: c.display_name }))}
              selectedValues={selectedCreatorIds}
              onChange={handleCreatorChange}
              placeholder="All creators"
            />
          </div>
        )}

        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500">Sort:</span>
          <select
            value={sort}
            onChange={handleSortChange}
            className="rounded-lg bg-transparent px-2 py-1 text-sm text-gray-600 hover:bg-gray-100 focus:outline-none"
          >
            <option value="updated_at">Updated</option>
            <option value="created_at">Created</option>
          </select>
        </div>
      </div>

      {showInitialLoading ? (
        <div className="flex justify-center py-12">
          <Spinner size="lg" />
        </div>
      ) : error ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
          Failed to load sessions: {error.message}
        </div>
      ) : (
        <div className={isFetching ? 'opacity-50 transition-opacity' : ''}>
          <SessionList sessions={sessions} />
        </div>
      )}

      {(page > 1 || hasMore) && (
        <div className="mt-6 flex items-center justify-between">
          <Button
            variant="secondary"
            size="sm"
            onClick={goToPrevPage}
            disabled={page === 1}
          >
            <ChevronLeft className="mr-1 h-4 w-4" />
            Previous
          </Button>
          <span className="text-sm text-gray-500">Page {page}</span>
          <Button
            variant="secondary"
            size="sm"
            onClick={goToNextPage}
            disabled={!hasMore}
          >
            Next
            <ChevronRight className="ml-1 h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}
