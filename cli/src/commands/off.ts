import { uninstallHooks, uninstallMcpServer } from "../hooks/installer.js";
import { loadConfig } from "../config/manager.js";

export async function offCommand(): Promise<void> {
  // Check if config exists
  const config = loadConfig();
  if (!config) {
    console.log("Agentrace is not configured. Run 'npx agentrace init' first.");
    return;
  }

  const result = uninstallHooks();
  if (result.success) {
    console.log(`✓ Hooks disabled. Your credentials are still saved.`);
    console.log(`  Run 'npx agentrace on' to re-enable.`);
  } else {
    console.error(`✗ ${result.message}`);
  }

  const mcpResult = uninstallMcpServer();
  if (mcpResult.success) {
    console.log(`✓ ${mcpResult.message}`);
  } else {
    console.error(`✗ ${mcpResult.message}`);
  }
}
