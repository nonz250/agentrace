import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Timeline, expandEvents, extractMessageBlocks } from './Timeline'
import { MessageNav } from './MessageNav'
import type { Event } from '@/types/event'

interface TimelineContainerProps {
  events: Event[]
  projectPath?: string
}

export function TimelineContainer({ events, projectPath }: TimelineContainerProps) {
  const [activeBlockId, setActiveBlockId] = useState<string | null>(null)

  // Expand events and extract message blocks (user + assistant)
  const displayBlocks = useMemo(() => expandEvents(events, projectPath), [events, projectPath])
  const messageBlocks = useMemo(() => extractMessageBlocks(displayBlocks), [displayBlocks])

  // Create refs for all message blocks (user + assistant)
  const blockRefs = useMemo(() => {
    const refs = new Map<string, React.RefObject<HTMLDivElement | null>>()
    messageBlocks.forEach((block) => {
      refs.set(block.id, { current: null })
    })
    return refs
  }, [messageBlocks])

  // Mutable refs container for IntersectionObserver
  const blockRefsRef = useRef<Map<string, React.RefObject<HTMLDivElement | null>>>(blockRefs)
  useEffect(() => {
    blockRefsRef.current = blockRefs
  }, [blockRefs])

  // Scroll to hash on initial load and hash change
  useEffect(() => {
    const scrollToHash = () => {
      const hash = window.location.hash
      if (hash && hash.startsWith('#event-')) {
        // Small delay to ensure DOM is ready
        setTimeout(() => {
          const element = document.getElementById(hash.slice(1))
          if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' })
          }
        }, 100)
      }
    }

    // Initial load
    scrollToHash()

    // Listen for hash changes
    window.addEventListener('hashchange', scrollToHash)
    return () => window.removeEventListener('hashchange', scrollToHash)
  }, [events]) // Re-run when events change (data loaded)

  // Set up IntersectionObserver to track which message block is currently visible
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

    // Observe all message block elements
    blockRefsRef.current.forEach((ref) => {
      if (ref.current) {
        observer.observe(ref.current)
      }
    })

    return () => observer.disconnect()
  }, [messageBlocks])

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
        <MessageNav
          messageBlocks={messageBlocks}
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
