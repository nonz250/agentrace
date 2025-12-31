# PlanDocument リソース実装計画

## 概要

Claude Codeが実装や変更の計画を記録・管理するための新しいリソース「PlanDocument」を追加する。

## データモデル

### PlanDocument
| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | UUID |
| description | string | 短い説明 |
| body | string | 計画のMarkdownテキスト |
| git_remote_url | string | 作成時のgitリポジトリURL |
| created_at | time | 作成日時 |
| updated_at | time | 更新日時 |

※ collaborators は PlanDocumentEvent の user_id から動的に集計（APIレスポンス時）

### PlanDocumentEvent
| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | UUID |
| plan_document_id | string | FK |
| session_id | *string | nullable - Sessionとの関連 |
| user_id | *string | 変更を行ったユーザー |
| patch | string | diff-match-patch形式のパッチ |
| created_at | time | 作成日時 |

## MCP Tools（CLI統合）

`npx agentrace mcp-server` で起動、stdioでMCPプロトコル処理

| Tool | 説明 | 引数 |
|------|------|------|
| list_plans | リポジトリのPlan一覧取得 | git_remote_url |
| read_plan | Plan読み込み | id |
| create_plan | Plan作成 | description, body, git_remote_url |
| update_plan | Plan更新（パッチ生成） | id, body, session_id? |

## API エンドポイント

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| GET | `/api/plans` | Bearer/Session | 一覧（?git_remote_url=でフィルタ） |
| GET | `/api/plans/:id` | Bearer/Session | 詳細 |
| GET | `/api/plans/:id/events` | Bearer/Session | 変更履歴 |
| POST | `/api/plans` | Bearer | 作成 |
| PATCH | `/api/plans/:id` | Bearer | 更新 |
| DELETE | `/api/plans/:id` | Bearer | 削除 |

## Web 画面

- `/plans` - PlanDocument一覧ページ
- `/plans/:id` - 詳細ページ（Markdown表示 + 変更履歴タブ）

---

## 実装ファイル一覧

### Phase 1: Server ドメイン・Repository

**新規作成:**
- `server/internal/domain/plan_document.go`
- `server/internal/domain/plan_document_event.go`
- `server/internal/repository/memory/plan_document.go`
- `server/internal/repository/memory/plan_document_event.go`
- `server/internal/repository/sqlite/plan_document.go`
- `server/internal/repository/sqlite/plan_document_event.go`
- `server/internal/repository/postgres/plan_document.go`
- `server/internal/repository/postgres/plan_document_event.go`
- `server/internal/repository/mongodb/plan_document.go`
- `server/internal/repository/mongodb/plan_document_event.go`
- `server/migrations/sqlite/002_plan_documents.sql`
- `server/migrations/postgres/002_plan_documents.up.sql`

**修正:**
- `server/internal/repository/interface.go` - インターフェース追加
- `server/internal/repository/memory/repositories.go` - Repository追加
- `server/internal/repository/sqlite/repositories.go` - Repository追加
- `server/internal/repository/sqlite/db.go` - マイグレーション追加
- `server/internal/repository/postgres/repositories.go` - Repository追加
- `server/internal/repository/postgres/db.go` - マイグレーション追加
- `server/internal/repository/mongodb/repositories.go` - Repository追加
- `server/migrations/embed.go` - SQL埋め込み

### Phase 2: Server API

**新規作成:**
- `server/internal/api/plan_document.go`

**修正:**
- `server/internal/api/router.go` - ルート追加
- `docs/server/api.md` - PlanDocument APIドキュメント追記

### Phase 3: CLI MCP Server

**新規作成:**
- `cli/src/commands/mcp-server.ts` - MCPサーバーコマンド実装
- `cli/src/mcp/plan-document-client.ts` - PlanDocument APIクライアント

**修正:**
- `cli/package.json` - 依存追加（@modelcontextprotocol/server, zod, diff-match-patch-es）
- `cli/src/index.ts` - mcp-serverコマンド追加
- `cli/src/hooks/installer.ts` - installMcpServer/uninstallMcpServer関数追加
- `cli/src/commands/init.ts` - MCPサーバー設定のインストール処理追加
- `cli/src/commands/on.ts` - MCPサーバー有効化も追加
- `cli/src/commands/off.ts` - MCPサーバー無効化も追加
- `cli/src/commands/uninstall.ts` - MCPサーバー設定削除も追加
- `docs/cli/commands.md` - mcp-serverコマンドのドキュメント追記
- `docs/cli/configuration.md` - mcpServers設定のドキュメント追記

**init時のMCPサーバー設定:**

`~/.claude/settings.json` に以下を追加:
```json
{
  "mcpServers": {
    "agentrace": {
      "command": "npx",
      "args": ["agentrace", "mcp-server"]
    }
  }
}
```

開発モード（--dev）の場合:
```json
{
  "mcpServers": {
    "agentrace": {
      "command": "npx",
      "args": ["tsx", "/path/to/cli/src/index.ts", "mcp-server"]
    }
  }
}
```

### Phase 4: Web フロントエンド

**新規作成:**
- `web/src/types/plan-document.ts` - 型定義
- `web/src/api/plan-documents.ts` - APIクライアント
- `web/src/pages/PlansPage.tsx` - Plan一覧ページ
- `web/src/pages/PlanDetailPage.tsx` - Plan詳細ページ（履歴タブ含む）
- `web/src/components/plans/PlanList.tsx` - Plan一覧コンポーネント
- `web/src/components/plans/PlanCard.tsx` - Planカードコンポーネント
- `web/src/components/plans/PlanEventHistory.tsx` - 変更履歴コンポーネント

**修正:**
- `web/src/App.tsx` - ルート追加（/plans, /plans/:id）
- `web/src/pages/HomePage.tsx` - Plansセクション追加（Recent Plansを表示）
- `docs/web/README.md` - Plansページのドキュメント追記

