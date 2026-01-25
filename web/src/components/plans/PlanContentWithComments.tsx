import { useState, useRef, useEffect, useCallback, memo } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { MessageSquare, MessageSquarePlus, X, Send, CheckCircle, User, Trash2, MessageCircle, MoreVertical, AlertTriangle } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { Button } from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import { CopyButton } from '@/components/ui/CopyButton'
import { useAuth } from '@/hooks/useAuth'
import * as commentsApi from '@/api/plan-comments'
import type { PlanCommentThread } from '@/types/plan-document'

interface SelectionInfo {
  text: string
  rect: DOMRect
  contextBefore: string
  contextAfter: string
}

// Separate component for comment form to isolate re-renders during typing
interface CommentFormProps {
  planId: string
  selection: SelectionInfo
  onSuccess: () => void
  onCancel: () => void
}

// Memoized markdown renderer to prevent re-rendering when unrelated state changes
interface MarkdownContentProps {
  body: string
}

const MarkdownContent = memo(function MarkdownContent({ body }: MarkdownContentProps) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        code({ className, children, ...props }) {
          const match = /language-(\w+)/.exec(className || '')
          const code = String(children).replace(/\n$/, '')
          return match ? (
            <SyntaxHighlighter
              language={match[1]}
              style={oneLight}
              customStyle={{
                fontSize: '0.875rem',
                borderRadius: '0.5rem',
                margin: '1rem 0',
                padding: '1rem',
              }}
            >
              {code}
            </SyntaxHighlighter>
          ) : (
            <code className={className} {...props}>
              {children}
            </code>
          )
        },
      }}
    >
      {body}
    </ReactMarkdown>
  )
})

const CommentForm = memo(function CommentForm({ planId, selection, onSuccess, onCancel }: CommentFormProps) {
  const [content, setContent] = useState('')
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (data: { target_text: string; context_before: string; context_after: string; content: string }) =>
      commentsApi.createPlanThread(planId, {
        target_text: data.target_text,
        context_before: data.context_before,
        context_after: data.context_after,
        content: data.content,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
      onSuccess()
    },
  })

  const handleSubmit = () => {
    if (!content.trim()) return
    createMutation.mutate({
      target_text: selection.text,
      context_before: selection.contextBefore,
      context_after: selection.contextAfter,
      content: content.trim(),
    })
  }

  return (
    <div
      className="comment-form absolute z-[5] bg-white border border-gray-200 rounded-lg shadow-xl p-4 w-80"
      style={{
        top: selection.rect.top + 8,
        left: selection.rect.left + 12,
      }}
    >
      <div className="mb-3">
        <div className="text-xs text-gray-500 mb-1">Commenting on:</div>
        <div className="text-sm bg-gray-50 border-l-2 border-gray-300 px-2 py-1 italic text-gray-600 line-clamp-2">
          "{selection.text}"
        </div>
      </div>
      <Textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="Write your comment..."
        rows={3}
        autoFocus
      />
      <div className="flex justify-end gap-2 mt-3">
        <Button variant="ghost" size="sm" onClick={onCancel} disabled={createMutation.isPending}>
          <X className="mr-1 h-4 w-4" />
          Cancel
        </Button>
        <Button
          size="sm"
          onClick={handleSubmit}
          disabled={!content.trim() || createMutation.isPending}
          loading={createMutation.isPending}
        >
          <Send className="mr-1 h-4 w-4" />
          Send
        </Button>
      </div>
    </div>
  )
})

interface PlanContentWithCommentsProps {
  planId: string
  body: string
  threads: PlanCommentThread[]
}

