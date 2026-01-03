import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { GitBranch, Users, Clock, FileText, History, Pencil, X, Save, Copy, Check, FolderEdit } from 'lucide-react'
import { format } from 'date-fns'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { PlanEventHistory } from '@/components/plans/PlanEventHistory'
import { PlanStatusBadge } from '@/components/plans/PlanStatusBadge'
import { Breadcrumb, type BreadcrumbItem } from '@/components/ui/Breadcrumb'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { Select } from '@/components/ui/Select'
import { ProjectSelect } from '@/components/ui/ProjectSelect'
import { useAuth } from '@/hooks/useAuth'
import * as plansApi from '@/api/plan-documents'
import type { PlanDocumentStatus } from '@/types/plan-document'
import { parseRepoName, getRepoUrl, isDefaultProject, getProjectDisplayName } from '@/lib/project-utils'

type TabType = 'content' | 'history'

const STATUS_OPTIONS: { value: PlanDocumentStatus; label: string }[] = [
  { value: 'scratch', label: 'Scratch' },
  { value: 'draft', label: 'Draft' },
  { value: 'planning', label: 'Planning' },
  { value: 'pending', label: 'Pending' },
  { value: 'implementation', label: 'Implementation' },
  { value: 'complete', label: 'Complete' },
]

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

  const handleCancelStatusEdit = () => {
    setIsEditingStatus(false)
  }

  const handleSaveStatusEdit = () => {
    if (plan && editStatus !== plan.status) {
      statusMutation.mutate(editStatus, {
        onSuccess: () => setIsEditingStatus(false),
      })
    } else {
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
  const formattedDate = format(new Date(plan.updated_at), 'yyyy/MM/dd HH:mm')
  const projectDisplayName = getProjectDisplayName(plan.project)

  // Build breadcrumb items - always show project from URL
  const breadcrumbItems: BreadcrumbItem[] = [
    { label: projectDisplayName || '(no project)', href: `/projects/${projectId}` },
    { label: 'Plans', href: `/projects/${projectId}/plans` },
  ]
  // Plan name: description truncated
  const planName = plan.description.length > 30 ? plan.description.slice(0, 30) + '...' : plan.description
  breadcrumbItems.push({ label: planName })

  return (
    <div>
      <Breadcrumb items={breadcrumbItems} project={plan.project ?? undefined} />

      <div className="mb-6">
        {/* Title: Description + Status + Actions */}
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <h1 className="text-lg font-medium text-gray-900">{plan.description}</h1>
            {isEditingStatus ? (
              <span className="flex items-center gap-1">
                <Select
                  value={editStatus}
                  onChange={(e) => setEditStatus(e.target.value as PlanDocumentStatus)}
                  disabled={statusMutation.isPending}
                  className="!py-1 !px-2 text-xs min-w-[130px]"
                >
                  {STATUS_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </Select>
                <Button variant="ghost" size="sm" onClick={handleCancelStatusEdit} disabled={statusMutation.isPending} className="!p-1">
                  <X className="h-3 w-3" />
                </Button>
                <Button variant="ghost" size="sm" onClick={handleSaveStatusEdit} disabled={statusMutation.isPending} className="!p-1">
                  <Save className="h-3 w-3" />
                </Button>
              </span>
            ) : (
              <span className="flex items-center gap-1 group">
                <PlanStatusBadge status={plan.status} />
                {user && (
                  <Button variant="ghost" size="sm" onClick={handleStartStatusEdit} className="!p-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Pencil className="h-3 w-3" />
                  </Button>
                )}
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
            {formattedDate}
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
            <div className="prose prose-sm max-w-none prose-headings:text-gray-900 prose-p:text-gray-700 prose-a:text-blue-600 prose-code:rounded prose-code:bg-gray-100 prose-code:px-1 prose-code:py-0.5 prose-code:text-gray-800 prose-pre:bg-gray-900 prose-pre:text-gray-100">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{plan.body}</ReactMarkdown>
            </div>
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
