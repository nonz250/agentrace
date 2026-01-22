import { deleteConfig, deleteLocalConfig } from "../config/manager.js";
import { uninstallHooks, uninstallMcpServer, uninstallPreToolUseHook } from "../hooks/installer.js";

export interface UninstallOptions {
  local?: boolean;
}

export async function uninstallCommand(options: UninstallOptions = {}): Promise<void> {
  const projectDir = options.local ? process.cwd() : undefined;

  if (options.local) {
    console.log("Uninstalling AgenTrace (local settings only)...\n");
  } else {
    console.log("Uninstalling AgenTrace...\n");
  }

  // Remove hooks
  const hookResult = uninstallHooks({
    local: options.local,
    projectDir,
  });
  if (hookResult.success) {
    console.log(`✓ ${hookResult.message}`);
  } else {
    console.error(`✗ ${hookResult.message}`);
  }

  // Remove PreToolUse hook
  const preToolUseResult = uninstallPreToolUseHook({
    local: options.local,
    projectDir,
    // Don't remove the hook script when uninstalling local settings
    removeScript: !options.local,
  });
  if (preToolUseResult.success) {
    console.log(`✓ ${preToolUseResult.message}`);
  } else {
    console.error(`✗ ${preToolUseResult.message}`);
  }

  // Remove MCP server
  const mcpResult = uninstallMcpServer({
    local: options.local,
    projectDir,
  });
  if (mcpResult.success) {
    console.log(`✓ ${mcpResult.message}`);
  } else {
    console.error(`✗ ${mcpResult.message}`);
  }

  // Remove config
  if (options.local && projectDir) {
    // Remove local config directory
    const configRemoved = deleteLocalConfig(projectDir);
    if (configRemoved) {
      console.log("✓ Local config removed (.agentrace/)");
    } else {
      console.log("✓ No local config to remove");
    }
  } else {
    // Remove global config
    const configRemoved = deleteConfig();
    if (configRemoved) {
      console.log("✓ Config removed");
    } else {
      console.log("✓ No config to remove");
    }
  }

  console.log("\nUninstall complete!");
}
