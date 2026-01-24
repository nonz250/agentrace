import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { User, Pencil, Trash2, Check, X, CheckCircle } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import type { PlanComment } from '@/types/plan-document'

interface PlanCommentItemProps {
  comment: PlanComment
  currentUserId?: string
  onUpdate?: (commentId: string, content: string) => Promise<void>
  onDelete?: (commentId: string) => Promise<void>
  onResolve?: (commentId: string) => Promise<void>
}

export function PlanCommentItem({
  comment,
  currentUserId,
  onUpdate,
  onDelete,
  onResolve,
}: PlanCommentItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editContent, setEditContent] = useState(comment.content)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const isOwner = currentUserId === comment.user_id
  const isActive = comment.status === 'active'
  const isResolved = comment.status === 'resolved'
  const isOutdated = comment.status === 'outdated'

  const handleEdit = () => {
    setEditContent(comment.content)
    setIsEditing(true)
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
    setEditContent(comment.content)
  }

  const handleSaveEdit = async () => {
    if (!onUpdate || !editContent.trim()) return
    setIsSubmitting(true)
    try {
      await onUpdate(comment.id, editContent)
      setIsEditing(false)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (!onDelete) return
    if (!confirm('Are you sure you want to delete this comment?')) return
    setIsSubmitting(true)
    try {
      await onDelete(comment.id)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleResolve = async () => {
    if (!onResolve) return
    setIsSubmitting(true)
    try {
      await onResolve(comment.id)
    } finally {
      setIsSubmitting(false)
    }
  }

  const relativeTime = formatDistanceToNow(new Date(comment.created_at), { addSuffix: true })

  return (
    <div
      className={`rounded-lg border p-3 ${
        isOutdated
          ? 'border-gray-200 bg-gray-50 opacity-60'
          : isResolved
          ? 'border-green-200 bg-green-50'
          : 'border-blue-200 bg-blue-50'
      }`}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2 text-sm">
          <User className="h-4 w-4 text-gray-500" />
          <span className="font-medium text-gray-700">{comment.user_name}</span>
          <span className="text-gray-400">{relativeTime}</span>
          {isResolved && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs bg-green-100 text-green-700">
              <CheckCircle className="h-3 w-3" />
              Resolved
            </span>
          )}
          {isOutdated && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs bg-gray-200 text-gray-600">
              Outdated
            </span>
          )}
        </div>
        {isOwner && isActive && !isEditing && (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" onClick={handleEdit} disabled={isSubmitting} className="!p-1">
              <Pencil className="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="sm" onClick={handleDelete} disabled={isSubmitting} className="!p-1 text-red-500 hover:text-red-700">
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        )}
      </div>

      {/* Target text quote */}
      <div className="mb-2 px-2 py-1 border-l-2 border-gray-300 bg-white/50 text-sm text-gray-600 italic">
        "{comment.target_text}"
        {isOutdated && (
          <span className="block mt-1 text-xs text-orange-600 not-italic">
            This text has been modified or removed from the document.
          </span>
        )}
      </div>

      {/* Content */}
      {isEditing ? (
        <div className="space-y-2">
          <Textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            rows={3}
            className="text-sm"
          />
          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="sm" onClick={handleCancelEdit} disabled={isSubmitting}>
              <X className="mr-1 h-3.5 w-3.5" />
              Cancel
            </Button>
            <Button size="sm" onClick={handleSaveEdit} disabled={isSubmitting || !editContent.trim()}>
              <Check className="mr-1 h-3.5 w-3.5" />
              Save
            </Button>
          </div>
        </div>
      ) : (
        <div className="text-sm text-gray-700 whitespace-pre-wrap">{comment.content}</div>
      )}

      {/* Resolve button */}
      {isActive && !isEditing && onResolve && (
        <div className="mt-2 flex justify-end">
          <Button variant="ghost" size="sm" onClick={handleResolve} disabled={isSubmitting}>
            <CheckCircle className="mr-1 h-3.5 w-3.5" />
            Resolve
          </Button>
        </div>
      )}
    </div>
  )
}