---

## 技術詳細

### collaborators集計
- APIレスポンス時に PlanDocumentEvent の user_id を DISTINCT で取得
- User情報と結合してレスポンスに含める

### パッチ形式
- CLI: `diff-match-patch-es` パッケージ
- パッチ生成はCLI側で実行、Server側はパッチを保存のみ

### 依存パッケージ（CLI）
```json
{
  "@modelcontextprotocol/server": "^2.0.0",
  "zod": "^3.25.0",
  "diff-match-patch-es": "^1.0.0"
}
```

---

## 実装順序

1. **Phase 1**: Server ドメイン・Repository（memory + sqlite） ✅ 完了
2. **Phase 2**: Server API ハンドラー・ルーティング ✅ 完了
3. **Phase 3**: CLI MCP Server コマンド ✅ 完了
4. **Phase 4**: Web フロントエンド ✅ 完了

各Phase完了後に動作確認を行う。

---

## 実装ログ

### Phase 1 完了

**作成したファイル:**
- `server/internal/domain/plan_document.go`
- `server/internal/domain/plan_document_event.go`
- `server/internal/repository/memory/plan_document.go`
- `server/internal/repository/memory/plan_document_event.go`
- `server/internal/repository/sqlite/plan_document.go`
- `server/internal/repository/sqlite/plan_document_event.go`
- `server/internal/repository/postgres/plan_document.go`
- `server/internal/repository/postgres/plan_document_event.go`
- `server/internal/repository/mongodb/plan_document.go`
- `server/internal/repository/mongodb/plan_document_event.go`
- `server/migrations/sqlite/002_plan_documents.sql`
- `server/migrations/postgres/002_plan_documents.up.sql`

**修正したファイル:**
- `server/internal/repository/interface.go`
- `server/internal/repository/memory/repositories.go`
- `server/internal/repository/sqlite/repositories.go`
- `server/internal/repository/sqlite/db.go`
- `server/internal/repository/postgres/repositories.go`
- `server/internal/repository/postgres/db.go`
- `server/internal/repository/mongodb/repositories.go`
- `server/migrations/embed.go`

**確認:** `go build ./...` 成功

### Phase 2 完了

**作成したファイル:**
- `server/internal/api/plan_document.go`

**修正したファイル:**
- `server/internal/api/router.go`
- `docs/server/api.md`

**確認:** `go build ./...` 成功

### Phase 3 完了

**作成したファイル:**
- `cli/src/commands/mcp-server.ts`
- `cli/src/mcp/plan-document-client.ts`

**修正したファイル:**
- `cli/package.json` - 依存追加（@modelcontextprotocol/sdk, zod, diff-match-patch-es）
- `cli/src/index.ts` - mcp-serverコマンド追加
- `cli/src/hooks/installer.ts` - installMcpServer/uninstallMcpServer関数追加
- `cli/src/commands/init.ts` - MCPサーバー設定のインストール処理追加
- `cli/src/commands/on.ts` - MCPサーバー有効化追加
- `cli/src/commands/off.ts` - MCPサーバー無効化追加
- `cli/src/commands/uninstall.ts` - MCPサーバー設定削除追加
- `docs/cli/commands.md` - mcp-serverコマンドのドキュメント追記
- `docs/cli/configuration.md` - mcpServers設定のドキュメント追記

**確認:** `npm run build` 成功

### Phase 4 完了

**作成したファイル:**
- `web/src/types/plan-document.ts`
- `web/src/api/plan-documents.ts`
- `web/src/pages/PlansPage.tsx`
- `web/src/pages/PlanDetailPage.tsx`
- `web/src/components/plans/PlanList.tsx`
- `web/src/components/plans/PlanCard.tsx`
- `web/src/components/plans/PlanEventHistory.tsx`

**修正したファイル:**
- `web/src/App.tsx` - ルート追加（/plans, /plans/:id）
- `web/src/pages/HomePage.tsx` - Recent Plansセクション追加
- `docs/web/README.md` - Plansページのドキュメント追記
- `web/package.json` - remark-gfm 依存追加

**確認:** `npm run build` 成功

---

## バグ修正

### user_id がPlanDocumentEventに保存されない問題

**問題:**
PlanDocument作成・更新時に、Bearer認証（APIキー）から特定したユーザーIDが `PlanDocumentEvent.user_id` に保存されていなかった。

**原因:**
`server/internal/api/plan_document.go` で `ctx.Value("user_id")` を使用してユーザーIDを取得しようとしていたが、認証ミドルウェアは異なるコンテキストキーを使用していた。

- ミドルウェア: `context.WithValue(ctx, userIDContextKey, user.ID)`
  - `userIDContextKey` は `contextKey("userID")` （カスタム型）
- plan_document.go: `ctx.Value("user_id").(string)`
  - キーの型が異なる（`contextKey` vs `string`）
  - キーの値も異なる（`"userID"` vs `"user_id"`）

**修正:**
`ctx.Value("user_id").(string)` を `GetUserIDFromContext(ctx)` に変更。
ミドルウェアで提供されているヘルパー関数を使用することで正しくユーザーIDを取得できるようになった。

**修正ファイル:**
- `server/internal/api/plan_document.go` - Create, Update 両ハンドラーを修正

### update_plan の session_id を必須化

**変更:**
MCPツールの `update_plan` で `session_id` パラメータを必須に変更。
AIエージェントは自分のセッションIDを渡すことで、変更履歴にセッション情報が紐付けられる。

**修正ファイル:**
- `cli/src/commands/mcp-server.ts` - UpdatePlanSchemaのsession_idを必須に変更、説明を追加
- `docs/cli/commands.md` - ドキュメント更新
