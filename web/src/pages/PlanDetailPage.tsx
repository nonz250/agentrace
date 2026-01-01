import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, GitBranch, Users, Clock, FileText, History } from 'lucide-react'
import { format } from 'date-fns'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { PlanEventHistory } from '@/components/plans/PlanEventHistory'
import { PlanStatusBadge } from '@/components/plans/PlanStatusBadge'
import { Spinner } from '@/components/ui/Spinner'
import * as plansApi from '@/api/plan-documents'

// Parse git remote URL to get repo name (e.g., "owner/repo")
function parseGitRepoName(remoteUrl: string): string | null {
  if (!remoteUrl) return null

  // Handle SSH URL format: ssh://git@github.com/owner/repo.git
  let match = remoteUrl.match(/ssh:\/\/git@[^/]+\/(.+?)(?:\.git)?$/)
  if (match) return match[1]

  // Handle SSH format: git@github.com:owner/repo.git
  match = remoteUrl.match(/git@[^:]+:(.+?)(?:\.git)?$/)
  if (match) return match[1]

  // Handle HTTPS format: https://github.com/owner/repo.git
  match = remoteUrl.match(/https?:\/\/[^/]+\/(.+?)(?:\.git)?$/)
  if (match) return match[1]

  return null
}

// Get GitHub/GitLab URL from remote URL
function getRepoUrl(remoteUrl: string): string | null {
  const repoName = parseGitRepoName(remoteUrl)
  if (!repoName) return null

  if (remoteUrl.includes('github.com')) {
    return `https://github.com/${repoName}`
  }
  if (remoteUrl.includes('gitlab.com')) {
    return `https://gitlab.com/${repoName}`
  }

  return null
}

type TabType = 'content' | 'history'

export function PlanDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<TabType>('content')

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

  const repoName = parseGitRepoName(plan.git_remote_url)
  const repoUrl = getRepoUrl(plan.git_remote_url)
  const collaboratorNames = plan.collaborators.map((c) => c.display_name).join(', ')
  const formattedDate = format(new Date(plan.updated_at), 'yyyy/MM/dd HH:mm')

  return (
    <div>
      <button
        onClick={() => navigate(-1)}
        className="mb-6 inline-flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
      >
        <ArrowLeft className="h-4 w-4" />
        Back
      </button>

      <div className="mb-6">
        {/* Title: Description + Status */}
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-medium text-gray-900">{plan.description}</h1>
          <PlanStatusBadge status={plan.status} />
        </div>
        {/* Metadata: repo, collaborators, updated_at */}
        <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400">
          {repoName && (
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
          <div className="prose prose-sm max-w-none prose-headings:text-gray-900 prose-p:text-gray-700 prose-a:text-blue-600 prose-code:rounded prose-code:bg-gray-100 prose-code:px-1 prose-code:py-0.5 prose-code:text-gray-800 prose-pre:bg-gray-900 prose-pre:text-gray-100">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{plan.body}</ReactMarkdown>
          </div>
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
