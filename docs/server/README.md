# Server

Agentrace のバックエンドサーバー。データ永続化、認証、REST API を提供。

## 技術スタック

- Go + Gorilla Mux
- Repository パターン（Memory / SQLite / PostgreSQL / MongoDB）

## ディレクトリ構成

```
server/
├── cmd/server/main.go           # エントリーポイント（DI・DB初期化）
├── internal/
│   ├── api/                     # HTTPハンドラーレイヤー
│   │   ├── router.go            # ルーティング定義
│   │   ├── middleware.go        # 認証・ロギング
│   │   ├── auth.go              # 認証ハンドラー
│   │   ├── ingest.go            # データ受信ハンドラー
│   │   ├── session.go           # セッション取得ハンドラー
│   │   └── github_oauth.go      # GitHub OAuth
│   ├── config/config.go         # 環境変数パース
│   ├── domain/                  # ドメインモデル
│   │   ├── user.go
│   │   ├── session.go
│   │   ├── event.go
│   │   ├── apikey.go
│   │   ├── websession.go
│   │   ├── password_credential.go
│   │   └── oauth_connection.go
│   └── repository/              # データアクセス層
│       ├── interface.go         # インターフェース定義
│       ├── factory.go           # Factory パターン
│       ├── memory/              # オンメモリ実装
│       ├── sqlite/              # SQLite実装
│       ├── postgres/            # PostgreSQL実装
│       └── mongodb/             # MongoDB実装
└── migrations/                  # スキーママイグレーション
    ├── embed.go                 # SQLファイル埋め込み
    ├── sqlite/001_initial.sql
    └── postgres/001_initial.up.sql
```

## レイヤードアーキテクチャ

```
┌─────────────────────────────────┐
│     API Layer (internal/api/)   │  ← HTTP リクエスト/レスポンス
└─────────────────────────────────┘
              ↓
┌─────────────────────────────────┐
│  Domain Layer (internal/domain/)│  ← ビジネスモデル定義
└─────────────────────────────────┘
              ↓
┌─────────────────────────────────┐
│  Repository Layer (internal/repo)│ ← データアクセス抽象化
└─────────────────────────────────┘
              ↓
         Database
```

## 設計方針

### Repository パターン

データアクセスをインターフェースで抽象化し、複数のDB実装を切り替え可能:

```go
type Repositories struct {
    Session            SessionRepository
    Event              EventRepository
    User               UserRepository
    APIKey             APIKeyRepository
    WebSession         WebSessionRepository
    PasswordCredential PasswordCredentialRepository
    OAuthConnection    OAuthConnectionRepository
}
```

### ドメインモデル

| モデル | 主要フィールド | 備考 |
|--------|---------------|------|
| User | ID, Email, DisplayName | |
| Session | ClaudeSessionID, ProjectPath, GitInfo | UserID は nullable |
| Event | SessionID, EventType, Payload | Payload は map[string]interface{} |
| APIKey | KeyHash, KeyPrefix | bcrypt ハッシュ |
| WebSession | Token, ExpiresAt | 7日間有効 |
| PasswordCredential | UserID, PasswordHash | bcrypt ハッシュ |
| OAuthConnection | Provider, ProviderID | GitHub連携用 |

### イベントフィルタリング

セッション詳細APIでは内部イベントを除外:
- `file-history-snapshot`: Claude Code 内部のファイル履歴追跡
- `system`: init, mcp_server_status, stop_hook_summary 等

## 関連ドキュメント

- [API エンドポイント](./api.md)
- [認証フロー](./authentication.md)
- [環境変数・DB設定](./configuration.md)
