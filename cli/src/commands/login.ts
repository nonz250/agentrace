import { createWebSession } from "../utils/http.js";
import * as readline from "node:readline";

export async function loginCommand(): Promise<void> {
  console.log("Creating login session...\n");

  const result = await createWebSession();

  if (!result.ok) {
    console.error(`Error: ${result.error}`);
    process.exit(1);
  }

  console.log(`Login URL: ${result.data.url}\n`);
  console.log("Press Enter to open in browser, or Ctrl+C to cancel.\n");

  // Wait for Enter key
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  await new Promise<void>((resolve) => {
    rl.question("", () => {
      rl.close();
      resolve();
    });
  });

  // Open browser
  const { exec } = await import("node:child_process");
  const url = result.data.url;

  const command =
    process.platform === "darwin"
      ? `open "${url}"`
      : process.platform === "win32"
        ? `start "${url}"`
        : `xdg-open "${url}"`;

  exec(command, (error) => {
    if (error) {
      console.error(`Failed to open browser: ${error.message}`);
      console.log(`Please open this URL manually: ${url}`);
    } else {
      console.log("Opened in browser");
    }
  });
}
