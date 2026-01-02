import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { Select } from '@/components/ui/Select'
import { Button } from '@/components/ui/Button'
import { createPlan } from '@/api/plan-documents'
import { getProjects } from '@/api/projects'
import type { PlanDocumentStatus } from '@/types/plan-document'

interface CreatePlanModalProps {
  open: boolean
  onClose: () => void
  onSuccess?: () => void
}

const STATUS_OPTIONS: { value: PlanDocumentStatus; label: string }[] = [
  { value: 'scratch', label: 'Scratch' },
  { value: 'draft', label: 'Draft' },
  { value: 'planning', label: 'Planning' },
  { value: 'pending', label: 'Pending' },
  { value: 'implementation', label: 'Implementation' },
  { value: 'complete', label: 'Complete' },
]

export function CreatePlanModal({ open, onClose, onSuccess }: CreatePlanModalProps) {
  const queryClient = useQueryClient()
  const [description, setDescription] = useState('')
  const [body, setBody] = useState('')
  const [projectId, setProjectId] = useState('')
  const [status, setStatus] = useState<PlanDocumentStatus>('scratch')

  const { data: projectsData } = useQuery({
    queryKey: ['projects'],
    queryFn: () => getProjects(),
    enabled: open,
  })

  const createMutation = useMutation({
    mutationFn: () => createPlan({
      description,
      body,
      project_id: projectId || undefined,
      status,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['plans'] })
      handleClose()
      onSuccess?.()
    },
  })

  const handleClose = () => {
    setDescription('')
    setBody('')
    setProjectId('')
    setStatus('scratch')
    onClose()
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (description.trim()) {
      createMutation.mutate()
    }
  }

  return (
    <Modal open={open} onClose={handleClose} title="Create Plan" className="max-w-2xl">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Select
          label="Project"
          value={projectId}
          onChange={(e) => setProjectId(e.target.value)}
        >
          <option value="">Select a project (optional)</option>
          {projectsData?.projects.map((project) => (
            <option key={project.id} value={project.id}>
              {project.canonical_git_repository || 'Default Project'}
            </option>
          ))}
        </Select>

        <Select
          label="Status"
          value={status}
          onChange={(e) => setStatus(e.target.value as PlanDocumentStatus)}
        >
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </Select>

        <Input
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Brief description of the plan"
          required
        />

        <Textarea
          label="Body"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Plan details in Markdown format"
          rows={10}
        />

        <div className="flex justify-end gap-3 pt-4">
          <Button type="button" variant="ghost" onClick={handleClose}>
            Cancel
          </Button>
          <Button
            type="submit"
            loading={createMutation.isPending}
            disabled={!description.trim()}
          >
            Create Plan
          </Button>
        </div>
      </form>
    </Modal>
  )
}
