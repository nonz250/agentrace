# Server 開発ガイド

Go + Gorilla Mux によるバックエンドサーバー。

## ディレクトリ構成

```
server/
├── cmd/server/main.go       # エントリーポイント（DI・DB初期化）
├── internal/
│   ├── api/                 # HTTPハンドラーレイヤー
│   │   ├── router.go        # ルーティング定義
│   │   ├── middleware.go    # 認証・ロギング
│   │   └── *.go             # 各種ハンドラー
│   ├── config/config.go     # 環境変数パース
│   ├── domain/              # ドメインモデル
│   └── repository/          # データアクセス層
│       ├── interface.go     # インターフェース定義
│       ├── factory.go       # Factory パターン
│       ├── testsuite/       # 共通テストスイート
│       ├── memory/          # オンメモリ実装
│       ├── sqlite/          # SQLite実装
│       ├── postgres/        # PostgreSQL実装
│       ├── dynamodb/        # AWS DynamoDB実装
│       └── turso/           # Turso（libSQL）実装
└── migrations/              # スキーママイグレーション
    ├── embed.go             # SQLファイルの埋め込み
    ├── runner.go            # migration実行ロジック
    ├── sqlite/initial.sql   # SQLite用スキーマ
    └── postgres/initial.up.sql # PostgreSQL用スキーマ
```

## Migration 管理

### 概要

- **initial.sql**: v0.0.1-alpha時点のスキーマ（バージョン管理前）
- **v0.0.1以降**: セマンティックバージョンで `schema_migrations` テーブルにより管理

### ファイル構成

```
migrations/
├── embed.go                           # SQLファイル埋め込み + Migration型定義
├── runner.go                          # migration実行ロジック
├── sqlite/
│   ├── initial.sql                    # 初期スキーマ（v0.0.1-alpha）
│   └── v0.0.1_add_something.sql       # v0.0.1用migration（例）
└── postgres/
    ├── initial.up.sql                 # 初期スキーマ（v0.0.1-alpha）
    └── v0.0.1_add_something.up.sql    # v0.0.1用migration（例）
```

### 新しいmigrationの追加手順

v0.0.1 など新しいバージョン用のmigrationを追加する場合:

#### 1. SQLファイルを作成

```sql
-- migrations/sqlite/v0.0.1_add_new_table.sql
CREATE TABLE IF NOT EXISTS new_table (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

-- migrations/postgres/v0.0.1_add_new_table.up.sql
CREATE TABLE IF NOT EXISTS new_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);
```

#### 2. embed.go に追加

```go
// 新しいmigrationファイルを埋め込み
//go:embed sqlite/v0.0.1_add_new_table.sql
var SQLiteMigration_0_0_1 string

//go:embed postgres/v0.0.1_add_new_table.up.sql
var PostgresMigration_0_0_1 string

// Migrations() に追加
func SQLiteMigrations() []Migration {
    return []Migration{
        {Version: "0.0.1", SQL: SQLiteMigration_0_0_1},
    }
}

func PostgresMigrations() []Migration {
    return []Migration{
        {Version: "0.0.1", SQL: PostgresMigration_0_0_1},
    }
}
```

### 実行フロー

アプリケーション起動時に `migrations.Runner.Run()` が以下を実行:

1. **initial.sql を適用**: 全てIF NOT EXISTSで冪等
2. **既存DB用の後方互換処理**: messageカラム等の存在チェック
3. **schema_migrations から適用済みバージョンを取得**
4. **未適用のセマンティックバージョンmigrationを順次実行**
5. **適用したバージョンを schema_migrations に記録**

### 注意事項

- **バージョン形式**: セマンティックバージョン（例: "0.0.1", "0.1.0", "1.0.0"）
- **ソート**: `golang.org/x/mod/semver` でバージョン順にソートされる
- **冪等性**: 可能な限り IF NOT EXISTS を使用
- **SQLite/PostgreSQL両方**: 必ず両方のSQLファイルを作成すること
- **Turso**: SQLite互換のため、SQLite用のmigrationが使用される

