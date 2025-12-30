# Agentrace

Claude Codeのやりとりをチームでレビューできるサービス

## 想定する利用シーン

| シーン | 説明 |
| ------ | ---- |
| ローカル | 個人でローカル起動してシングルユーザで使用 |
| イントラネット | 社内サーバーにホストしてチームで使用 |
| 開発 | DEV_MODEでデバッグログを出力しながら開発 |

※ インターネット公開は想定しない（GitHub OAuthはオプションで対応）

## 技術スタック

| コンポーネント | 技術 |
| -------------- | ---- |
| CLI | Node.js / TypeScript（npx配布） |
| バックエンド | Go + Gorilla Mux |
| データ層 | Repository パターン（Memory / SQLite / PostgreSQL / MongoDB） |
| フロントエンド | React + Vite + Tailwind CSS |

## プロジェクト構成

```
agentrace/
├── cli/                       # npx agentrace CLI
├── server/                    # バックエンドサーバー
└── web/                       # フロントエンド
```

## アーキテクチャ

```text
┌─────────────────────────────────────────────────────────────┐
│ ユーザー登録（Web）                                          │
├─────────────────────────────────────────────────────────────┤
│  ブラウザで http://server:8080 にアクセス                    │
│      ↓                                                      │
│  「Register」→ email + password 入力 → APIキー発行          │
│      ↓                                                      │
│  APIキーをコピー（この1回のみ表示）                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ CLI セットアップ（ブラウザ連携）                             │
├─────────────────────────────────────────────────────────────┤
│  $ npx agentrace init --url http://server:8080              │
│      ↓                                                      │
│  1. CLIがワンタイムトークンを生成                            │
│  2. ブラウザで http://server:8080/setup?token=xxx を開く     │
│  3. CLIはローカルでHTTPサーバーを起動して待機                │
│      ↓                                                      │
│  Web: 未ログインなら登録/ログイン → セットアップ画面        │
│      ↓                                                      │
│  「Setup CLI」ボタン押下                                     │
│      ↓                                                      │
│  1. POST /api/keys でAPIキー生成                             │
│  2. CLIのコールバックURLにAPIキーをPOST                      │
│      ↓                                                      │
│  CLI: APIキー受信 → config保存 → hooks追加 → 完了           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ データ送信（自動・差分更新）                                 │
├─────────────────────────────────────────────────────────────┤
│  Claude Code → Stop hook 発火                               │
│      ↓                                                      │
│  npx agentrace send                                         │
│      ↓ 環境変数 CLAUDE_PROJECT_DIR でプロジェクトパス取得   │
│      ↓ transcript_path から差分読み取り                     │
│      ↓ 初回送信時: git remote URL, branch も取得            │
│      ↓ HTTP POST /api/ingest (Bearer認証)                   │
│  Agentrace Server                                           │
│      ↓ APIKey → User 解決、UserIDをセッションに紐付け       │
│      ↓ git情報をセッションに保存（初回のみ）                │
│  Database（Memory / SQLite / PostgreSQL / MongoDB）         │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Webログイン                                                  │
├─────────────────────────────────────────────────────────────┤
│  方法1: $ npx agentrace login → URL発行 → ブラウザで開く    │
│  方法2: Webでemail + passwordを入力してログイン              │
│  方法3: GitHub OAuthでログイン（設定時のみ）                 │
│      ↓                                                      │
│  セッションCookie発行 → ダッシュボードへ                    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ レビュー                                                     │
├─────────────────────────────────────────────────────────────┤
│  Web UI（セッション認証）                                    │
│      ↓ REST API                                             │
│  Agentrace Server                                           │
│      ↓                                                      │
│  セッション一覧 → 詳細 → イベントタイムライン               │
│  （全ユーザーのセッションが閲覧可能）                        │
└─────────────────────────────────────────────────────────────┘
```

## ディレクトリ構成

