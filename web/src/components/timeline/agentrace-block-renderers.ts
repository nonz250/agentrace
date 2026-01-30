import type { DisplayBlock } from 'cc-transcript-react'
import { AGENTRACE_MCP_TOOLS } from './extract-plan-links'
import { AgentraceToolBlock } from './AgentraceToolBlock'
import { createElement } from 'react'

/** Build customBlockRenderers map for agentrace MCP tools. */
export function buildAgentraceBlockRenderers(): Record<string, (block: DisplayBlock) => React.ReactNode | null> {
  const renderers: Record<string, (block: DisplayBlock) => React.ReactNode | null> = {}
  for (const toolName of AGENTRACE_MCP_TOOLS) {
    renderers[toolName] = (block) => createElement(AgentraceToolBlock, { block })
  }
  return renderers
}
