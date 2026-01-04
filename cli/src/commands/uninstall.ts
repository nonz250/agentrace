import { deleteConfig } from "../config/manager.js";
import { uninstallHooks, uninstallMcpServer, uninstallPreToolUseHook } from "../hooks/installer.js";

export async function uninstallCommand(): Promise<void> {
  console.log("Uninstalling AgenTrace...\n");

  // Remove hooks
  const hookResult = uninstallHooks();
  if (hookResult.success) {
    console.log(`✓ ${hookResult.message}`);
  } else {
    console.error(`✗ ${hookResult.message}`);
  }

  // Remove PreToolUse hook
  const preToolUseResult = uninstallPreToolUseHook();
  if (preToolUseResult.success) {
    console.log(`✓ ${preToolUseResult.message}`);
  } else {
    console.error(`✗ ${preToolUseResult.message}`);
  }

  // Remove MCP server
  const mcpResult = uninstallMcpServer();
  if (mcpResult.success) {
    console.log(`✓ ${mcpResult.message}`);
  } else {
    console.error(`✗ ${mcpResult.message}`);
  }

  // Remove config
  const configRemoved = deleteConfig();
  if (configRemoved) {
    console.log("✓ Config removed");
  } else {
    console.log("✓ No config to remove");
  }

  console.log("\nUninstall complete!");
}
