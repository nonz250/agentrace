import { execSync } from "child_process";
import { loadConfig } from "../config/manager.js";
import { getNewLines, saveCursor, hasCursor } from "../config/cursor.js";
import { sendIngest } from "../utils/http.js";

interface HookInput {
  session_id?: string;
  transcript_path?: string;
  cwd?: string;
}

function getGitRemoteUrl(cwd: string): string | null {
  try {
    const url = execSync("git remote get-url origin", {
      cwd,
      encoding: "utf-8",
      stdio: ["pipe", "pipe", "pipe"],
    }).trim();
    return url || null;
  } catch {
    return null; // Not a git repo or no remote
  }
}

function getGitBranch(cwd: string): string | null {
  try {
    const branch = execSync("git branch --show-current", {
      cwd,
      encoding: "utf-8",
      stdio: ["pipe", "pipe", "pipe"],
    }).trim();
    return branch || null;
  } catch {
    return null;
  }
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
  } catch {
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

  const sessionId = data.session_id;
  const transcriptPath = data.transcript_path;

  if (!sessionId || !transcriptPath) {
    console.error("[agentrace] Warning: Missing session_id or transcript_path");
    process.exit(0);
  }

  // Get new lines from transcript
  const { lines, totalLineCount } = getNewLines(transcriptPath, sessionId);

  if (lines.length === 0) {
    // No new lines to send
    process.exit(0);
  }

  // Parse JSONL lines
  const transcriptLines: unknown[] = [];
  for (const line of lines) {
    try {
      transcriptLines.push(JSON.parse(line));
    } catch {
      // Skip invalid JSON lines
    }
  }

  if (transcriptLines.length === 0) {
    process.exit(0);
  }

  // Use CLAUDE_PROJECT_DIR (stable project root) instead of cwd (can change during builds)
  const projectDir = process.env.CLAUDE_PROJECT_DIR || data.cwd;

  // Extract git info only on first send (when cursor doesn't exist yet)
  let gitRemoteUrl: string | undefined;
  let gitBranch: string | undefined;
  if (projectDir && !hasCursor(sessionId)) {
    gitRemoteUrl = getGitRemoteUrl(projectDir) ?? undefined;
    gitBranch = getGitBranch(projectDir) ?? undefined;
  }

  // Send to server
  const result = await sendIngest({
    session_id: sessionId,
    transcript_lines: transcriptLines,
    cwd: projectDir,
    git_remote_url: gitRemoteUrl,
    git_branch: gitBranch,
  });

  if (result.ok) {
    // Update cursor on success
    saveCursor(sessionId, totalLineCount);
  } else {
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
