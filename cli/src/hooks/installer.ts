import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";

const CLAUDE_SETTINGS_PATH = path.join(os.homedir(), ".claude", "settings.json");
// MCP servers are configured in ~/.claude.json, NOT in settings.json
const CLAUDE_CONFIG_PATH = path.join(os.homedir(), ".claude.json");

interface ClaudeHook {
  type: string;
  command: string;
}

interface ClaudeHookMatcher {
  matcher?: string;
  hooks: ClaudeHook[];
}

interface McpServerConfig {
  command: string;
  args: string[];
}

interface ClaudeSettings {
  hooks?: {
    Stop?: ClaudeHookMatcher[];
    [key: string]: ClaudeHookMatcher[] | undefined;
  };
  [key: string]: unknown;
}

// ~/.claude.json structure (separate from settings.json)
interface ClaudeConfig {
  mcpServers?: {
    [key: string]: McpServerConfig;
  };
  [key: string]: unknown;
}

const DEFAULT_COMMAND = "npx agentrace send";

function createAgentraceHook(command: string): ClaudeHook {
  return {
    type: "command",
    command,
  };
}

function isAgentraceHook(hook: ClaudeHook): boolean {
  // Match both production ("agentrace send") and dev mode ("index.ts send")
  return hook.command?.includes("agentrace send") || hook.command?.includes("index.ts send");
}

export interface InstallHooksOptions {
  command?: string;
}

export function installHooks(options: InstallHooksOptions = {}): { success: boolean; message: string } {
  const command = options.command || DEFAULT_COMMAND;
  const agentraceHook = createAgentraceHook(command);
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

    // Add Stop hook only (transcript diff is sent on each Stop)
    if (!settings.hooks.Stop) {
      settings.hooks.Stop = [];
    }

    const hasStopHook = settings.hooks.Stop.some((matcher) =>
      matcher.hooks?.some(isAgentraceHook)
    );

    if (hasStopHook) {
      return { success: true, message: "Hooks already installed (skipped)" };
    }

    settings.hooks.Stop.push({
      hooks: [agentraceHook],
    });

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

    const hasStopHook = settings.hooks?.Stop?.some((matcher) =>
      matcher.hooks?.some(isAgentraceHook)
    );

    return !!hasStopHook;
  } catch {
    return false;
  }
}

// MCP Server installer functions

const MCP_SERVER_NAME = "agentrace";

export interface InstallMcpServerOptions {
  command?: string;
  args?: string[];
}

export function installMcpServer(options: InstallMcpServerOptions = {}): { success: boolean; message: string } {
  const command = options.command || "npx";
  const args = options.args || ["agentrace", "mcp-server"];

  try {
    let config: ClaudeConfig = {};

    // Load existing config if file exists
    // MCP servers are configured in ~/.claude.json (NOT settings.json)
    if (fs.existsSync(CLAUDE_CONFIG_PATH)) {
      const content = fs.readFileSync(CLAUDE_CONFIG_PATH, "utf-8");
      config = JSON.parse(content);
    }

    // Initialize mcpServers structure if not present
    if (!config.mcpServers) {
      config.mcpServers = {};
    }

    // Check if already installed
    if (config.mcpServers[MCP_SERVER_NAME]) {
      // Update existing config
      config.mcpServers[MCP_SERVER_NAME] = { command, args };
      fs.writeFileSync(CLAUDE_CONFIG_PATH, JSON.stringify(config, null, 2));
      return { success: true, message: "MCP server config updated" };
    }

    // Add MCP server config
    config.mcpServers[MCP_SERVER_NAME] = { command, args };

    // Write config
    fs.writeFileSync(CLAUDE_CONFIG_PATH, JSON.stringify(config, null, 2));

    return { success: true, message: `MCP server added to ${CLAUDE_CONFIG_PATH}` };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { success: false, message: `Failed to install MCP server: ${message}` };
  }
}

export function uninstallMcpServer(): { success: boolean; message: string } {
  try {
    if (!fs.existsSync(CLAUDE_CONFIG_PATH)) {
      return { success: true, message: "No config file found" };
    }

    const content = fs.readFileSync(CLAUDE_CONFIG_PATH, "utf-8");
    const config: ClaudeConfig = JSON.parse(content);

    if (!config.mcpServers || !config.mcpServers[MCP_SERVER_NAME]) {
      return { success: true, message: "MCP server not configured" };
    }

    // Remove agentrace MCP server
    delete config.mcpServers[MCP_SERVER_NAME];

    // Clean up empty mcpServers object
    if (Object.keys(config.mcpServers).length === 0) {
      delete config.mcpServers;
    }

    fs.writeFileSync(CLAUDE_CONFIG_PATH, JSON.stringify(config, null, 2));

    return {
      success: true,
      message: `Removed MCP server from ${CLAUDE_CONFIG_PATH}`,
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { success: false, message: `Failed to uninstall MCP server: ${message}` };
  }
}

export function checkMcpServerInstalled(): boolean {
  try {
    if (!fs.existsSync(CLAUDE_CONFIG_PATH)) {
      return false;
    }

    const content = fs.readFileSync(CLAUDE_CONFIG_PATH, "utf-8");
    const config: ClaudeConfig = JSON.parse(content);

    return !!config.mcpServers?.[MCP_SERVER_NAME];
  } catch {
    return false;
  }
}
