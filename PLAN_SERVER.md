# Server 実装計画

## 概要

Go で実装するバックエンドサーバー。Repository パターンでデータ層を抽象化し、複数のデータベースを切り替え可能にする。

### 対応データベース

| DB | DB_TYPE | 利用シーン |
| -- | ------- | ---------- |
| オンメモリ | `memory` | 開発・テスト |
| SQLite3 | `sqlite` | ローカル/小規模運用 |
| PostgreSQL | `postgres` | イントラネット/本番運用 |
| MongoDB | `mongodb` | AWS (DocumentDB) 環境 |

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
│   │   ├── factory.go        # DB_TYPEに応じたRepository生成
│   │   ├── memory/           # オンメモリ実装
│   │   │   ├── session.go
│   │   │   ├── event.go
│   │   │   ├── user.go       # Step 2
│   │   │   ├── apikey.go     # Step 2
│   │   │   ├── websession.go # Step 2
│   │   │   └── repositories.go
│   │   ├── sqlite/           # SQLite3実装（Step 4）
│   │   │   └── ...
│   │   ├── postgres/         # PostgreSQL実装（Step 4）
│   │   │   └── ...
│   │   └── mongodb/          # MongoDB実装（Step 4）
│   │       └── ...
│   ├── api/                  # HTTP ハンドラ
│   │   ├── router.go
│   │   ├── middleware.go
│   │   ├── ingest.go
│   │   ├── session.go
│   │   └── auth.go           # Step 2
│   └── ws/                   # WebSocket（Step 5）
│       └── hub.go
├── migrations/               # マイグレーション（Step 4）
│   ├── sqlite/
│   │   └── 001_initial.sql
│   └── postgres/
│       ├── 001_initial.up.sql
│       └── 001_initial.down.sql
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

## データモデル - Step 4

### SQLite3

```sql
-- ユーザー
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);

-- APIキー
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    last_used_at TEXT,
    created_at TEXT DEFAULT (datetime('now'))
);

-- Webセッション
CREATE TABLE web_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);

-- セッション（Claude Codeセッション）
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    claude_session_id TEXT,
    project_path TEXT,
    started_at TEXT,
    ended_at TEXT,
    created_at TEXT DEFAULT (datetime('now'))
);

-- イベント（transcript行をJSONで保存）
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE,
    event_type TEXT,
    payload TEXT,  -- JSON文字列
    created_at TEXT DEFAULT (datetime('now'))
);

-- インデックス
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_web_sessions_token ON web_sessions(token);
CREATE INDEX idx_sessions_claude_id ON sessions(claude_session_id);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_events_session ON events(session_id);
CREATE INDEX idx_events_created ON events(created_at);
```

### PostgreSQL

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

### MongoDB (DocumentDB)

```javascript
// コレクション: users
{
  _id: ObjectId,
  name: String,
  created_at: Date
}

// コレクション: api_keys
{
  _id: ObjectId,
  user_id: ObjectId,  // users._id参照
  name: String,
  key_hash: String,
  key_prefix: String,
  last_used_at: Date | null,
  created_at: Date
}
// インデックス: { key_hash: 1 }

// コレクション: web_sessions
{
  _id: ObjectId,
  user_id: ObjectId,
  token: String,
  expires_at: Date,
  created_at: Date
}
// インデックス: { token: 1 }, unique
// TTLインデックス: { expires_at: 1 }, expireAfterSeconds: 0

// コレクション: sessions
{
  _id: ObjectId,
  user_id: ObjectId | null,
  claude_session_id: String,
  project_path: String,
  started_at: Date,
  ended_at: Date | null,
  created_at: Date
}
// インデックス: { claude_session_id: 1 }, { user_id: 1 }

// コレクション: events
{
  _id: ObjectId,
  session_id: ObjectId,
  event_type: String,
  payload: Object,  // そのまま埋め込み
  created_at: Date
}
// インデックス: { session_id: 1 }, { created_at: 1 }
```

## 環境変数

| 変数名 | 説明 | デフォルト |
| ------ | ---- | ---------- |
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類 (`memory` / `sqlite` / `postgres` / `mongodb`) | memory |
| `DATABASE_URL` | DB接続文字列（下記参照） | - |
| `DEV_MODE` | デバッグログ有効化 | false |

**DATABASE_URL の形式:**

| DB_TYPE | DATABASE_URL 例 |
| ------- | --------------- |
| sqlite | `./data/agentrace.db` または `/var/lib/agentrace/data.db` |
| postgres | `postgres://user:pass@localhost:5432/agentrace?sslmode=disable` |
| mongodb | `mongodb://user:pass@localhost:27017/agentrace` または DocumentDB接続文字列 |

**デバッグモード:**
- `DEV_MODE=true` でリクエストログを出力

## 依存パッケージ

- `github.com/gorilla/mux` - ルーティング
- `github.com/google/uuid` - UUID生成
- `golang.org/x/crypto/bcrypt` - APIキーハッシュ
- `github.com/gorilla/websocket` - WebSocket（Step 5）
- `github.com/mattn/go-sqlite3` - SQLite3 ドライバ（Step 4）
- `github.com/lib/pq` - PostgreSQL ドライバ（Step 4）
- `go.mongodb.org/mongo-driver` - MongoDB ドライバ（Step 4）

## 実装順序

### Step 1: 最小動作版（オンメモリDB） ✅ 完了

1. Repository インターフェース定義（Session, Event）
2. オンメモリ Repository 実装
3. POST /api/ingest（固定APIキー認証、transcript行配列対応）
4. GET /api/sessions, /api/sessions/:id
5. DEV_MODE時のリクエストログ出力

### Step 2: 認証機能 ✅ 完了

1. User, APIKey, WebSession ドメインモデル
2. User, APIKey, WebSession Repository（memory）
3. POST /auth/register - 名前入力でユーザー＆APIキー作成
4. POST /auth/login - APIキーでログイン（Cookie発行）
5. GET /auth/session - トークンでログイン（CLI経由）
6. POST /api/auth/web-session - Webログイントークン発行
7. POST /api/auth/logout - ログアウト
8. GET /api/me, GET /api/users - ユーザー情報取得
9. GET /api/keys, POST /api/keys, DELETE /api/keys/:id - APIキー管理
10. Bearer認証ミドルウェア更新（APIKey → User解決）
11. Session認証ミドルウェア追加
12. セッションにUserID紐付け

### Step 3: Web UI（web/で実装）

（サーバー側の追加実装は特になし）

### Step 4: 複数データベース対応 ✅ 完了

1. Repository ファクトリ実装（DB_TYPEに応じた切り替え） ✅
2. SQLite3 Repository 実装 ✅
   - マイグレーション
   - 各Repository（Session, Event, User, APIKey, WebSession）
3. PostgreSQL Repository 実装 ✅
   - マイグレーション（up/down）
   - 各Repository
4. MongoDB Repository 実装 ✅
   - インデックス作成
   - 各Repository
5. 動作確認 ✅

### Step 5: リアルタイム機能

1. WebSocket Hub 実装
2. イベント保存時に配信