```
agentrace/
├── cli/                         # Node.js CLI (TypeScript)
│   ├── src/
│   │   ├── index.ts             # エントリーポイント
│   │   ├── commands/
│   │   │   ├── init.ts          # 初期設定（ブラウザ連携）
│   │   │   ├── login.ts         # Webログイン
│   │   │   ├── send.ts          # イベント送信（hooks用）
│   │   │   ├── on.ts            # hooks有効化
│   │   │   ├── off.ts           # hooks無効化
│   │   │   └── uninstall.ts     # 設定削除
│   │   ├── config/
│   │   │   ├── manager.ts       # ~/.agentrace/config.json 管理
│   │   │   └── cursor.ts        # 送信済み行数管理
│   │   ├── hooks/
│   │   │   └── installer.ts     # ~/.claude/settings.json 編集
│   │   └── utils/
│   │       ├── http.ts          # HTTP クライアント
│   │       ├── callback-server.ts  # ローカルコールバックサーバー
│   │       └── browser.ts       # ブラウザ起動
│   └── package.json
│
├── server/                      # Go バックエンド
│   ├── cmd/server/main.go       # エントリーポイント
│   ├── migrations/              # DBマイグレーション
│   │   ├── sqlite/
│   │   └── postgres/
│   └── internal/
│       ├── api/                 # HTTP ハンドラ
│       │   ├── router.go
│       │   ├── middleware.go
│       │   ├── ingest.go
│       │   ├── session.go
│       │   ├── auth.go
│       │   └── github_oauth.go  # GitHub OAuth
│       ├── config/config.go     # 環境変数管理
│       ├── domain/              # ドメインモデル
│       │   ├── session.go
│       │   ├── event.go
│       │   ├── user.go
│       │   ├── password_credential.go  # パスワード認証情報
│       │   ├── oauth_connection.go     # OAuth連携
│       │   ├── apikey.go
│       │   └── websession.go
│       └── repository/          # データアクセス層
│           ├── interface.go
│           ├── factory.go
│           ├── memory/          # オンメモリ実装
│           ├── sqlite/          # SQLite実装
│           ├── postgres/        # PostgreSQL実装
│           └── mongodb/         # MongoDB実装
│
└── web/                         # React フロントエンド
    ├── src/
    │   ├── main.tsx             # エントリーポイント
    │   ├── App.tsx              # ルーティング
    │   ├── api/                 # API クライアント
    │   │   ├── client.ts        # fetch ラッパー
    │   │   ├── auth.ts          # 認証 API
    │   │   ├── sessions.ts      # セッション API
    │   │   └── keys.ts          # APIキー API
    │   ├── components/          # UIコンポーネント
    │   │   ├── layout/          # Layout, Header
    │   │   ├── sessions/        # SessionCard, SessionList
    │   │   ├── timeline/        # Timeline, ContentBlockCard, EventCard, UserMessage, AssistantMessage, ToolUse
    │   │   ├── settings/        # ApiKeyList, ApiKeyForm
    │   │   └── ui/              # Button, Input, Card, Modal, CopyButton, Spinner
    │   ├── hooks/               # カスタムHooks
    │   │   └── useAuth.ts       # 認証状態管理
    │   ├── lib/                 # ユーティリティ
    │   │   └── cn.ts            # clsx + tailwind-merge
    │   ├── pages/               # ページコンポーネント
    │   │   ├── WelcomePage.tsx
    │   │   ├── RegisterPage.tsx
    │   │   ├── LoginPage.tsx
    │   │   ├── SetupPage.tsx     # CLIセットアップ
    │   │   ├── SessionListPage.tsx
    │   │   ├── SessionDetailPage.tsx
    │   │   └── SettingsPage.tsx
    │   └── types/               # 型定義
    │       ├── auth.ts
    │       ├── session.ts
    │       └── event.ts
    └── package.json
```

## 開発環境での動作確認

### 1. サーバー起動

```bash
cd server
DEV_MODE=true WEB_URL=http://localhost:5173 go run ./cmd/server
```

- `DEV_MODE=true` でリクエストログを出力
- `WEB_URL` でフロントエンドURLを指定（CLI init時のリダイレクト先）

### 2. ユーザー登録

```bash
# curlでユーザー登録（APIキーが返される）
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com", "password": "yourpassword"}'
# => {"user": {...}, "api_key": "agtr_xxxxx"}
```

### 3. CLI初期化（開発モード）

