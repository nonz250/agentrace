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

| Command                                      | Description                            |
| -------------------------------------------- | -------------------------------------- |
| `agentrace init --url <url>`                 | Initial setup + hooks installation     |
| `agentrace init --url <url> --proxy <proxy>` | Setup with proxy                       |
| `agentrace init --url <url> --local`         | Setup for current project only         |
| `agentrace login`                            | Open the web dashboard                 |
| `agentrace send`                             | Send transcript diff (used by hooks)   |
| `agentrace on`                               | Enable hooks                           |
| `agentrace on --local`                       | Enable hooks for current project       |
| `agentrace off`                              | Disable hooks                          |
| `agentrace off --local`                      | Disable hooks for current project      |
| `agentrace uninstall`                        | Remove hooks and configuration         |
| `agentrace uninstall --local`                | Remove project-local settings only     |
| `agentrace doctor`                           | Check configuration and server status  |

## Command Details

### init

Sets up the server connection and installs Claude Code hooks.

```bash
# Global setup (all projects)
npx agentrace init --url http://localhost:9080

# Project-local setup (current project only)
npx agentrace init --url http://localhost:9080 --local

# Project-local with separate config file
npx agentrace init --url http://localhost:9080 --local --separate-local-config
```

**Process flow:**

1. Opens the server's registration/login page in browser
2. After registration, API key is automatically retrieved
3. Claude Code hooks are configured

**Options:**

| Option | Description |
| ------ | ----------- |
| `--url <url>` | Server URL (required) |
| `--proxy <url>` | HTTP/HTTPS proxy URL |
| `--local` | Install hooks/MCP for current project only |
| `--separate-local-config` | Store config in project directory (requires --local) |

### login

Issues a login URL for the web dashboard and opens it in browser.

```bash
npx agentrace login
```

### on / off

Toggle hooks enabled/disabled. Configuration is preserved.

```bash
# Temporarily stop sending (global)
npx agentrace off

# Resume sending (global)
npx agentrace on

# For project-local settings
npx agentrace off --local
npx agentrace on --local
```

### uninstall

Completely removes hooks and configuration files.

```bash
# Remove global settings
npx agentrace uninstall

# Remove project-local settings only
npx agentrace uninstall --local
```

### doctor

Displays configuration status and checks server connectivity.

```bash
npx agentrace doctor
```

**Output example:**

```
AgenTrace Doctor

================

CLI Version: 0.1.0

Configuration:
  Global config: /Users/you/.agentrace/config.json
    Status: ✓ Found

  Local config: /path/to/project/.agentrace/config.json
    Status: ✓ Found

  Active config: Local (/path/to/project/.agentrace/config.json)
  Server URL: http://localhost:9080
  API Key: agtr_xxx****xxxx

Server Status:
  Connection: ✓ Connected
  Server Version: v0.1.0
```

### send

This command is automatically called by Claude Code's Stop hook. You normally don't need to run it manually.

## Configuration Files

### Config Loading Priority

The CLI loads configuration in the following order:

1. **Local config**: Searches from the current directory up to parent directories for `.agentrace/config.json`
2. **Global config**: `~/.agentrace/config.json`

This means even when you're in a subdirectory of a project, the project's local config will be used.

Use `agentrace doctor` to see which config is currently active.

### File Locations

Configuration is stored in the following locations:

| File                 | Location                          |
| -------------------- | --------------------------------- |
| AgenTrace config     | `~/.agentrace/config.json`        |
| Cursor data          | `~/.agentrace/cursors/`           |
| Claude Code hooks    | `~/.claude/settings.json`         |
| MCP servers          | `~/.claude.json`                  |

### Project-Local Configuration

With the `--local` option, hooks and MCP are configured per-project:

| File                 | Location                                       |
| -------------------- | ---------------------------------------------- |
| Claude Code hooks    | `{project}/.claude/settings.json`              |
| MCP servers          | `~/.claude.json` (projects.{path}.mcpServers)  |
| Config (optional)    | `{project}/.agentrace/config.json`             |

To also store the config locally, use `--separate-local-config`:

```bash
npx agentrace init --url http://localhost:9080 --local --separate-local-config
```

**Note:** Add `.agentrace/` to your `.gitignore` when using `--separate-local-config` to avoid committing API keys.

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

## Proxy Configuration

If you need to connect through an HTTP proxy, you can configure it in several ways:

### 1. Using the `--proxy` Option

```bash
npx agentrace init --url http://localhost:9080 --proxy http://proxy.example.com:8080
```

For proxies requiring authentication:

```bash
npx agentrace init --url http://localhost:9080 --proxy http://user:pass@proxy.example.com:8080
```

### 2. Using Environment Variables

If `proxy_url` is not set in the config file, the CLI will check the following environment variables (in order):

1. `HTTPS_PROXY`
2. `https_proxy`
3. `HTTP_PROXY`
4. `http_proxy`

```bash
export HTTPS_PROXY=http://proxy.example.com:8080
npx agentrace send
```

**Priority:** Config file > Environment variables

## Requirements

- Node.js 18 or later
- Claude Code installed

## License

MIT
