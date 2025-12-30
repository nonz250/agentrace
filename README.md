# AgenTrace

A service for reviewing Claude Code conversations with your team.

## Quick Start

### 1. Start the Server

```bash
docker run -d --name agentrace -p 9080:9080 -v $(pwd)/data:/data satetsu888/agentrace:latest
```

Open http://localhost:9080 in your browser.

### 2. Setup CLI

```bash
npx agentrace init --url http://localhost:9080
```

This will:
1. Open your browser for registration/login
2. Automatically configure API key
3. Install Claude Code hooks

Once setup is complete, your Claude Code sessions will be automatically sent to the server.

### 3. View Sessions

Open http://localhost:9080 to browse and review your sessions.

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
| `GITHUB_CLIENT_ID` | (empty) | GitHub OAuth Client ID |
| `GITHUB_CLIENT_SECRET` | (empty) | GitHub OAuth Client Secret |

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
