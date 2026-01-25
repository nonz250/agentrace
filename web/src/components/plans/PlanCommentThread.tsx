import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { User, Pencil, Trash2, Check, X, CheckCircle, MessageCircle, Send, MoreVertical } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import type { PlanCommentThread, PlanCommentMessage } from '@/types/plan-document'

interface PlanCommentThreadProps {
  thread: PlanCommentThread
  currentUserId?: string
  onResolve?: (threadId: string) => Promise<void>
  onDeleteThread?: (threadId: string) => Promise<void>
  onAddMessage?: (threadId: string, content: string) => Promise<void>
  onUpdateMessage?: (threadId: string, messageId: string, content: string) => Promise<void>
  onDeleteMessage?: (threadId: string, messageId: string) => Promise<void>
}

interface MessageItemProps {
  message: PlanCommentMessage
  isOwner: boolean
  isActive: boolean
  onUpdate?: (content: string) => Promise<void>
  onDelete?: () => Promise<void>
}

function MessageItem({ message, isOwner, isActive, onUpdate, onDelete }: MessageItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editContent, setEditContent] = useState(message.content)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleEdit = () => {
    setEditContent(message.content)
    setIsEditing(true)
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
    setEditContent(message.content)
  }

  const handleSaveEdit = async () => {
    if (!onUpdate || !editContent.trim()) return
    setIsSubmitting(true)
    try {
      await onUpdate(editContent)
      setIsEditing(false)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (!onDelete) return
    if (!confirm('Are you sure you want to delete this message?')) return
    setIsSubmitting(true)
    try {
      await onDelete()
    } finally {
      setIsSubmitting(false)
    }
  }

  const relativeTime = formatDistanceToNow(new Date(message.created_at), { addSuffix: true })

  return (
    <div className="py-2 border-t border-gray-200 first:border-t-0">
      <div className="flex items-center justify-between mb-1">
        <div className="flex items-center gap-2 text-sm">
          <User className="h-3.5 w-3.5 text-gray-500" />
          <span className="font-medium text-gray-700">{message.user_name}</span>
          <span className="text-gray-400 text-xs">{relativeTime}</span>
        </div>
        {isOwner && isActive && !isEditing && (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" onClick={handleEdit} disabled={isSubmitting} className="!p-1">
              <Pencil className="h-3 w-3" />
            </Button>
            <Button variant="ghost" size="sm" onClick={handleDelete} disabled={isSubmitting} className="!p-1 text-red-500 hover:text-red-700">
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        )}
      </div>

      {isEditing ? (
        <div className="space-y-2">
          <Textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            rows={2}
            className="text-sm"
          />
          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="sm" onClick={handleCancelEdit} disabled={isSubmitting}>
              <X className="mr-1 h-3 w-3" />
              Cancel
            </Button>
            <Button size="sm" onClick={handleSaveEdit} disabled={isSubmitting || !editContent.trim()}>
              <Check className="mr-1 h-3 w-3" />
              Save
            </Button>
          </div>
        </div>
      ) : (
        <div className="text-sm text-gray-700 whitespace-pre-wrap pl-5">{message.content}</div>
      )}
    </div>
  )
}

export function PlanCommentThread({
  thread,
  currentUserId,
  onResolve,
  onDeleteThread,
  onAddMessage,
  onUpdateMessage,
  onDeleteMessage,
}: PlanCommentThreadProps) {
  const [replyContent, setReplyContent] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [showReplyForm, setShowReplyForm] = useState(false)
  const [showThreadMenu, setShowThreadMenu] = useState(false)

  const isActive = thread.status === 'active'
  const isResolved = thread.status === 'resolved'
  const isOutdated = thread.status === 'outdated'

  // First message creator is the thread owner
  const firstMessage = thread.messages[0]
  const isThreadOwner = currentUserId && firstMessage && currentUserId === firstMessage.user_id

  const handleResolve = async () => {
    if (!onResolve) return
    setIsSubmitting(true)
    try {
      await onResolve(thread.id)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDeleteThread = async () => {
    if (!onDeleteThread) return
    if (!confirm('Are you sure you want to delete this entire thread?')) return
    setIsSubmitting(true)
    try {
      await onDeleteThread(thread.id)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleAddReply = async () => {
    if (!onAddMessage || !replyContent.trim()) return
    setIsSubmitting(true)
    try {
      await onAddMessage(thread.id, replyContent)
      setReplyContent('')
      setShowReplyForm(false)
    } finally {
      setIsSubmitting(false)
    }
  }

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
      {/* Header with quote and actions */}
      <div className="flex items-start justify-between gap-2 mb-2">
        <div className="flex-1">
          {isResolved && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs bg-green-100 text-green-700 mb-1">
              <CheckCircle className="h-3 w-3" />
              Resolved
            </span>
          )}
          {isOutdated && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs bg-gray-200 text-gray-600 mb-1">
              Outdated
            </span>
          )}
          <div className="px-2 py-1 border-l-2 border-gray-300 bg-white/50 text-sm text-gray-600 italic">
            "{thread.target_text}"
            {isOutdated && (
              <span className="block mt-1 text-xs text-orange-600 not-italic">
                This text has been modified or removed from the document.
              </span>
            )}
          </div>
        </div>
        {isThreadOwner && isActive && (
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
                      handleDeleteThread()
                    }}
                    disabled={isSubmitting}
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
      <div className="space-y-0">
        {thread.messages.map((message) => (
          <MessageItem
            key={message.id}
            message={message}
            isOwner={currentUserId === message.user_id}
            isActive={isActive}
            onUpdate={onUpdateMessage ? (content) => onUpdateMessage(thread.id, message.id, content) : undefined}
            onDelete={onDeleteMessage ? () => onDeleteMessage(thread.id, message.id) : undefined}
          />
        ))}
      </div>

      {/* Reply form or button */}
      {isActive && onAddMessage && (
        <div className="mt-2 pt-2 border-t border-gray-200">
          {showReplyForm ? (
            <div className="space-y-2">
              <Textarea
                value={replyContent}
                onChange={(e) => setReplyContent(e.target.value)}
                placeholder="Write a reply..."
                rows={2}
                className="text-sm"
              />
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="sm" onClick={() => { setShowReplyForm(false); setReplyContent('') }} disabled={isSubmitting}>
                  Cancel
                </Button>
                <Button size="sm" onClick={handleAddReply} disabled={isSubmitting || !replyContent.trim()}>
                  <Send className="mr-1 h-3 w-3" />
                  Reply
                </Button>
              </div>
            </div>
          ) : (
            <div className="flex justify-between">
              <Button variant="ghost" size="sm" onClick={() => setShowReplyForm(true)} disabled={isSubmitting}>
                <MessageCircle className="mr-1 h-3.5 w-3.5" />
                Reply
              </Button>
              {onResolve && (
                <Button variant="ghost" size="sm" onClick={handleResolve} disabled={isSubmitting}>
                  <CheckCircle className="mr-1 h-3.5 w-3.5" />
                  Resolve
                </Button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
