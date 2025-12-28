import { EventCard } from './EventCard'
import type { Event } from '@/types/event'

interface TimelineProps {
  events: Event[]
}

export function Timeline({ events }: TimelineProps) {
  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No events yet.</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {events.map((event) => (
        <EventCard key={event.id} event={event} />
      ))}
    </div>
  )
}
