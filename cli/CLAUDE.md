# CLI 開発ガイド

Claude Code の transcript を Agentrace サーバーに送信する CLI ツール。

## 技術スタック

- Node.js / TypeScript
- Commander.js（CLI フレームワーク）
- npx 配布

## ディレクトリ構成

```
cli/src/
├── index.ts             # エントリーポイント（Commander.js）
├── commands/            # コマンド実装
│   ├── init.ts          # 初期セットアップ（ブラウザ連携）
│   ├── login.ts         # Webログイン
│   ├── send.ts          # transcript送信（hooks用）
│   ├── mcp-server.ts    # MCPサーバー
│   ├── on.ts            # hooks有効化
│   ├── off.ts           # hooks無効化
│   └── uninstall.ts     # 完全アンインストール
├── config/              # 設定管理
│   ├── manager.ts       # ~/.agentrace/config.json CRUD
│   └── cursor.ts        # 差分追跡（送信済み行数）
├── hooks/               # Claude Code hooks連携
│   └── installer.ts     # ~/.claude/settings.json 編集
└── utils/               # ユーティリティ
    ├── http.ts          # HTTP APIクライアント
    ├── callback-server.ts # ローカルHTTP callbackサーバー
    └── browser.ts       # ブラウザ起動
```

## 設計方針

### 責務分離

| レイヤー | 責務 |
|---------|------|
| commands/ | ユーザーコマンドの処理フロー |
| config/ | データ永続化（設定、カーソル位置） |
| hooks/ | Claude Code連携（settings.json編集） |
| utils/ | 外部サービス連携（HTTP、ブラウザ） |

### エラーハンドリング

- **send コマンド**: すべてのエラーで `exit(0)` → hooks をブロックしない
- **init コマンド**: 致命的エラーで `exit(1)` → ユーザーに再試行を促す

### 差分送信の仕組み

1. `~/.agentrace/cursors/{session_id}.json` で送信済み行数を管理
2. JSONL を読み込み、カーソル位置以降の行のみ抽出
3. 送信成功後にカーソル位置を更新

### Git 情報の取得

- 初回送信時のみ取得（パフォーマンス）
- `CLAUDE_PROJECT_DIR` 環境変数を優先
- 未設定時は stdin の `cwd` にフォールバック

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `init --url <url>` | 初期設定 + hooks + MCP インストール |
| `init --url <url> --dev` | 開発モード（ローカルCLIパス使用） |
| `login` | Webログイン URL 発行 |
| `send` | transcript 差分送信（hooks用） |
| `mcp-server` | MCPサーバー起動（stdio通信） |
| `on` / `off` | hooks + MCP 有効化/無効化 |
| `uninstall` | hooks/MCP/config 削除 |

## 設定ファイル

### ~/.agentrace/config.json

```json
{
  "server_url": "http://localhost:8080",
  "api_key": "agtr_xxxxxxxxxxxxxxxxxxxxxxxx"
}
```

### ~/.agentrace/cursors/{session_id}.json

```json
{
  "lineCount": 123,
  "lastUpdated": "2024-01-01T00:00:00.000Z"
}
```

## Claude Code 設定

### ~/.claude/settings.json（hooks設定）

`init` または `on` コマンドで自動追加:

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "npx agentrace send"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "npx agentrace send"
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "npx agentrace send"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "npx agentrace send"
          }
        ]
      }
    ]
  }
}
```

### ~/.claude.json（MCP設定）

MCPサーバーは `settings.json` ではなく `~/.claude.json` に設定:

```json
{
  "mcpServers": {
    "agentrace": {
      "command": "npx",
      "args": ["agentrace", "mcp-server"]
    }
  }
}
```

## MCPサーバー

Claude Code の MCP (Model Context Protocol) サーバーとして動作し、PlanDocument の操作を提供。

### 提供するツール

| ツール | 説明 | 引数 |
|--------|------|------|
| `list_plans` | リポジトリのPlan一覧取得 | `git_remote_url` |
| `read_plan` | Plan読み込み | `id` |
| `create_plan` | Plan作成（project は session から自動取得） | `description`, `body` |
| `update_plan` | Plan更新（パッチ自動生成） | `id`, `body` |
| `set_plan_status` | Planステータス変更 | `id`, `status` |

**注**: `create_plan` と `update_plan` の `session_id` は PreToolUse hook により自動注入されます。

### 使用例

```
# Planの一覧を取得
list_plans(git_remote_url: "https://github.com/user/repo.git")

# 新規Planを作成（session_id は自動注入）
create_plan(
  description: "実装計画",
  body: "# 実装ステップ\n\n1. ..."
)
```

### PreToolUse Hook

`init` コマンドで `~/.agentrace/hooks/inject-session-id.sh` がインストールされ、`~/.claude/settings.json` に登録されます。

このhookは `mcp__agentrace__create_plan` と `mcp__agentrace__update_plan` の呼び出し時に `session_id` を自動注入します。

## Hooks の仕組み

transcript送信は以下の4つのタイミングで発火:

1. **UserPromptSubmit**: ユーザーがメッセージを送信した直後（10秒待機後に送信）
2. **Stop**: Claude Codeが応答を完了した時
3. **SubagentStop**: Taskエージェント（explore, plan等）が完了した時
4. **PostToolUse**: ツール使用完了後（リアルタイム更新用）

どのイベントでも同じ処理:
1. `~/.claude/settings.json` の該当hookを実行
2. stdin に JSON を渡して `npx agentrace send` を実行
3. CLI が stdin から JSON を読み取り、差分をサーバーに送信

### stdin JSON 形式

```json
{
  "session_id": "uuid",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/working/directory"
}
```

## 開発モード

`--dev` オプションを付けると、hooks/MCPコマンドが変わる:

| モード | コマンド |
|--------|----------|
| 本番 | `npx agentrace send` |
| 開発 | `npx tsx /path/to/cli/src/index.ts send` |

## 開発時の起動

```bash
npm install
npx tsx src/index.ts init --url http://localhost:8080 --dev
```
