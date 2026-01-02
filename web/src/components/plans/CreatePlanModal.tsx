import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { Select } from '@/components/ui/Select'
import { Button } from '@/components/ui/Button'
import { createPlan } from '@/api/plan-documents'
import { getProjects } from '@/api/projects'

interface CreatePlanModalProps {
  open: boolean
  onClose: () => void
  onSuccess?: () => void
}

export function CreatePlanModal({ open, onClose, onSuccess }: CreatePlanModalProps) {
  const queryClient = useQueryClient()
  const [description, setDescription] = useState('')
  const [body, setBody] = useState('')
  const [projectId, setProjectId] = useState('')

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
