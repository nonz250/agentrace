# Agentrace

Claude Codeのやりとりをチームでレビューできるサービス

## ディレクトリ構成

```
agentrace/
├── cli/                         # Node.js CLI (TypeScript)
│   ├── src/
│   │   ├── index.ts             # エントリーポイント
│   │   ├── commands/
│   │   │   ├── init.ts          # 初期設定
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
└── server/                      # Go バックエンド
    ├── cmd/server/main.go       # エントリーポイント
    └── internal/
        ├── api/                 # HTTP ハンドラ
        │   ├── router.go
        │   ├── middleware.go
        │   ├── ingest.go
        │   └── session.go
        ├── config/config.go     # 環境変数管理
        ├── domain/              # ドメインモデル
        │   ├── session.go
        │   └── event.go
        └── repository/          # データアクセス層
            ├── interface.go
            └── memory/          # オンメモリ実装
```

## 開発環境での動作確認

### 1. サーバー起動

```bash
cd server
API_KEY_FIXED=test-key go run ./cmd/server
```

- `API_KEY_FIXED` 設定時は開発モードとしてリクエストログを出力

### 2. CLI初期化（開発モード）

```bash
cd cli
npm install
npx tsx src/index.ts init --dev
# Server URL: http://localhost:8080
# API Key: test-key
```

- `--dev` オプションでローカルCLIパスを使用

### 3. 動作確認

Claude Codeで操作すると、Stopイベントごとにtranscript差分がサーバーに送信される

```bash
# セッション一覧取得
curl -H "Authorization: Bearer test-key" http://localhost:8080/api/sessions

# セッション詳細取得
curl -H "Authorization: Bearer test-key" http://localhost:8080/api/sessions/{id}
```

## CLIコマンド

| コマンド | 説明 |
|---------|------|
| `agentrace init` | 設定 + hooks インストール |
| `agentrace init --dev` | 開発モード（ローカルCLIパス使用） |
| `agentrace send` | transcript差分送信（hooks用） |
| `agentrace uninstall` | hooks/config 削除 |

## API エンドポイント

| Method | Path | 説明 |
|--------|------|------|
| POST | `/api/ingest` | transcript行を受信 |
| GET | `/api/sessions` | セッション一覧 |
| GET | `/api/sessions/:id` | セッション詳細（イベント含む） |
| GET | `/health` | ヘルスチェック |

## 環境変数（サーバー）

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | データベース種類 | memory |
| `API_KEY_FIXED` | 固定APIキー（開発用） | - |

## データフロー

1. Claude Code が応答完了 → Stop hook 発火
2. CLI: stdin から session_id, transcript_path を取得
3. CLI: transcript_path のJSONLを読み、前回からの差分を抽出
4. CLI: 差分をサーバーに POST /api/ingest
5. Server: 各行を Event として保存
6. CLI: カーソル位置を更新（~/.agentrace/cursors/{session_id}.json）
