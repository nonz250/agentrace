import { execSync } from "child_process";
import { loadConfig } from "../config/manager.js";
import { getNewLines, saveCursor, hasCursor } from "../config/cursor.js";
import { sendIngest } from "../utils/http.js";
import {
  findSessionFile,
  extractCwdFromTranscript,
} from "../utils/session-finder.js";

interface HookInput {
  session_id?: string;
  transcript_path?: string;
  cwd?: string;
  hook_event_name?: string;
}

interface SendTranscriptParams {
  sessionId: string;
  transcriptPath: string;
  cwd?: string;
  isHook: boolean;
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

/**
 * Core logic for sending transcript data to the server.
 * Shared between hook-based and manual invocations.
 */
async function sendTranscript(params: SendTranscriptParams): Promise<void> {
  const { sessionId, transcriptPath, cwd, isHook } = params;

  const exitWithError = (message: string) => {
    console.error(message);
    process.exit(isHook ? 0 : 1);
  };

  // Check if config exists
  const config = loadConfig();
  if (!config) {
    exitWithError(
      "[agentrace] Warning: Config not found. Run 'npx agentrace init' first."
    );
    return;
  }

  // Get new lines from transcript
  const { lines, totalLineCount } = getNewLines(transcriptPath, sessionId);

  if (lines.length === 0) {
    if (!isHook) {
      console.log("[agentrace] No new lines to send.");
    }
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
    if (!isHook) {
      console.log("[agentrace] No valid transcript lines to send.");
    }
    process.exit(0);
  }

  // Extract git info only on first send (when cursor doesn't exist yet)
  let gitRemoteUrl: string | undefined;
  let gitBranch: string | undefined;
  if (cwd && !hasCursor(sessionId)) {
    gitRemoteUrl = getGitRemoteUrl(cwd) ?? undefined;
    gitBranch = getGitBranch(cwd) ?? undefined;
  }

  // Send to server
  const result = await sendIngest({
    session_id: sessionId,
    transcript_lines: transcriptLines,
    cwd: cwd,
    git_remote_url: gitRemoteUrl,
    git_branch: gitBranch,
  });

  if (result.ok) {
    // Update cursor on success
    saveCursor(sessionId, totalLineCount);
    if (!isHook) {
      console.log(
        `[agentrace] Sent ${transcriptLines.length} lines for session ${sessionId}`
      );
    }
  } else {
    exitWithError(`[agentrace] Warning: ${result.error}`);
    return;
  }

  process.exit(0);
}

/**
 * Hook-based send command.
 * Reads session info from stdin (provided by Claude Code hooks).
 */
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

  // For UserPromptSubmit, wait for transcript to be written
  // (Claude hasn't started processing yet, so transcript may not be updated)
  if (data.hook_event_name === "UserPromptSubmit") {
    await sleep(10000);
  }

  // Use CLAUDE_PROJECT_DIR (stable project root) instead of cwd (can change during builds)
  const projectDir = process.env.CLAUDE_PROJECT_DIR || data.cwd;

  await sendTranscript({
    sessionId,
    transcriptPath,
    cwd: projectDir,
    isHook: true,
  });
}

/**
 * Manual send command.
 * Finds session file by ID and sends to server.
 */
export async function sendManualCommand(options: {
  sessionId: string;
}): Promise<void> {
  const { sessionId } = options;

  // Check if config exists
  const config = loadConfig();
  if (!config) {
    console.error(
      "[agentrace] Error: Config not found. Run 'npx agentrace init' first."
    );
    process.exit(1);
  }

  // Find session file
  const transcriptPath = findSessionFile(sessionId);
  if (!transcriptPath) {
    console.error(
      `[agentrace] Error: Session file not found for ID: ${sessionId}`
    );
    console.error("  Searched in: ~/.claude/projects/");
    process.exit(1);
  }

  // Extract cwd from transcript
  const cwd = extractCwdFromTranscript(transcriptPath) ?? undefined;

  console.log(`[agentrace] Found session file: ${transcriptPath}`);

  await sendTranscript({
    sessionId,
    transcriptPath,
    cwd,
    isHook: false,
  });
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
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