```bash
cd cli
npm install
npx tsx src/index.ts init --url http://localhost:8080 --dev
# ブラウザが開く → ログイン/登録 → 自動的にAPIキー取得
```

- `--url` でサーバーURLを指定（必須）
- `--dev` オプションでローカルCLIパスを使用

### 4. Web UI起動

```bash
cd web
npm install
npm run dev
```

- http://localhost:5173 でアクセス
- Viteのプロキシ設定でAPIリクエストは自動的に localhost:8080 に転送

### 5. 動作確認

Claude Codeで操作すると、Stopイベントごとにtranscript差分がサーバーに送信される

```bash
# セッション一覧取得
curl -H "Authorization: Bearer agtr_xxxxx" http://localhost:8080/api/sessions

# セッション詳細取得
curl -H "Authorization: Bearer agtr_xxxxx" http://localhost:8080/api/sessions/{id}

# Webダッシュボードにログイン
npx tsx src/index.ts login
```

## CLIコマンド

| コマンド | 説明 |
|---------|------|
| `agentrace init --url <url>` | ブラウザ連携で設定 + hooks インストール |
| `agentrace init --url <url> --dev` | 開発モード（ローカルCLIパス使用） |
| `agentrace login` | WebログインURL発行 |
| `agentrace send` | transcript差分送信（hooks用） |
| `agentrace on` | hooks有効化（認証情報は保持） |
| `agentrace on --dev` | hooks有効化（開発モード） |
| `agentrace off` | hooks無効化（認証情報は保持） |
| `agentrace uninstall` | hooks/config 削除 |

### 設定ファイル

**~/.agentrace/config.json**

```json
{
  "server_url": "http://localhost:8080",
  "api_key": "agtr_xxxxxxxxxxxxxxxxxxxxxxxx"
}
```

### Hooks 設定

**~/.claude/settings.json に追加**

```json
{
  "hooks": {
    "Stop": [{
      "hooks": [{
        "type": "command",
        "command": "npx agentrace send"
      }]
    }]
  }
}
```

## API エンドポイント

### データ受信（CLI用）

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| POST | `/api/ingest` | Bearer | transcript行を受信 |
| GET | `/api/sessions` | Bearer/Session | セッション一覧 |
| GET | `/api/sessions/:id` | Bearer/Session | セッション詳細（イベント含む） |
| GET | `/health` | なし | ヘルスチェック |

### 認証

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| POST | `/auth/register` | なし | ユーザー登録（email, password → APIキー発行） |
| POST | `/auth/login` | なし | メール/パスワードでログイン |
| POST | `/auth/login/apikey` | なし | APIキーでログイン |
| GET | `/auth/session` | なし | トークンでログイン（CLI経由） |
| GET | `/auth/config` | なし | OAuth設定確認（有効なプロバイダー） |
| GET | `/auth/github` | なし | GitHub OAuth開始（リダイレクト） |
| GET | `/auth/github/callback` | なし | GitHub OAuthコールバック |
| POST | `/api/auth/web-session` | Bearer | Webログイントークン発行 |
| POST | `/api/auth/logout` | Session | ログアウト |
| GET | `/api/me` | Session | 自分の情報 |
| GET | `/api/users` | Session | ユーザー一覧 |
| GET | `/api/keys` | Session | 自分のAPIキー一覧 |
| POST | `/api/keys` | Session | 新しいAPIキー発行 |
| DELETE | `/api/keys/:id` | Session | APIキー削除 |

### APIリクエスト/レスポンス形式

**POST /api/ingest リクエスト**

```json
{
  "session_id": "string",
  "transcript_lines": [{"type": "...", ...}],
  "cwd": "string (作業ディレクトリ)",
  "git_remote_url": "string (git remote origin URL)",
  "git_branch": "string (現在のブランチ名)"
}
```

**Session レスポンス**

```json
{
  "id": "string",
  "user_id": "string | null",
  "user_name": "string | null",
  "claude_session_id": "string",
  "project_path": "string",
  "git_remote_url": "string",
  "git_branch": "string",
  "started_at": "ISO 8601",
  "ended_at": "ISO 8601 | null",
  "event_count": "number",
  "created_at": "ISO 8601"
}
```

### イベントフィルタリング

セッション詳細APIでは、内部イベントをフィルタリングして返す：

