# Server 実装計画

## 概要

Go で実装するバックエンドサーバー。Repository パターンでデータ層を抽象化し、オンメモリ/PostgreSQL を切り替え可能にする。

## ディレクトリ構成

```text
server/
├── cmd/
│   └── server/
│       └── main.go           # エントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go         # 環境変数・設定管理
│   ├── domain/               # ドメインモデル
│   │   ├── user.go
│   │   ├── workspace.go
│   │   ├── apikey.go
│   │   ├── session.go
│   │   └── event.go
│   ├── repository/           # データアクセス層
│   │   ├── interface.go      # インターフェース定義
│   │   ├── memory/           # オンメモリ実装
│   │   │   ├── user.go
│   │   │   ├── session.go
│   │   │   └── event.go
│   │   └── postgres/         # PostgreSQL実装
│   │       ├── user.go
│   │       ├── session.go
│   │       └── event.go
│   ├── service/              # ビジネスロジック
│   │   ├── auth.go
│   │   ├── ingest.go
│   │   └── session.go
│   ├── api/                  # HTTP ハンドラ
│   │   ├── router.go
│   │   ├── middleware.go
│   │   ├── auth.go
│   │   ├── ingest.go
│   │   └── session.go
│   └── ws/                   # WebSocket
│       └── hub.go
├── migrations/               # PostgreSQL マイグレーション
│   ├── 001_initial.up.sql
│   └── 001_initial.down.sql
├── go.mod
└── go.sum
```

## Repository パターン

### インターフェース定義

```go
// internal/repository/interface.go

package repository

import (
    "context"
    "server/internal/domain"
)

// ユーザー（認証とは分離）
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id string) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

// 認証情報（パスワード）
type CredentialRepository interface {
    Create(ctx context.Context, userID string, passwordHash string) error
    FindByUserID(ctx context.Context, userID string) (*domain.Credential, error)
}

// 認証情報（OAuth）
type OAuthProviderRepository interface {
    Create(ctx context.Context, provider *domain.OAuthProvider) error
    FindByProvider(ctx context.Context, provider string, providerUserID string) (*domain.OAuthProvider, error)
    FindByUserID(ctx context.Context, userID string) ([]*domain.OAuthProvider, error)
}

// APIキー
type APIKeyRepository interface {
    Create(ctx context.Context, key *domain.APIKey) error
    FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
    FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error)
}

// ワークスペース
type WorkspaceRepository interface {
    Create(ctx context.Context, workspace *domain.Workspace) error
    FindByID(ctx context.Context, id string) (*domain.Workspace, error)
    FindByUserID(ctx context.Context, userID string) ([]*domain.Workspace, error)
}

// セッション
type SessionRepository interface {
    Create(ctx context.Context, session *domain.Session) error
    FindByID(ctx context.Context, id string) (*domain.Session, error)
    FindByWorkspace(ctx context.Context, workspaceID string, limit int, offset int) ([]*domain.Session, error)
    FindOrCreateByClaude(ctx context.Context, claudeSessionID string, workspaceID string, userID string) (*domain.Session, error)
}

// イベント
type EventRepository interface {
    Create(ctx context.Context, event *domain.Event) error
    FindBySession(ctx context.Context, sessionID string) ([]*domain.Event, error)
}

// Repositories は全リポジトリをまとめる
type Repositories struct {
    User          UserRepository
    Credential    CredentialRepository
    OAuthProvider OAuthProviderRepository
    APIKey        APIKeyRepository
    Workspace     WorkspaceRepository
    Session       SessionRepository
    Event         EventRepository
}
```

### 実装切り替え

```go
// internal/repository/factory.go

func NewRepositories(cfg *config.Config) (*Repositories, error) {
    switch cfg.DBType {
    case "memory":
        return memory.NewRepositories(), nil
    case "postgres":
        db, err := sql.Open("postgres", cfg.DatabaseURL)
        if err != nil {
            return nil, err
        }
        return postgres.NewRepositories(db), nil
    default:
        return memory.NewRepositories(), nil
    }
}
```

## ドメインモデル

