import { useState, useRef, useEffect, useCallback, memo } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { MessageSquare, MessageSquarePlus, X, Send, CheckCircle, User, Trash2 } from 'lucide-react'
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
import type { PlanComment } from '@/types/plan-document'

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
      commentsApi.createPlanComment(planId, {
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
        top: Math.max(selection.rect.top, 16),
        left: selection.rect.left + 12,
        transform: 'translateY(-50%)',
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
  comments: PlanComment[]
}

export function PlanContentWithComments({ planId, body, comments }: PlanContentWithCommentsProps) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const containerRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)

  const [selection, setSelection] = useState<SelectionInfo | null>(null)
  const [isAddingComment, setIsAddingComment] = useState(false)
  const [activeComment, setActiveComment] = useState<PlanComment | null>(null)
  const [activeCommentRect, setActiveCommentRect] = useState<DOMRect | null>(null)

  const resolveMutation = useMutation({
    mutationFn: (commentId: string) => commentsApi.resolvePlanComment(planId, commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
      setActiveComment(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (commentId: string) => commentsApi.deletePlanComment(planId, commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
      setActiveComment(null)
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

  // Helper function to extract context from the raw markdown body
  const extractContextFromBody = useCallback((targetText: string, occurrenceIndex: number): { contextBefore: string; contextAfter: string } => {
    const contextLength = 100
    const occurrences = findAllOccurrences(body, targetText)

    if (occurrenceIndex < 0 || occurrenceIndex >= occurrences.length) {
      // Fallback: use first occurrence or empty context
      if (occurrences.length > 0) {
        const pos = occurrences[0]
        return {
          contextBefore: body.substring(Math.max(0, pos - contextLength), pos),
          contextAfter: body.substring(pos + targetText.length, pos + targetText.length + contextLength),
        }
      }
      return { contextBefore: '', contextAfter: '' }
    }

    const pos = occurrences[occurrenceIndex]
    return {
      contextBefore: body.substring(Math.max(0, pos - contextLength), pos),
      contextAfter: body.substring(pos + targetText.length, pos + targetText.length + contextLength),
    }
  }, [body, findAllOccurrences])

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

      // Check if selection is within our content area
      const range = sel.getRangeAt(0)
      if (!contentRef.current?.contains(range.commonAncestorContainer)) {
        setSelection(null)
        return
      }

      // Check if selection spans across block-level elements
      // by examining the commonAncestorContainer
      const commonAncestor = range.commonAncestorContainer
      const commonAncestorElement = commonAncestor.nodeType === Node.ELEMENT_NODE
        ? commonAncestor as Element
        : commonAncestor.parentElement

      // If the common ancestor is the content container itself, selection spans multiple top-level blocks
      if (commonAncestorElement === contentRef.current) {
        setSelection(null)
        return
      }

      // If the common ancestor is a list, table, or similar container (not a leaf element),
      // selection spans multiple items
      const containerTags = ['UL', 'OL', 'TABLE', 'TBODY', 'THEAD', 'TFOOT']
      if (commonAncestorElement && containerTags.includes(commonAncestorElement.tagName)) {
        setSelection(null)
        return
      }

      // Check if selection spans across inline formatting elements (code, strong, em, etc.)
      // This happens when selecting text like "Repository: `server" which crosses a code boundary
      const getInlineFormattingParent = (node: Node): Element | null => {
        let current = node.nodeType === Node.ELEMENT_NODE ? node as Element : node.parentElement
        const inlineTags = ['CODE', 'STRONG', 'EM', 'A', 'MARK', 'SPAN']
        while (current && current !== contentRef.current) {
          if (inlineTags.includes(current.tagName)) {
            return current
          }
          current = current.parentElement
        }
        return null
      }

      const startInlineParent = getInlineFormattingParent(range.startContainer)
      const endInlineParent = getInlineFormattingParent(range.endContainer)

      // If one end is in an inline element and the other is not, or they're in different inline elements
      if (startInlineParent !== endInlineParent) {
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
      // Calculate position relative to container for absolute positioning
      const containerRect = containerRef.current?.getBoundingClientRect()
      const relativeX = e.clientX - (containerRect?.left || 0)
      const relativeY = e.clientY - (containerRect?.top || 0)

      // Create a rect based on container-relative position
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

  // Store comments in a ref for event delegation
  const commentsRef = useRef<PlanComment[]>([])
  commentsRef.current = comments

  // Handle click on highlight via event delegation
  const handleContentClick = useCallback((e: React.MouseEvent) => {
    const target = e.target as HTMLElement
    const highlight = target.closest('.comment-highlight') as HTMLElement | null
    if (highlight && highlight.dataset.commentId) {
      e.stopPropagation()
      const comment = commentsRef.current.find(c => c.id === highlight.dataset.commentId)
      if (comment) {
        setActiveComment(comment)
        // Calculate position relative to container for absolute positioning
        const containerRect = containerRef.current?.getBoundingClientRect()
        const relativeX = e.clientX - (containerRect?.left || 0)
        const relativeY = e.clientY - (containerRect?.top || 0)
        setActiveCommentRect({
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

  // Find and highlight comments in the rendered content
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

      // Add new highlights for active comments
      const activeComments = comments.filter((c) => c.status === 'active' && c.position?.found)

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

      for (const comment of activeComments) {
        // Determine which occurrence this comment refers to using the position from server
        // Server returns byte offset in the body, we need to find the same Nth occurrence in rendered content
        const startByteOffset = comment.position!.start_offset
        const startCharOffset = byteOffsetToCharOffset(body, startByteOffset)

        // Count how many occurrences of target_text appear before this position in body
        const bodyBeforeTarget = body.substring(0, startCharOffset)
        const occurrencesBeforeInBody = findAllOccurrences(bodyBeforeTarget, comment.target_text).length

        // Find all occurrences in rendered content
        const occurrencesInRendered = findAllOccurrences(fullRenderedText, comment.target_text)

        // Get the position of the same Nth occurrence in rendered content
        // If there are fewer occurrences in rendered than in body, use the last one
        const targetOccurrenceIndex = Math.min(occurrencesBeforeInBody, occurrencesInRendered.length - 1)
        if (targetOccurrenceIndex < 0 || occurrencesInRendered.length === 0) {
          continue // Can't find this occurrence in rendered content
        }
        const targetPosInRendered = occurrencesInRendered[targetOccurrenceIndex]

        let highlighted = false
        const textNodes = getTextNodesWithPositions(contentRef.current!)

        // Find the text node that contains this position
        for (const { node: textNode, start: nodeStart, end: nodeEnd } of textNodes) {
          if (highlighted) break

          // Check if this node contains our target position
          if (targetPosInRendered >= nodeStart && targetPosInRendered < nodeEnd) {
            const text = textNode.textContent || ''
            const indexInNode = targetPosInRendered - nodeStart

            // Verify the text matches at this position
            if (text.substring(indexInNode, indexInNode + comment.target_text.length) !== comment.target_text) {
              continue
            }

            if (textNode.parentNode) {
              try {
                const before = text.substring(0, indexInNode)
                const target = text.substring(indexInNode, indexInNode + comment.target_text.length)
                const after = text.substring(indexInNode + comment.target_text.length)

                const highlight = document.createElement('mark')
                highlight.className = 'comment-highlight bg-yellow-100 hover:bg-yellow-200 cursor-pointer rounded px-0.5 transition-colors'
                highlight.dataset.commentId = comment.id
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

        // If not found in single node, try cross-element matching using the same calculated position
        if (!highlighted && targetPosInRendered >= 0) {
          // Re-get text nodes (DOM might have changed)
          const freshTextNodes = getTextNodesWithPositions(contentRef.current!)
          const matchEnd = targetPosInRendered + comment.target_text.length

          // Find all text nodes that overlap with our match
          for (const { node: textNode, start, end } of freshTextNodes) {
            if (end <= targetPosInRendered || start >= matchEnd) continue // No overlap
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
              highlight.dataset.commentId = comment.id
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

    // Apply highlights after a small delay to ensure DOM is stable
    const timer = setTimeout(applyHighlights, 10)
    return () => clearTimeout(timer)
  }, [comments, body, byteOffsetToCharOffset, findAllOccurrences])

  // Get active comments count
  const activeCommentsCount = comments.filter((c) => c.status === 'active').length

  return (
    <div ref={containerRef} className="relative">
      <CopyButton text={body} className="absolute top-0 right-0 z-10" />

      {/* Comments indicator */}
      {activeCommentsCount > 0 && (
        <div className="absolute top-0 right-12 flex items-center gap-1 text-xs text-yellow-600 bg-yellow-50 px-2 py-1 rounded">
          <MessageSquare className="h-3 w-3" />
          {activeCommentsCount} comment{activeCommentsCount > 1 ? 's' : ''}
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
            top: selection.rect.top,
            left: selection.rect.left + 12,
            transform: 'translateY(-50%)',
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

      {/* Active comment popup */}
      {activeComment && activeCommentRect && (
        <div
          className="absolute z-[5] bg-white border border-gray-200 rounded-lg shadow-xl p-4 w-80"
          style={{
            top: Math.max(activeCommentRect.top, 16),
            left: activeCommentRect.left + 12,
            transform: 'translateY(-50%)',
          }}
        >
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2 text-sm">
              <User className="h-4 w-4 text-gray-500" />
              <span className="font-medium">{activeComment.user_name}</span>
              <span className="text-gray-400 text-xs">
                {formatDistanceToNow(new Date(activeComment.created_at), { addSuffix: true })}
              </span>
            </div>
            <Button variant="ghost" size="sm" onClick={() => setActiveComment(null)} className="!p-1">
              <X className="h-4 w-4" />
            </Button>
          </div>
          <div className="text-sm bg-gray-50 border-l-2 border-gray-300 px-2 py-1 mb-2 italic text-gray-600">
            "{activeComment.target_text}"
          </div>
          <div className="text-sm text-gray-700 whitespace-pre-wrap mb-3">
            {activeComment.content}
          </div>
          <div className="flex justify-end gap-2">
            {user?.id === activeComment.user_id && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  if (confirm('Delete this comment?')) {
                    deleteMutation.mutate(activeComment.id)
                  }
                }}
                disabled={deleteMutation.isPending}
                className="text-red-500 hover:text-red-700"
              >
                <Trash2 className="mr-1 h-4 w-4" />
                Delete
              </Button>
            )}
            {user && activeComment.status === 'active' && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => resolveMutation.mutate(activeComment.id)}
                disabled={resolveMutation.isPending}
              >
                <CheckCircle className="mr-1 h-4 w-4" />
                Resolve
              </Button>
            )}
          </div>
        </div>
      )}

      {/* Backdrop for comment popup */}
      {activeComment && (
        <div
          className="fixed inset-0 z-[4]"
          onClick={() => setActiveComment(null)}
        />
      )}
    </div>
  )
}
