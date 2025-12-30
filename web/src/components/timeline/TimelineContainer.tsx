import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Timeline, expandEvents, extractUserBlocks } from './Timeline'
import { UserMessageNav } from './UserMessageNav'
import type { Event } from '@/types/event'

interface TimelineContainerProps {
  events: Event[]
  projectPath?: string
}

export function TimelineContainer({ events, projectPath }: TimelineContainerProps) {
  const [activeBlockId, setActiveBlockId] = useState<string | null>(null)

  // Expand events and extract user blocks
  const displayBlocks = useMemo(() => expandEvents(events, projectPath), [events, projectPath])
  const userBlocks = useMemo(() => extractUserBlocks(displayBlocks), [displayBlocks])

  // Create refs for all user blocks
  const blockRefs = useMemo(() => {
    const refs = new Map<string, React.RefObject<HTMLDivElement | null>>()
    userBlocks.forEach((block) => {
      refs.set(block.id, { current: null })
    })
    return refs
  }, [userBlocks])

  // Mutable refs container for IntersectionObserver
  const blockRefsRef = useRef<Map<string, React.RefObject<HTMLDivElement | null>>>(blockRefs)
  useEffect(() => {
    blockRefsRef.current = blockRefs
  }, [blockRefs])

  // Set up IntersectionObserver to track which user block is currently visible
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        // Find the first intersecting entry
        const intersecting = entries.find((entry) => entry.isIntersecting)
        if (intersecting) {
          const blockId = intersecting.target.getAttribute('data-block-id')
          if (blockId) {
            setActiveBlockId(blockId)
          }
        }
      },
      {
        // Trigger when block enters the upper portion of the viewport
        rootMargin: '-100px 0px -70% 0px',
        threshold: 0,
      }
    )

    // Observe all user block elements
    blockRefsRef.current.forEach((ref) => {
      if (ref.current) {
        observer.observe(ref.current)
      }
    })

    return () => observer.disconnect()
  }, [userBlocks])

  // Handle navigation to a specific block
  const handleNavigate = useCallback((blockId: string) => {
    const ref = blockRefs.get(blockId)
    if (ref?.current) {
      ref.current.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
      })
    }
  }, [blockRefs])

  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No events yet.</p>
      </div>
    )
  }

  return (
    <div className="flex gap-6">
      {/* Left sidebar - hidden on small screens */}
      <aside className="hidden w-48 flex-shrink-0 lg:block">
        <UserMessageNav
          userBlocks={userBlocks}
          activeBlockId={activeBlockId}
          onNavigate={handleNavigate}
        />
      </aside>

      {/* Main timeline */}
      <main className="min-w-0 flex-1">
        <Timeline
          events={events}
          projectPath={projectPath}
          blockRefs={blockRefs}
        />
      </main>
    </div>
  )
}
