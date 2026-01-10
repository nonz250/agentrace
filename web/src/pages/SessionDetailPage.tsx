import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Clock, Folder, GitBranch, MessageSquare, User, Pencil, X, Save, FolderEdit } from 'lucide-react'
import { format, formatDistanceToNow } from 'date-fns'
import { TimelineContainer } from '@/components/timeline/TimelineContainer'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ProjectSelect } from '@/components/ui/ProjectSelect'
import { FavoriteButton } from '@/components/ui/FavoriteButton'
import { useAuth } from '@/hooks/useAuth'
import * as sessionsApi from '@/api/sessions'
import { parseRepoName, getRepoUrl, isDefaultProject, getProjectDisplayName } from '@/lib/project-utils'

// Extract directory name from absolute path
function getDirectoryName(path: string): string {
  if (!path) return ''
  return path.split('/').pop() || path
}

export function SessionDetailPage() {
  const { projectId, id } = useParams<{ projectId: string; id: string }>()
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const [isEditingTitle, setIsEditingTitle] = useState(false)
  const [editTitle, setEditTitle] = useState('')
  const [isEditingProject, setIsEditingProject] = useState(false)
  const [editProjectId, setEditProjectId] = useState('')

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

  const updateProjectMutation = useMutation({
    mutationFn: (projectId: string) => sessionsApi.updateSession(id!, { project_id: projectId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['session', id] })
      queryClient.invalidateQueries({ queryKey: ['sessions', 'list'] })
      setIsEditingProject(false)
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

  const handleStartProjectEdit = () => {
    setEditProjectId(session?.project?.id || '')
    setIsEditingProject(true)
  }

  const handleCancelProjectEdit = () => {
    setIsEditingProject(false)
    setEditProjectId('')
  }

  const handleSaveProjectEdit = () => {
    updateProjectMutation.mutate(editProjectId)
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

  // Build breadcrumb items - always show project from URL
  const breadcrumbItems: BreadcrumbItem[] = [
    { label: projectDisplayName || '(no project)', href: `/projects/${projectId}` },
    { label: 'Sessions', href: `/projects/${projectId}/sessions` },
  ]
  // Session name: date
  const sessionName = format(new Date(session.started_at), 'yyyy/MM/dd HH:mm')
  breadcrumbItems.push({ label: sessionName })

  return (
    <div>
      <Breadcrumb items={breadcrumbItems} project={session.project ?? undefined} />

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
              {user && (
                <FavoriteButton
                  targetType="session"
                  targetId={session.id}
                  isFavorited={session.is_favorited}
                />
              )}
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
        {/* Metadata: project, repo, branch, path, user, events */}
        <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400">
          {/* Project selector */}
          {isEditingProject ? (
            <span className="flex items-center gap-1">
              <FolderEdit className="h-3 w-3" />
              <ProjectSelect
                value={editProjectId}
                onChange={setEditProjectId}
                disabled={updateProjectMutation.isPending}
                className="!py-0.5 !px-1 text-xs min-w-[150px]"
              />
              <Button variant="ghost" size="sm" onClick={handleCancelProjectEdit} disabled={updateProjectMutation.isPending} className="!p-0.5">
                <X className="h-3 w-3" />
              </Button>
              <Button variant="ghost" size="sm" onClick={handleSaveProjectEdit} disabled={updateProjectMutation.isPending} className="!p-0.5">
                <Save className="h-3 w-3" />
              </Button>
            </span>
          ) : (
            <span className="flex items-center gap-1 group">
              <GitBranch className="h-3 w-3" />
              {hasProject && repoUrl ? (
                <a
                  href={repoUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-gray-600 hover:underline"
                >
                  {repoName}
                </a>
              ) : hasProject && repoName ? (
                repoName
              ) : (
                <span className="text-gray-300">(no project)</span>
              )}
              {session.git_branch && <span>: {session.git_branch}</span>}
              {user && (
                <Button variant="ghost" size="sm" onClick={handleStartProjectEdit} className="!p-0.5 hidden group-hover:inline-flex">
                  <Pencil className="h-3 w-3" />
                </Button>
              )}
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
          <span className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {formatDistanceToNow(new Date(session.updated_at), { addSuffix: true })}
          </span>
        </div>
      </div>

      <TimelineContainer events={session.events || []} projectPath={session.project_path} />
    </div>
  )
}