## レイヤードアーキテクチャ

```
API Layer (internal/api/)      ← HTTP リクエスト/レスポンス
         ↓
Domain Layer (internal/domain/) ← ビジネスモデル定義
         ↓
Repository Layer (internal/repository/) ← データアクセス抽象化
         ↓
      Database
```

## 設計方針

### Repository パターン

- データアクセスをインターフェースで抽象化
- Memory / SQLite / PostgreSQL / DynamoDB / Turso を切り替え可能
- 新しいエンティティ追加時は `repository/interface.go` にインターフェース追加
- ページネーションはカーソルベース（`next_cursor` フィールド）を使用

### ドメインモデル

| モデル | 備考 |
|--------|------|
| User | ID, Email, DisplayName |
| Session | ClaudeSessionID, ProjectPath, GitInfo（UserID は nullable） |
| Event | SessionID, EventType, Payload（map[string]interface{}） |
| APIKey | bcrypt ハッシュ |
| WebSession | 7日間有効 |
| PlanDocument | git_remote_url でリポジトリに紐付け |

### イベントフィルタリング

セッション詳細APIでは内部イベントを除外:
- `file-history-snapshot`: Claude Code 内部のファイル履歴追跡
- `system`: init, mcp_server_status, stop_hook_summary 等

### セッションタイトルの自動生成

`/api/ingest` でイベントを受信する際、最初の有効なユーザーメッセージからセッションタイトルを自動生成する（`internal/api/ingest.go`）。

**対象**: `type: "user"` かつタイトル未設定のセッション

**スキップ対象**（以下のメッセージはタイトル生成に使用しない）:
- `isMeta: true` のメタメッセージ（Caveat等）
- `<command-name>` で始まるコマンドメッセージ（`/clear` 等）
- `<local-command-stdout>` で始まるローカルコマンド出力
- `<system-reminder>` で始まるシステムメッセージ
- `/` で始まるスラッシュコマンド
- `Caveat:` で始まるメッセージ

**content形式**: Claude Code のメッセージは `message.content` が文字列形式（API形式の配列ではない）

## 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類 | memory |
| `DATABASE_URL` | DB接続文字列 | - |
| `DEV_MODE` | デバッグログ有効化 | false |
| `WEB_URL` | フロントエンドURL（CORS許可・リダイレクト用） | - |
| `GITHUB_CLIENT_ID` | GitHub OAuth Client ID | - |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth Client Secret | - |

### データベース設定

| DB_TYPE | DATABASE_URL 例 | 利用シーン |
|---------|-----------------|------------|
| memory | - | 開発・テスト |
| sqlite | `./data/agentrace.db` | ローカル/小規模運用 |
| postgres | `postgres://user:pass@localhost:5432/agentrace?sslmode=disable` | 本番運用 |
| dynamodb | `dynamodb://us-east-1/agentrace_` または `dynamodb://localhost:8000/agentrace_` | AWS環境/サーバーレス |
| turso | `libsql://[db-name]-[org].turso.io?authToken=[token]` | エッジ環境/サーバーレス |

## 認証方式

| 方式 | 用途 | 有効期間 |
|------|------|---------|
| Bearer 認証 | CLI → Server | 無期限（APIキー） |
| Session 認証 | Web → Server | 7日間（Cookie） |

### ミドルウェア

| ミドルウェア | 説明 |
|-------------|------|
| CORS | `WEB_URL`で指定されたオリジンからのクロスオリジンリクエストを許可 |
| RequestLogger | `DEV_MODE=true`時にリクエストをログ出力 |

### 認証ミドルウェア

| エンドポイント | ミドルウェア |
|---------------|-------------|
| `/api/ingest` | AuthenticateBearer |
| `/api/auth/web-session` | AuthenticateBearer |
| `/api/me`, `/api/keys`, `/api/users` | AuthenticateSession |
| `/api/sessions`, `/api/plans` (GET) | OptionalBearerOrSession |
| `/api/plans` (POST/PATCH/DELETE) | AuthenticateBearerOrSession |

