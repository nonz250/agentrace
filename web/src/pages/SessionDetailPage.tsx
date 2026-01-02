import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Folder, GitBranch, MessageSquare, User, Pencil, X, Save } from 'lucide-react'
import { format } from 'date-fns'
import { TimelineContainer } from '@/components/timeline/TimelineContainer'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuth } from '@/hooks/useAuth'
import * as sessionsApi from '@/api/sessions'
import { parseRepoName, getRepoUrl, isDefaultProject, getProjectDisplayName } from '@/lib/project-utils'

// Extract directory name from absolute path
function getDirectoryName(path: string): string {
  if (!path) return ''
  return path.split('/').pop() || path
}

export function SessionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const [isEditingTitle, setIsEditingTitle] = useState(false)
  const [editTitle, setEditTitle] = useState('')

  const { data: session, isLoading, error } = useQuery({
    queryKey: ['session', id],
    queryFn: () => sessionsApi.getSession(id!),
    enabled: !!id,
  })

  const updateMutation = useMutation({
    mutationFn: (title: string) => sessionsApi.updateSessionTitle(id!, title),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['session', id] })
      queryClient.invalidateQueries({ queryKey: ['sessions', 'list'] })
      setIsEditingTitle(false)
    },
  })

  const handleStartEdit = () => {
    setEditTitle(session?.title || '')
    setIsEditingTitle(true)
  }

  const handleCancelEdit = () => {
    setIsEditingTitle(false)
    setEditTitle('')
  }

  const handleSaveEdit = () => {
    updateMutation.mutate(editTitle)
  }

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
        Failed to load session: {error.message}
      </div>
    )
  }

  if (!session) {
    return (
      <div className="rounded-xl border border-gray-200 bg-white p-8 text-center">
        <p className="text-gray-500">Session not found.</p>
      </div>
    )
  }

  const repoName = parseRepoName(session.project)
  const repoUrl = getRepoUrl(session.project)
  const hasProject = !isDefaultProject(session.project)
  const projectDisplayName = getProjectDisplayName(session.project)

  // Build breadcrumb items
  const breadcrumbItems: BreadcrumbItem[] = []
  if (hasProject && session.project) {
    breadcrumbItems.push({ label: projectDisplayName || '(no project)', href: `/projects/${session.project.id}` })
    breadcrumbItems.push({ label: 'Sessions', href: `/sessions?project_id=${session.project.id}` })
  } else {
    breadcrumbItems.push({ label: 'Sessions', href: '/sessions' })
  }
  // Session name: date
  const sessionName = format(new Date(session.started_at), 'yyyy/MM/dd HH:mm')
  breadcrumbItems.push({ label: sessionName })

  return (
    <div>
      <Breadcrumb items={breadcrumbItems} />

      <div className="mb-6">
        {/* Title: Date + Title */}
        <div className="flex items-center gap-3">
          {isEditingTitle ? (
            <div className="flex flex-1 items-center gap-2">
              <span className="text-lg font-medium text-gray-900">
                {format(new Date(session.started_at), 'yyyy/MM/dd HH:mm')}
              </span>
              <Input
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                placeholder="Session title"
                className="flex-1 min-w-[400px]"
              />
              <Button variant="ghost" size="sm" onClick={handleCancelEdit} disabled={updateMutation.isPending}>
                <X className="h-4 w-4" />
              </Button>
              <Button size="sm" onClick={handleSaveEdit} loading={updateMutation.isPending}>
                <Save className="h-4 w-4" />
              </Button>
            </div>
          ) : (
            <>
              <h1 className="text-lg font-medium text-gray-900">
                {format(new Date(session.started_at), 'yyyy/MM/dd HH:mm')}
                {session.title && <span className="ml-2">{session.title}</span>}
              </h1>
              {user && (
                <Button variant="ghost" size="sm" onClick={handleStartEdit}>
                  <Pencil className="h-4 w-4" />
                </Button>
              )}
            </>
          )}
        </div>
        {/* Metadata: repo, branch, path, user, events */}
        <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400">
          {hasProject && repoName && (
            <span className="flex items-center gap-1">
              <GitBranch className="h-3 w-3" />
              {repoUrl ? (
                <a
                  href={repoUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-gray-600 hover:underline"
                >
                  {repoName}
                </a>
              ) : (
                repoName
              )}
              {session.git_branch && <span>: {session.git_branch}</span>}
            </span>
          )}
          {!hasProject && session.project_path && (
            <span className="flex items-center gap-1" title={session.project_path}>
              <Folder className="h-3 w-3 flex-shrink-0" />
              <span className="font-mono">{getDirectoryName(session.project_path)}</span>
            </span>
          )}
          {session.user_name && (
            <span className="flex items-center gap-1">
              <User className="h-3 w-3" />
              {session.user_name}
            </span>
          )}
          <span className="flex items-center gap-1">
            <MessageSquare className="h-3 w-3" />
            {session.events?.length || 0}
          </span>
        </div>
      </div>

      <h2 className="mb-4 text-lg font-semibold text-gray-900">Timeline</h2>
      <TimelineContainer events={session.events || []} projectPath={session.project_path} />
    </div>
  )
}
