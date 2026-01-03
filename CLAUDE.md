# Agentrace

Claude Codeのやりとりをチームでレビューできるサービス

## 想定する利用シーン

| シーン | 説明 |
|--------|------|
| ローカル | 個人でローカル起動してシングルユーザで使用 |
| イントラネット | 社内サーバーにホストしてチームで使用 |

※ ClaudeCode の動作ログはコードそのものや実行環境に関する情報などが含まれているため、インターネット上に公開する形のホスティングは想定しない

## 技術スタック

| コンポーネント | 技術 |
|----------------|------|
| CLI | Node.js / TypeScript（npx配布） |
| バックエンド | Go + Gorilla Mux |
| データ層 | Repository パターン（Memory / SQLite / PostgreSQL / MongoDB） |
| フロントエンド | React + Vite + Tailwind CSS |

## プロジェクト構成

```
agentrace/
├── cli/      # npx agentrace CLI
├── server/   # バックエンドサーバー
└── web/      # フロントエンド
```

## タスク別の必読ドキュメント

作業を始める前に、該当するCLAUDE.mdをReadツールで必ず読んでください。

### Server (Go) の変更時

必読: `server/CLAUDE.md`

### Web (React) の変更時

必読: `web/CLAUDE.md`

### CLI の変更時

必読: `cli/CLAUDE.md`

## 全体アーキテクチャ

```text
┌──────────────────────────────────────────────────────────┐
│                     Claude Code                          │
│  Stop hook → npx agentrace send                         │
└──────────────────────────────────────────────────────────┘
                         ↓ POST /api/ingest
┌──────────────────────────────────────────────────────────┐
│                   Agentrace Server                       │
│  Repository パターン（Memory/SQLite/PostgreSQL/MongoDB） │
└──────────────────────────────────────────────────────────┘
                         ↓ REST API
┌──────────────────────────────────────────────────────────┐
│                    Agentrace Web                         │
│  プロジェクト一覧 → Sessions/Plans → 詳細表示           │
└──────────────────────────────────────────────────────────┘
```

## クイックスタート

### Docker を使う場合

```bash
docker run -d --name agentrace -p 9080:9080 -v $(pwd)/data:/data satetsu888/agentrace:latest
npx agentrace init --url http://localhost:9080
```

### Docker を使わない場合（開発）

```bash
# Server（WEB_URLはCORS許可とリダイレクト用）
cd server && DEV_MODE=true DB_TYPE=sqlite DATABASE_URL=./dev.db WEB_URL=http://localhost:5173 go run ./cmd/server

# Web（VITE_API_URLは.env.developmentで設定済み）
cd web && npm install && npm run dev

# CLI
cd cli && npm install && npx tsx src/index.ts init --url http://localhost:8080 --dev
```

環境変数:
- `WEB_URL`: サーバーがCORSで許可するフロントエンドのオリジン
- `VITE_API_URL`: Webフロントエンドが接続するAPIサーバーのURL（開発時は `.env.development` で設定済み）

## Docker デプロイ

### Docker Hub

- リポジトリ: https://hub.docker.com/r/satetsu888/agentrace
- 対応アーキテクチャ: `linux/amd64`, `linux/arm64`

### ビルド

```bash
# ローカルビルド
docker build -t agentrace:latest .

# マルチアーキテクチャビルド & push
docker buildx build --platform linux/amd64,linux/arm64 -t satetsu888/agentrace:latest --push .
```

### Docker 構成

```
agentrace/
├── Dockerfile              # マルチステージビルド（node→go→runtime）
├── docker-compose.yml      # 簡易起動用
└── docker/
    ├── nginx.conf          # nginx設定（:9080で静的ファイル+APIプロキシ）
    ├── supervisord.conf    # プロセス管理（nginx + Go server）
    └── entrypoint.sh       # 起動スクリプト
```

## 将来の拡張（スコープ外）

- リアルタイム機能（WebSocket）
- コメント機能
- セッションの再開機能
- Slack/Discord通知
- 統計ダッシュボード
- セッションのエクスポート
- Google OAuth
