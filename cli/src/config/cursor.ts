import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";

const CURSORS_DIR = path.join(os.homedir(), ".agentrace", "cursors");

interface CursorData {
  lineCount: number;
  lastUpdated: string;
}

function getCursorPath(sessionId: string): string {
  return path.join(CURSORS_DIR, `${sessionId}.json`);
}

export function getCursor(sessionId: string): number {
  try {
    const cursorPath = getCursorPath(sessionId);
    if (!fs.existsSync(cursorPath)) {
      return 0;
    }
    const content = fs.readFileSync(cursorPath, "utf-8");
    const data: CursorData = JSON.parse(content);
    return data.lineCount;
  } catch {
    return 0;
  }
}

export function saveCursor(sessionId: string, lineCount: number): void {
  if (!fs.existsSync(CURSORS_DIR)) {
    fs.mkdirSync(CURSORS_DIR, { recursive: true });
  }

  const data: CursorData = {
    lineCount,
    lastUpdated: new Date().toISOString(),
  };

  const cursorPath = getCursorPath(sessionId);
  fs.writeFileSync(cursorPath, JSON.stringify(data, null, 2));
}

export function readTranscriptLines(transcriptPath: string): string[] {
  try {
    if (!fs.existsSync(transcriptPath)) {
      return [];
    }
    const content = fs.readFileSync(transcriptPath, "utf-8");
    return content.split("\n").filter((line) => line.trim() !== "");
  } catch {
    return [];
  }
}

export function getNewLines(
  transcriptPath: string,
  sessionId: string
): { lines: string[]; totalLineCount: number } {
  const allLines = readTranscriptLines(transcriptPath);
  const cursor = getCursor(sessionId);

  const newLines = allLines.slice(cursor);
  return {
    lines: newLines,
    totalLineCount: allLines.length,
  };
}
