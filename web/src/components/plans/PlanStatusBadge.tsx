import type { PlanDocumentStatus } from '@/types/plan-document'
import { statusConfig } from '@/lib/plan-status'

interface PlanStatusBadgeProps {
  status: PlanDocumentStatus
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
