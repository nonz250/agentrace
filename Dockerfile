# =============================================================================
# Stage 1: Build Frontend
# =============================================================================
FROM node:22-alpine AS node-builder

WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci

COPY web/ ./
# VITE_API_URL can be empty for relative paths (nginx will proxy)
ENV VITE_API_URL=""
RUN npm run build

# =============================================================================
# Stage 2: Build Go Backend
# =============================================================================
FROM golang:1.24-bookworm AS go-builder

# Version can be set at build time
ARG VERSION=dev

# Install build dependencies for CGO (sqlite3)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

# Build with CGO enabled for sqlite3 and version injection
ENV CGO_ENABLED=1
RUN go build -ldflags "-X github.com/satetsu888/agentrace/server/internal/version.Version=${VERSION}" -o /app/agentrace-server ./cmd/server

# =============================================================================
# Stage 3: Runtime
# =============================================================================
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    nginx \
    libsqlite3-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN useradd -r -s /bin/false agentrace

# Create data directory for SQLite
RUN mkdir -p /data && chown agentrace:agentrace /data

# Copy built artifacts
COPY --from=go-builder /app/agentrace-server /usr/local/bin/
COPY --from=node-builder /app/web/dist /var/www/html

# Copy configuration files
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Environment variables with defaults
ENV PORT=8080
ENV DB_TYPE=sqlite
ENV DATABASE_URL=/data/agentrace.db
ENV DEV_MODE=false
ENV WEB_URL=""
ENV GITHUB_CLIENT_ID=""
ENV GITHUB_CLIENT_SECRET=""

# Expose the nginx port
EXPOSE 9080

# Volume for SQLite database persistence
VOLUME ["/data"]

ENTRYPOINT ["/entrypoint.sh"]
