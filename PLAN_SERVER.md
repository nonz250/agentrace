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
│   │   ├── session.go
│   │   └── event.go
│   ├── repository/           # データアクセス層
│   │   ├── interface.go      # インターフェース定義
│   │   ├── memory/           # オンメモリ実装
│   │   │   ├── session.go
│   │   │   ├── event.go
│   │   │   └── repositories.go
│   │   └── postgres/         # PostgreSQL実装（Step 4）
│   │       ├── session.go
│   │       └── event.go
│   ├── api/                  # HTTP ハンドラ
│   │   ├── router.go
│   │   ├── middleware.go
│   │   ├── ingest.go
│   │   └── session.go
│   └── ws/                   # WebSocket（Step 5）
│       └── hub.go
├── migrations/               # PostgreSQL マイグレーション（Step 4）
│   ├── 001_initial.up.sql
│   └── 001_initial.down.sql
├── go.mod
└── go.sum
```

## Repository パターン

### インターフェース定義（現在の実装）

```go
// internal/repository/interface.go

package repository

import (
    "context"
    "github.com/satetsu888/agentrace/server/internal/domain"
)

// セッション
type SessionRepository interface {
    Create(ctx context.Context, session *domain.Session) error
    FindByID(ctx context.Context, id string) (*domain.Session, error)
    FindAll(ctx context.Context, limit int, offset int) ([]*domain.Session, error)
    FindOrCreateByClaudeSessionID(ctx context.Context, claudeSessionID string) (*domain.Session, error)
}

// イベント
type EventRepository interface {
    Create(ctx context.Context, event *domain.Event) error
    FindBySessionID(ctx context.Context, sessionID string) ([]*domain.Event, error)
}

// Repositories は全リポジトリをまとめる
type Repositories struct {
    Session SessionRepository
    Event   EventRepository
}
```

### Step 2以降で追加予定

```go
// ユーザー（認証とは分離）
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id string) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
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
```

## ドメインモデル

### 現在の実装

```go
// internal/domain/session.go
type Session struct {
    ID              string
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
    Payload   map[string]interface{}  // transcript行をそのまま保存
    CreatedAt time.Time
}
```

### Step 2以降で追加予定

```go
// internal/domain/user.go
type User struct {
    ID        string
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// internal/domain/workspace.go
type Workspace struct {
    ID        string
    Name      string
    CreatedAt time.Time
}

// internal/domain/apikey.go
type APIKey struct {
    ID          string
    UserID      string
    WorkspaceID string
    KeyHash     string
    KeyPrefix   string
    Name        string
    LastUsedAt  *time.Time
    CreatedAt   time.Time
}
```

## API 設計

### Step 1: データ受信（CLI用） ✅ 完了

| Method | Path | 説明 |
| ------ | ---- | ---- |
| POST | `/api/ingest` | transcript行受信（Bearer認証） |
| GET | `/api/sessions` | セッション一覧 |
| GET | `/api/sessions/:id` | セッション詳細（イベント含む） |
| GET | `/health` | ヘルスチェック |

**POST /api/ingest リクエスト:**

```json
{
  "session_id": "claude-session-id",
  "transcript_lines": [
    {"type": "user", "message": {...}},
    {"type": "assistant", "message": {...}}
  ],
  "cwd": "/path/to/project"
}
```

**POST /api/ingest レスポンス:**

```json
{
  "ok": true,
  "events_created": 5
}
```

### Step 2: 認証（Web用）

| Method | Path | 説明 |
| ------ | ---- | ---- |
| POST | `/auth/register` | ユーザー登録 |
| POST | `/auth/login` | ログイン → セッションCookie |
| POST | `/auth/logout` | ログアウト |
| GET | `/auth/oauth/:provider` | OAuth開始 |
| GET | `/auth/oauth/:provider/callback` | OAuthコールバック |

### Step 2: CLI セットアップ用

| Method | Path | 説明 |
| ------ | ---- | ---- |
| GET | `/setup` | セットアップ画面（ログイン後APIキー発行） |
| POST | `/api/keys` | APIキー生成 |
| DELETE | `/api/keys/:id` | APIキー削除 |

### Step 3: REST API（フロント用）

| Method | Path | 説明 |
| ------ | ---- | ---- |
| GET | `/api/workspaces` | ワークスペース一覧 |

### Step 5: WebSocket（フロント用）

| Path | 説明 |
| ---- | ---- |
| `/ws/live` | リアルタイム配信（新規イベント通知） |

## データモデル（PostgreSQL）- Step 4

```sql
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

-- イベント（transcript行をJSONBで保存）
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    event_type VARCHAR(50),
    payload JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- インデックス
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

**開発モード判定:**
- `API_KEY_FIXED` が設定されている場合、開発モードとしてリクエストログを出力

## 依存パッケージ

- `github.com/gorilla/mux` - ルーティング
- `github.com/google/uuid` - UUID生成
- `github.com/gorilla/websocket` - WebSocket（Step 5）
- `github.com/lib/pq` - PostgreSQL ドライバ（Step 4）
- `golang.org/x/crypto/bcrypt` - パスワードハッシュ（Step 2）

## 実装順序

### Step 1: 最小動作版（オンメモリDB） ✅ 完了

1. Repository インターフェース定義（Session, Event）
2. オンメモリ Repository 実装
3. POST /api/ingest（固定APIキー認証、transcript行配列対応）
4. GET /api/sessions, /api/sessions/:id
5. 開発モード時のリクエストログ出力

### Step 2: 認証とセットアップUI

1. User, Credential, APIKey Repository
2. POST /auth/register, /auth/login
3. GET /setup（HTML）
4. POST /api/keys

### Step 3: Web UI（web/で実装）

（サーバー側の追加実装は特になし）

### Step 4: PostgreSQL対応

1. PostgreSQL Repository 実装
2. マイグレーション実行
3. DB_TYPE 環境変数で切り替え

### Step 5: リアルタイム機能

1. WebSocket Hub 実装
2. イベント保存時に配信
