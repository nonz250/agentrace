import { Card } from '@/components/ui/Card'
import { GitBranch, Folder, MessageSquare } from 'lucide-react'
import { format } from 'date-fns'
import type { Session } from '@/types/session'
import { parseRepoName, getRepoUrl, isDefaultProject } from '@/lib/project-utils'

interface SessionCardProps {
  session: Session
  onClick: () => void
}

// Extract directory name from absolute path
function getDirectoryName(path: string): string {
  if (!path) return ''
  return path.split('/').pop() || path
}

export function SessionCard({ session, onClick }: SessionCardProps) {
  const repoName = parseRepoName(session.project)
  const repoUrl = getRepoUrl(session.project)
  const hasProject = !isDefaultProject(session.project)
  const formattedDate = format(new Date(session.started_at), 'yyyy/MM/dd HH:mm')

  return (
    <Card hover onClick={onClick}>
      <div className="min-w-0">
        {/* Title: Date and User */}
        <p className="text-sm font-medium text-gray-900">
          {formattedDate}
          <span className="ml-2 text-gray-600">{session.user_name || 'Unknown'}</span>
        </p>
        {/* Metadata: repo, branch, path, events */}
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
                  onClick={(e) => e.stopPropagation()}
                >
                  {repoName}
                </a>
              ) : (
                repoName
              )}
              {session.git_branch && (
                <span>: {session.git_branch}</span>
              )}
            </span>
          )}
          {!hasProject && session.project_path && (
            <span className="flex items-center gap-1 truncate" title={session.project_path}>
              <Folder className="h-3 w-3 flex-shrink-0" />
              <span className="truncate font-mono">{getDirectoryName(session.project_path)}</span>
            </span>
          )}
          <span className="flex items-center gap-1">
            <MessageSquare className="h-3 w-3" />
            {session.event_count}
          </span>
        </div>
      </div>
    </Card>
  )
}
