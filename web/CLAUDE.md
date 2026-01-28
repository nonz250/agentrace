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
├── utils/               # ユーティリティ関数
│   ├── patch.ts         # パッチ適用（reconstructContent）
│   └── line-diff.ts     # 行単位diff計算
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
| タイプ | デフォルト | 備考 |
|--------|-----------|------|
| text | 展開 | |
| thinking | 折りたたみ | |
| tool_group | 折りたたみ | Todo/AskUserQuestionは展開 |
| tool_group (Todo) | 展開 | 専用UI表示 |
| tool_group (AskUserQuestion) | 展開 | 専用UI表示 |
| agentrace_tool | 展開 | プランカード表示 |
| compact_summary | 展開 | |

### ツール専用表示

特定のツールは専用UIで表示される（`ContentBlockCard.tsx`）:

| ツール | 表示形式 |
|--------|---------|
| TodoWrite / TodoRead | チェックボックス形式のタスクリスト |
| AskUserQuestion | 質問・選択肢カード + 回答表示 |
| AgenTrace MCP tools | プランカード（APIから詳細取得） |

#### Todoツール表示
- pending: 灰色の空チェックボックス (Circle)
- in_progress: 青いドット付き円 (CircleDot)
- completed: 緑のチェック (CheckCircle2) + 打ち消し線

#### AskUserQuestionツール表示
- ヘッダー + 質問テキスト
- 選択肢: ラジオボタン風表示（Circle / CircleDot）
- 選択された回答: 青背景でハイライト
- Other回答: Pencilアイコン + 別枠表示

## Plan表示

### Markdownコードハイライト

`PlanDetailPage.tsx`と`ContentBlockCard.tsx`で`react-syntax-highlighter`を使用:

```tsx
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'

<ReactMarkdown
  components={{
    code({ className, children, ...props }) {
      const match = /language-(\w+)/.exec(className || '')
      const code = String(children).replace(/\n$/, '')
      return match ? (
        <SyntaxHighlighter language={match[1]} style={oneLight}>
          {code}
        </SyntaxHighlighter>
      ) : (
        <code className={className} {...props}>{children}</code>
      )
    },
  }}
>
```

### バージョン切り替え

`PlanDetailPage.tsx`でPlanの過去バージョンを閲覧可能:

- **VersionSelector**: バージョン選択コンポーネント（`components/plans/VersionSelector.tsx`）
- **patch.ts**: パッチ適用ユーティリティ（`utils/patch.ts`）

#### パッチ形式

2種類のパッチ形式を処理:

| 形式 | 生成元 | 例 |
|------|--------|-----|
| 初期パッチ | サーバー（`bodyToInitialPatch`） | `+line1\n+line2` |
| 更新パッチ | CLI（`diff-match-patch`） | `@@ -1,3 +1,4 @@...` |

`reconstructContent()`関数で両形式を判定し、イベントを順番に適用してテキストを復元。

### Side-by-Side Diff（変更履歴表示）

`PlanEventHistory.tsx`の「View changes」でモーダル表示:

| コンポーネント | 役割 |
|----------------|------|
| `DiffModal.tsx` | フルスクリーンモーダル（95vw、ESC/背景クリックで閉じる） |
| `SideBySideDiff.tsx` | 2カラムside-by-side表示（Before/After横並び） |
| `line-diff.ts` | 行単位diff計算、ハンクマージ、コンテキスト行付与 |

#### 表示仕様

- Before（左）/ After（右）の2カラム表示
- 削除行: 赤背景（左側のみ）
- 追加行: 緑背景（右側のみ）
- 行番号表示
- 連続する削除+追加は高さを揃えてペアリング
- 変更前後5行のコンテキスト表示
- 近いハンク（5行以内）は1つにマージ

#### diff計算の流れ

1. `reconstructContent()` でbefore/afterの全文を取得
2. `computeLineDiff()` でdiff-match-patchの `diff_linesToChars_` を使い行単位diff計算
3. 変更箇所をハンクにグループ化（コンテキスト行付き）
4. `SideBySideDiff` で2カラム表示

### コメント機能

`PlanContentWithComments.tsx`でPlan本文へのインラインコメントを実装。

#### テキスト選択の制限

以下のケースではテキスト選択が無効化される:

| ケース | 理由 |
|--------|------|
| 改行を含む選択 | マークダウンの行頭記号（`-`, `#`, `>`等）によりマッチング困難 |
| テーブルのセルを跨ぐ選択 | セル区切り（`\|`）がレンダリング後に消えるためマッチング困難 |
| リンク内のテキスト選択 | クリック時のリンク遷移とコメント表示が競合するため |

#### インライン記法を跨ぐコメント

以下のインライン記法を跨ぐテキスト選択は許可され、正しくマッチング・ハイライト表示される:

| 記法 | 例 |
|------|-----|
| インラインコード | `` `code` `` |
| 太字 | `**bold**` / `__bold__` |
| 斜体 | `*italic*` / `_italic_` |
| 取り消し線 | `~~strike~~` |

サーバー側（`plan_comment_thread.go`）とクライアント側（`PlanContentWithComments.tsx`）の両方で`stripInlineMarkdown`関数を使用し、マークダウン記法を除去したテキストでマッチングを行う。

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
