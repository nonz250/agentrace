import { useNavigate } from 'react-router-dom'
import { PlanCard } from './PlanCard'
import type { PlanDocument } from '@/types/plan-document'

interface PlanListProps {
  plans: PlanDocument[]
}

export function PlanList({ plans }: PlanListProps) {
  const navigate = useNavigate()

  if (plans.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center">
        <p className="text-gray-500">No plans yet.</p>
        <p className="mt-1 text-sm text-gray-400">
          Plans will appear here once created via Claude Code MCP tools.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {plans.map((plan) => (
        <PlanCard
          key={plan.id}
          plan={plan}
          onClick={() => navigate(`/projects/${plan.project?.id}/plans/${plan.id}`)}
        />
      ))}
    </div>
  )
}
