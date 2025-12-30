# CLI 設定

## 設定ファイル

### ~/.agentrace/config.json

メイン設定ファイル。`init` コマンドで自動生成される。

```json
{
  "server_url": "http://localhost:8080",
  "api_key": "agtr_xxxxxxxxxxxxxxxxxxxxxxxx"
}
```

| フィールド | 説明 |
|-----------|------|
| `server_url` | Agentrace サーバーの URL |
| `api_key` | Bearer 認証用の API キー |

### ~/.agentrace/cursors/{session_id}.json

セッションごとの送信済み行数を管理。差分送信に使用。

```json
{
  "lineCount": 123,
  "lastUpdated": "2024-01-01T00:00:00.000Z"
}
```

## Hooks 設定

### ~/.claude/settings.json

Claude Code の hooks 設定。`init` または `on` コマンドで自動追加される。

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
    ]
  }
}
```

### Hooks の仕組み

1. Claude Code が応答完了（Stop イベント発火）
2. `~/.claude/settings.json` の `hooks.Stop` を実行
3. stdin に JSON を渡して `npx agentrace send` を実行
4. CLI が stdin から JSON を読み取り、差分をサーバーに送信

### stdin JSON 形式

Claude Code から渡される JSON:

```json
{
  "session_id": "uuid",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/working/directory"
}
```

## 環境変数

### CLAUDE_PROJECT_DIR

Claude Code 起動時のプロジェクトパス。Git 情報の取得に使用。

| 変数 | 取得元 | 特徴 |
|------|--------|------|
| `CLAUDE_PROJECT_DIR` | 環境変数 | Claude Code起動時のパス（固定） |
| `cwd` | stdin JSON | ビルドコマンド等で変わる可能性あり |

CLI は `CLAUDE_PROJECT_DIR` を優先して使用し、未設定時は `cwd` にフォールバック。

## ファイルパス一覧

| パス | 説明 |
|------|------|
| `~/.agentrace/config.json` | メイン設定 |
| `~/.agentrace/cursors/` | セッションごとのカーソル位置 |
| `~/.claude/settings.json` | Claude Code hooks 設定 |
