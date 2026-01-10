import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { SessionList } from '@/components/sessions/SessionList'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { useSortPreference } from '@/hooks/useSortPreference'
import * as sessionsApi from '@/api/sessions'
import * as projectsApi from '@/api/projects'
import { getProjectDisplayName } from '@/lib/project-utils'

const PAGE_SIZE = 20

export function SessionsPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const [page, setPage] = useState(1)
  const { sort, updateSort } = useSortPreference('sessions')
  const offset = (page - 1) * PAGE_SIZE

  const handleSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    updateSort(e.target.value as 'updated_at' | 'created_at')
    setPage(1) // Reset to first page when sort changes
  }

  const { data: projectData } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectsApi.getProject(projectId!),
    enabled: !!projectId,
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['sessions', 'list', page, projectId, sort],
    queryFn: () => sessionsApi.getSessions({ projectId: projectId || undefined, limit: PAGE_SIZE, offset, sort }),
  })

  const sessions = data?.sessions || []
  const hasMore = sessions.length === PAGE_SIZE

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
        Failed to load sessions: {error.message}
      </div>
    )
  }

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
      <SessionList sessions={sessions} />

      {(page > 1 || hasMore) && (
        <div className="mt-6 flex items-center justify-between">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="mr-1 h-4 w-4" />
            Previous
          </Button>
          <span className="text-sm text-gray-500">Page {page}</span>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => p + 1)}
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
