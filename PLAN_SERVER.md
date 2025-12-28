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
│   │   ├── event.go
│   │   ├── user.go           # Step 2
│   │   ├── apikey.go         # Step 2
│   │   └── websession.go     # Step 2
│   ├── repository/           # データアクセス層
│   │   ├── interface.go      # インターフェース定義
│   │   ├── memory/           # オンメモリ実装
│   │   │   ├── session.go
│   │   │   ├── event.go
│   │   │   ├── user.go       # Step 2
│   │   │   ├── apikey.go     # Step 2
│   │   │   ├── websession.go # Step 2
│   │   │   └── repositories.go
│   │   └── postgres/         # PostgreSQL実装（Step 4）
│   │       ├── session.go
│   │       ├── event.go
│   │       ├── user.go
│   │       ├── apikey.go
│   │       └── websession.go
│   ├── api/                  # HTTP ハンドラ
│   │   ├── router.go
│   │   ├── middleware.go
│   │   ├── ingest.go
│   │   ├── session.go
│   │   └── auth.go           # Step 2
│   └── ws/                   # WebSocket（Step 5）
│       └── hub.go
├── migrations/               # PostgreSQL マイグレーション（Step 4）
│   ├── 001_initial.up.sql
│   └── 001_initial.down.sql
├── go.mod
└── go.sum
```

## Repository パターン

### インターフェース定義（Step 1 実装済み）

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

### Step 2で追加

```go
// ユーザー
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id string) (*domain.User, error)
    FindAll(ctx context.Context) ([]*domain.User, error)
}

// APIキー
type APIKeyRepository interface {
    Create(ctx context.Context, key *domain.APIKey) error
    FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
    FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error)
}

