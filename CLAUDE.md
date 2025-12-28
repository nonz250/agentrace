# Agentrace

Claude Codeのやりとりをチームでレビューできるサービス

## 想定する利用シーン

| シーン | 説明 |
| ------ | ---- |
| ローカル | 個人でローカル起動してシングルユーザで使用 |
| イントラネット | 社内サーバーにホストしてチームで使用 |
| 開発 | DEV_MODEでデバッグログを出力しながら開発 |

## ディレクトリ構成

```
agentrace/
├── cli/                         # Node.js CLI (TypeScript)
│   ├── src/
│   │   ├── index.ts             # エントリーポイント
│   │   ├── commands/
│   │   │   ├── init.ts          # 初期設定
│   │   │   ├── login.ts         # Webログイン
│   │   │   ├── send.ts          # イベント送信（hooks用）
│   │   │   └── uninstall.ts     # 設定削除
│   │   ├── config/
│   │   │   ├── manager.ts       # ~/.agentrace/config.json 管理
│   │   │   └── cursor.ts        # 送信済み行数管理
│   │   ├── hooks/
│   │   │   └── installer.ts     # ~/.claude/settings.json 編集
│   │   └── utils/
│   │       └── http.ts          # HTTP クライアント
│   └── package.json
│
├── server/                      # Go バックエンド
│   ├── cmd/server/main.go       # エントリーポイント
│   └── internal/
│       ├── api/                 # HTTP ハンドラ
│       │   ├── router.go
│       │   ├── middleware.go
│       │   ├── ingest.go
│       │   ├── session.go
│       │   └── auth.go
│       ├── config/config.go     # 環境変数管理
│       ├── domain/              # ドメインモデル
│       │   ├── session.go
│       │   ├── event.go
│       │   ├── user.go
│       │   ├── apikey.go
│       │   └── websession.go
│       └── repository/          # データアクセス層
│           ├── interface.go
│           └── memory/          # オンメモリ実装
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
    │   │   ├── timeline/        # EventCard, AssistantMessage, etc.
    │   │   ├── settings/        # ApiKeyList, ApiKeyForm
    │   │   └── ui/              # Button, Input, Card, etc.
    │   ├── hooks/               # カスタムHooks
    │   │   └── useAuth.ts       # 認証状態管理
    │   ├── lib/                 # ユーティリティ
    │   │   └── cn.ts            # clsx + tailwind-merge
    │   ├── pages/               # ページコンポーネント
    │   │   ├── WelcomePage.tsx
    │   │   ├── RegisterPage.tsx
    │   │   ├── LoginPage.tsx
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
DEV_MODE=true go run ./cmd/server
```

- `DEV_MODE=true` でリクエストログを出力

### 2. ユーザー登録

```bash
# curlでユーザー登録（APIキーが返される）
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Your Name"}'
# => {"user": {...}, "api_key": "agtr_xxxxx"}
```

### 3. CLI初期化（開発モード）

```bash
cd cli
npm install
npx tsx src/index.ts init --dev
# Server URL: http://localhost:8080
# API Key: (上記で取得したAPIキー)
```

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
| `agentrace init` | 設定 + hooks インストール |
| `agentrace init --dev` | 開発モード（ローカルCLIパス使用） |
| `agentrace login` | WebログインURL発行 |
| `agentrace send` | transcript差分送信（hooks用） |
| `agentrace uninstall` | hooks/config 削除 |

## API エンドポイント

### Step 1（実装済み）

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| POST | `/api/ingest` | Bearer | transcript行を受信 |
| GET | `/api/sessions` | Bearer/Session | セッション一覧 |
| GET | `/api/sessions/:id` | Bearer/Session | セッション詳細（イベント含む） |
| GET | `/health` | なし | ヘルスチェック |

### Step 2（実装済み）

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| POST | `/auth/register` | なし | ユーザー登録（名前→APIキー発行） |
| POST | `/auth/login` | なし | APIキーでログイン |
| GET | `/auth/session` | なし | トークンでログイン（CLI経由） |
| POST | `/api/auth/web-session` | Bearer | Webログイントークン発行 |
| POST | `/api/auth/logout` | Session | ログアウト |
| GET | `/api/me` | Session | 自分の情報 |
| GET | `/api/users` | Session | ユーザー一覧 |
| GET | `/api/keys` | Session | 自分のAPIキー一覧 |
| POST | `/api/keys` | Session | 新しいAPIキー発行 |
| DELETE | `/api/keys/:id` | Session | APIキー削除 |

## 環境変数（サーバー）

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類（`memory` / `sqlite` / `postgres` / `mongodb`） | memory |
| `DATABASE_URL` | DB接続文字列（Step 4） | - |
| `DEV_MODE` | デバッグログ有効化 | false |

**DATABASE_URL の形式:**

| DB_TYPE | DATABASE_URL 例 |
|---------|-----------------|
| sqlite | `./data/agentrace.db` |
| postgres | `postgres://user:pass@localhost:5432/agentrace?sslmode=disable` |
| mongodb | `mongodb://user:pass@localhost:27017/agentrace` |

## 認証フロー

### ユーザー登録（Web）

1. ブラウザで http://server:8080 にアクセス
2. 「Register」→ 名前入力 → APIキー発行
3. APIキーをコピー（この1回のみ表示）

### CLIセットアップ

1. `npx agentrace init`
2. Server URLとAPIキーを入力
3. hooks自動設定

### Webログイン

- 方法1: `npx agentrace login` → URL発行 → ブラウザで開く
- 方法2: WebでAPIキーを入力してログイン

### 複数APIキー

- 各ユーザーは複数のAPIキーを発行可能（別デバイス用など）
- Webの設定画面（/settings）でAPIキーの管理
- キー発行時に名前を付けられる（例: "MacBook Pro", "Work PC"）

## データフロー

1. Claude Code が応答完了 → Stop hook 発火
2. CLI: stdin から session_id, transcript_path を取得
3. CLI: transcript_path のJSONLを読み、前回からの差分を抽出
4. CLI: 差分をサーバーに POST /api/ingest（Bearer認証）
5. Server: APIKey → User解決、UserIDをセッションに紐付け
6. Server: 各行を Event として保存
7. CLI: カーソル位置を更新（~/.agentrace/cursors/{session_id}.json）

## Web フロントエンド（Step 3）

### 技術スタック

| カテゴリ | 技術 |
| -------- | ---- |
| ビルドツール | Vite |
| UIライブラリ | React 18 |
| 言語 | TypeScript |
| スタイリング | Tailwind CSS |
| ルーティング | React Router v6 |
| 状態管理 | TanStack Query (React Query) |
| フォーム | React Hook Form |
| 日時処理 | date-fns |
| アイコン | Lucide React |
| コード表示 | react-syntax-highlighter |

### ソート仕様

| 対象 | ソートキー | 順序 |
| ---- | ---------- | ---- |
| セッション一覧 | StartedAt | 降順（新しい順） |
| イベント一覧 | payload.timestamp | 昇順（会話順） |

- イベントの時刻表示は `payload.timestamp` を優先（Claude Codeのオリジナルタイムスタンプ）
- `created_at` はサーバー保存時刻なのでフォールバックとして使用

### メッセージ表示

AssistantMessageコンポーネントは以下のブロックタイプに対応：

| ブロックタイプ | 表示 |
| -------------- | ---- |
| text | テキストそのまま表示 |
| thinking | 折りたたみ可能なUI（紫色、Brainアイコン） |
| tool_use | ツール名 + JSONハイライト表示 |
| tool_result | ツール結果のpre表示 |
| その他 | ブロックタイプ名 + JSON表示 |