```go
// internal/domain/user.go
type User struct {
    ID        string
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// internal/domain/session.go
type Session struct {
    ID              string
    WorkspaceID     string
    UserID          string
    ClaudeSessionID string
    ProjectPath     string
    StartedAt       time.Time
    EndedAt         *time.Time
    CreatedAt       time.Time
}

// internal/domain/event.go
type Event struct {
    ID        string
    SessionID string
    EventType string
    ToolName  string
    Payload   map[string]interface{}
    CreatedAt time.Time
}
```

## API 設計

### 認証（Web用）

| Method | Path | 説明 |
| ------ | ---- | ---- |
| POST | `/auth/register` | ユーザー登録 |
| POST | `/auth/login` | ログイン → セッションCookie |
| POST | `/auth/logout` | ログアウト |
| GET | `/auth/oauth/:provider` | OAuth開始 |
| GET | `/auth/oauth/:provider/callback` | OAuthコールバック |

### CLI セットアップ用

| Method | Path | 説明 |
| ------ | ---- | ---- |
| GET | `/setup` | セットアップ画面（ログイン後APIキー発行） |
| POST | `/api/keys` | APIキー生成 |
| DELETE | `/api/keys/:id` | APIキー削除 |

### データ受信（CLI用）

| Method | Path | 説明 |
| ------ | ---- | ---- |
| POST | `/api/ingest` | イベント受信（Bearer認証） |

**リクエスト:**

```json
{
  "session_id": "claude-session-id",
  "hook_event_name": "PostToolUse",
  "tool_name": "Bash",
  "tool_input": {},
  "tool_response": {},
  "cwd": "/path/to/project"
}
```

**レスポンス:**

```json
{
  "ok": true,
  "event_id": "uuid"
}
```

### REST API（フロント用）

| Method | Path | 説明 |
| ------ | ---- | ---- |
| GET | `/api/workspaces` | ワークスペース一覧 |
| GET | `/api/sessions` | セッション一覧（フィルタ可） |
| GET | `/api/sessions/:id` | セッション詳細（イベント含む） |

### WebSocket（フロント用）

| Path | 説明 |
| ---- | ---- |
| `/ws/live` | リアルタイム配信（新規イベント通知） |

## データモデル（PostgreSQL）

```sql
-- ユーザー（基本情報のみ）
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- パスワード認証（1:1、オプション）
CREATE TABLE user_credentials (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- OAuth認証（1:N、複数プロバイダ対応）
CREATE TABLE user_oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- ワークスペース
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE workspace_members (
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member',
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

-- APIキー
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(10) NOT NULL,
    name VARCHAR(255),
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- セッション
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    claude_session_id VARCHAR(255),
    project_path VARCHAR(1024),
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- イベント
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    event_type VARCHAR(50),
    tool_name VARCHAR(100),
    payload JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_user_oauth_user ON user_oauth_providers(user_id);
CREATE INDEX idx_user_oauth_provider ON user_oauth_providers(provider, provider_user_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_sessions_workspace ON sessions(workspace_id);
CREATE INDEX idx_sessions_claude_id ON sessions(claude_session_id);
CREATE INDEX idx_events_session ON events(session_id);
CREATE INDEX idx_events_created ON events(created_at);
```

## 環境変数

| 変数名 | 説明 | デフォルト |
| ------ | ---- | ---------- |
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類 | memory |
| `DATABASE_URL` | PostgreSQL接続文字列 | - |
| `API_KEY_FIXED` | 固定APIキー（開発用） | - |

## 依存パッケージ

- `github.com/gorilla/mux` - ルーティング
- `github.com/gorilla/websocket` - WebSocket
- `github.com/lib/pq` - PostgreSQL ドライバ
- `github.com/google/uuid` - UUID生成
- `golang.org/x/crypto/bcrypt` - パスワードハッシュ

## 実装順序

### Step 1: 最小動作版（オンメモリDB）

1. Repository インターフェース定義
2. オンメモリ Repository 実装（Session, Event, APIKey）
3. POST /api/ingest（固定APIキー認証）
4. GET /api/sessions, /api/sessions/:id

### Step 2: 認証とセットアップUI

1. User, Credential Repository
2. POST /auth/register, /auth/login
3. GET /setup（HTML）
4. POST /api/keys

### Step 3: PostgreSQL対応

1. PostgreSQL Repository 実装
2. マイグレーション実行
3. DB_TYPE 環境変数で切り替え

### Step 4: WebSocket

1. WebSocket Hub 実装
2. イベント保存時に配信
