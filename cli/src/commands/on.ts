import * as path from "node:path";
import { fileURLToPath } from "node:url";
import { installHooks } from "../hooks/installer.js";
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
}
