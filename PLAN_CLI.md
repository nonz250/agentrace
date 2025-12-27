# CLI 実装計画

## 概要

`npx agentrace` で利用可能なCLIツール。Claude Codeのhooksと連携してイベントをサーバーに送信する。

## ディレクトリ構成

```text
cli/
├── src/
│   ├── index.ts              # エントリーポイント（commander.js）
│   ├── commands/
│   │   ├── init.ts           # 初期設定（APIキー取得）
│   │   ├── send.ts           # イベント送信（hooks用）
│   │   ├── status.ts         # 接続状態確認
│   │   └── logout.ts         # 設定削除
│   ├── config/
│   │   └── manager.ts        # ~/.agentrace/config.json 管理
│   ├── hooks/
│   │   └── installer.ts      # ~/.claude/settings.json 編集
│   └── utils/
│       └── http.ts           # HTTP クライアント
├── package.json
├── tsconfig.json
└── README.md
```

## コマンド一覧

| コマンド | 説明 |
| -------- | ---- |
| `npx agentrace init` | ログイン + APIキー取得 + hooks自動設定 |
| `npx agentrace send` | イベント送信（hooks用、stdinからJSON受取） |
| `npx agentrace status` | 接続状態・認証状態確認 |
| `npx agentrace logout` | 設定削除 + hooks削除 |

## コマンド詳細

### init

**Step 1（最小動作版）:**

```text
$ npx agentrace init
? Server URL: https://agentrace.example.com
? API Key: agtr_xxxxxxxxxxxx
✓ Config saved to ~/.agentrace/config.json
✓ Hooks added to ~/.claude/settings.json
Setup complete!
```

**Step 2（ブラウザ連携版）:**

```text
$ npx agentrace init
Opening browser for authentication...
Waiting for callback on http://localhost:19876 ...
✓ Authenticated as user@example.com
✓ Config saved to ~/.agentrace/config.json
✓ Hooks added to ~/.claude/settings.json
Setup complete!
```

### send

hooks から呼び出される。stdin から JSON を受け取り、サーバーに POST する。

```text
[Claude Code]
    ↓ hook 呼び出し（stdin に JSON）
[npx agentrace send]
    ↓ config.json から API キー読み込み
    ↓ HTTP POST /api/ingest
[Server]
```

**入力（stdin）:**

```json
{
  "session_id": "abc123",
  "hook_event_name": "PostToolUse",
  "tool_name": "Bash",
  "tool_input": { "command": "ls -la" },
  "tool_response": { "stdout": "..." }
}
```

**エラーハンドリング:**

- 設定ファイルがない → 警告を stderr に出力、exit 0（hooks をブロックしない）
- サーバー接続エラー → 警告を stderr に出力、exit 0
- 認証エラー → 警告を stderr に出力、exit 0

### status

```text
$ npx agentrace status
Server: https://agentrace.example.com
Status: Connected
User: user@example.com
Workspace: My Team
Hooks: Installed
```

### logout

```text
$ npx agentrace logout
✓ Removed ~/.agentrace/config.json
✓ Removed hooks from ~/.claude/settings.json
Logged out.
```

## 設定ファイル

**~/.agentrace/config.json**

```json
{
  "server_url": "https://agentrace.example.com",
  "api_key": "agtr_xxxxxxxxxxxxxxxxxxxx",
  "workspace_id": "workspace-uuid"
}
```

## Hooks 設定

**~/.claude/settings.json に追加**

```json
{
  "hooks": {
    "PostToolUse": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": "npx agentrace send"
      }]
    }],
    "Stop": [{
      "hooks": [{
        "type": "command",
        "command": "npx agentrace send"
      }]
    }]
  }
}
```

## 依存パッケージ

- `commander` - CLI フレームワーク
- `inquirer` - 対話的プロンプト
- `open` - ブラウザを開く
- `node-fetch` または標準 fetch - HTTP クライアント

## 実装順序

### Step 1: 最小動作版

1. `init` - 手動でAPIキーを入力、config.json保存、hooks設定
2. `send` - stdinからJSON読み取り、POST送信

### Step 2: ブラウザ連携

1. `init` - ローカルHTTPサーバー起動、ブラウザ認証、コールバック受信
2. `status` - サーバー接続確認
3. `logout` - 設定削除

## package.json

```json
{
  "name": "agentrace",
  "version": "0.1.0",
  "bin": {
    "agentrace": "./dist/index.js"
  },
  "scripts": {
    "build": "tsc",
    "dev": "tsx src/index.ts"
  },
  "dependencies": {
    "commander": "^12.0.0",
    "inquirer": "^9.0.0",
    "open": "^10.0.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "tsx": "^4.0.0",
    "@types/node": "^20.0.0"
  }
}
```
