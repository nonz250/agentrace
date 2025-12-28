# CLI 実装計画

## 概要

`npx agentrace` で利用可能なCLIツール。Claude Codeのhooksと連携してtranscript（会話履歴）をサーバーに送信する。

## ディレクトリ構成

```text
cli/
├── src/
│   ├── index.ts              # エントリーポイント（commander.js）
│   ├── commands/
│   │   ├── init.ts           # 初期設定（APIキー取得）
│   │   ├── send.ts           # transcript差分送信（hooks用）
│   │   └── uninstall.ts      # 設定削除
│   ├── config/
│   │   ├── manager.ts        # ~/.agentrace/config.json 管理
│   │   └── cursor.ts         # 送信済み行数管理
│   ├── hooks/
│   │   └── installer.ts      # ~/.claude/settings.json 編集
│   └── utils/
│       └── http.ts           # HTTP クライアント
├── package.json
└── tsconfig.json
```

## コマンド一覧

| コマンド | 説明 |
| -------- | ---- |
| `npx agentrace init` | APIキー入力 + hooks設定 |
| `npx agentrace init --dev` | 開発モード（ローカルCLIパス使用） |
| `npx agentrace send` | transcript差分送信（hooks用） |
| `npx agentrace uninstall` | 設定削除 + hooks削除 |

## コマンド詳細

### init

**Step 1（最小動作版）:** ✅ 完了

```text
$ npx agentrace init
Agentrace Setup

? Server URL: http://localhost:8080
? API Key: test-key
✓ Config saved to ~/.agentrace/config.json
✓ Hooks added to ~/.claude/settings.json
Setup complete!
```

**開発モード:**

```text
$ npx agentrace init --dev
Agentrace Setup

[Dev Mode] Using local CLI for hooks

? Server URL: http://localhost:8080
? API Key: test-key
✓ Config saved to ~/.agentrace/config.json
  Hook command: npx tsx /path/to/cli/src/index.ts send
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

Stop hook から呼び出される。transcript_path から差分を読み取り、サーバーに POST する。

```text
[Claude Code] → 応答完了（Stop hook発火）
    ↓ stdin に JSON（session_id, transcript_path）
[npx agentrace send]
    ↓ transcript_path のJSONLを読み込み
    ↓ カーソル位置から差分を抽出
    ↓ HTTP POST /api/ingest
[Server]
    ↓ 成功
[npx agentrace send]
    ↓ カーソル位置を更新
```

**入力（stdin）:**

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/.../.claude/projects/.../session.jsonl",
  "cwd": "/path/to/project"
}
```

**カーソル管理:**

```text
~/.agentrace/cursors/{session_id}.json
{
  "lineCount": 42,
  "lastUpdated": "2025-12-28T..."
}
```

**エラーハンドリング:**

- 設定ファイルがない → 警告を stderr に出力、exit 0（hooks をブロックしない）
- サーバー接続エラー → 警告を stderr に出力、exit 0
- 認証エラー → 警告を stderr に出力、exit 0

### uninstall

```text
$ npx agentrace uninstall
Uninstalling Agentrace...

✓ Removed hooks from ~/.claude/settings.json
✓ Config removed
Uninstall complete!
```

## 設定ファイル

**~/.agentrace/config.json**

```json
{
  "server_url": "http://localhost:8080",
  "api_key": "test-key"
}
```

## Hooks 設定

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

## 依存パッケージ

- `commander` - CLI フレームワーク
- `inquirer` - 対話的プロンプト

## 実装順序

### Step 1: 最小動作版 ✅ 完了

1. `init` - 手動でAPIキーを入力、config.json保存、hooks設定
2. `init --dev` - 開発モード（ローカルCLIパス使用）
3. `send` - transcript差分読み取り、POST送信
4. `uninstall` - hooks/config削除
5. カーソル管理による差分送信

### Step 2: ブラウザ連携

1. `init` - ローカルHTTPサーバー起動、ブラウザ認証、コールバック受信
2. `status` - サーバー接続確認（オプション）
