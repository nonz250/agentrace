#!/usr/bin/env node

import { Command } from "commander";
import { initCommand } from "./commands/init.js";
import { sendCommand } from "./commands/send.js";

const program = new Command();

program.name("agentrace").description("CLI for Agentrace").version("0.1.0");

program
  .command("init")
  .description("Initialize agentrace configuration and hooks")
  .option("--dev", "Use local CLI path for development")
  .action(async (options: { dev?: boolean }) => {
    await initCommand({ dev: options.dev });
  });

program
  .command("send")
  .description("Send event to server (used by hooks)")
  .action(async () => {
    await sendCommand();
  });

program.parse();
