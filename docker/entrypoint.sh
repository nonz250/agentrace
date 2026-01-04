#!/bin/bash
set -e

# Ensure data directory has correct permissions
chown -R agentrace:agentrace /data

# Export environment variables with defaults
export PORT=${PORT:-8080}
export DB_TYPE=${DB_TYPE:-sqlite}
export DATABASE_URL=${DATABASE_URL:-/data/agentrace.db}
export DEV_MODE=${DEV_MODE:-false}
export WEB_URL=${WEB_URL:-}
export GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID:-}
export GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET:-}

echo "Starting AgenTrace..."
echo "  DB_TYPE: $DB_TYPE"
echo "  DATABASE_URL: $DATABASE_URL"
echo "  DEV_MODE: $DEV_MODE"

exec "$@"
