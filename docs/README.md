# Agentrace ドキュメント

Claude Code のやりとりをチームでレビューできるサービス。

## ドキュメント構成

```
docs/
├── README.md              # このファイル（概要、全体アーキテクチャ）
├── cli/                   # CLI 関連
│   ├── README.md          # CLI 概要・設計
│   ├── commands.md        # コマンド一覧・使い方
│   └── configuration.md   # 設定ファイル・hooks設定
├── server/                # Server 関連
│   ├── README.md          # Server 概要・設計
│   ├── api.md             # API エンドポイント
│   ├── authentication.md  # 認証フロー
│   └── configuration.md   # 環境変数・DB設定
├── web/                   # Web 関連
│   └── README.md          # Web 概要・設計・コンポーネント
└── deployment/            # デプロイ関連
    └── docker.md          # Docker 設定
```

## 全体アーキテクチャ

```
┌──────────────────────────────────────────────────────────┐
│                     Claude Code                          │
│  Stop hook → npx agentrace send                         │
└──────────────────────────────────────────────────────────┘
                         ↓ POST /api/ingest
┌──────────────────────────────────────────────────────────┐
│                   Agentrace Server                       │
│  Go + Gorilla Mux                                        │
│  Repository パターン（Memory/SQLite/PostgreSQL/MongoDB） │
└──────────────────────────────────────────────────────────┘
                         ↓ REST API
┌──────────────────────────────────────────────────────────┐
│                    Agentrace Web                         │
│  React + TanStack Query                                  │
│  セッション一覧 → 詳細 → タイムライン表示                │
└──────────────────────────────────────────────────────────┘
```

## データフロー

### 初期セットアップ

```
$ npx agentrace init --url http://server:8080
    ↓
ブラウザで /setup → 登録/ログイン → APIキー生成
    ↓
callback でAPIキー受信 → config保存 → hooks設定
```

### データ送信（Claude Code 使用時）

```
Claude Code Stop event
    ↓
npx agentrace send（stdin から session_id, transcript_path）
    ↓
差分抽出（前回カーソル位置から）
    ↓
POST /api/ingest（Bearer認証）
    ↓
Session + Event として DB 保存
```

### レビュー（Web）

```
/api/sessions で一覧取得
    ↓
/api/sessions/:id で詳細取得
    ↓
Timeline コンポーネントでイベント展開・グループ化
```

## 設計原則

### 1. 責務分離

各コンポーネントは明確な責務を持つ:
- **CLI**: Claude Code との連携、transcript 差分送信
- **Server**: データ永続化、認証、API提供
- **Web**: ユーザーインターフェース

### 2. Repository パターン（Server）

データアクセスをインターフェースで抽象化し、複数のDB実装を切り替え可能:
- Memory: 開発・テスト用
- SQLite: ローカル・小規模運用
- PostgreSQL: 本番運用
- MongoDB: AWS DocumentDB 環境

### 3. hooks による非破壊的連携（CLI）

Claude Code の動作をブロックしない設計:
- send コマンドのエラーは常に `exit(0)`
- 差分送信で帯域節約
- カーソル管理で重複送信防止

### 4. TanStack Query による状態管理（Web）

サーバーキャッシュとUIを分離:
- `staleTime: 30秒` でキャッシュ有効期間を管理
- `queryKey` でキャッシュを識別

## セキュリティ考慮事項

- APIキーは bcrypt でハッシュ化して保存
- パスワードも bcrypt でハッシュ化
- callback URL は localhost のみ許可
- returnTo パラメータは `/` で始まる相対パスのみ許可
- インターネット公開は想定しない（イントラネット用）

## 開発環境

### Docker を使う場合

```bash
docker build -t agentrace:latest .
docker run -d -p 9080:9080 -v $(pwd)/data:/data agentrace:latest
npx agentrace init --url http://localhost:9080
```

### Docker を使わない場合

```bash
# Server
cd server && DEV_MODE=true DB_TYPE=sqlite DATABASE_URL=./dev.db go run ./cmd/server

# Web
cd web && npm install && npm run dev

# CLI
cd cli && npm install && npx tsx src/index.ts init --url http://localhost:8080 --dev
```
