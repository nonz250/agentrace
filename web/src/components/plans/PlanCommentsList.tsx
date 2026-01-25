import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { MessageSquare, Plus } from 'lucide-react'
import { PlanCommentThread as PlanCommentThreadComponent } from './PlanCommentThread'
import { Button } from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import { Spinner } from '@/components/ui/Spinner'
import { useAuth } from '@/hooks/useAuth'
import * as commentsApi from '@/api/plan-comments'
import type { PlanCommentThread } from '@/types/plan-document'

interface PlanCommentsListProps {
  planId: string
  threads: PlanCommentThread[]
  isLoading: boolean
}

export function PlanCommentsList({ planId, threads, isLoading }: PlanCommentsListProps) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [showAddForm, setShowAddForm] = useState(false)
  const [newContent, setNewContent] = useState('')
  const [selectedText, setSelectedText] = useState('')

  const createThreadMutation = useMutation({
    mutationFn: () =>
      commentsApi.createPlanThread(planId, {
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

  const deleteThreadMutation = useMutation({
    mutationFn: (threadId: string) => commentsApi.deletePlanThread(planId, threadId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const resolveThreadMutation = useMutation({
    mutationFn: (threadId: string) => commentsApi.resolvePlanThread(planId, threadId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const addMessageMutation = useMutation({
    mutationFn: ({ threadId, content }: { threadId: string; content: string }) =>
      commentsApi.addThreadMessage(planId, threadId, { content }),
    onSuccess: (newMessage, { threadId }) => {
      // Optimistically update the cache with the new message
      queryClient.setQueryData<PlanCommentThread[]>(['plan', planId, 'comments'], (oldThreads) => {
        if (!oldThreads || !newMessage) return oldThreads
        return oldThreads.map((thread) =>
          thread.id === threadId
            ? { ...thread, messages: [...thread.messages, newMessage] }
            : thread
        )
      })
      // Also invalidate to ensure consistency with server
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const updateMessageMutation = useMutation({
    mutationFn: ({ threadId, messageId, content }: { threadId: string; messageId: string; content: string }) =>
      commentsApi.updateThreadMessage(planId, threadId, messageId, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const deleteMessageMutation = useMutation({
    mutationFn: ({ threadId, messageId }: { threadId: string; messageId: string }) =>
      commentsApi.deleteThreadMessage(planId, threadId, messageId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plan', planId, 'comments'] })
    },
  })

  const handleResolve = async (threadId: string) => {
    await resolveThreadMutation.mutateAsync(threadId)
  }

  const handleDeleteThread = async (threadId: string) => {
    await deleteThreadMutation.mutateAsync(threadId)
  }

  const handleAddMessage = async (threadId: string, content: string) => {
    await addMessageMutation.mutateAsync({ threadId, content })
  }

  const handleUpdateMessage = async (threadId: string, messageId: string, content: string) => {
    await updateMessageMutation.mutateAsync({ threadId, messageId, content })
  }

  const handleDeleteMessage = async (threadId: string, messageId: string) => {
    await deleteMessageMutation.mutateAsync({ threadId, messageId })
  }

  const handleSubmitNew = () => {
    if (selectedText.trim() && newContent.trim()) {
      createThreadMutation.mutate()
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  // Group threads by status
  const activeThreads = threads.filter((t) => t.status === 'active')
  const resolvedThreads = threads.filter((t) => t.status === 'resolved')
  const outdatedThreads = threads.filter((t) => t.status === 'outdated')

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
                  disabled={createThreadMutation.isPending}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleSubmitNew}
                  disabled={!selectedText.trim() || !newContent.trim() || createThreadMutation.isPending}
                  loading={createThreadMutation.isPending}
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

      {/* Active threads */}
      {activeThreads.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-gray-700 flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            Active Threads ({activeThreads.length})
          </h3>
          {activeThreads.map((thread) => (
            <PlanCommentThreadComponent
              key={thread.id}
              thread={thread}
              currentUserId={user?.id}
              onResolve={handleResolve}
              onDeleteThread={handleDeleteThread}
              onAddMessage={handleAddMessage}
              onUpdateMessage={handleUpdateMessage}
              onDeleteMessage={handleDeleteMessage}
            />
          ))}
        </div>
      )}

      {/* Resolved threads */}
      {resolvedThreads.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-gray-500 flex items-center gap-2">
            Resolved ({resolvedThreads.length})
          </h3>
          {resolvedThreads.map((thread) => (
            <PlanCommentThreadComponent
              key={thread.id}
              thread={thread}
              currentUserId={user?.id}
            />
          ))}
        </div>
      )}

      {/* Outdated threads */}
      {outdatedThreads.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-medium text-gray-400 flex items-center gap-2">
            Outdated ({outdatedThreads.length})
          </h3>
          {outdatedThreads.map((thread) => (
            <PlanCommentThreadComponent
              key={thread.id}
              thread={thread}
              currentUserId={user?.id}
            />
          ))}
        </div>
      )}

      {/* Empty state */}
      {threads.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          <MessageSquare className="h-8 w-8 mx-auto mb-2 opacity-50" />
          <p>No comments yet.</p>
          {user && <p className="text-sm">Be the first to add a comment!</p>}
        </div>
      )}
    </div>
  )
}
