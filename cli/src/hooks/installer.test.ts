import { describe, it, expect, beforeEach, afterEach } from "vitest";
import * as fs from "node:fs";
import * as path from "node:path";
import * as os from "node:os";
import {
  installHooks,
  uninstallHooks,
  checkHooksInstalled,
  installMcpServer,
  uninstallMcpServer,
  checkMcpServerInstalled,
} from "./installer.js";

describe("hooks/installer", () => {
  describe("local hooks", () => {
    let tempProjectDir: string;

    beforeEach(() => {
      tempProjectDir = fs.mkdtempSync(path.join(os.tmpdir(), "agentrace-project-"));
    });

    afterEach(() => {
      if (fs.existsSync(tempProjectDir)) {
        fs.rmSync(tempProjectDir, { recursive: true });
      }
    });

    it("installHooks creates local settings.json", () => {
      const result = installHooks({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(result.success).toBe(true);

      const settingsPath = path.join(tempProjectDir, ".claude", "settings.json");
      expect(fs.existsSync(settingsPath)).toBe(true);

      const settings = JSON.parse(fs.readFileSync(settingsPath, "utf-8"));
      expect(settings.hooks).toBeDefined();
      expect(settings.hooks.Stop).toBeDefined();
      expect(settings.hooks.UserPromptSubmit).toBeDefined();
      expect(settings.hooks.SubagentStop).toBeDefined();
      expect(settings.hooks.PostToolUse).toBeDefined();
    });

    it("installHooks with custom command", () => {
      const customCommand = "npx custom-cli send";
      const result = installHooks({
        command: customCommand,
        local: true,
        projectDir: tempProjectDir,
      });

      expect(result.success).toBe(true);

      const settingsPath = path.join(tempProjectDir, ".claude", "settings.json");
      const settings = JSON.parse(fs.readFileSync(settingsPath, "utf-8"));

      expect(settings.hooks.Stop[0].hooks[0].command).toBe(customCommand);
    });

    it("checkHooksInstalled returns true for local hooks", () => {
      installHooks({
        local: true,
        projectDir: tempProjectDir,
      });

      const installed = checkHooksInstalled({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(installed).toBe(true);
    });

    it("checkHooksInstalled returns false when no local hooks", () => {
      const installed = checkHooksInstalled({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(installed).toBe(false);
    });

    it("uninstallHooks removes local hooks", () => {
      installHooks({
        local: true,
        projectDir: tempProjectDir,
      });

      const result = uninstallHooks({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(result.success).toBe(true);

      const installed = checkHooksInstalled({
        local: true,
        projectDir: tempProjectDir,
      });
      expect(installed).toBe(false);
    });
  });

  describe("local MCP server", () => {
    let tempProjectDir: string;
    let claudeJsonPath: string;
    let originalClaudeJson: string | null = null;

    beforeEach(() => {
      tempProjectDir = fs.mkdtempSync(path.join(os.tmpdir(), "agentrace-project-"));
      claudeJsonPath = path.join(os.homedir(), ".claude.json");

      // Backup original ~/.claude.json if exists
      if (fs.existsSync(claudeJsonPath)) {
        originalClaudeJson = fs.readFileSync(claudeJsonPath, "utf-8");
      }
    });

    afterEach(() => {
      if (fs.existsSync(tempProjectDir)) {
        fs.rmSync(tempProjectDir, { recursive: true });
      }

      // Restore original ~/.claude.json
      if (originalClaudeJson !== null) {
        fs.writeFileSync(claudeJsonPath, originalClaudeJson);
      } else if (fs.existsSync(claudeJsonPath)) {
        // If original didn't exist but now exists, we need to clean up
        // But we should be careful not to delete user's actual config
        // So we just leave it as is
      }
    });

    it("installMcpServer adds to projects for local scope", () => {
      const result = installMcpServer({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(result.success).toBe(true);
      expect(result.message).toContain("local scope");

      const claudeJson = JSON.parse(fs.readFileSync(claudeJsonPath, "utf-8"));
      expect(claudeJson.projects).toBeDefined();
      expect(claudeJson.projects[tempProjectDir]).toBeDefined();
      expect(claudeJson.projects[tempProjectDir].mcpServers).toBeDefined();
      expect(claudeJson.projects[tempProjectDir].mcpServers.agentrace).toBeDefined();
    });

    it("checkMcpServerInstalled returns true for local MCP", () => {
      installMcpServer({
        local: true,
        projectDir: tempProjectDir,
      });

      const installed = checkMcpServerInstalled({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(installed).toBe(true);
    });

    it("checkMcpServerInstalled returns false when no local MCP", () => {
      const installed = checkMcpServerInstalled({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(installed).toBe(false);
    });

    it("uninstallMcpServer removes local MCP", () => {
      installMcpServer({
        local: true,
        projectDir: tempProjectDir,
      });

      const result = uninstallMcpServer({
        local: true,
        projectDir: tempProjectDir,
      });

      expect(result.success).toBe(true);

      const installed = checkMcpServerInstalled({
        local: true,
        projectDir: tempProjectDir,
      });
      expect(installed).toBe(false);
    });

    it("installMcpServer with custom command and args", () => {
      const result = installMcpServer({
        command: "node",
        args: ["custom-server.js"],
        local: true,
        projectDir: tempProjectDir,
      });

      expect(result.success).toBe(true);

      const claudeJson = JSON.parse(fs.readFileSync(claudeJsonPath, "utf-8"));
      const mcpConfig = claudeJson.projects[tempProjectDir].mcpServers.agentrace;
      expect(mcpConfig.command).toBe("node");
      expect(mcpConfig.args).toEqual(["custom-server.js"]);
    });
  });
});