export function PlanContentWithComments({ planId, body, threads }: PlanContentWithCommentsProps) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const containerRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)

  const [selection, setSelection] = useState<SelectionInfo | null>(null)
  const [isAddingComment, setIsAddingComment] = useState(false)
  const [activeThread, setActiveThread] = useState<PlanCommentThread | null>(null)
  const [activeThreadRect, setActiveThreadRect] = useState<DOMRect | null>(null)
  const [replyContent, setReplyContent] = useState('')
  const [showThreadMenu, setShowThreadMenu] = useState(false)
  const [showThreadList, setShowThreadList] = useState(false)

  const resolveThreadMutation = useMutation({
    mutationFn: (threadId: string) => commentsApi.resolvePlanThread(planId, threadId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
      setActiveThread(null)
    },
  })

  const deleteThreadMutation = useMutation({
    mutationFn: (threadId: string) => commentsApi.deletePlanThread(planId, threadId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
      setActiveThread(null)
    },
  })

  const addMessageMutation = useMutation({
    mutationFn: ({ threadId, content }: { threadId: string; content: string }) =>
      commentsApi.addThreadMessage(planId, threadId, { content }),
    onSuccess: (newMessage) => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
      setReplyContent('')
      // Update activeThread with the new message for immediate UI feedback
      if (activeThread && newMessage) {
        setActiveThread({
          ...activeThread,
          messages: [...activeThread.messages, newMessage],
        })
      }
    },
  })

  // Helper function to get character offset of a node within the content container
  const getCharacterOffsetInContainer = useCallback((container: Node, targetNode: Node, targetOffset: number): number => {
    let offset = 0
    const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, null)
    let node: Node | null
    while ((node = walker.nextNode())) {
      if (node === targetNode) {
        return offset + targetOffset
      }
      offset += (node as Text).textContent?.length || 0
    }
    return offset
  }, [])

  // Helper function to find all occurrences of a substring
  const findAllOccurrences = useCallback((text: string, searchStr: string): number[] => {
    const indices: number[] = []
    let idx = 0
    while ((idx = text.indexOf(searchStr, idx)) !== -1) {
      indices.push(idx)
      idx += 1
    }
    return indices
  }, [])

  // Helper function to strip inline markdown formatting from text
  // Returns the stripped text and a position map from stripped positions to original positions
  const stripInlineMarkdown = useCallback((text: string): { stripped: string; posMap: number[] } => {
    const posMap: number[] = []
    let result = ''
    let i = 0

    while (i < text.length) {
      const ch = text[i]

      // Handle inline code: `code`
      if (ch === '`') {
        // Find closing backtick
        let j = i + 1
        while (j < text.length && text[j] !== '`') {
          j++
        }
        if (j < text.length) {
          // Found closing backtick, extract content
          for (let k = i + 1; k < j; k++) {
            result += text[k]
            posMap.push(k)
          }
          i = j + 1
          continue
        }
      }

      // Handle links: [text](url) or [text](url "title")
      if (ch === '[') {
        // Find closing bracket
        let j = i + 1
        let bracketDepth = 1
        while (j < text.length && bracketDepth > 0) {
          if (text[j] === '[') {
            bracketDepth++
          } else if (text[j] === ']') {
            bracketDepth--
          }
          j++
        }
        // j now points to character after ']'
        if (j < text.length && text[j] === '(') {
          // Find closing parenthesis
          let k = j + 1
          let parenDepth = 1
          while (k < text.length && parenDepth > 0) {
            if (text[k] === '(') {
              parenDepth++
            } else if (text[k] === ')') {
              parenDepth--
            }
            k++
          }
          // k now points to character after ')'
          if (parenDepth === 0) {
            // Valid link found, extract link text (between [ and ])
            for (let l = i + 1; l < j - 1; l++) {
              result += text[l]
              posMap.push(l)
            }
            i = k
            continue
          }
        }
      }

      // Handle bold/italic: ** or * or __ or _
      if (ch === '*' || ch === '_') {
        // Check for double marker
        if (i + 1 < text.length && text[i + 1] === ch) {
          i += 2
          continue
        }
        i++
        continue
      }

      // Handle strikethrough: ~~
      if (ch === '~' && i + 1 < text.length && text[i + 1] === '~') {
        i += 2
        continue
      }

      // Regular character
      result += ch
      posMap.push(i)
      i++
    }

    return { stripped: result, posMap }
  }, [])


  // Helper function to extract context from the raw markdown body
  // Uses stripped text matching to handle inline formatting like `code`
  const extractContextFromBody = useCallback((targetText: string, occurrenceIndex: number): { contextBefore: string; contextAfter: string } => {
    const contextLength = 100

    // First try exact match (for text without inline formatting)
    const exactOccurrences = findAllOccurrences(body, targetText)
    if (occurrenceIndex >= 0 && occurrenceIndex < exactOccurrences.length) {
      const pos = exactOccurrences[occurrenceIndex]
      return {
        contextBefore: body.substring(Math.max(0, pos - contextLength), pos),
        contextAfter: body.substring(pos + targetText.length, pos + targetText.length + contextLength),
      }
    }

    // Try matching in stripped text (for text spanning inline formatting)
    const { stripped, posMap } = stripInlineMarkdown(body)
    const strippedOccurrences = findAllOccurrences(stripped, targetText)

    if (occurrenceIndex >= 0 && occurrenceIndex < strippedOccurrences.length) {
      const strippedPos = strippedOccurrences[occurrenceIndex]
      const strippedEndPos = strippedPos + targetText.length

      // Map stripped positions back to original positions
      if (strippedPos < posMap.length && strippedEndPos <= posMap.length) {
        const origStartPos = posMap[strippedPos]
        const origEndPos = posMap[strippedEndPos - 1] + 1

        return {
          contextBefore: body.substring(Math.max(0, origStartPos - contextLength), origStartPos),
          contextAfter: body.substring(origEndPos, origEndPos + contextLength),
        }
      }
    }

    // Fallback: use first occurrence in stripped text
    if (strippedOccurrences.length > 0) {
      const strippedPos = strippedOccurrences[0]
      const strippedEndPos = strippedPos + targetText.length

      if (strippedPos < posMap.length && strippedEndPos <= posMap.length) {
        const origStartPos = posMap[strippedPos]
        const origEndPos = posMap[strippedEndPos - 1] + 1

        return {
          contextBefore: body.substring(Math.max(0, origStartPos - contextLength), origStartPos),
          contextAfter: body.substring(origEndPos, origEndPos + contextLength),
        }
      }
    }

    return { contextBefore: '', contextAfter: '' }
  }, [body, findAllOccurrences, stripInlineMarkdown])

  // Handle text selection
  const handleMouseUp = useCallback((e: React.MouseEvent) => {
    if (!user || isAddingComment) return

    // Small delay to let the selection finalize
    setTimeout(() => {
      const sel = window.getSelection()
      if (!sel || sel.isCollapsed) {
        return
      }

      const text = sel.toString().trim()
      if (!text || text.length < 2) {
        setSelection(null)
        return
      }

      // Reject selections that contain newlines (spanning multiple blocks/lines)
      // Markdown line breaks make matching unreliable due to list markers, indentation, etc.
      if (text.includes('\n')) {
        setSelection(null)
        return
      }

      // Check if selection is within our content area
      const range = sel.getRangeAt(0)
      if (!contentRef.current?.contains(range.commonAncestorContainer)) {
        setSelection(null)
        return
      }

      // Check if selection spans across block-level elements
      const commonAncestor = range.commonAncestorContainer
      const commonAncestorElement = commonAncestor.nodeType === Node.ELEMENT_NODE
        ? commonAncestor as Element
        : commonAncestor.parentElement

      // If the common ancestor is the content container itself, selection spans multiple top-level blocks
      if (commonAncestorElement === contentRef.current) {
        setSelection(null)
        return
      }

      // If the common ancestor is a list, table, or similar container, selection spans multiple items
      // TR is included to prevent selecting across table cells (which would break markdown matching)
      const containerTags = ['UL', 'OL', 'TABLE', 'TBODY', 'THEAD', 'TFOOT', 'TR']
      if (commonAncestorElement && containerTags.includes(commonAncestorElement.tagName)) {
        setSelection(null)
        return
      }

      // Note: We allow selections that span across inline formatting elements (CODE, STRONG, EM, etc.)
      // The server-side and client-side logic handles inline markdown stripping for proper matching

      // However, reject selections inside links (A elements) to avoid click handling conflicts
      const isInsideLink = (node: Node): boolean => {
        let current = node.nodeType === Node.ELEMENT_NODE ? node as Element : node.parentElement
        while (current && current !== contentRef.current) {
          if (current.tagName === 'A') {
            return true
          }
          current = current.parentElement
        }
        return false
      }

      if (isInsideLink(range.startContainer) || isInsideLink(range.endContainer)) {
        setSelection(null)
        return
      }

      // Find which occurrence of the target text this is in the rendered content
      const fullRenderedText = contentRef.current.textContent || ''
      const startNode = range.startContainer
      const startOffset = range.startOffset
      const charOffset = getCharacterOffsetInContainer(contentRef.current, startNode, startOffset)

      // Count how many occurrences of the text appear before this position in the rendered content
      const textBeforeSelection = fullRenderedText.substring(0, charOffset)
      const occurrencesBeforeInRendered = findAllOccurrences(textBeforeSelection, text).length

      // Extract context from the raw markdown body for the same Nth occurrence
      const { contextBefore, contextAfter } = extractContextFromBody(text, occurrencesBeforeInRendered)

      // Use mouse position for more reliable popup placement
      const containerRect = containerRef.current?.getBoundingClientRect()
      const relativeX = e.clientX - (containerRect?.left || 0)
      const relativeY = e.clientY - (containerRect?.top || 0)

      const rect = {
        top: relativeY,
        bottom: relativeY,
        left: relativeX,
        right: relativeX,
        width: 0,
        height: 0,
        x: relativeX,
        y: relativeY,
        toJSON: () => ({}),
      } as DOMRect

      setSelection({ text, rect, contextBefore, contextAfter })
    }, 10)
  }, [user, isAddingComment, getCharacterOffsetInContainer, findAllOccurrences, extractContextFromBody])

  // Close selection popup when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      if (selection && !target.closest('.selection-popup') && !target.closest('.comment-form')) {
        if (!isAddingComment) {
          setSelection(null)
        }
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [selection, isAddingComment])

  // Handle starting comment
  const handleStartComment = () => {
    setIsAddingComment(true)
  }

  // Handle comment form close
  const handleCommentFormClose = () => {
    setIsAddingComment(false)
    setSelection(null)
  }

  // Store threads in a ref for event delegation
  const threadsRef = useRef<PlanCommentThread[]>([])
  threadsRef.current = threads

  // Handle click on highlight via event delegation
  const handleContentClick = useCallback((e: React.MouseEvent) => {
    const target = e.target as HTMLElement
    const highlight = target.closest('.comment-highlight') as HTMLElement | null
    if (highlight && highlight.dataset.threadId) {
      e.stopPropagation()
      const thread = threadsRef.current.find(t => t.id === highlight.dataset.threadId)
      if (thread) {
        setActiveThread(thread)
        // Calculate position relative to container for absolute positioning
        const containerRect = containerRef.current?.getBoundingClientRect()
        const relativeX = e.clientX - (containerRect?.left || 0)
        const relativeY = e.clientY - (containerRect?.top || 0)
        setActiveThreadRect({
          top: relativeY,
          bottom: relativeY,
          left: relativeX,
          right: relativeX,
          width: 0,
          height: 0,
          x: relativeX,
          y: relativeY,
          toJSON: () => ({}),
        } as DOMRect)
      }
    }
  }, [])

  // Helper: Convert byte offset to character offset in UTF-8 string
  const byteOffsetToCharOffset = useCallback((str: string, byteOffset: number): number => {
    const encoder = new TextEncoder()
    let charIndex = 0
    let byteIndex = 0
    for (const char of str) {
      if (byteIndex >= byteOffset) break
      byteIndex += encoder.encode(char).length
      charIndex++
    }
    return charIndex
  }, [])

  // Find and highlight threads in the rendered content
  useEffect(() => {
    if (!contentRef.current) return

    const applyHighlights = () => {
      if (!contentRef.current) return

      // Remove existing highlights
      const existingHighlights = contentRef.current.querySelectorAll('.comment-highlight')
      existingHighlights.forEach((el) => {
        const parent = el.parentNode
        if (parent) {
          parent.replaceChild(document.createTextNode(el.textContent || ''), el)
          parent.normalize()
        }
      })

      // Add new highlights for active threads
      const activeThreads = threads.filter((t) => t.status === 'active' && t.position?.found)

      // Get all text content from the container for cross-element matching
      const getTextNodesWithPositions = (root: Node): { node: Text; start: number; end: number }[] => {
        const result: { node: Text; start: number; end: number }[] = []
        let currentPos = 0
        const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, null)
        let node: Node | null
        while ((node = walker.nextNode())) {
          const textNode = node as Text
          const len = textNode.textContent?.length || 0
          result.push({ node: textNode, start: currentPos, end: currentPos + len })
          currentPos += len
        }
        return result
      }

      // Get full text content (without markup)
      const fullRenderedText = contentRef.current.textContent || ''

      for (const thread of activeThreads) {
        // Determine which occurrence this thread refers to using the position from server
        const startByteOffset = thread.position!.start_offset
        const startCharOffset = byteOffsetToCharOffset(body, startByteOffset)

        // Count how many occurrences of target_text appear before this position in body
        // Use stripped text to handle cases where target_text spans across inline formatting
        const bodyBeforeTarget = body.substring(0, startCharOffset)
        const { stripped: strippedBodyBefore } = stripInlineMarkdown(bodyBeforeTarget)
        const occurrencesBeforeInBody = findAllOccurrences(strippedBodyBefore, thread.target_text).length

        // Find all occurrences in rendered content
        const occurrencesInRendered = findAllOccurrences(fullRenderedText, thread.target_text)

        // Get the position of the same Nth occurrence in rendered content
        const targetOccurrenceIndex = Math.min(occurrencesBeforeInBody, occurrencesInRendered.length - 1)
        if (targetOccurrenceIndex < 0 || occurrencesInRendered.length === 0) {
          continue
        }
        const targetPosInRendered = occurrencesInRendered[targetOccurrenceIndex]

        let highlighted = false
        const textNodes = getTextNodesWithPositions(contentRef.current!)

        // Find the text node that contains this position
        for (const { node: textNode, start: nodeStart, end: nodeEnd } of textNodes) {
          if (highlighted) break

          if (targetPosInRendered >= nodeStart && targetPosInRendered < nodeEnd) {
            const text = textNode.textContent || ''
            const indexInNode = targetPosInRendered - nodeStart

            if (text.substring(indexInNode, indexInNode + thread.target_text.length) !== thread.target_text) {
              continue
            }

            if (textNode.parentNode) {
              try {
                const before = text.substring(0, indexInNode)
                const target = text.substring(indexInNode, indexInNode + thread.target_text.length)
                const after = text.substring(indexInNode + thread.target_text.length)

                const highlight = document.createElement('mark')
                highlight.className = 'comment-highlight bg-yellow-100 hover:bg-yellow-200 cursor-pointer rounded px-0.5 transition-colors'
                highlight.dataset.threadId = thread.id
                highlight.textContent = target

                const parent = textNode.parentNode
                const frag = document.createDocumentFragment()

                if (before) {
                  frag.appendChild(document.createTextNode(before))
                }
                frag.appendChild(highlight)
                if (after) {
                  frag.appendChild(document.createTextNode(after))
                }

                parent.replaceChild(frag, textNode)
                highlighted = true
              } catch {
                // Skip on error
              }
            }
          }
        }

        // Cross-element matching fallback
        if (!highlighted && targetPosInRendered >= 0) {
          const freshTextNodes = getTextNodesWithPositions(contentRef.current!)
          const matchEnd = targetPosInRendered + thread.target_text.length

          for (const { node: textNode, start, end } of freshTextNodes) {
            if (end <= targetPosInRendered || start >= matchEnd) continue
            if (!textNode.parentNode) continue

            const text = textNode.textContent || ''
            const overlapStart = Math.max(0, targetPosInRendered - start)
            const overlapEnd = Math.min(text.length, matchEnd - start)

            if (overlapStart >= overlapEnd) continue

            try {
              const before = text.substring(0, overlapStart)
              const target = text.substring(overlapStart, overlapEnd)
              const after = text.substring(overlapEnd)

              const highlight = document.createElement('mark')
              highlight.className = 'comment-highlight bg-yellow-100 hover:bg-yellow-200 cursor-pointer rounded px-0.5 transition-colors'
              highlight.dataset.threadId = thread.id
              highlight.textContent = target

              const parent = textNode.parentNode
              const frag = document.createDocumentFragment()

              if (before) {
                frag.appendChild(document.createTextNode(before))
              }
              frag.appendChild(highlight)
              if (after) {
                frag.appendChild(document.createTextNode(after))
              }

              parent.replaceChild(frag, textNode)
              highlighted = true
            } catch {
              // Skip on error
            }
          }
        }
      }
    }

    const timer = setTimeout(applyHighlights, 10)
    return () => clearTimeout(timer)
  }, [threads, body, byteOffsetToCharOffset, findAllOccurrences])

  // Get thread counts
  const activeThreadsCount = threads.filter((t) => t.status === 'active').length
  const outdatedThreadsCount = threads.filter((t) => t.status === 'outdated').length
  const outdatedThreads = threads.filter((t) => t.status === 'outdated')

  // Get first message creator for thread owner check
  const isThreadOwner = activeThread && activeThread.messages.length > 0 && user?.id === activeThread.messages[0].user_id

  return (
    <div ref={containerRef} className="relative">
      <CopyButton text={body} className="absolute top-0 right-0 z-10" />

      {/* Threads indicator - clickable to show thread list */}
      {(activeThreadsCount > 0 || outdatedThreadsCount > 0) && (
        <div className="absolute top-0 right-12">
          <button
            className="flex items-center gap-1 text-xs text-yellow-600 bg-yellow-50 hover:bg-yellow-100 px-2 py-1 rounded transition-colors cursor-pointer"
            onClick={() => setShowThreadList(!showThreadList)}
          >
            <MessageSquare className="h-3 w-3" />
            {activeThreadsCount > 0 && (
              <span>{activeThreadsCount} thread{activeThreadsCount > 1 ? 's' : ''}</span>
            )}
            {outdatedThreadsCount > 0 && (
              <span className="text-gray-400">
                {activeThreadsCount > 0 && ' · '}
                {outdatedThreadsCount} outdated
              </span>
            )}
          </button>

          {/* Thread list dropdown */}
          {showThreadList && (
            <>
              <div
                className="fixed inset-0 z-[3]"
                onClick={() => setShowThreadList(false)}
              />
              <div className="absolute right-0 top-full mt-1 z-[4] bg-white border border-gray-200 rounded-lg shadow-lg py-1 w-72 max-h-80 overflow-y-auto">
                {/* Active Threads */}
                {activeThreadsCount > 0 && (
                  <>
                    <div className="px-3 py-2 text-xs font-medium text-gray-500 border-b border-gray-100">
                      Active Threads
                    </div>
                    {threads
                      .filter((t) => t.status === 'active')
                      .map((thread) => (
                        <button
                          key={thread.id}
                          className="w-full px-3 py-2 text-left hover:bg-gray-50 transition-colors"
                          onClick={() => {
                            setShowThreadList(false)

                            // Find the highlight element for this thread
                            const highlightEl = contentRef.current?.querySelector(
                              `.comment-highlight[data-thread-id="${thread.id}"]`
                            ) as HTMLElement | null

                            if (highlightEl) {
                              // Scroll the highlight into view
                              highlightEl.scrollIntoView({ behavior: 'smooth', block: 'center' })

                              // Open the thread UI after a short delay to let scroll finish
                              setTimeout(() => {
                                const containerRect = containerRef.current?.getBoundingClientRect()
                                const highlightRect = highlightEl.getBoundingClientRect()
                                if (containerRect) {
                                  setActiveThread(thread)
                                  setActiveThreadRect({
                                    top: highlightRect.top - containerRect.top + highlightRect.height,
                                    bottom: highlightRect.bottom - containerRect.top,
                                    left: highlightRect.left - containerRect.left,
                                    right: highlightRect.right - containerRect.left,
                                    width: highlightRect.width,
                                    height: highlightRect.height,
                                    x: highlightRect.left - containerRect.left,
                                    y: highlightRect.top - containerRect.top,
                                    toJSON: () => ({}),
                                  } as DOMRect)
                                }
                              }, 100)
                            }
                          }}
                        >
                          <div className="text-sm text-gray-700 line-clamp-2">
                            "{thread.target_text}"
                          </div>
                          <div className="text-xs text-gray-400 mt-1">
                            {thread.messages.length} {thread.messages.length === 1 ? 'reply' : 'replies'}
                          </div>
                        </button>
                      ))}
                  </>
                )}

                {/* Outdated Threads */}
                {outdatedThreadsCount > 0 && (
                  <>
                    <div className="px-3 py-2 text-xs font-medium text-gray-400 border-b border-gray-100 flex items-center gap-1">
                      <AlertTriangle className="h-3 w-3" />
                      Outdated ({outdatedThreadsCount})
                    </div>
                    {outdatedThreads.map((thread) => (
                      <button
                        key={thread.id}
                        className="w-full px-3 py-2 text-left hover:bg-gray-50 transition-colors opacity-60"
                        onClick={() => {
                          setShowThreadList(false)
                          // Open popup at a fixed position (no scrolling since text doesn't exist)
                          const containerRect = containerRef.current?.getBoundingClientRect()
                          if (containerRect) {
                            setActiveThread(thread)
                            setActiveThreadRect({
                              top: 40,
                              bottom: 60,
                              left: 0,
                              right: 200,
                              width: 200,
                              height: 20,
                              x: 0,
                              y: 40,
                              toJSON: () => ({}),
                            } as DOMRect)
                          }
                        }}
                      >
                        <div className="text-sm text-gray-500 line-clamp-2">
                          "{thread.target_text}"
                        </div>
                        <div className="text-xs text-gray-400 mt-1">
                          {thread.messages.length} {thread.messages.length === 1 ? 'reply' : 'replies'} · text no longer exists
                        </div>
                      </button>
                    ))}
                  </>
                )}
              </div>
            </>
          )}
        </div>
      )}

      {/* Content with selection detection */}
      <div
        ref={contentRef}
        onMouseUp={handleMouseUp}
        onClick={handleContentClick}
        className="prose prose-sm max-w-none prose-headings:text-gray-900 prose-p:text-gray-700 prose-a:text-blue-600 prose-code:rounded prose-code:bg-gray-100 prose-code:px-1 prose-code:py-0.5 prose-code:text-gray-800 prose-code:before:content-none prose-code:after:content-none prose-pre:bg-transparent prose-pre:p-0 prose-pre:my-0"
      >
        <MarkdownContent body={body} />
      </div>

      {/* Selection popup - Add Comment button */}
      {selection && user && !isAddingComment && (
        <button
          className="selection-popup absolute z-[5] bg-white border border-gray-200 rounded-full shadow-lg p-2 hover:bg-gray-50 transition-colors"
          style={{
            top: selection.rect.top + 8,
            left: selection.rect.left + 12,
          }}
          onClick={handleStartComment}
          title="Add Comment"
        >
          <MessageSquarePlus className="h-4 w-4 text-gray-600" />
        </button>
      )}

      {/* Comment form popup */}
      {selection && isAddingComment && (
        <CommentForm
          planId={planId}
          selection={selection}
          onSuccess={handleCommentFormClose}
          onCancel={handleCommentFormClose}
        />
      )}

      {/* Active thread popup */}
      {activeThread && activeThreadRect && (
        <div
          className="absolute z-[5] bg-white border border-gray-200 rounded-lg shadow-xl p-4 w-96 max-h-[80vh] overflow-y-auto"
          style={{
            top: activeThreadRect.top + 8,
            left: activeThreadRect.left + 12,
          }}
        >
          {/* Header with quote and actions */}
          <div className="flex items-start justify-between gap-2 mb-3">
            <div className="flex-1">
              {activeThread.status === 'resolved' && (
                <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs bg-green-100 text-green-700 mb-1">
                  <CheckCircle className="h-3 w-3" />
                  Resolved
                </span>
              )}
              {activeThread.status === 'outdated' && (
                <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs bg-gray-200 text-gray-600 mb-1">
                  <AlertTriangle className="h-3 w-3" />
                  Outdated
                </span>
              )}
              <div className="text-sm bg-gray-50 border-l-2 border-gray-300 px-2 py-1 italic text-gray-600">
                "{activeThread.target_text}"
                {activeThread.status === 'outdated' && (
                  <span className="block mt-1 text-xs text-orange-600 not-italic">
                    This text has been modified or removed from the document.
                  </span>
                )}
              </div>
            </div>
            {/* Dropdown menu for thread actions */}
            {isThreadOwner && (activeThread.status === 'active' || activeThread.status === 'outdated') && (
              <div className="relative flex-shrink-0">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowThreadMenu(!showThreadMenu)}
                  className="!p-1"
                >
                  <MoreVertical className="h-4 w-4 text-gray-500" />
                </Button>
                {showThreadMenu && (
                  <>
                    <div
                      className="fixed inset-0 z-10"
                      onClick={() => setShowThreadMenu(false)}
                    />
                    <div className="absolute right-0 top-full mt-1 z-20 bg-white border border-gray-200 rounded-lg shadow-lg py-1 min-w-32">
                      <button
                        className="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                        onClick={() => {
                          setShowThreadMenu(false)
                          if (confirm('Delete this thread?')) {
                            deleteThreadMutation.mutate(activeThread.id)
                          }
                        }}
                        disabled={deleteThreadMutation.isPending}
                      >
                        <Trash2 className="h-4 w-4" />
                        Delete
                      </button>
                    </div>
                  </>
                )}
              </div>
            )}
          </div>

          {/* Messages */}
          <div className="space-y-2 mb-3">
            {activeThread.messages.map((message) => (
              <div key={message.id} className="border-t border-gray-100 pt-2 first:border-t-0 first:pt-0">
                <div className="flex items-center gap-2 text-sm mb-1">
                  <User className="h-3.5 w-3.5 text-gray-500" />
                  <span className="font-medium">{message.user_name}</span>
                  <span className="text-gray-400 text-xs">
                    {formatDistanceToNow(new Date(message.created_at), { addSuffix: true })}
                  </span>
                </div>
                <div className="text-sm text-gray-700 whitespace-pre-wrap pl-5">
                  {message.content}
                </div>
              </div>
            ))}
          </div>

          {/* Reply form */}
          {activeThread.status === 'active' && user && (
            <div className="border-t border-gray-200 pt-3">
              <Textarea
                value={replyContent}
                onChange={(e) => setReplyContent(e.target.value)}
                placeholder="Write a reply..."
                rows={2}
                className="text-sm mb-2"
              />
              <div className="flex justify-end gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => resolveThreadMutation.mutate(activeThread.id)}
                  disabled={resolveThreadMutation.isPending}
                >
                  <CheckCircle className="mr-1 h-3.5 w-3.5" />
                  Resolve
                </Button>
                <Button
                  size="sm"
                  onClick={() => addMessageMutation.mutate({ threadId: activeThread.id, content: replyContent })}
                  disabled={!replyContent.trim() || addMessageMutation.isPending}
                  loading={addMessageMutation.isPending}
                >
                  <MessageCircle className="mr-1 h-3.5 w-3.5" />
                  Reply
                </Button>
              </div>
            </div>
          )}

          {/* Actions for outdated threads */}
          {activeThread.status === 'outdated' && user && (
            <div className="border-t border-gray-200 pt-3 flex justify-end">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => resolveThreadMutation.mutate(activeThread.id)}
                disabled={resolveThreadMutation.isPending}
              >
                <CheckCircle className="mr-1 h-3.5 w-3.5" />
                Resolve
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Backdrop for thread popup */}
      {activeThread && (
        <div
          className="fixed inset-0 z-[4]"
          onClick={() => { setActiveThread(null); setShowThreadMenu(false) }}
        />
      )}
    </div>
  )
}
