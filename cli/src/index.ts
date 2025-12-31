#!/usr/bin/env node

import { Command } from "commander";
import { initCommand } from "./commands/init.js";
import { loginCommand } from "./commands/login.js";
import { sendCommand } from "./commands/send.js";
import { uninstallCommand } from "./commands/uninstall.js";
import { onCommand } from "./commands/on.js";
import { offCommand } from "./commands/off.js";
import { mcpServerCommand } from "./commands/mcp-server.js";

const program = new Command();

program.name("agentrace").description("CLI for Agentrace").version("0.1.0");

program
  .command("init")
  .description("Initialize agentrace configuration and hooks")
  .requiredOption("--url <url>", "Server URL (required)")
  .option("--dev", "Use local CLI path for development")
  .action(async (options: { url: string; dev?: boolean }) => {
    await initCommand({ url: options.url, dev: options.dev });
  });

program
  .command("login")
  .description("Open web dashboard in browser")
  .action(async () => {
    await loginCommand();
  });

program
  .command("send")
  .description("Send event to server (used by hooks)")
  .action(async () => {
    await sendCommand();
  });

program
  .command("uninstall")
  .description("Remove agentrace hooks and config")
  .action(async () => {
    await uninstallCommand();
  });

program
  .command("on")
  .description("Enable agentrace hooks (credentials preserved)")
  .option("--dev", "Use local CLI path for development")
  .action(async (options: { dev?: boolean }) => {
    await onCommand({ dev: options.dev });
  });

program
  .command("off")
  .description("Disable agentrace hooks (credentials preserved)")
  .action(async () => {
    await offCommand();
  });

program
  .command("mcp-server")
  .description("Run MCP server for Claude Code integration (stdio)")
  .action(async () => {
    await mcpServerCommand();
  });

program.parse();
