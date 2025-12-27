import { loadConfig } from "../config/manager.js";
import { sendIngest, type IngestPayload } from "../utils/http.js";

interface HookInput {
  session_id?: string;
  hook_event_name?: string;
  tool_name?: string;
  tool_input?: unknown;
  tool_response?: unknown;
  cwd?: string;
}

export async function sendCommand(): Promise<void> {
  // Check if config exists
  const config = loadConfig();
  if (!config) {
    console.error(
      "[agentrace] Warning: Config not found. Run 'npx agentrace init' first."
    );
    process.exit(0); // Exit 0 to not block hooks
  }

  // Read stdin
  let input = "";
  try {
    input = await readStdin();
  } catch (error) {
    console.error("[agentrace] Warning: Failed to read stdin");
    process.exit(0);
  }

  if (!input.trim()) {
    console.error("[agentrace] Warning: Empty input");
    process.exit(0);
  }

  // Parse JSON
  let data: HookInput;
  try {
    data = JSON.parse(input);
  } catch {
    console.error("[agentrace] Warning: Invalid JSON input");
    process.exit(0);
  }

  // Prepare payload
  const payload: IngestPayload = {
    session_id: data.session_id || "unknown",
    hook_event_name: data.hook_event_name || "unknown",
    tool_name: data.tool_name,
    tool_input: data.tool_input,
    tool_response: data.tool_response,
    cwd: data.cwd,
  };

  // Send to server
  const result = await sendIngest(payload);
  if (!result.ok) {
    console.error(`[agentrace] Warning: ${result.error}`);
  }

  // Always exit 0 to not block hooks
  process.exit(0);
}

function readStdin(): Promise<string> {
  return new Promise((resolve, reject) => {
    let data = "";

    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => {
      data += chunk;
    });
    process.stdin.on("end", () => {
      resolve(data);
    });
    process.stdin.on("error", reject);

    // Set timeout to avoid hanging
    setTimeout(() => {
      resolve(data);
    }, 5000);
  });
}
