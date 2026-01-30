export interface PlanLinkInfo {
  id: string
  changedStatus?: string
}

export const AGENTRACE_MCP_TOOLS = [
  'mcp__agentrace__search_plans',
  'mcp__agentrace__read_plan',
  'mcp__agentrace__create_plan',
  'mcp__agentrace__update_plan',
  'mcp__agentrace__set_plan_status',
]

export function extractToolResultText(resultContent: unknown): string | null {
  if (!resultContent) return null
  const content = resultContent as Record<string, unknown>
  const data = content?.content
  if (typeof data === 'string') return data
  if (Array.isArray(data)) {
    const textParts = data
      .map((c) => {
        if (typeof c === 'string') return c
        if (c?.type === 'text' && typeof c.text === 'string') return c.text
        return null
      })
      .filter(Boolean)
    return textParts.length > 0 ? textParts.join('\n') : null
  }
  return null
}

export function extractPlanLinks(
  toolName: string,
  input: Record<string, unknown> | undefined,
  resultText: string | null
): PlanLinkInfo[] {
  const tool = toolName.replace('mcp__agentrace__', '')

  switch (tool) {
    case 'search_plans': {
      if (!resultText) return []
      if (resultText.includes('No plans found')) return []
      try {
        const plans = JSON.parse(resultText)
        if (!Array.isArray(plans)) return []
        return plans.map((p: { id: string }) => ({ id: p.id }))
      } catch {
        return []
      }
    }
    case 'read_plan':
      if (input?.id) return [{ id: input.id as string }]
      return []
    case 'create_plan':
    case 'update_plan': {
      if (!resultText) {
        if (input?.id) return [{ id: input.id as string }]
        return []
      }
      const idMatch = resultText.match(/ID:\s*([^\n]+)/)
      if (idMatch) return [{ id: idMatch[1].trim() }]
      if (input?.id) return [{ id: input.id as string }]
      return []
    }
    case 'set_plan_status': {
      const changedStatus = input?.status as string | undefined
      if (!resultText) {
        if (input?.id) return [{ id: input.id as string, changedStatus }]
        return []
      }
      const idMatch = resultText.match(/ID:\s*([^\n]+)/)
      if (idMatch) return [{ id: idMatch[1].trim(), changedStatus }]
      if (input?.id) return [{ id: input.id as string, changedStatus }]
      return []
    }
    default:
      return []
  }
}
