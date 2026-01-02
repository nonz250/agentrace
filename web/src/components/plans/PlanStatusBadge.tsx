import type { PlanDocumentStatus } from '@/types/plan-document'

interface PlanStatusBadgeProps {
  status: PlanDocumentStatus
}

const statusConfig: Record<PlanDocumentStatus, { label: string; className: string }> = {
  scratch: {
    label: 'Scratch',
    className: 'bg-orange-100 text-orange-700',
  },
  draft: {
    label: 'Draft',
    className: 'bg-gray-100 text-gray-600',
  },
  planning: {
    label: 'Planning',
    className: 'bg-blue-100 text-blue-700',
  },
  pending: {
    label: 'Pending',
    className: 'bg-yellow-100 text-yellow-700',
  },
  implementation: {
    label: 'Implementation',
    className: 'bg-purple-100 text-purple-700',
  },
  complete: {
    label: 'Complete',
    className: 'bg-green-100 text-green-700',
  },
}

export function PlanStatusBadge({ status }: PlanStatusBadgeProps) {
  const config = statusConfig[status] || statusConfig.draft

  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${config.className}`}
    >
      {config.label}
    </span>
  )
}
