import { Card } from '@/components/ui/Card'
import { PlanStatusBadge } from './PlanStatusBadge'
import { GitBranch, Users, Clock } from 'lucide-react'
import { format } from 'date-fns'
import type { PlanDocument } from '@/types/plan-document'
import { parseRepoName, getRepoUrl, isDefaultProject } from '@/lib/project-utils'

interface PlanCardProps {
  plan: PlanDocument
  onClick: () => void
}

export function PlanCard({ plan, onClick }: PlanCardProps) {
  const repoName = parseRepoName(plan.project)
  const repoUrl = getRepoUrl(plan.project)
  const hasProject = !isDefaultProject(plan.project)
  const formattedDate = format(new Date(plan.updated_at), 'yyyy/MM/dd HH:mm')
  const collaboratorNames = plan.collaborators.map((c) => c.display_name).join(', ')

  return (
    <Card hover onClick={onClick}>
      <div className="min-w-0">
        {/* Title: Description + Status */}
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-gray-900 truncate">{plan.description}</p>
          <PlanStatusBadge status={plan.status} />
        </div>
        {/* Metadata: repo, collaborators, updated_at */}
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
            </span>
          )}
          {!hasProject && (
            <span className="flex items-center gap-1 text-gray-300">
              <GitBranch className="h-3 w-3" />
              (no project)
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
