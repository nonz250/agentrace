import * as path from "node:path";
import { fileURLToPath } from "node:url";
import { installHooks, installMcpServer, installPreToolUseHook } from "../hooks/installer.js";
import { loadConfig } from "../config/manager.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export interface OnOptions {
  dev?: boolean;
}

export async function onCommand(options: OnOptions = {}): Promise<void> {
  // Check if config exists
  const config = loadConfig();
  if (!config) {
    console.log("Agentrace is not configured. Run 'npx agentrace init' first.");
    return;
  }

  // Determine hook command
  let hookCommand: string | undefined;
  if (options.dev) {
    // Use local CLI path for development
    const cliRoot = path.resolve(__dirname, "../..");
    const indexPath = path.join(cliRoot, "src/index.ts");
    hookCommand = `npx tsx ${indexPath} send`;
  }

  const result = installHooks({ command: hookCommand });
  if (result.success) {
    console.log(`✓ Hooks enabled. Session data will be sent to ${config.server_url}`);
  } else {
    console.error(`✗ ${result.message}`);
  }

  // Install MCP server
  let mcpCommand: string | undefined;
  let mcpArgs: string[] | undefined;
  if (options.dev) {
    const cliRoot = path.resolve(__dirname, "../..");
    const indexPath = path.join(cliRoot, "src/index.ts");
    mcpCommand = "npx";
    mcpArgs = ["tsx", indexPath, "mcp-server"];
  }
  const mcpResult = installMcpServer({ command: mcpCommand, args: mcpArgs });
  if (mcpResult.success) {
    console.log(`✓ ${mcpResult.message}`);
  } else {
    console.error(`✗ ${mcpResult.message}`);
  }

  // Install PreToolUse hook for session_id injection
  const preToolUseResult = installPreToolUseHook();
  if (preToolUseResult.success) {
    console.log(`✓ ${preToolUseResult.message}`);
  } else {
    console.error(`✗ ${preToolUseResult.message}`);
  }
}
