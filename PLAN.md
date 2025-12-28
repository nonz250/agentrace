# Agentrace 実装計画

## 概要

Claude Codeのやりとりをチームでレビューできるサービス

## 想定する利用シーン

| シーン | 説明 |
| ------ | ---- |
| ローカル | 個人でローカル起動してシングルユーザで使用 |
| イントラネット | 社内サーバーにホストしてチームで使用 |
| 開発 | DEV_MODEでデバッグログを出力しながら開発 |

※ インターネット公開は想定しない（OAuth等は不要）

## 技術スタック

| コンポーネント | 技術 |
| -------------- | ---- |
| CLI | Node.js / TypeScript（npx配布） |
| バックエンド | Go + Gorilla |
| データ層 | Repository パターン（オンメモリ / PostgreSQL） |
| フロントエンド | React + Vite |

## 詳細プラン

- [PLAN_CLI.md](./PLAN_CLI.md) - CLI実装詳細
- [PLAN_SERVER.md](./PLAN_SERVER.md) - サーバー実装詳細
- [PLAN_WEB.md](./PLAN_WEB.md) - フロントエンド実装詳細

## プロジェクト構成

```text
agentrace/
├── cli/                       # npx agentrace CLI
├── server/                    # バックエンドサーバー
└── web/                       # フロントエンド
```

## アーキテクチャ

```text
┌─────────────────────────────────────────────────────────────┐
│ ユーザー登録（Web）                                          │
├─────────────────────────────────────────────────────────────┤
│  ブラウザで http://server:8080 にアクセス                    │
│      ↓                                                      │
│  「Register」→ 名前入力 → APIキー発行                       │
│      ↓                                                      │
│  APIキーをコピー（この1回のみ表示）                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ CLI セットアップ                                             │
├─────────────────────────────────────────────────────────────┤
│  $ npx agentrace init                                       │
│      ↓                                                      │
│  Server URL と APIキーを入力                                 │
│      ↓                                                      │
│  ~/.agentrace/config.json に保存                            │
│  ~/.claude/settings.json に hooks 追加                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ データ送信（自動・差分更新）                                 │
├─────────────────────────────────────────────────────────────┤
│  Claude Code → Stop hook 発火                               │
│      ↓                                                      │
│  npx agentrace send                                         │
│      ↓ transcript_path から差分読み取り                     │
│      ↓ HTTP POST /api/ingest (Bearer認証)                   │
│  Agentrace Server                                           │
│      ↓ APIKey → User 解決、UserIDをセッションに紐付け       │
│  Database（Memory / PostgreSQL）                            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Webログイン                                                  │
├─────────────────────────────────────────────────────────────┤
│  方法1: $ npx agentrace login → URL発行 → ブラウザで開く    │
│  方法2: WebでAPIキーを入力してログイン                       │
│      ↓                                                      │
│  セッションCookie発行 → ダッシュボードへ                    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ レビュー                                                     │
├─────────────────────────────────────────────────────────────┤
│  Web UI（セッション認証）                                    │
│      ↓ REST API / WebSocket                                 │
│  Agentrace Server                                           │
│      ↓                                                      │
│  セッション一覧 → 詳細 → イベントタイムライン               │
│  （全ユーザーのセッションが閲覧可能）                        │
└─────────────────────────────────────────────────────────────┘
```

## 実装順序

### Step 1: 最小動作版（E2E疎通・オンメモリDB） ✅ 完了

**目標**: Claude Codeのhooksからデータがサーバーに送信され、メモリに保存される

1. **server/** ✅
   - Repository インターフェース定義
   - オンメモリ Repository 実装
   - POST /api/ingest（固定APIキー認証）
   - GET /api/sessions, /api/sessions/:id
   - DEV_MODE時のリクエストログ出力

2. **cli/** ✅
   - `npx agentrace init` - APIキー入力（手動）+ hooks設定
   - `npx agentrace init --dev` - 開発モード（ローカルCLIパス使用）
   - `npx agentrace send` - transcript差分読み取り→POST
   - `npx agentrace uninstall` - hooks/config削除
   - カーソル管理による差分送信

3. **動作確認** ✅
   - 実際のClaude Codeでhooksが動作することを確認

### Step 2: 認証機能 ✅ 完了

**目標**: ユーザー登録とAPIキー発行、Webログイン

1. **server/** ✅
   - User, APIKey, WebSession ドメインモデル
   - 各種Repository（memory実装）
   - POST `/auth/register` - ユーザー登録＆APIキー発行
   - POST `/auth/login` - APIキーでログイン（Cookie発行）
   - GET `/auth/session` - トークンでログイン（CLI経由）
   - POST `/api/auth/web-session` - Webログイントークン発行
   - POST `/api/auth/logout` - ログアウト
   - GET `/api/me` - 自分の情報取得
   - GET `/api/users` - ユーザー一覧
   - GET `/api/keys` - 自分のAPIキー一覧
   - POST `/api/keys` - 新しいAPIキー発行
   - DELETE `/api/keys/:id` - APIキー削除
   - Bearer認証ミドルウェア（APIKey → User解決）
   - Session認証ミドルウェア（Cookie → User解決）
   - セッションにUserID紐付け

2. **cli/** ✅
   - `npx agentrace login` - WebログインURL発行→ブラウザで開く

### Step 3: Web UI

**目標**: セッション一覧・詳細の表示

1. **web/**
   - Vite + React セットアップ
   - ログイン/登録ページ
   - セッション一覧ページ
   - セッション詳細ページ（タイムライン表示）

### Step 4: PostgreSQL対応

**目標**: データの永続化

1. **server/**
   - PostgreSQL Repository 実装
   - マイグレーションファイル
   - DB_TYPE環境変数で切り替え

### Step 5: リアルタイム機能

**目標**: 新規イベントのリアルタイム表示

1. **server/** - WebSocket Hub実装、イベント保存時に配信
2. **web/** - WebSocket接続、リアルタイム更新

## 将来の拡張（スコープ外）

- コメント機能（セッション/イベントへのコメント）
- セッションの再開機能（コンテキストをClaude Codeに渡す）
- Slack/Discord通知
- 統計ダッシュボード
- セッションのエクスポート（Markdown等）
