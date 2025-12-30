import { Card } from '@/components/ui/Card'
import { Folder, User, Clock, GitBranch } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import type { Session } from '@/types/session'

interface SessionCardProps {
  session: Session
  onClick: () => void
}

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

export function SessionCard({ session, onClick }: SessionCardProps) {
  const repoName = parseGitRepoName(session.git_remote_url)
  const repoUrl = getRepoUrl(session.git_remote_url)

  return (
    <Card hover onClick={onClick}>
      <div className="flex items-start gap-3">
        <Folder className="mt-0.5 h-5 w-5 flex-shrink-0 text-gray-400" />
        <div className="min-w-0 flex-1">
          <p className="truncate font-mono text-sm text-gray-900">
            {getDirectoryName(session.project_path)}
          </p>
          <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-500">
            {session.git_remote_url && (
              <span className="flex items-center gap-1">
                <GitBranch className="h-4 w-4" />
                {repoUrl && repoName ? (
                  <a
                    href={repoUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:underline"
                    onClick={(e) => e.stopPropagation()}
                  >
                    {repoName}
                  </a>
                ) : (
                  repoName || session.git_remote_url
                )}
                {session.git_branch && (
                  <span className="text-gray-400">({session.git_branch})</span>
                )}
              </span>
            )}
            <span className="flex items-center gap-1">
              <User className="h-4 w-4" />
              {session.user_name || 'Unknown'}
            </span>
            <span className="flex items-center gap-1">
              <Clock className="h-4 w-4" />
              {formatDistanceToNow(new Date(session.started_at), {
                addSuffix: true,
              })}
            </span>
            <span>{session.event_count} events</span>
          </div>
        </div>
      </div>
    </Card>
  )
}