// Webセッション
type WebSessionRepository interface {
    Create(ctx context.Context, session *domain.WebSession) error
    FindByToken(ctx context.Context, token string) (*domain.WebSession, error)
    Delete(ctx context.Context, id string) error
    DeleteExpired(ctx context.Context) error
}
```

## ドメインモデル

### Step 1 実装済み

```go
// internal/domain/session.go
type Session struct {
    ID              string
    UserID          *string    // Step 2で追加（nullable）
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

### Step 2で追加

```go
// internal/domain/user.go
type User struct {
    ID        string
    Name      string
    CreatedAt time.Time
}

// internal/domain/apikey.go
type APIKey struct {
    ID         string
    UserID     string
    Name       string     // キーの名前（例: "MacBook Pro", "Work PC"）
    KeyHash    string     // bcrypt hash
    KeyPrefix  string     // "agtr_xxxx..." (表示用、先頭12文字程度)
    LastUsedAt *time.Time
    CreatedAt  time.Time
}

// internal/domain/websession.go
type WebSession struct {
    ID        string
    UserID    string
    Token     string
    ExpiresAt time.Time
    CreatedAt time.Time
}
```

## API 設計

### Step 1: データ受信（CLI用） ✅ 完了

| Method | Path | 認証 | 説明 |
| ------ | ---- | ---- | ---- |
| POST | `/api/ingest` | Bearer | transcript行受信 |
| GET | `/api/sessions` | Bearer | セッション一覧 |
| GET | `/api/sessions/:id` | Bearer | セッション詳細（イベント含む） |
| GET | `/health` | なし | ヘルスチェック |

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

### Step 2: 認証

| Method | Path | 認証 | 説明 |
| ------ | ---- | ---- | ---- |
| POST | `/auth/register` | なし | ユーザー登録（名前入力→APIキー発行） |
| POST | `/auth/login` | なし | APIキー入力→セッションCookie発行 |
| GET | `/auth/session` | なし | トークンでログイン（CLI経由、Cookie発行） |
| POST | `/api/auth/web-session` | Bearer | Webログイントークン発行 |
| POST | `/api/auth/logout` | Session | ログアウト |
| GET | `/api/me` | Session | 自分の情報 |
| GET | `/api/users` | Session | ユーザー一覧 |
| GET | `/api/keys` | Session | 自分のAPIキー一覧 |
| POST | `/api/keys` | Session | 新しいAPIキー発行 |
| DELETE | `/api/keys/:id` | Session | APIキー削除 |

**POST /auth/register リクエスト:**

```json
{
  "name": "Taro"
}
```

**POST /auth/register レスポンス:**

```json
{
  "user": {
    "id": "xxx",
    "name": "Taro",
    "created_at": "2025-12-28T..."
  },
  "api_key": "agtr_xxxxxxxxxxxxxxxxxxxxxxxx"
}
```
※ `api_key` は生の値で、この1回のみ返される
※ 同時に `Set-Cookie: session=xxx` でログイン状態にする

**POST /auth/login リクエスト:**

```json
{
  "api_key": "agtr_xxxxxxxxxxxxxxxxxxxxxxxx"
}
```

**POST /auth/login レスポンス:**

```json
{
  "user": {
    "id": "xxx",
    "name": "Taro",
    "created_at": "2025-12-28T..."
  }
}
```
※ `Set-Cookie: session=xxx` でログイン状態にする

**POST /api/auth/web-session レスポンス:**

```json
{
  "url": "http://server:8080/auth/session?token=xxxxx",
  "expires_at": "2025-12-28T..."
}
```

**GET /api/keys レスポンス:**

```json
{
  "keys": [
    {
      "id": "xxx",
      "name": "MacBook Pro",
      "key_prefix": "agtr_xxxx...",
      "last_used_at": "2025-12-28T...",
      "created_at": "2025-12-28T..."
    }
  ]
}
```

**POST /api/keys リクエスト:**

```json
{
  "name": "Work PC"
}
```

**POST /api/keys レスポンス:**

```json
{
  "key": {
    "id": "xxx",
    "name": "Work PC",
    "key_prefix": "agtr_yyyy...",
    "created_at": "2025-12-28T..."
  },
  "api_key": "agtr_yyyyyyyyyyyyyyyyyyyyyyyy"
}
```
※ `api_key` は生の値で、この1回のみ返される

### Step 5: WebSocket（フロント用）

| Path | 説明 |
| ---- | ---- |
| `/ws/live` | リアルタイム配信（新規イベント通知） |

## 認証フロー

### Bearer認証（CLI用）

```
リクエスト:
  Authorization: Bearer agtr_xxxxxxxx

サーバー処理:
  1. APIKeyをbcryptでハッシュ化
  2. DB検索で一致するAPIKeyを探す
  3. 一致すればUserIDを取得
  4. コンテキストにUserを設定
```

### Session認証（Web用）

```
リクエスト:
  Cookie: session=xxxxx

サーバー処理:
  1. WebSessionテーブルからトークンで検索
  2. 有効期限チェック
  3. UserIDを取得
  4. コンテキストにUserを設定
```

## データモデル（PostgreSQL）- Step 4

```sql
-- ユーザー
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- APIキー
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Webセッション
CREATE TABLE web_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- セッション（Claude Codeセッション）
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_web_sessions_token ON web_sessions(token);
CREATE INDEX idx_sessions_claude_id ON sessions(claude_session_id);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_events_session ON events(session_id);
CREATE INDEX idx_events_created ON events(created_at);
```

## 環境変数

| 変数名 | 説明 | デフォルト |
| ------ | ---- | ---------- |
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類 | memory |
| `DATABASE_URL` | PostgreSQL接続文字列 | - |
| `DEV_MODE` | デバッグログ有効化 | false |

**デバッグモード:**
- `DEV_MODE=true` でリクエストログを出力

## 依存パッケージ

- `github.com/gorilla/mux` - ルーティング
- `github.com/google/uuid` - UUID生成
- `golang.org/x/crypto/bcrypt` - APIキーハッシュ
- `github.com/gorilla/websocket` - WebSocket（Step 5）
- `github.com/lib/pq` - PostgreSQL ドライバ（Step 4）

## 実装順序

### Step 1: 最小動作版（オンメモリDB） ✅ 完了

1. Repository インターフェース定義（Session, Event）
2. オンメモリ Repository 実装
3. POST /api/ingest（固定APIキー認証、transcript行配列対応）
4. GET /api/sessions, /api/sessions/:id
5. DEV_MODE時のリクエストログ出力

### Step 2: 認証機能

1. User, APIKey, WebSession ドメインモデル
2. User, APIKey, WebSession Repository（memory）
3. POST /auth/register - 名前入力でユーザー＆APIキー作成
4. POST /auth/login - APIキーでログイン（Cookie発行）
5. GET /auth/session - トークンでログイン（CLI経由）
6. POST /api/auth/web-session - Webログイントークン発行
7. Bearer認証ミドルウェア更新（APIKey → User解決）
8. Session認証ミドルウェア追加
9. セッションにUserID紐付け

### Step 3: Web UI（web/で実装）

（サーバー側の追加実装は特になし）

### Step 4: PostgreSQL対応

1. PostgreSQL Repository 実装
2. マイグレーション実行
3. DB_TYPE 環境変数で切り替え

### Step 5: リアルタイム機能

1. WebSocket Hub 実装
2. イベント保存時に配信
