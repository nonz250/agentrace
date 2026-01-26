import { uninstallHooks, uninstallMcpServer, uninstallPreToolUseHook } from "../hooks/installer.js";
import { loadConfigWithFallback } from "../config/manager.js";

export interface OffOptions {
  local?: boolean;
}

export async function offCommand(options: OffOptions = {}): Promise<void> {
  const projectDir = options.local ? process.cwd() : undefined;

  // Check if config exists (local config takes precedence over global)
  const config = loadConfigWithFallback(process.cwd());
  if (!config) {
    console.log("AgenTrace is not configured. Run 'npx agentrace init' first.");
    return;
  }

  if (options.local) {
    console.log("[Local Mode] Disabling hooks/MCP for this project only\n");
  }

  const result = uninstallHooks({
    local: options.local,
    projectDir,
  });
  if (result.success) {
    console.log(`✓ Hooks disabled. Your credentials are still saved.`);
    console.log(`  Run 'npx agentrace on${options.local ? " --local" : ""}' to re-enable.`);
  } else {
    console.error(`✗ ${result.message}`);
  }

  // Remove PreToolUse hook
  const preToolUseResult = uninstallPreToolUseHook({
    local: options.local,
    projectDir,
  });
  if (preToolUseResult.success) {
    console.log(`✓ ${preToolUseResult.message}`);
  } else {
    console.error(`✗ ${preToolUseResult.message}`);
  }

  const mcpResult = uninstallMcpServer({
    local: options.local,
    projectDir,
  });
  if (mcpResult.success) {
    console.log(`✓ ${mcpResult.message}`);
  } else {
    console.error(`✗ ${mcpResult.message}`);
  }
}
