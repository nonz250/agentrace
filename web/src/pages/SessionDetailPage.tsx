import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, Folder, GitBranch, MessageSquare } from 'lucide-react'
import { format } from 'date-fns'
import { TimelineContainer } from '@/components/timeline/TimelineContainer'
import { Spinner } from '@/components/ui/Spinner'
import * as sessionsApi from '@/api/sessions'

// Extract directory name from absolute path
function getDirectoryName(path: string): string {
  if (!path) return ''
  return path.split('/').pop() || path
}

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
function getGitHubUrl(remoteUrl: string): string | null {
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

export function SessionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: session, isLoading, error } = useQuery({
    queryKey: ['session', id],
    queryFn: () => sessionsApi.getSession(id!),
    enabled: !!id,
  })

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
        {/* Title: Date and User */}
        <h1 className="text-lg font-medium text-gray-900">
          {format(new Date(session.started_at), 'yyyy/MM/dd HH:mm')}
          <span className="ml-2 text-gray-600">{session.user_name || 'Unknown'}</span>
        </h1>
        {/* Metadata: repo, branch, path, events */}
        <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400">
          {(() => {
            const repoName = parseGitRepoName(session.git_remote_url)
            const repoUrl = getGitHubUrl(session.git_remote_url)
            if (repoName) {
              return (
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
              )
            }
            return null
          })()}
          {!parseGitRepoName(session.git_remote_url) && session.project_path && (
            <span className="flex items-center gap-1" title={session.project_path}>
              <Folder className="h-3 w-3 flex-shrink-0" />
              <span className="font-mono">{getDirectoryName(session.project_path)}</span>
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
