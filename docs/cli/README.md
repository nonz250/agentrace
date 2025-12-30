# CLI

Claude Code の transcript を Agentrace サーバーに送信する CLI ツール。

## 技術スタック

- Node.js / TypeScript
- Commander.js（CLI フレームワーク）
- npx 配布

## ディレクトリ構成

```
cli/
├── src/
│   ├── index.ts                 # エントリーポイント（Commander.js）
│   ├── commands/                # コマンド実装
│   │   ├── init.ts              # 初期セットアップ（ブラウザ連携）
│   │   ├── login.ts             # Webログイン
│   │   ├── send.ts              # transcript送信（hooks用）
│   │   ├── on.ts                # hooks有効化
│   │   ├── off.ts               # hooks無効化
│   │   └── uninstall.ts         # 完全アンインストール
│   ├── config/                  # 設定管理
│   │   ├── manager.ts           # ~/.agentrace/config.json CRUD
│   │   └── cursor.ts            # 差分追跡（送信済み行数）
│   ├── hooks/                   # Claude Code hooks連携
│   │   └── installer.ts         # ~/.claude/settings.json 編集
│   └── utils/                   # ユーティリティ
│       ├── http.ts              # HTTP APIクライアント
│       ├── callback-server.ts   # ローカルHTTP callbackサーバー
│       └── browser.ts           # ブラウザ起動
└── package.json
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

- **send コマンド**: すべてのエラーで `exit(0)` → hooks をブロックしない設計
- **init コマンド**: 致命的エラーで `exit(1)` → ユーザーに再試行を促す
- 設定未読み込み時は `null` を返す（graceful fallback）

### 差分送信

transcript 全体を毎回送信するのではなく、前回送信位置からの差分のみを送信:

1. `~/.agentrace/cursors/{session_id}.json` で送信済み行数を管理
2. JSONL を読み込み、カーソル位置以降の行のみ抽出
3. 送信成功後にカーソル位置を更新

### Git 情報の取得

- 初回送信時のみ取得（パフォーマンス）
- `CLAUDE_PROJECT_DIR` 環境変数を優先（Claude Code 起動時のパス）
- 未設定時は stdin の `cwd` にフォールバック

## 関連ドキュメント

- [コマンド一覧](./commands.md)
- [設定ファイル](./configuration.md)
