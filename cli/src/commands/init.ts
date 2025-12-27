import * as path from "node:path";
import { fileURLToPath } from "node:url";
import inquirer from "inquirer";
import { saveConfig, getConfigPath } from "../config/manager.js";
import { installHooks } from "../hooks/installer.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export interface InitOptions {
  dev?: boolean;
}

export async function initCommand(options: InitOptions = {}): Promise<void> {
  console.log("Agentrace Setup\n");

  if (options.dev) {
    console.log("[Dev Mode] Using local CLI for hooks\n");
  }

  const answers = await inquirer.prompt([
    {
      type: "input",
      name: "serverUrl",
      message: "Server URL:",
      default: "http://localhost:8080",
      validate: (input: string) => {
        try {
          new URL(input);
          return true;
        } catch {
          return "Please enter a valid URL";
        }
      },
    },
    {
      type: "input",
      name: "apiKey",
      message: "API Key:",
      validate: (input: string) => {
        if (!input.trim()) {
          return "API Key is required";
        }
        return true;
      },
    },
  ]);

  // Save config
  saveConfig({
    server_url: answers.serverUrl,
    api_key: answers.apiKey,
  });
  console.log(`✓ Config saved to ${getConfigPath()}`);

  // Determine hook command
  let hookCommand: string | undefined;
  if (options.dev) {
    // Use local CLI path for development
    const cliRoot = path.resolve(__dirname, "../..");
    const indexPath = path.join(cliRoot, "src/index.ts");
    hookCommand = `npx tsx ${indexPath} send`;
    console.log(`  Hook command: ${hookCommand}`);
  }

  // Install hooks
  const hookResult = installHooks({ command: hookCommand });
  if (hookResult.success) {
    console.log(`✓ ${hookResult.message}`);
  } else {
    console.error(`✗ ${hookResult.message}`);
  }

  console.log("\nSetup complete!");
}
