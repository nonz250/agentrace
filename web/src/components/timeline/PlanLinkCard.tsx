import { FileText, ExternalLink, Loader2, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getPlan } from '@/api/plan-documents'
import { PlanStatusBadge } from '@/components/plans/PlanStatusBadge'
import type { PlanDocumentStatus } from '@/types/plan-document'
import type { PlanLinkInfo } from './extract-plan-links'

export function PlanLinkCard({ plan }: { plan: PlanLinkInfo }) {
  const { data: planData, isLoading, isError } = useQuery({
    queryKey: ['plan', plan.id],
    queryFn: () => getPlan(plan.id),
    staleTime: 30 * 1000,
  })

  const shortId = plan.id.length > 8 ? plan.id.slice(0, 8) + '...' : plan.id

  if (isLoading) {
    return (
      <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-3">
        <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
        <span className="text-sm text-gray-500">Loading plan...</span>
      </div>
    )
  }

  if (isError || !planData) {
    return (
      <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-3">
        <FileText className="h-4 w-4 text-gray-400" />
        <span className="text-sm text-gray-500">Plan {shortId}</span>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {plan.changedStatus && (
        <div className="flex items-center gap-2 text-sm text-gray-600">
          <span>Status changed</span>
          <ArrowRight className="h-3 w-3 text-gray-400" />
          <PlanStatusBadge status={plan.changedStatus as PlanDocumentStatus} />
        </div>
      )}
      <Link
        to={`/plans/${plan.id}`}
        className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3 transition-colors hover:border-gray-300 hover:bg-gray-50"
      >
        <div className="flex items-center gap-3 min-w-0">
          <FileText className="h-4 w-4 flex-shrink-0 text-gray-400" />
          <div className="flex items-center gap-2 min-w-0">
            <span className="truncate text-sm font-medium text-gray-900">
              {planData.description}
            </span>
            <PlanStatusBadge status={planData.status} />
          </div>
        </div>
        <ExternalLink className="h-4 w-4 flex-shrink-0 text-gray-400" />
      </Link>
    </div>
  )
}
