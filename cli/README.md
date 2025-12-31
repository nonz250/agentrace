# AgenTrace CLI

A CLI tool to send Claude Code sessions to the AgenTrace server.

## Overview

AgenTrace is a self-hosted service that enables teams to review Claude Code conversations. Since Claude Code logs contain source code and environment information, AgenTrace is designed to run on your local machine or internal network rather than on the public internet.

This CLI uses Claude Code's hooks feature to automatically send session data to your AgenTrace server.

## Setup

### 1. Start the Server

```bash
# Using Docker (recommended)
docker run -d --name agentrace -p 9080:9080 -v $(pwd)/data:/data satetsu888/agentrace:latest
```

Docker Hub: <https://hub.docker.com/r/satetsu888/agentrace>

### 2. Initialize the CLI

```bash
npx agentrace init --url http://localhost:9080
```

A browser window will open displaying the registration/login page. After registration, hooks are automatically configured.

That's it! When you use Claude Code, sessions will be automatically sent to AgenTrace.

## Commands

| Command                      | Description                            |
| ---------------------------- | -------------------------------------- |
| `agentrace init --url <url>` | Initial setup + hooks installation     |
| `agentrace login`            | Open the web dashboard                 |
| `agentrace send`             | Send transcript diff (used by hooks)   |
| `agentrace on`               | Enable hooks                           |
| `agentrace off`              | Disable hooks                          |
| `agentrace uninstall`        | Remove hooks and configuration         |

## Command Details

### init

Sets up the server connection and installs Claude Code hooks.

```bash
npx agentrace init --url http://localhost:9080
```

**Process flow:**

1. Opens the server's registration/login page in browser
2. After registration, API key is automatically retrieved
3. Claude Code hooks are configured

### login

Issues a login URL for the web dashboard and opens it in browser.

```bash
npx agentrace login
```

### on / off

Toggle hooks enabled/disabled. Configuration is preserved.

```bash
# Temporarily stop sending
npx agentrace off

# Resume sending
npx agentrace on
```

### uninstall

Completely removes hooks and configuration files.

```bash
npx agentrace uninstall
```

### send

This command is automatically called by Claude Code's Stop hook. You normally don't need to run it manually.

## Configuration Files

Configuration is stored in the following locations:

| File                 | Location                          |
| -------------------- | --------------------------------- |
| AgenTrace config     | `~/.config/agentrace/config.json` |
| Cursor data          | `~/.config/agentrace/cursors/`    |
| Claude Code hooks    | `~/.claude/settings.json`         |

## How It Works

```text
┌─────────────────┐
│   Claude Code   │
│  (Stop hook)    │
└────────┬────────┘
         │ npx agentrace send
         ↓
┌─────────────────┐
│  AgenTrace CLI  │
│  Extract & Send │
└────────┬────────┘
         │ POST /api/ingest
         ↓
┌─────────────────┐
│ AgenTrace Server│
│   Save to DB    │
└─────────────────┘
```

- Only the transcript diff is sent to the server when a Claude Code conversation ends
- Errors do not block Claude Code's operation by design

## Requirements

- Node.js 18 or later
- Claude Code installed

## License

MIT
