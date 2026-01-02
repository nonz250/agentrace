-- Users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Password Credentials table (separate from users for flexibility)
CREATE TABLE IF NOT EXISTS password_credentials (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- OAuth Connections table (for social login)
CREATE TABLE IF NOT EXISTS oauth_connections (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(provider, provider_id)
);

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    last_used_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Web Sessions table
CREATE TABLE IF NOT EXISTS web_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    canonical_git_repository TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(canonical_git_repository)
);

-- Default project (no project)
INSERT OR IGNORE INTO projects (id, canonical_git_repository)
VALUES ('00000000-0000-0000-0000-000000000000', '');

-- Sessions table (Claude Code sessions)
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    project_id TEXT NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000' REFERENCES projects(id),
    claude_session_id TEXT,
    project_path TEXT,
    git_branch TEXT,
    started_at TEXT,
    ended_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Events table
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    event_type TEXT,
    payload TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Plan Documents table
CREATE TABLE IF NOT EXISTS plan_documents (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000' REFERENCES projects(id),
    description TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'planning',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Plan Document Events table
CREATE TABLE IF NOT EXISTS plan_document_events (
    id TEXT PRIMARY KEY,
    plan_document_id TEXT NOT NULL REFERENCES plan_documents(id) ON DELETE CASCADE,
    claude_session_id TEXT,
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL DEFAULT 'body_change',
    patch TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_password_credentials_user ON password_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_connections_user ON oauth_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_connections_provider ON oauth_connections(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_web_sessions_token ON web_sessions(token);
CREATE INDEX IF NOT EXISTS idx_web_sessions_expires ON web_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_projects_canonical ON projects(canonical_git_repository);
CREATE INDEX IF NOT EXISTS idx_sessions_claude_id ON sessions(claude_session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_sessions_started ON sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id);
CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at);
CREATE INDEX IF NOT EXISTS idx_plan_documents_project ON plan_documents(project_id);
CREATE INDEX IF NOT EXISTS idx_plan_documents_updated ON plan_documents(updated_at);
CREATE INDEX IF NOT EXISTS idx_plan_documents_status ON plan_documents(status);
CREATE INDEX IF NOT EXISTS idx_plan_document_events_doc ON plan_document_events(plan_document_id);
CREATE INDEX IF NOT EXISTS idx_plan_document_events_claude_session ON plan_document_events(claude_session_id);
CREATE INDEX IF NOT EXISTS idx_plan_document_events_user ON plan_document_events(user_id);
CREATE INDEX IF NOT EXISTS idx_plan_document_events_type ON plan_document_events(event_type);
