import { Link } from 'react-router-dom'
import { format } from 'date-fns'
import { User, Clock, ExternalLink, ArrowRight, FileEdit } from 'lucide-react'
import { DiffView } from './DiffView'
import type { PlanDocumentEvent } from '@/types/plan-document'

interface PlanEventHistoryProps {
  events: PlanDocumentEvent[]
}

export function PlanEventHistory({ events }: PlanEventHistoryProps) {
  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No history yet.</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {events.map((event, index) => {
        const formattedDate = format(new Date(event.created_at), 'yyyy/MM/dd HH:mm:ss')
        const isInitial = index === 0 && (!event.patch || event.patch === '')
        const isStatusChange = event.event_type === 'status_change'

        return (
          <div
            key={event.id}
            className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
          >
            <div className="flex items-start justify-between">
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
                  {event.user_name && (
                    <span className="flex items-center gap-1 text-gray-700">
                      <User className="h-4 w-4" />
                      {event.user_name}
                    </span>
                  )}
                  <span className="flex items-center gap-1 text-gray-400">
                    <Clock className="h-4 w-4" />
                    {formattedDate}
                  </span>
                  {event.session_id && (
                    <Link
                      to={`/sessions/${event.session_id}`}
                      className="flex items-center gap-1 text-blue-500 hover:text-blue-700"
                    >
                      <ExternalLink className="h-4 w-4" />
                      Session
                    </Link>
                  )}
                </div>
                <div className="mt-2">
                  {isInitial ? (
                    <span className="text-xs text-gray-500 bg-gray-100 px-2 py-1 rounded">
                      Initial creation
                    </span>
                  ) : isStatusChange ? (
                    <div className="flex items-center gap-2 text-sm">
                      <span className="flex items-center gap-1 text-purple-600 bg-purple-50 px-2 py-1 rounded">
                        <ArrowRight className="h-3 w-3" />
                        Status changed: {event.patch}
                      </span>
                    </div>
                  ) : event.patch ? (
                    <div>
                      <span className="flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded inline-flex mb-2">
                        <FileEdit className="h-3 w-3" />
                        Content updated
                      </span>
                      <details className="text-xs">
                        <summary className="cursor-pointer text-gray-500 hover:text-gray-700">
                          View patch
                        </summary>
                        <DiffView patch={event.patch} />
                      </details>
                    </div>
                  ) : (
                    <span className="text-xs text-gray-500">No patch recorded</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
