import { Card } from '@/components/ui/Card'
import { GitBranch, Users, Clock } from 'lucide-react'
import { format } from 'date-fns'
import type { PlanDocument } from '@/types/plan-document'

interface PlanCardProps {
  plan: PlanDocument
  onClick: () => void
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

export function PlanCard({ plan, onClick }: PlanCardProps) {
  const repoName = parseGitRepoName(plan.git_remote_url)
  const repoUrl = getRepoUrl(plan.git_remote_url)
  const formattedDate = format(new Date(plan.updated_at), 'yyyy/MM/dd HH:mm')
  const collaboratorNames = plan.collaborators.map((c) => c.display_name).join(', ')

  return (
    <Card hover onClick={onClick}>
      <div className="min-w-0">
        {/* Title: Description */}
        <p className="text-sm font-medium text-gray-900 truncate">{plan.description}</p>
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
                  onClick={(e) => e.stopPropagation()}
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
    </Card>
  )
}
