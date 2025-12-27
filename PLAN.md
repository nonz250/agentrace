# Agentrace 実装計画

## 概要

Claude Codeのやりとりをチームでレビューできるサービス

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
│ セットアップ                                                 │
├─────────────────────────────────────────────────────────────┤
│  $ npx agentrace init                                       │
│      ↓                                                      │
│  ブラウザでログイン → APIキー取得                            │
│      ↓                                                      │
│  ~/.agentrace/config.json に保存                            │
│  ~/.claude/settings.json に hooks 追加                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ データ送信（自動）                                           │
├─────────────────────────────────────────────────────────────┤
│  Claude Code → Hook呼び出し                                 │
│      ↓                                                      │
│  npx agentrace send（stdin から JSON）                      │
│      ↓ HTTP POST                                            │
│  Agentrace Server /api/ingest                               │
│      ↓                                                      │
│  Database（Memory / PostgreSQL）                            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ レビュー                                                     │
├─────────────────────────────────────────────────────────────┤
│  Web UI                                                     │
│      ↓ REST API / WebSocket                                 │
│  Agentrace Server                                           │
│      ↓                                                      │
│  セッション一覧 → 詳細 → イベントタイムライン               │
└─────────────────────────────────────────────────────────────┘
```

## 実装順序

### Step 1: 最小動作版（E2E疎通・オンメモリDB）

**目標**: Claude Codeのhooksからデータがサーバーに送信され、メモリに保存される

1. **server/**
   - Repository インターフェース定義
   - オンメモリ Repository 実装
   - POST /api/ingest（固定APIキー認証）
   - GET /api/sessions, /api/sessions/:id

2. **cli/**
   - `npx agentrace init` - APIキー入力（手動）+ hooks設定
   - `npx agentrace send` - stdinからJSON読み取り→POST

3. **動作確認**
   - 実際のClaude Codeでhooksが動作することを確認

### Step 2: 認証とセットアップUI

1. **server/** - ユーザー登録/ログイン、/setup画面
2. **cli/** - ブラウザ連携でAPIキー自動取得

### Step 3: PostgreSQL対応

1. **server/** - PostgreSQL Repository 実装、マイグレーション

### Step 4: Web UI

1. **web/** - ログイン、セッション一覧・詳細

### Step 5: リアルタイム機能

1. **server/** - WebSocket配信
2. **web/** - リアルタイム更新

## 将来の拡張（スコープ外）

- コメント機能（セッション/イベントへのコメント）
- セッションの再開機能（コンテキストをClaude Codeに渡す）
- Slack/Discord通知
- 統計ダッシュボード
- チーム招待機能
- セッションのエクスポート（Markdown等）
