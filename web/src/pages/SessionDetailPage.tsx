import { useState, useRef, useEffect, useMemo, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Clock, Folder, GitBranch, MessageSquare, User, Pencil, X, Save, FolderEdit, MoreVertical, Trash2 } from 'lucide-react'
import { format, formatDistanceToNow } from 'date-fns'
import { ClaudeCodeTranscript, filterHiddenEvents, expandEvents, extractMessageBlocks } from 'cc-transcript-react'
import type { TranscriptEvent } from 'cc-transcript-react'
import 'cc-transcript-react/styles.css'
import { buildAgentraceBlockRenderers } from '@/components/timeline/agentrace-block-renderers'
import { MessageNav } from '@/components/timeline/MessageNav'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
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
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const [isEditingTitle, setIsEditingTitle] = useState(false)
  const [editTitle, setEditTitle] = useState('')
  const [isEditingProject, setIsEditingProject] = useState(false)
  const [editProjectId, setEditProjectId] = useState('')
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  // Close menu when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

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

  const deleteMutation = useMutation({
    mutationFn: () => sessionsApi.deleteSession(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions', 'list'] })
      navigate(`/projects/${projectId}/sessions`)
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

  const handleDeleteClick = () => {
    setIsMenuOpen(false)
    setIsDeleteDialogOpen(true)
  }

  const handleConfirmDelete = () => {
    deleteMutation.mutate()
  }

  // Check if current user can delete this session
  const canDelete = user && session?.user_id === user.id

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
  // Session ID (shortened) with copy button
  const shortId = session.id.slice(0, 8) + '...'
  breadcrumbItems.push({ label: shortId, copyText: session.id })

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
              {canDelete && (
                <div className="relative ml-auto" ref={menuRef}>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setIsMenuOpen(!isMenuOpen)}
                  >
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                  {isMenuOpen && (
                    <div className="absolute right-0 mt-1 w-40 rounded-lg border border-gray-200 bg-white shadow-lg z-10">
                      <button
                        className="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                        onClick={handleDeleteClick}
                      >
                        <Trash2 className="h-4 w-4" />
                        Delete
                      </button>
                    </div>
                  )}
                </div>
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

      <TranscriptTimeline events={session.events || []} projectPath={session.project_path} />

      {/* Delete confirmation dialog */}
      <Modal
        open={isDeleteDialogOpen}
        onClose={() => setIsDeleteDialogOpen(false)}
        title="Delete Session"
      >
        <div className="space-y-4">
          <p className="text-gray-600">
            Are you sure you want to delete this session? This action cannot be undone.
          </p>
          {deleteMutation.isError && (
            <p className="text-sm text-red-600">
              Failed to delete session. Please try again.
            </p>
          )}
          <div className="flex justify-end gap-3">
            <Button
              variant="ghost"
              onClick={() => setIsDeleteDialogOpen(false)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleConfirmDelete}
              loading={deleteMutation.isPending}
              className="bg-red-600 hover:bg-red-700 text-white"
            >
              Delete
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}

// Wrapper to convert agentrace Event[] to TranscriptEvent[] and render via library with sidebar navigation
function TranscriptTimeline({ events, projectPath }: { events: import('@/types/event').Event[]; projectPath?: string }) {
  const [activeBlockId, setActiveBlockId] = useState<string | null>(null)
  const isInitialScrollDoneRef = useRef(false)
  const hashUpdateTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const transcriptEvents: TranscriptEvent[] = useMemo(
    () => events.map((e) => ({
      key: e.id,
      uuid: e.uuid,
      event_type: e.event_type,
      payload: e.payload,
      created_at: e.created_at,
    })),
    [events]
  )

  const customBlockRenderers = useMemo(() => buildAgentraceBlockRenderers(), [])

  // Compute message blocks for sidebar navigation
  const messageBlocks = useMemo(() => {
    const filtered = filterHiddenEvents(transcriptEvents)
    const displayBlocks = expandEvents(filtered, projectPath)
    return extractMessageBlocks(displayBlocks)
  }, [transcriptEvents, projectPath])

  // Scroll to hash on initial load and hash change
  useEffect(() => {
    const scrollToHash = () => {
      const hash = window.location.hash
      if (hash && hash.startsWith('#event-')) {
        setTimeout(() => {
          const element = document.getElementById(hash.slice(1))
          if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' })
          }
          isInitialScrollDoneRef.current = true
        }, 100)
      } else {
        isInitialScrollDoneRef.current = true
      }
    }

    scrollToHash()
    window.addEventListener('hashchange', scrollToHash)
    return () => window.removeEventListener('hashchange', scrollToHash)
  }, [events])

  // Update URL hash when active block changes (debounced)
  useEffect(() => {
    if (!isInitialScrollDoneRef.current || !activeBlockId) return

    if (hashUpdateTimerRef.current) {
      clearTimeout(hashUpdateTimerRef.current)
    }

    hashUpdateTimerRef.current = setTimeout(() => {
      const newHash = `#event-${activeBlockId}`
      if (window.location.hash !== newHash) {
        window.history.replaceState(null, '', newHash)
      }
    }, 300)

    return () => {
      if (hashUpdateTimerRef.current) {
        clearTimeout(hashUpdateTimerRef.current)
      }
    }
  }, [activeBlockId])

  // IntersectionObserver to track which message block is currently visible
  useEffect(() => {
    const messageBlockIds = messageBlocks.map(b => b.id)
    if (messageBlockIds.length === 0) return

    const observer = new IntersectionObserver(
      (entries) => {
        const intersecting = entries.find((entry) => entry.isIntersecting)
        if (intersecting) {
          const blockId = intersecting.target.getAttribute('data-block-id')
          if (blockId) {
            setActiveBlockId(blockId)
          }
        }
      },
      {
        rootMargin: '-100px 0px -70% 0px',
        threshold: 0,
      }
    )

    // Observe elements rendered by ClaudeCodeTranscript (they have id="event-{blockId}")
    for (const blockId of messageBlockIds) {
      const element = document.getElementById(`event-${blockId}`)
      if (element) {
        element.setAttribute('data-block-id', blockId)
        observer.observe(element)
      }
    }

    return () => observer.disconnect()
  }, [messageBlocks])

  const handleNavigate = useCallback((blockId: string) => {
    const element = document.getElementById(`event-${blockId}`)
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }, [])

  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No events yet.</p>
      </div>
    )
  }

  return (
    <div className="flex gap-6">
      {/* Left sidebar - hidden on small screens */}
      <aside className="hidden w-48 flex-shrink-0 lg:block">
        <MessageNav
          messageBlocks={messageBlocks}
          activeBlockId={activeBlockId}
          onNavigate={handleNavigate}
        />
      </aside>

      {/* Main timeline */}
      <main className="min-w-0 flex-1">
        <ClaudeCodeTranscript
          events={transcriptEvents}
          projectPath={projectPath}
          colorScheme="light"
          customBlockRenderers={customBlockRenderers}
        />
      </main>
    </div>
  )
}
