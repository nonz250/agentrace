import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { GitBranch, Users, Clock, FileText, History, Pencil, X, Save, Copy, Check, FolderEdit } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { PlanEventHistory } from '@/components/plans/PlanEventHistory'
import { PlanContentWithComments } from '@/components/plans/PlanContentWithComments'
import { PlanStatusBadge } from '@/components/plans/PlanStatusBadge'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { Select } from '@/components/ui/Select'
import { ProjectSelect } from '@/components/ui/ProjectSelect'
import { FavoriteButton } from '@/components/ui/FavoriteButton'
import { useAuth } from '@/hooks/useAuth'
import * as plansApi from '@/api/plan-documents'
import * as commentsApi from '@/api/plan-comments'
import type { PlanDocumentStatus } from '@/types/plan-document'
import { parseRepoName, getRepoUrl, isDefaultProject, getProjectDisplayName } from '@/lib/project-utils'
import { statusConfig } from '@/lib/plan-status'

type TabType = 'content' | 'history'

export function PlanDetailPage() {
  const { projectId, id } = useParams<{ projectId: string; id: string }>()
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState<TabType>('content')
  const [isEditing, setIsEditing] = useState(false)
  const [editDescription, setEditDescription] = useState('')
  const [editBody, setEditBody] = useState('')
  const [copied, setCopied] = useState(false)
  const [isEditingProject, setIsEditingProject] = useState(false)
  const [editProjectId, setEditProjectId] = useState('')
  const [isEditingStatus, setIsEditingStatus] = useState(false)
  const [editStatus, setEditStatus] = useState<PlanDocumentStatus>('scratch')

  const { data: plan, isLoading: isPlanLoading, error: planError } = useQuery({
    queryKey: ['plan', id],
    queryFn: () => plansApi.getPlan(id!),
    enabled: !!id,
  })

  const { data: eventsData, isLoading: isEventsLoading } = useQuery({
    queryKey: ['plan', id, 'events'],
    queryFn: () => plansApi.getPlanEvents(id!),
    enabled: !!id && activeTab === 'history',
  })

  const { data: threadsData } = useQuery({
    queryKey: ['plan', id, 'comments'],
    queryFn: () => commentsApi.getPlanThreads(id!),
    enabled: !!id,
  })

  const statusMutation = useMutation({
    mutationFn: (status: PlanDocumentStatus) => plansApi.setPlanStatus(id!, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', id] })
      queryClient.invalidateQueries({ queryKey: ['plan', id, 'events'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: () => plansApi.updatePlan(id!, {
      description: editDescription,
      body: editBody,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', id] })
      queryClient.invalidateQueries({ queryKey: ['plan', id, 'events'] })
      setIsEditing(false)
    },
  })

  const updateProjectMutation = useMutation({
    mutationFn: (projectId: string) => plansApi.updatePlan(id!, { project_id: projectId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', id] })
      queryClient.invalidateQueries({ queryKey: ['plans', 'list'] })
      setIsEditingProject(false)
    },
  })

  const handleStartEdit = () => {
    if (plan) {
      setEditDescription(plan.description)
      setEditBody(plan.body)
      setIsEditing(true)
    }
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
    setEditDescription('')
    setEditBody('')
  }

  const handleSaveEdit = () => {
    if (editDescription.trim()) {
      updateMutation.mutate()
    }
  }

  const handleStartStatusEdit = () => {
    if (plan) {
      setEditStatus(plan.status)
      setIsEditingStatus(true)
    }
  }

  const handleStatusChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newStatus = e.target.value as PlanDocumentStatus
    if (plan && newStatus !== plan.status) {
      statusMutation.mutate(newStatus, {
        onSuccess: () => setIsEditingStatus(false),
      })
    } else {
      setIsEditingStatus(false)
    }
  }

  const handleStatusBlur = () => {
    if (!statusMutation.isPending) {
      setIsEditingStatus(false)
    }
  }

  const handleCopyId = async () => {
    if (!plan) return
    const text = `${plan.description}\nplan document: ${plan.id}`
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleStartProjectEdit = () => {
    setEditProjectId(plan?.project?.id || '')
    setIsEditingProject(true)
  }

  const handleCancelProjectEdit = () => {
    setIsEditingProject(false)
    setEditProjectId('')
  }

  const handleSaveProjectEdit = () => {
    updateProjectMutation.mutate(editProjectId)
  }

  if (isPlanLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  if (planError) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
        Failed to load plan: {planError.message}
      </div>
    )
  }

  if (!plan) {
    return (
      <div className="rounded-xl border border-gray-200 bg-white p-8 text-center">
        <p className="text-gray-500">Plan not found.</p>
      </div>
    )
  }

  const repoName = parseRepoName(plan.project)
  const repoUrl = getRepoUrl(plan.project)
  const hasProject = !isDefaultProject(plan.project)
  const collaboratorNames = plan.collaborators.map((c) => c.display_name).join(', ')
  const relativeTime = formatDistanceToNow(new Date(plan.updated_at), { addSuffix: true })
  const projectDisplayName = getProjectDisplayName(plan.project)

  // Build breadcrumb items - always show project from URL
  const breadcrumbItems: BreadcrumbItem[] = [
    { label: projectDisplayName || '(no project)', href: `/projects/${projectId}` },
    { label: 'Plans', href: `/projects/${projectId}/plans` },
  ]
  // Plan ID (shortened) with copy button
  const shortId = plan.id.slice(0, 8) + '...'
  breadcrumbItems.push({ label: shortId, copyText: plan.id })

  return (
    <div>
      <Breadcrumb items={breadcrumbItems} project={plan.project ?? undefined} />

      <div className="mb-6">
        {/* Title: Description + Status + Actions */}
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            {user && (
              <FavoriteButton
                targetType="plan"
                targetId={plan.id}
                isFavorited={plan.is_favorited}
              />
            )}
            <h1 className="text-lg font-medium text-gray-900">{plan.description}</h1>
            {isEditingStatus ? (
              <Select
                value={editStatus}
                onChange={handleStatusChange}
                onBlur={handleStatusBlur}
                disabled={statusMutation.isPending}
                className="!py-1 !px-2 text-xs min-w-[130px]"
                autoFocus
              >
                {Object.entries(statusConfig).map(([status, config]) => (
                  <option key={status} value={status}>
                    {config.label}
                  </option>
                ))}
              </Select>
            ) : (
              <span
                className={user ? "cursor-pointer hover:opacity-80" : ""}
                onClick={user ? handleStartStatusEdit : undefined}
              >
                <PlanStatusBadge status={plan.status} />
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={handleCopyId} title="Copy plan ID for AI agents">
              {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
            </Button>
            {user && !isEditing && (
              <Button variant="secondary" size="sm" onClick={handleStartEdit}>
                <Pencil className="mr-1 h-4 w-4" />
                Edit
              </Button>
            )}
          </div>
        </div>
        {/* Metadata: project, collaborators, updated_at */}
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
              {user && (
                <Button variant="ghost" size="sm" onClick={handleStartProjectEdit} className="!p-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Pencil className="h-3 w-3" />
                </Button>
              )}
            </span>
          )}
          {collaboratorNames && (
            <span className="flex items-center gap-1">
              <Users className="h-3 w-3" />
              {collaboratorNames}
            </span>
          )}
          <span className="flex items-center gap-1">
            <Clock className="h-3 w-3" />
            {relativeTime}
          </span>
        </div>
      </div>

      {/* Tabs */}
      <div className="mb-4 border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('content')}
            className={`flex items-center gap-2 border-b-2 px-1 py-2 text-sm font-medium ${
              activeTab === 'content'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
            }`}
          >
            <FileText className="h-4 w-4" />
            Content
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`flex items-center gap-2 border-b-2 px-1 py-2 text-sm font-medium ${
              activeTab === 'history'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
            }`}
          >
            <History className="h-4 w-4" />
            History
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'content' && (
        <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          {isEditing ? (
            <div className="space-y-4">
              <Input
                label="Description"
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                placeholder="Brief description of the plan"
              />
              <Textarea
                label="Body"
                value={editBody}
                onChange={(e) => setEditBody(e.target.value)}
                placeholder="Plan details in Markdown format"
                rows={15}
              />
              <div className="flex justify-end gap-3 pt-2">
                <Button variant="ghost" onClick={handleCancelEdit} disabled={updateMutation.isPending}>
                  <X className="mr-1 h-4 w-4" />
                  Cancel
                </Button>
                <Button onClick={handleSaveEdit} loading={updateMutation.isPending}>
                  <Save className="mr-1 h-4 w-4" />
                  Save
                </Button>
              </div>
            </div>
          ) : (
            <PlanContentWithComments
              planId={id!}
              body={plan.body}
              threads={threadsData?.threads || []}
            />
          )}
        </div>
      )}

      {activeTab === 'history' && (
        <>
          {isEventsLoading ? (
            <div className="flex justify-center py-12">
              <Spinner size="lg" />
            </div>
          ) : (
            <PlanEventHistory events={eventsData?.events || []} />
          )}
        </>
      )}
    </div>
  )
}