## APIエンドポイント

### データ受信（CLI用）

| Method | Path | 説明 |
|--------|------|------|
| POST | `/api/ingest` | transcript行を受信 |
| GET | `/api/sessions` | セッション一覧 |
| GET | `/api/sessions/:id` | セッション詳細（イベント含む） |

### PlanDocument

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/plans` | Plan一覧（?git_remote_url=でフィルタ） |
| GET | `/api/plans/:id` | Plan詳細 |
| GET | `/api/plans/:id/events` | Plan変更履歴 |
| POST | `/api/plans` | Plan作成 |
| PATCH | `/api/plans/:id` | Plan更新 |
| DELETE | `/api/plans/:id` | Plan削除 |

#### PlanDocument ステータス

ステータスは線形のワークフローではなく、状況に応じて使い分ける。

| ステータス | 説明 |
|-----------|------|
| scratch | 走り書きのメモ。AIと相談しながらプランを作っていく起点 |
| draft | プランをまだ十分に検討できていない状態（任意） |
| planning | プランを検討中 |
| pending | 十分検討したが実装には進まない状態（任意） |
| implementation | 実装作業中 |
| complete | 完了 |

**基本フロー**: scratch → planning → implementation → complete

draft と pending は必要に応じて使用する補助的なステータス。

### 認証

| Method | Path | 説明 |
|--------|------|------|
| POST | `/auth/register` | ユーザー登録 |
| POST | `/auth/login` | ログイン |
| GET | `/auth/github` | GitHub OAuth開始 |
| GET | `/auth/github/callback` | GitHub OAuthコールバック |
| POST | `/api/auth/web-session` | Webログイントークン発行 |
| GET | `/api/me` | 自分の情報 |
| GET | `/api/keys` | 自分のAPIキー一覧 |
| POST | `/api/keys` | 新しいAPIキー発行 |

## API変更時の注意

- 認証ミドルウェアの適用は `internal/api/router.go` で設定
- エラーレスポンス: `{"error": "メッセージ"}` 形式
- ステータスコード: 400（入力エラー）, 401（認証失敗）, 403（権限不足）, 404（未検索）, 409（競合）

## 開発時の起動

```bash
DEV_MODE=true DB_TYPE=sqlite DATABASE_URL=./db.sqlite3 WEB_URL=http://localhost:5173 go run ./cmd/server
```

- `WEB_URL`を設定することで、フロントエンド（localhost:5173）からのCORSリクエストを許可

## テスト

### Repository テスト

共通テストスイート（`testsuite/`）を使用して、全実装で同じ振る舞いを保証。

```bash
# 通常テスト（memory, sqlite）
go test ./internal/repository/memory/...
go test ./internal/repository/sqlite/...

# 統合テスト（要DB接続、ビルドタグ integration）
DATABASE_URL="postgres://..." go test -tags=integration ./internal/repository/postgres/...
DATABASE_URL="dynamodb://localhost:8000/test_" go test -tags=integration ./internal/repository/dynamodb/...
TURSO_URL="libsql://..." go test -tags=integration ./internal/repository/turso/...
```

### テストスイート構成

| スイート | テスト対象 |
|---------|-----------|
| ProjectRepositorySuite | Project CRUD、FindOrCreate |
| SessionRepositorySuite | Session CRUD、各種Update |
| EventRepositorySuite | Event作成、重複チェック |
| UserRepositorySuite | User CRUD |
| APIKeyRepositorySuite | APIKey CRUD、LastUsedAt更新 |
| WebSessionRepositorySuite | WebSession CRUD、期限切れ削除 |
| PasswordCredentialRepositorySuite | パスワード認証情報 CRUD |
| OAuthConnectionRepositorySuite | OAuth連携 CRUD |
| PlanDocumentRepositorySuite | Plan CRUD、Find クエリ |
| PlanDocumentEventRepositorySuite | Plan変更イベント、共同作業者取得 |
| UserFavoriteRepositorySuite | お気に入り CRUD |
