import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { MessageSquare, Plus } from 'lucide-react'
import { PlanCommentItem } from './PlanCommentItem'
import { Button } from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import { Spinner } from '@/components/ui/Spinner'
import { useAuth } from '@/hooks/useAuth'
import * as commentsApi from '@/api/plan-comments'
import type { PlanComment } from '@/types/plan-document'

interface PlanCommentsListProps {
  planId: string
  comments: PlanComment[]
  isLoading: boolean
}

export function PlanCommentsList({ planId, comments, isLoading }: PlanCommentsListProps) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [showAddForm, setShowAddForm] = useState(false)
  const [newContent, setNewContent] = useState('')
  const [selectedText, setSelectedText] = useState('')

  const updateMutation = useMutation({
    mutationFn: ({ commentId, content }: { commentId: string; content: string }) =>
      commentsApi.updatePlanComment(planId, commentId, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (commentId: string) => commentsApi.deletePlanComment(planId, commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const resolveMutation = useMutation({
    mutationFn: (commentId: string) => commentsApi.resolvePlanComment(planId, commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const createMutation = useMutation({
    mutationFn: () =>
      commentsApi.createPlanComment(planId, {
        target_text: selectedText,
        context_before: '',
        context_after: '',
        content: newContent,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
      setShowAddForm(false)
      setNewContent('')
      setSelectedText('')
    },
  })

  const handleUpdate = async (commentId: string, content: string) => {
    await updateMutation.mutateAsync({ commentId, content })
  }

  const handleDelete = async (commentId: string) => {
    await deleteMutation.mutateAsync(commentId)
  }

  const handleResolve = async (commentId: string) => {
    await resolveMutation.mutateAsync(commentId)
  }

  const handleSubmitNew = () => {
    if (selectedText.trim() && newContent.trim()) {
      createMutation.mutate()
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  // Group comments by status
  const activeComments = comments.filter((c) => c.status === 'active')
  const resolvedComments = comments.filter((c) => c.status === 'resolved')
  const outdatedComments = comments.filter((c) => c.status === 'outdated')

  return (
    <div className="space-y-6">
      {/* Add comment section */}
      {user && (
        <div className="rounded-xl border border-gray-200 bg-white p-4">
          {showAddForm ? (
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Target text (copy from the document)
                </label>
                <Textarea
                  value={selectedText}
                  onChange={(e) => setSelectedText(e.target.value)}
                  placeholder="Paste the text you want to comment on..."
                  rows={2}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Comment
                </label>
                <Textarea
                  value={newContent}
                  onChange={(e) => setNewContent(e.target.value)}
                  placeholder="Write your comment..."
                  rows={3}
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button
                  variant="ghost"
                  onClick={() => {
                    setShowAddForm(false)
                    setNewContent('')
                    setSelectedText('')
                  }}
                  disabled={createMutation.isPending}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleSubmitNew}
                  disabled={!selectedText.trim() || !newContent.trim() || createMutation.isPending}
                  loading={createMutation.isPending}
                >
                  Add Comment
                </Button>
              </div>
            </div>
          ) : (
            <Button variant="secondary" onClick={() => setShowAddForm(true)}>
              <Plus className="mr-1 h-4 w-4" />
              Add Comment
            </Button>
          )}
        </div>
      )}

      {/* Active comments */}
      {activeComments.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-gray-700 flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            Active Comments ({activeComments.length})
          </h3>
          {activeComments.map((comment) => (
            <PlanCommentItem
              key={comment.id}
              comment={comment}
              currentUserId={user?.id}
              onUpdate={handleUpdate}
              onDelete={handleDelete}
              onResolve={handleResolve}
            />
          ))}
        </div>
      )}

      {/* Resolved comments */}
      {resolvedComments.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-gray-500 flex items-center gap-2">
            Resolved ({resolvedComments.length})
          </h3>
          {resolvedComments.map((comment) => (
            <PlanCommentItem
              key={comment.id}
              comment={comment}
              currentUserId={user?.id}
            />
          ))}
        </div>
      )}

      {/* Outdated comments */}
      {outdatedComments.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-gray-400 flex items-center gap-2">
            Outdated ({outdatedComments.length})
          </h3>
          {outdatedComments.map((comment) => (
            <PlanCommentItem
              key={comment.id}
              comment={comment}
              currentUserId={user?.id}
            />
          ))}
        </div>
      )}

      {/* Empty state */}
      {comments.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          <MessageSquare className="h-8 w-8 mx-auto mb-2 opacity-50" />
          <p>No comments yet.</p>
          {user && <p className="text-sm">Be the first to add a comment!</p>}
        </div>
      )}
    </div>
  )
}