| フィルタ対象 | 理由 |
| ------------ | ---- |
| `file-history-snapshot` | Claude Code内部のファイル履歴追跡 |
| `system` | システムイベント（init, mcp_server_status, stop_hook_summary等） |

## 環境変数（サーバー）

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類（`memory` / `sqlite` / `postgres` / `mongodb`） | memory |
| `DATABASE_URL` | DB接続文字列 | - |
| `DEV_MODE` | デバッグログ有効化 | false |
| `WEB_URL` | フロントエンドURL（開発時のリダイレクト用） | - |
| `GITHUB_CLIENT_ID` | GitHub OAuth Client ID | - |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth Client Secret | - |

### DATABASE_URL の形式

| DB_TYPE | DATABASE_URL 例 |
|---------|-----------------|
| sqlite | `./data/agentrace.db` |
| postgres | `postgres://user:pass@localhost:5432/agentrace?sslmode=disable` |
| mongodb | `mongodb://user:pass@localhost:27017/agentrace` |

### 対応データベース

| DB | DB_TYPE | 利用シーン |
| -- | ------- | ---------- |
| オンメモリ | `memory` | 開発・テスト |
| SQLite3 | `sqlite` | ローカル/小規模運用 |
| PostgreSQL | `postgres` | イントラネット/本番運用 |
| MongoDB | `mongodb` | AWS (DocumentDB) 環境 |

## 認証フロー

### ユーザー登録（Web）

1. ブラウザで http://server:8080 にアクセス
2. 「Register」→ email + password 入力 → APIキー発行
3. APIキーをコピー（この1回のみ表示）

### CLIセットアップ（ブラウザ連携）

```bash
npx agentrace init --url http://server:8080
```

1. CLIがワンタイムトークンを生成し、ローカルHTTPサーバー起動
2. ブラウザで `/setup?token=xxx&callback=http://localhost:xxxxx/callback` を開く
3. 未ログインなら登録/ログイン画面を経由
4. セットアップ画面で「Setup CLI」ボタン押下
5. WebがAPIキーを生成し、CLIのコールバックURLにPOST
6. CLIがAPIキーを受信、config保存、hooks追加

セキュリティ:
- トークンは `crypto.randomUUID()` で生成（推測困難）
- コールバックURLは `localhost` のみ許可
- タイムアウト5分

### Webログイン

- 方法1: `npx agentrace login` → URL発行 → ブラウザで開く
- 方法2: Webでemail + passwordを入力してログイン
- 方法3: GitHub OAuthでログイン（`GITHUB_CLIENT_ID/SECRET` 設定時のみ）

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

### GitHub OAuth

環境変数 `GITHUB_CLIENT_ID` と `GITHUB_CLIENT_SECRET` が設定されている場合のみ有効。

```
1. ユーザーが「Continue with GitHub」をクリック
2. GET /auth/github → GitHub認証画面にリダイレクト
3. GitHub認証完了 → GET /auth/github/callback
4. GitHubユーザー情報を取得
5. 既存ユーザー（email一致 or OAuth連携済み）ならログイン
   新規ならアカウント作成
6. セッションCookie発行、ダッシュボードへリダイレクト
```

OAuthConnectionテーブルでGitHubのユーザーIDとローカルユーザーを紐付け管理。

### 複数APIキー

- 各ユーザーは複数のAPIキーを発行可能（別デバイス用など）
- Webの設定画面（/settings）でAPIキーの管理
- キー発行時に名前を付けられる（例: "MacBook Pro", "Work PC"）

## データフロー

1. Claude Code が応答完了 → Stop hook 発火
2. CLI: stdin から session_id, transcript_path を取得
3. CLI: 環境変数 `CLAUDE_PROJECT_DIR` からプロジェクトルートを取得（フォールバック: stdin の cwd）
4. CLI: transcript_path のJSONLを読み、前回からの差分を抽出
5. CLI: 初回送信時のみ、プロジェクトルートで `git remote get-url origin` と `git branch --show-current` を実行してgit情報を取得
6. CLI: 差分とgit情報をサーバーに POST /api/ingest（Bearer認証）
7. Server: APIKey → User解決、UserIDをセッションに紐付け
8. Server: project_path, git_remote_url, git_branch をセッションに保存（初回のみ）
9. Server: 各行を Event として保存
10. CLI: カーソル位置を更新（~/.agentrace/cursors/{session_id}.json）

