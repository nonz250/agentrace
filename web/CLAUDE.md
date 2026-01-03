# Web 開発ガイド

React + Vite + Tailwind CSS によるフロントエンド。

## 技術スタック

- Vite 7 / React 19 / TypeScript
- Tailwind CSS 3
- React Router v7
- TanStack Query v5 + AuthContext
- Lucide React（アイコン）
- react-syntax-highlighter / react-markdown

## ディレクトリ構成

```
web/src/
├── api/                 # APIクライアント層
│   ├── client.ts        # fetch ラッパー
│   └── *.ts             # 各種API（auth, sessions, plans, keys）
├── components/
│   ├── ui/              # 基本UIコンポーネント
│   ├── layout/          # レイアウト
│   ├── sessions/        # セッション表示
│   ├── timeline/        # イベントタイムライン
│   ├── plans/           # Plan表示
│   ├── settings/        # 設定画面
│   └── members/         # メンバー表示
├── hooks/               # カスタムフック
├── lib/                 # ユーティリティ
├── pages/               # ページコンポーネント
├── types/               # 型定義
├── App.tsx              # ルーティング・AuthProvider
└── main.tsx             # エントリーポイント
```

## 設計方針

### 状態管理
- **AuthContext**: グローバル認証状態（user, isLoading, refetch）
- **TanStack Query**: サーバーキャッシュ（staleTime: 30秒）

### Query Key パターン
| データ | queryKey |
|--------|----------|
| セッション一覧 | `['sessions', 'list', page]` |
| セッション詳細 | `['session', id]` |
| Plan一覧 | `['plans', 'list', page]` |
| Plan詳細 | `['plan', id]` |

### コンポーネント階層
```
ページ (pages/)
  └─ useQuery / useMutation
     └─ コンテナ (sessions/, timeline/)
        └─ 機能コンポーネント
           └─ 基本UI (ui/)
```

## タイムライン表示

### イベントグルーピング
- **Tool グループ化**: `tool_use` と `tool_result` を `tool_use_id` で紐付け
- **ローカルコマンド グループ化**: `/compact` 等とメタメッセージ・サマリーをまとめる

### ブロックタイプ
| タイプ | デフォルト |
|--------|-----------|
| text | 展開 |
| thinking | 折りたたみ |
| tool_group | 折りたたみ |
| compact_summary | 展開 |

## ルーティング

### URL構造

Session と Plan は Project 配下のリソースとして構成される。

| パス | 説明 |
|------|------|
| `/` | プロジェクト一覧（トップページ） |
| `/projects/:projectId` | プロジェクト詳細（Recent Plans/Sessions） |
| `/projects/:projectId/sessions` | プロジェクト内のセッション一覧 |
| `/projects/:projectId/sessions/:id` | セッション詳細 |
| `/projects/:projectId/plans` | プロジェクト内のプラン一覧 |
| `/projects/:projectId/plans/:id` | プラン詳細 |
| `/sessions/:id` | セッション詳細へリダイレクト（後方互換） |
| `/plans/:id` | プラン詳細へリダイレクト（後方互換） |

### 認証

| パス | 認証 |
|------|------|
| `/welcome`, `/register`, `/login`, `/setup` | Public |
| `/`, `/projects/**` | 認証なしでも閲覧可 |
| `/members` | 認証なしでも閲覧可 |
| `/settings` | Protected（要認証） |

### ナビゲーション

- **ヘッダー**: プロジェクト配下のページでのみ Sessions/Plans リンクを表示
- **パンくずリスト**: Project > Sessions/Plans > 詳細 の階層構造で表示

## 環境変数

| 変数 | 説明 | デフォルト |
|------|------|-----------|
| `VITE_API_URL` | APIサーバーのURL | `http://localhost:8080` |

- 開発時: `.env.development` で設定
- 本番時: 同一オリジンの場合は設定不要（`window.location.origin` が使用される）

## 開発時の起動

```bash
npm install && npm run dev
```

- http://localhost:5173
- APIリクエストは`VITE_API_URL`（`.env.development`で設定）に直接送信
- サーバー側で`WEB_URL`を設定してCORSを許可する必要あり
