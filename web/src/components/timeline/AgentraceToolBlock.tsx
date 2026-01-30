import type { DisplayBlock } from 'cc-transcript-react'
import { extractToolResultText, extractPlanLinks } from './extract-plan-links'
import { PlanLinkCard } from './PlanLinkCard'

export function AgentraceToolBlock({ block }: { block: DisplayBlock }) {
  const content = block.content as Record<string, unknown> | undefined
  const toolName = content?.name as string | undefined
  const input = content?.input as Record<string, unknown> | undefined

  const resultContent = block.toolResultBlock?.content as Record<string, unknown> | undefined
  const resultText = extractToolResultText(resultContent)
  const planLinks = toolName ? extractPlanLinks(toolName, input, resultText) : []

  if (planLinks.length === 0) {
    return (
      <div className="text-sm text-gray-500 italic">
        Operation completed
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {planLinks.map((plan) => (
        <PlanLinkCard key={plan.id} plan={plan} />
      ))}
    </div>
  )
}
