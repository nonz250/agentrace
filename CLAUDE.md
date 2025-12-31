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
├── web/      # フロントエンド
└── docs/     # ドキュメント
```

## ドキュメント

詳細な設計・設定・APIについては以下のドキュメントを参照してください。
何か新しい機能を追加する場合や、機能の調査を行う場合には、まずこれらのドキュメントを確認してください。

| ドキュメント | 内容 |
|-------------|------|
| [docs/README.md](docs/README.md) | 全体概要、アーキテクチャ、データフロー |
| [docs/cli/](docs/cli/) | CLI設計、コマンド一覧、設定ファイル |
| [docs/server/](docs/server/) | Server設計、APIエンドポイント、認証、環境変数 |
| [docs/web/](docs/web/) | Web設計、コンポーネント、タイムライン表示 |
| [docs/deployment/](docs/deployment/) | Docker設定 |

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
│  セッション一覧 → 詳細 → タイムライン表示                │
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
# Server
cd server && DEV_MODE=true DB_TYPE=sqlite DATABASE_URL=./dev.db go run ./cmd/server

# Web
cd web && npm install && npm run dev

# CLI
cd cli && npm install && npx tsx src/index.ts init --url http://localhost:8080 --dev
```

## CLIコマンド

| コマンド | 説明 |
|---------|------|
| `agentrace init --url <url>` | 初期設定 + hooks インストール |
| `agentrace login` | WebログインURL発行 |
| `agentrace send` | transcript差分送信（hooks用） |
| `agentrace on` / `off` | hooks有効化/無効化 |
| `agentrace uninstall` | hooks/config 削除 |

詳細: [docs/cli/commands.md](docs/cli/commands.md)

## 環境変数（Server）

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `PORT` | サーバーポート | 8080 |
| `DB_TYPE` | `memory` / `sqlite` / `postgres` / `mongodb` | memory |
| `DATABASE_URL` | DB接続文字列 | - |
| `DEV_MODE` | デバッグログ | false |
| `GITHUB_CLIENT_ID` | GitHub OAuth | - |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth | - |

詳細: [docs/server/configuration.md](docs/server/configuration.md)

## 将来の拡張（スコープ外）

- リアルタイム機能（WebSocket）
- コメント機能
- セッションの再開機能
- Slack/Discord通知
- 統計ダッシュボード
- セッションのエクスポート
- Google OAuth
