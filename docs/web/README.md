# Web

Agentrace のフロントエンド。セッション一覧、詳細表示、設定画面を提供。

## 技術スタック

| カテゴリ | 技術 |
|---------|------|
| ビルドツール | Vite 7 |
| UIライブラリ | React 19 |
| 言語 | TypeScript |
| スタイリング | Tailwind CSS 3 |
| ルーティング | React Router v7 |
| 状態管理 | TanStack Query v5 + AuthContext |
| 日時処理 | date-fns |
| アイコン | Lucide React |
| コード表示 | react-syntax-highlighter |
| Markdown | react-markdown + @tailwindcss/typography |

## ディレクトリ構成

```
web/src/
├── api/                     # APIクライアント層
│   ├── client.ts            # fetch ラッパー
│   ├── auth.ts              # 認証API
│   ├── sessions.ts          # セッションAPI
│   ├── plan-documents.ts    # PlanDocument API
│   └── keys.ts              # APIキーAPI
├── components/              # UIコンポーネント
│   ├── ui/                  # 基本UIコンポーネント
│   ├── layout/              # レイアウト
│   ├── sessions/            # セッション表示
│   ├── timeline/            # イベントタイムライン
│   ├── plans/               # Plan表示
│   ├── settings/            # 設定画面
│   └── members/             # メンバー表示
├── hooks/                   # カスタムフック
│   └── useAuth.ts           # 認証操作
├── lib/                     # ユーティリティ
│   └── cn.ts                # Tailwind クラス統合
├── pages/                   # ページコンポーネント
├── types/                   # 型定義
├── App.tsx                  # ルーティング・AuthProvider
└── main.tsx                 # エントリーポイント
```

## 設計方針

### 状態管理

**AuthContext（グローバル認証状態）**

```typescript
interface AuthContextType {
  user: User | null
  isLoading: boolean
  refetch: () => Promise<void>
}
```

**TanStack Query（サーバーキャッシュ）**

```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000,  // 30秒
      retry: false,
      refetchOnWindowFocus: false,
    },
  },
})
```

### Query Key パターン

| データ | queryKey |
|--------|----------|
| 最新セッション | `['sessions', 'recent']` |
| セッション一覧 | `['sessions', 'list', page]` |
| セッション詳細 | `['session', id]` |
| 最新Plan | `['plans', 'recent']` |
| Plan一覧 | `['plans', 'list', page]` |
| Plan詳細 | `['plan', id]` |
| Plan履歴 | `['plan', id, 'events']` |
| APIキー一覧 | `['keys']` |
| ユーザー一覧 | `['users']` |

### コンポーネント階層

```
ページ (pages/)
  └─ useQuery / useMutation
     └─ コンテナ (sessions/, timeline/)
        └─ 機能コンポーネント (SessionCard, ContentBlockCard)
           └─ 基本UI (ui/)
```

## タイムライン表示

### イベントのグルーピング

**Tool グループ化**
- `tool_use` ブロックと対応する `tool_result` を1つのカードにまとめる
- `tool_use.id` と `tool_result.tool_use_id` で紐付け
- ファイル操作ツール（Read, Edit, Write, Glob, Grep等）はファイルパスを表示

**ローカルコマンド グループ化**
- `/compact` 等のローカルコマンドと関連イベントを1つのカードにまとめる
- 対象: メタメッセージ（`payload.isMeta`）、サマリー（`payload.isCompactSummary`）、コマンド出力
- コマンドの検出: コンテンツが `<command-name>/` で始まる
- 出力の検出: コンテンツに `<local-command-stdout>` を含む

### メッセージ表示

ContentBlockCard コンポーネントは以下のブロックタイプに対応：

| ブロックタイプ | 表示 | デフォルト |
|---------------|------|-----------|
| text | Markdown対応テキスト表示 | 展開 |
| thinking | 折りたたみ可能なUI（紫色） | 折りたたみ |
| tool_group | ツール呼び出しと結果をグループ化 | 折りたたみ |
| tool_use | ツール名 + JSONハイライト表示 | 折りたたみ |
| tool_result | ツール結果表示 | 折りたたみ |
| local_command_group | ローカルコマンドと関連イベント | 折りたたみ |
| compact_summary | compactコマンドのサマリー（amber背景） | 展開 |
| local_command_output | コマンド出力表示 | 展開 |

### タイムスタンプ

- 優先順位: `payload.timestamp` > `created_at`
- `payload.timestamp` は Claude Code のオリジナルタイムスタンプ
- `created_at` はサーバー保存時刻（フォールバック）

### ソート仕様

| 対象 | ソートキー | 順序 |
|------|-----------|------|
| セッション一覧 | StartedAt | 降順（新しい順） |
| イベント一覧 | payload.timestamp | 昇順（会話順） |

## セッション表示形式

セッション一覧（SessionCard）と詳細ページは同じ形式で表示：

**タイトル行**（目立つ表示）
- 開始時刻: `YYYY/MM/DD HH:MM` 形式
- ユーザー名

**メタデータ行**（灰色小文字）
- GitBranch: リポジトリ名:ブランチ（GitHub/GitLabリンク付き）
- Folder: ディレクトリ名
- MessageSquare: イベント数

## ルーティング

| パス | ページ | 認証 |
|------|--------|------|
| `/welcome` | ウェルカム | Public |
| `/register` | ユーザー登録 | Public |
| `/login` | ログイン | Public |
| `/setup` | CLIセットアップ | Public |
| `/` | ホーム（最新セッション + 最新Plans） | Protected |
| `/sessions` | セッション一覧 | Protected |
| `/sessions/:id` | セッション詳細 | Protected |
| `/plans` | Plan一覧 | Protected |
| `/plans/:id` | Plan詳細（Content + History タブ） | Protected |
| `/members` | メンバー一覧 | Protected |
| `/settings` | APIキー管理 | Protected |

## 開発時の起動

```bash
cd web
npm install
npm run dev
```

- http://localhost:5173 でアクセス
- Vite のプロキシ設定で API リクエストは自動的に localhost:8080 に転送
