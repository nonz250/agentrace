import { useNavigate } from 'react-router-dom'
import { SessionCard } from './SessionCard'
import type { Session } from '@/types/session'

interface SessionListProps {
  sessions: Session[]
}

export function SessionList({ sessions }: SessionListProps) {
  const navigate = useNavigate()

  if (sessions.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No sessions yet.</p>
        <p className="mt-1 text-sm text-gray-400">
          Sessions will appear here once Claude Code sends data.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {sessions.map((session) => (
        <SessionCard
          key={session.id}
          session={session}
          onClick={() => navigate(`/sessions/${session.id}`)}
        />
      ))}
    </div>
  )
}
