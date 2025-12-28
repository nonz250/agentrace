import { Card } from '@/components/ui/Card'
import { Folder, User, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import type { Session } from '@/types/session'

interface SessionCardProps {
  session: Session
  onClick: () => void
}

export function SessionCard({ session, onClick }: SessionCardProps) {
  return (
    <Card hover onClick={onClick}>
      <div className="flex items-start gap-3">
        <Folder className="mt-0.5 h-5 w-5 flex-shrink-0 text-gray-400" />
        <div className="min-w-0 flex-1">
          <p className="truncate font-mono text-sm text-gray-900">
            {session.project_path}
          </p>
          <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-500">
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