### プロジェクトパスについて

| 変数 | 取得元 | 特徴 |
| ---- | ------ | ---- |
| `CLAUDE_PROJECT_DIR` | 環境変数 | Claude Code起動時のパス（固定） |
| `cwd` | stdin JSON | ビルドコマンド等で変わる可能性あり |

CLIは `CLAUDE_PROJECT_DIR` を優先して使用し、未設定時は `cwd` にフォールバックする。

## Web フロントエンド

### 技術スタック

| カテゴリ | 技術 |
| -------- | ---- |
| ビルドツール | Vite 7 |
| UIライブラリ | React 19 |
| 言語 | TypeScript |
| スタイリング | Tailwind CSS 3 |
| ルーティング | React Router v7 |
| 状態管理 | TanStack Query (React Query) v5 + AuthContext |
| フォーム | React Hook Form |
| 日時処理 | date-fns |
| アイコン | Lucide React |
| コード表示 | react-syntax-highlighter |
| Markdown | react-markdown + @tailwindcss/typography |

### ソート仕様

| 対象 | ソートキー | 順序 |
| ---- | ---------- | ---- |
| セッション一覧 | StartedAt | 降順（新しい順） |
| イベント一覧 | payload.timestamp | 昇順（会話順） |

- イベントの時刻表示は `payload.timestamp` を優先（Claude Codeのオリジナルタイムスタンプ）
- `created_at` はサーバー保存時刻なのでフォールバックとして使用

### メッセージ表示

ContentBlockCardコンポーネントは以下のブロックタイプに対応：

| ブロックタイプ | 表示 |
| -------------- | ---- |
| text | Markdown対応テキスト表示（コードブロックはシンタックスハイライト） |
| thinking | 折りたたみ可能なUI（紫色、デフォルト折りたたみ） |
| tool_group | ツール呼び出しと結果をグループ化（Input/Result縦並び、デフォルト折りたたみ） |
| tool_use | ツール名 + JSONハイライト表示（デフォルト折りたたみ） |
| tool_result | ツール結果表示（デフォルト折りたたみ） |
| local_command_group | ローカルコマンド（/compact等）と関連イベントをグループ化（デフォルト折りたたみ） |
| compact_summary | compactコマンドのサマリー表示（amber背景） |
| local_command_output | コマンド出力表示 |
| その他 | ブロックタイプ名 + JSON表示 |

### セッション詳細表示

SessionDetailPageでは以下の情報を表示：

- **プロジェクトパス**: 作業ディレクトリ（フォルダアイコン付き）
- **Gitリポジトリ**: git remote URLからGitHub/GitLabリンクを生成（外部リンクアイコン付き）
- **Gitブランチ**: 現在のブランチ名
- **ユーザー**: セッションを作成したユーザー名
- **開始時刻**: セッション開始日時
- **イベントタイムライン**: 会話の全イベント

### イベントのグルーピング

タイムライン表示では関連イベントを自動的にグループ化：

**Tool グループ化**
- `tool_use`ブロックと対応する`tool_result`を1つのカードにまとめる
- `tool_use.id`と`tool_result.tool_use_id`で紐付け
- ファイル操作ツール（Read, Edit, Write, Glob, Grep等）はファイルパスを表示

**ローカルコマンド グループ化**
- `/compact`等のローカルコマンドと関連イベントを1つのカードにまとめる
- 対象: メタメッセージ（`payload.isMeta`）、サマリー（`payload.isCompactSummary`）、コマンド出力
- コマンドの検出: コンテンツが`<command-name>/`で始まる
- 出力の検出: コンテンツに`<local-command-stdout>`を含む

## 将来の拡張（スコープ外）

- リアルタイム機能（WebSocket）
- コメント機能（セッション/イベントへのコメント）
- セッションの再開機能（コンテキストをClaude Codeに渡す）
- Slack/Discord通知
- 統計ダッシュボード
- セッションのエクスポート（Markdown等）
- Google OAuth
