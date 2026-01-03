# AgenTrace

A self-hosted service for sharing Claude Code sessions and managing implementation plans with your team.

Since Claude Code logs contain source code and environment information, AgenTrace is designed to run on your local machine or internal network rather than on the public internet.

**Demo:** https://satetsu888.github.io/agentrace

## Features

- **Session Sharing**: Automatically capture and share Claude Code conversations with your team
- **Plan Management**: Create and track implementation plans via Claude Code's MCP tools

## Quick Start

### 1. Start the Server

```bash
docker run -d --name agentrace -p 9080:9080 -v $(pwd)/data:/data satetsu888/agentrace:latest
```

### 2. Setup CLI

```bash
npx agentrace init --url http://localhost:9080
```

This will:

1. Open your browser for registration/login
2. Automatically configure API key
3. Install Claude Code hooks and MCP server

Once setup is complete:
- Your Claude Code sessions will be automatically sent to the server
- Plan management tools will be available in Claude Code

### 3. Browse Dashboard

Open http://localhost:9080 to:
- Review shared sessions
- View and manage implementation plans

## Using Plan Management

After setup, Claude Code can use the following MCP tools:

| Tool | Description |
| ---- | ----------- |
| `list_plans` | List plans for a repository |
| `read_plan` | Read a plan by ID |
| `create_plan` | Create a new plan |
| `update_plan` | Update an existing plan |
| `set_plan_status` | Change plan status (scratch/planning/implementation/complete) |

Claude Code will automatically use these tools when discussing implementation plans.

## CLI Commands

| Command | Description |
| ------- | ----------- |
| `npx agentrace init --url <url>` | Initial setup with browser authentication |
| `npx agentrace login` | Open web dashboard in browser |
| `npx agentrace on` | Enable hooks (keeps credentials) |
| `npx agentrace off` | Disable hooks temporarily (keeps credentials) |
| `npx agentrace uninstall` | Remove all hooks and configuration |

### Temporarily Disable Tracking

If you want to pause session tracking without removing your configuration:

```bash
# Disable hooks
npx agentrace off

# Re-enable hooks later
npx agentrace on
```

## Environment Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `DB_TYPE` | sqlite | Database type |
| `DATABASE_URL` | /data/agentrace.db | Database path |
| `DEV_MODE` | false | Enable debug logging |
| `GITHUB_CLIENT_ID` | - | GitHub OAuth Client ID |
| `GITHUB_CLIENT_SECRET` | - | GitHub OAuth Client Secret |

```bash
# Example: Enable debug mode
docker run -d -p 9080:9080 -v $(pwd)/data:/data -e DEV_MODE=true satetsu888/agentrace:latest
```

## Cleanup

To completely remove AgenTrace:

### 1. Remove CLI Configuration and Hooks

```bash
npx agentrace uninstall
```

This removes:

- Claude Code hooks from `~/.claude/settings.json`
- MCP server configuration from `~/.claude/claude_desktop_config.json`
- Configuration from `~/.agentrace/`

### 2. Stop and Remove Docker Container

```bash
docker stop agentrace && docker rm agentrace
```

### 3. Remove Docker Image (Optional)

```bash
docker rmi satetsu888/agentrace:latest
```

### 4. Remove Data (Optional)

```bash
rm -rf ./data
```

## License

MIT
