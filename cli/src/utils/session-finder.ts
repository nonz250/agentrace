import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";

const CLAUDE_PROJECTS_DIR = path.join(os.homedir(), ".claude", "projects");

/**
 * Find a Claude session JSONL file by session ID.
 * Searches recursively in ~/.claude/projects/
 */
export function findSessionFile(sessionId: string): string | null {
  const targetFileName = `${sessionId}.jsonl`;

  if (!fs.existsSync(CLAUDE_PROJECTS_DIR)) {
    return null;
  }

  // Recursively search for the session file
  const result = searchDirectory(CLAUDE_PROJECTS_DIR, targetFileName);
  return result;
}

function searchDirectory(dir: string, targetFileName: string): string | null {
  try {
    const entries = fs.readdirSync(dir, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);

      if (entry.isDirectory()) {
        const result = searchDirectory(fullPath, targetFileName);
        if (result) {
          return result;
        }
      } else if (entry.isFile() && entry.name === targetFileName) {
        return fullPath;
      }
    }
  } catch {
    // Skip directories we can't read
  }

  return null;
}

interface TranscriptEntry {
  type?: string;
  cwd?: string;
}

/**
 * Extract the cwd from a transcript JSONL file.
 * Looks for the first entry with type="user" that has a cwd field.
 */
export function extractCwdFromTranscript(transcriptPath: string): string | null {
  try {
    if (!fs.existsSync(transcriptPath)) {
      return null;
    }

    const content = fs.readFileSync(transcriptPath, "utf-8");
    const lines = content.split("\n").filter((line) => line.trim() !== "");

    for (const line of lines) {
      try {
        const entry = JSON.parse(line) as TranscriptEntry;
        if (entry.type === "user" && entry.cwd) {
          return entry.cwd;
        }
      } catch {
        // Skip invalid JSON lines
      }
    }
  } catch {
    // File read error
  }

  return null;
}
