import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";

const CLAUDE_SETTINGS_PATH = path.join(os.homedir(), ".claude", "settings.json");

interface ClaudeHook {
  type: string;
  command: string;
}

interface ClaudeHookMatcher {
  matcher?: string;
  hooks: ClaudeHook[];
}

interface ClaudeSettings {
  hooks?: {
    PostToolUse?: ClaudeHookMatcher[];
    Stop?: ClaudeHookMatcher[];
    [key: string]: ClaudeHookMatcher[] | undefined;
  };
  [key: string]: unknown;
}

const AGENTRACE_HOOK: ClaudeHook = {
  type: "command",
  command: "npx agentrace send",
};

function isAgentraceHook(hook: ClaudeHook): boolean {
  return hook.command?.includes("agentrace send");
}

export function installHooks(): { success: boolean; message: string } {
  try {
    let settings: ClaudeSettings = {};

    // Load existing settings if file exists
    if (fs.existsSync(CLAUDE_SETTINGS_PATH)) {
      const content = fs.readFileSync(CLAUDE_SETTINGS_PATH, "utf-8");
      settings = JSON.parse(content);
    }

    // Initialize hooks structure if not present
    if (!settings.hooks) {
      settings.hooks = {};
    }

    // Add PostToolUse hook
    if (!settings.hooks.PostToolUse) {
      settings.hooks.PostToolUse = [];
    }

    // Check if agentrace hook already exists
    const hasPostToolUseHook = settings.hooks.PostToolUse.some((matcher) =>
      matcher.hooks?.some(isAgentraceHook)
    );

    if (!hasPostToolUseHook) {
      settings.hooks.PostToolUse.push({
        matcher: "*",
        hooks: [AGENTRACE_HOOK],
      });
    }

    // Add Stop hook
    if (!settings.hooks.Stop) {
      settings.hooks.Stop = [];
    }

    const hasStopHook = settings.hooks.Stop.some((matcher) =>
      matcher.hooks?.some(isAgentraceHook)
    );

    if (!hasStopHook) {
      settings.hooks.Stop.push({
        hooks: [AGENTRACE_HOOK],
      });
    }

    // Ensure directory exists
    const dir = path.dirname(CLAUDE_SETTINGS_PATH);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    // Write settings
    fs.writeFileSync(CLAUDE_SETTINGS_PATH, JSON.stringify(settings, null, 2));

    return { success: true, message: `Hooks added to ${CLAUDE_SETTINGS_PATH}` };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { success: false, message: `Failed to install hooks: ${message}` };
  }
}

export function uninstallHooks(): { success: boolean; message: string } {
  try {
    if (!fs.existsSync(CLAUDE_SETTINGS_PATH)) {
      return { success: true, message: "No settings file found" };
    }

    const content = fs.readFileSync(CLAUDE_SETTINGS_PATH, "utf-8");
    const settings: ClaudeSettings = JSON.parse(content);

    if (!settings.hooks) {
      return { success: true, message: "No hooks configured" };
    }

    // Remove agentrace hooks from PostToolUse
    if (settings.hooks.PostToolUse) {
      settings.hooks.PostToolUse = settings.hooks.PostToolUse.filter(
        (matcher) => !matcher.hooks?.some(isAgentraceHook)
      );
      if (settings.hooks.PostToolUse.length === 0) {
        delete settings.hooks.PostToolUse;
      }
    }

    // Remove agentrace hooks from Stop
    if (settings.hooks.Stop) {
      settings.hooks.Stop = settings.hooks.Stop.filter(
        (matcher) => !matcher.hooks?.some(isAgentraceHook)
      );
      if (settings.hooks.Stop.length === 0) {
        delete settings.hooks.Stop;
      }
    }

    // Clean up empty hooks object
    if (Object.keys(settings.hooks).length === 0) {
      delete settings.hooks;
    }

    fs.writeFileSync(CLAUDE_SETTINGS_PATH, JSON.stringify(settings, null, 2));

    return {
      success: true,
      message: `Removed hooks from ${CLAUDE_SETTINGS_PATH}`,
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { success: false, message: `Failed to uninstall hooks: ${message}` };
  }
}

export function checkHooksInstalled(): boolean {
  try {
    if (!fs.existsSync(CLAUDE_SETTINGS_PATH)) {
      return false;
    }

    const content = fs.readFileSync(CLAUDE_SETTINGS_PATH, "utf-8");
    const settings: ClaudeSettings = JSON.parse(content);

    const hasPostToolUseHook = settings.hooks?.PostToolUse?.some((matcher) =>
      matcher.hooks?.some(isAgentraceHook)
    );

    return !!hasPostToolUseHook;
  } catch {
    return false;
  }
}
