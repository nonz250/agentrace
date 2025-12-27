import inquirer from "inquirer";
import { saveConfig, getConfigPath } from "../config/manager.js";
import { installHooks } from "../hooks/installer.js";

export async function initCommand(): Promise<void> {
  console.log("Agentrace Setup\n");

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

  // Install hooks
  const hookResult = installHooks();
  if (hookResult.success) {
    console.log(`✓ ${hookResult.message}`);
  } else {
    console.error(`✗ ${hookResult.message}`);
  }

  console.log("\nSetup complete!");
}
