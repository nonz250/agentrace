import { exec } from "node:child_process";
import { platform } from "node:os";

export interface OpenBrowserResult {
  success: boolean;
  message: string;
}

export async function openBrowser(url: string): Promise<OpenBrowserResult> {
  return new Promise((resolve) => {
    const os = platform();
    let command: string;

    switch (os) {
      case "darwin":
        command = `open "${url}"`;
        break;
      case "win32":
        command = `start "" "${url}"`;
        break;
      default:
        // Linux and others
        command = `xdg-open "${url}"`;
        break;
    }

    exec(command, (error) => {
      if (error) {
        resolve({
          success: false,
          message: `Failed to open browser: ${error.message}`,
        });
      } else {
        resolve({
          success: true,
          message: "Browser opened",
        });
      }
    });
  });
}

export function buildSetupUrl(
  serverUrl: string,
  token: string,
  callbackUrl: string
): string {
  const url = new URL("/setup", serverUrl);
  url.searchParams.set("token", token);
  url.searchParams.set("callback", callbackUrl);
  return url.toString();
}
